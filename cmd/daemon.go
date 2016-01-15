package main

import (
	"github.com/spf13/cobra"

	"github.com/jimmidyson/wurzel/daemon"
)

var (
	daemonCmd = &cobra.Command{
		Use:   "daemon",
		Short: "Start a daemon with REST API to monitor your server remotely",
		Long:  `Start a daemon with REST API to monitor your server remotely.`,
		Run: func(cmd *cobra.Command, args []string) {
			daemon.Run(cgroups)
		},
	}
)

func init() {
	RootCmd.AddCommand(daemonCmd)
}
