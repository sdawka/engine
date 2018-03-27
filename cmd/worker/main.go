package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/battlesnakeio/engine/worker"
)

func main() {
	rand.Seed(time.Now().Unix())

	w := &worker.Worker{
		WorkStore:         worker.InMemStore("1", "2", "3"),
		PollInterval:      1 * time.Second,
		HeartbeatInterval: 300 * time.Millisecond,
		Perform: func(ctx context.Context, id string, workerID int) error {
			for i := 0; i < 10; i++ {
				log.Printf("[%d] performing work on %s", workerID, id)
				select {
				case <-ctx.Done():
					log.Println("perform closed")
					return nil
				case <-time.After(1 * time.Second):
				}
			}
			return nil
		},
	}

	go w.Run(1)
	go w.Run(2)
	go w.Run(3)
	w.Run(4)
}
