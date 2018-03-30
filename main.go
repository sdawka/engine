package main

import (
	"github.com/battlesnakeio/engine/api"
	"github.com/battlesnakeio/engine/controller"
)

func main() {
	server := api.New(controller.New())
	server.WaitForExit()
}
