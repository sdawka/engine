package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/battlesnakeio/engine/controller"
	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/battlesnakeio/engine/worker"
)

func init() { rand.Seed(time.Now().Unix()) }

func main() {
	var (
		listen  string
		workers int
	)
	flag.StringVar(&listen, "listen", ":3004", "Listen address.")
	flag.IntVar(&workers, "workers", 10, "Worker count.")
	flag.Parse()

	controller := controller.New(controller.InMemStore())
	go func() {
		if err := controller.Serve(listen); err != nil {
			log.Fatalf("controller failed to serve on (%s): %v", listen, err)
		}
	}()

	client, err := pb.Dial(listen)
	if err != nil {
		log.Fatalf("controller failed to dial (%s): %v", listen, err)
	}

	w := &worker.Worker{
		ControllerClient:  client,
		PollInterval:      1 * time.Second,
		HeartbeatInterval: 300 * time.Millisecond,
	}

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		client.Start(ctx, &pb.StartRequest{
			Game: &pb.Game{ID: fmt.Sprint(i)},
		})
	}

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
