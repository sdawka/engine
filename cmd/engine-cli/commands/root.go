package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "engine-cli",
	Short: "engine-cli helps run games on the battlesnake game engine",
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

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
