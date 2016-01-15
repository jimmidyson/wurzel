package main

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"

	"github.com/jimmidyson/wurzel/console"
)

var (
	consoleCmd = &cobra.Command{
		Use:   "console",
		Short: "Run a dashboard in your terminal",
		Long:  `Run a dashboard in your terminal.`,
		PreRun: func(cmd *cobra.Command, args []string) {
			if logFile != "" {
				f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
				if err != nil {
					log.Fatal(err)
				}
				if !logJSON {
					log.SetFormatter(&prefixed.TextFormatter{
						TimestampFormat: time.RFC3339Nano,
						DisableColors:   true,
					})
				}
				log.SetOutput(f)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			console.Run()
		},
	}
	logFile string
)

func init() {
	consoleCmd.Flags().StringVar(&logFile, "log-file", "wurzel.log", "file to log to")
	RootCmd.AddCommand(consoleCmd)
}
