package main

import (
	"flag"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/battlesnakeio/engine/pkg/controller/pb"
	"github.com/battlesnakeio/engine/pkg/worker"
	"golang.org/x/net/context"
)

func init() { rand.Seed(time.Now().Unix()) }

func main() {
	var (
		controllerAddr string
		workers        int
	)
	flag.StringVar(&controllerAddr, "controller-addr", "127.0.0.1:3004", "Address to dial the controller.")
	flag.IntVar(&workers, "workers", 10, "Worker count.")
	flag.Parse()

	client, err := pb.Dial(controllerAddr)
	if err != nil {
		log.Fatalf("controller failed to dial (%s): %v", controllerAddr, err)
	}

	w := &worker.Worker{
		ControllerClient:  client,
		PollInterval:      1 * time.Second,
		HeartbeatInterval: 300 * time.Millisecond,
	}

	ctx := context.Background()

	wg := &sync.WaitGroup{}
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			w.Run(ctx, i)
			wg.Done()
		}(i)
	}
	wg.Wait()
}
