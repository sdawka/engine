package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var controllerAddr = "127.0.0.1:3004"

var rootCmd = &cobra.Command{
	Use:   "engine",
	Short: "engine is the battlesnake game engine",
}

// Execute runs the root command,
func Execute() {
	rootCmd.AddCommand(apiCmd)
	rootCmd.AddCommand(controllerCmd)
	rootCmd.AddCommand(workerCmd)
	rootCmd.AddCommand(allCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
