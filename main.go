package main

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"

	"github.com/battlesnakeio/engine/api"
	"github.com/battlesnakeio/engine/controller"
	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/battlesnakeio/engine/worker"
	log "github.com/sirupsen/logrus"
)

func init() { rand.Seed(time.Now().Unix()) }

func main() {
	var (
		controllerAddr string
		apiAddr        string
		workers        int
		profileCPU     string
	)
	flag.StringVar(&controllerAddr, "controller", ":3004", "controller listen address.")
	flag.StringVar(&apiAddr, "api", ":3005", "api listen address")
	flag.IntVar(&workers, "workers", 10, "Worker count.")
	flag.StringVar(&profileCPU, "cpu", "", "path for cpu profile dump")
	flag.Parse()

	if len(profileCPU) > 0 {
		f, err := os.Create(profileCPU)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			log.Info("Writing out cpu profile")
			pprof.StopCPUProfile()
			os.Exit(1)
		}()

		defer func() {
			log.Info("Writing out cpu profile")
			pprof.StopCPUProfile()
		}()
	}

	c := controller.New(controller.InMemStore())
	go func() {
		log.Infof("controller listening on %s", controllerAddr)
		if err := c.Serve(controllerAddr); err != nil {
			log.Fatalf("controller failed to serve on (%s): %v", controllerAddr, err)
		}
	}()

	client, err := pb.Dial(controllerAddr)
	if err != nil {
		log.Fatalf("controller failed to dial (%s): %v", controllerAddr, err)
	}

	go func() {
		api := api.New(apiAddr, client)
		api.WaitForExit()
	}()

	w := &worker.Worker{
		ControllerClient: client,
		PollInterval:     1 * time.Second,
		RunGame:          worker.Runner,
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
