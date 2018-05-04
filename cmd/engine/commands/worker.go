package commands

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/battlesnakeio/engine/worker"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const pingRetryDelay = 1 * time.Second

var (
	workerThreads      = 10
	workerPollInterval = 1 * time.Second
	workerChaos        = false
)

func init() {
	workerCmd.Flags().IntVarP(&workerThreads, "threads", "t", workerThreads, "worker processor threads, this is the amount of concurrent games a worker can process")
	workerCmd.Flags().StringVarP(&controllerAddr, "controller-addr", "c", controllerAddr, "address of the controller")
	workerCmd.Flags().DurationVarP(&workerPollInterval, "poll-interval", "p", workerPollInterval, "worker poll interval")
	workerCmd.Flags().BoolVar(&workerChaos, "chaos", workerChaos, "introduce chaotic latency into the worker")
	allCmd.Flags().AddFlagSet(workerCmd.Flags())
}

// randTimeoutInterceptor provides a random amount of variance to all GRPC calls
// at the client level. This is part of the chaos mode for the workers. It means
// that calls will randomly go over the lock interval triggering some
// interesting situations.
func randTimeoutInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	var sleep time.Duration
	if rand.Intn(100) <= 20 {
		sleep = time.Duration(rand.Intn(5)) * time.Second
	} else {
		sleep = time.Duration(rand.Intn(50)) * time.Millisecond
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(sleep):
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "runs the engine worker",
	Run: func(c *cobra.Command, args []string) {
		var opts []grpc.DialOption
		if workerChaos {
			log.Warn("using chaos mode")
			opts = append(opts, grpc.WithUnaryInterceptor(randTimeoutInterceptor))
		}
		client, err := pb.Dial(controllerAddr, opts...)
		if err != nil {
			log.WithError(err).
				WithField("address", controllerAddr).
				Fatal("failed to dial controller")
		}

		w := &worker.Worker{
			ControllerClient: client,
			PollInterval:     workerPollInterval,
			RunGame:          worker.Runner,
		}

		// Begin pinging controller to push useful logs to an operator.
		go func() {
			for {
				ping, err := client.Ping(context.Background(), &pb.PingRequest{})
				if err == nil {
					log.WithField("version", ping.Version).
						Info("connection to controller healthy")
					break
				} else {
					log.WithError(err).Warn("controller connection unhealthy")
					time.Sleep(pingRetryDelay)
				}
			}
		}()

		ctx := context.Background()
		wg := &sync.WaitGroup{}
		wg.Add(workerThreads)

		for i := 0; i < workerThreads; i++ {
			go func(i int) {
				log.WithField("worker", i).Info("Battlesnake worker starting")
				w.Run(ctx, i)
				wg.Done()
			}(i)
		}
		wg.Wait()
	},
}
