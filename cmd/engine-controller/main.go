package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/battlesnakeio/engine/pkg/controller"
)

func init() { rand.Seed(time.Now().Unix()) }

func main() {
	var (
		listen string
	)
	flag.StringVar(&listen, "listen", ":3004", "Listen address.")
	flag.Parse()

	controller := controller.New(controller.InMemStore())
	if err := controller.Serve(listen); err != nil {
		log.Fatalf("controller failed to serve on (%s): %v", listen, err)
	}
}
