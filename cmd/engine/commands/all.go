package commands

import (
	"github.com/spf13/cobra"
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "runs all of the components",
	Run: func(c *cobra.Command, args []string) {
		go controllerCmd.Run(c, args)
		go apiCmd.Run(c, args)
		workerCmd.Run(c, args)
	},
}
