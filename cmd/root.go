// Copyright © 2015 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/formatters/logstash"
	"github.com/spf13/cobra"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	RootCmd = &cobra.Command{
		Use:   "wurzel",
		Short: "Combining & harvesting metrics from various sources",
		Long:  `Combining & harvesting metrics from various sources.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				log.SetLevel(log.DebugLevel)
			}
			var formatter log.Formatter = &prefixed.TextFormatter{
				TimestampFormat: time.RFC3339Nano,
			}
			if logJson {
				formatter = &logstash.LogstashFormatter{Type: "wurzel"}
			}
			log.SetFormatter(formatter)
			if logFile != "" {
				f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
				if err != nil {
					log.Fatal(err)
				}
				if !logJson {
					log.SetFormatter(&prefixed.TextFormatter{
						TimestampFormat: time.RFC3339Nano,
						DisableColors:   true,
					})
				}
				log.SetOutput(f)
			}
		},
	}
	verbose bool
	logFile string
	logJson bool
)

func init() {
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().BoolVar(&logJson, "log-json", false, "log in JSON format")
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
