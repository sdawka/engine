package main

import (
	"github.com/battlesnakeio/engine/cmd/engine-cli/commands"
)

func main() {
	// 	var echoTimes int

	// 	var cmdPrint = &cobra.Command{
	// 		Use:   "print [string to print]",
	// 		Short: "Print anything to the screen",
	// 		Long: `print is for printing anything back to the screen.
	// For many years people have printed back to the screen.`,
	// 		Args: cobra.MinimumNArgs(1),
	// 		Run: func(cmd *cobra.Command, args []string) {
	// 			fmt.Println("Print: " + strings.Join(args, " "))
	// 		},
	// 	}

	// 	var cmdEcho = &cobra.Command{
	// 		Use:   "echo [string to echo]",
	// 		Short: "Echo anything to the screen",
	// 		Long: `echo is for echoing anything back.
	// Echo works a lot like print, except it has a child command.`,
	// 		Args: cobra.MinimumNArgs(1),
	// 		Run: func(cmd *cobra.Command, args []string) {
	// 			fmt.Println("Print: " + strings.Join(args, " "))
	// 		},
	// 	}

	// 	var cmdTimes = &cobra.Command{
	// 		Use:   "times [# times] [string to echo]",
	// 		Short: "Echo anything to the screen more times",
	// 		Long: `echo things multiple times back to the user by providing
	// a count and a string.`,
	// 		Args: cobra.MinimumNArgs(1),
	// 		Run: func(cmd *cobra.Command, args []string) {
	// 			for i := 0; i < echoTimes; i++ {
	// 				fmt.Println("Echo: " + strings.Join(args, " "))
	// 			}
	// 		},
	// 	}

	// cmdTimes.Flags().IntVarP(&echoTimes, "times", "t", 1, "times to echo the input")

	commands.Execute()
	// var rootCmd = &cobra.Command{Use: "app"}
	// rootCmd.AddCommand(cmdPrint, cmdEcho)
	// cmdEcho.AddCommand(cmdTimes)
	// rootCmd.Execute()
	// tm.Clear() // Clear current screen

	// for {
	// 	// By moving cursor to top-left position we ensure that console output
	// 	// will be overwritten each time, instead of adding new.
	// 	tm.MoveCursor(1, 1)

	// 	tm.Println("Current Time:", time.Now().Format(time.RFC1123))

	// 	tm.Flush() // Call it every time at the end of rendering

	// 	time.Sleep(time.Second)
	// }
}
