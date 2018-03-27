package main

import "github.com/battlesnakeio/engine/api"

func main() {
	server := api.New()
	server.WaitForExit()
}
