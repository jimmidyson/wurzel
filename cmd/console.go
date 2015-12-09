package main

import (
	"github.com/spf13/cobra"

	"github.com/jimmidyson/wurzel/console"
)

var (
	consoleCmd = &cobra.Command{
		Use:   "console",
		Short: "Run a dashboard in your terminal",
		Long:  `Run a dashboard in your terminal.`,
		Run: func(cmd *cobra.Command, args []string) {
			console.Run()
		},
	}
)

func init() {
	consoleCmd.Flags().StringVar(&logFile, "log-file", "wurzel.log", "file to log to")
	RootCmd.AddCommand(consoleCmd)
}
