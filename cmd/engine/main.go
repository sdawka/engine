package main

import (
	"math/rand"
	"time"

	"github.com/battlesnakeio/engine/cmd/engine/commands"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	commands.Execute()
}
