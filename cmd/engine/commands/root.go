package commands

import (
	"fmt"
	"os"

	"github.com/battlesnakeio/engine/cmd/engine/commands/server"
	"github.com/battlesnakeio/engine/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "engine",
	Short:   "engine helps run games on the battlesnake game engine",
	Version: version.Version,
}

var (
	apiAddr string
)

// Execute runs the root command
func Execute() {

	rootCmd.Flags().StringVar(&apiAddr, "api-addr", "http://localhost:3005", "address of the api server")

	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(replayCmd)
	rootCmd.AddCommand(loadTestCmd)
	rootCmd.AddCommand(server.RootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
