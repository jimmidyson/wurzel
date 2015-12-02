package main

import (
	"github.com/jimmidyson/wurzel/console"
	"github.com/spf13/cobra"
)

// uiCmd represents the ui command
var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Run a dashboard in your terminal",
	Long:  `Run a dashboard in your terminal.`,
	Run: func(cmd *cobra.Command, args []string) {
		console.Run()
	},
}

func init() {
	RootCmd.AddCommand(consoleCmd)
}
