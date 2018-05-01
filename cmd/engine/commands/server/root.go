package server

import (
	"github.com/spf13/cobra"
)

var controllerAddr = "127.0.0.1:3004"

// RootCmd provides the root run command.
var RootCmd = &cobra.Command{
	Use:   "server",
	Short: "serve the battlesnake game engine",
	Run: func(c *cobra.Command, args []string) {
		go controllerCmd.Run(c, args)
		go apiCmd.Run(c, args)
		workerCmd.Run(c, args)
	},
}

// Execute runs the root command,
func init() {
	RootCmd.AddCommand(apiCmd)
	RootCmd.AddCommand(controllerCmd)
	RootCmd.AddCommand(workerCmd)
}
