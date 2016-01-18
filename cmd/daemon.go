package main

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jimmidyson/wurzel/daemon"
)

var (
	daemonCmd = &cobra.Command{
		Use:   "daemon",
		Short: "Start a daemon with REST API to monitor your server remotely",
		Long:  `Start a daemon with REST API to monitor your server remotely.`,
		Run: func(cmd *cobra.Command, args []string) {
			daemon.Run(strings.Split(viper.GetString("cgroups"), ","), viper.GetDuration("cgroups-stats-interval"))
		},
	}
)

func init() {
	RootCmd.AddCommand(daemonCmd)
}
