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
	"net/http"
	_ "net/http/pprof"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/formatters/logstash"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// RootCmd is the root command for the whole program.
	RootCmd = &cobra.Command{
		Use:   "wurzel",
		Short: "Combining & harvesting metrics from various sources",
		Long:  `Combining & harvesting metrics from various sources.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if viper.GetBool("verbose") {
				log.SetLevel(log.DebugLevel)
			}
			if viper.GetBool("log-json") {
				log.SetFormatter(&logstash.LogstashFormatter{
					Type:            "wurzel",
					TimestampFormat: time.RFC3339Nano,
				})
			}

			if viper.GetBool("debug") {
				go func() {
					log.WithFields(log.Fields{"endpoint": "debug", "address": viper.GetString("debug-address")}).Info("Listening")
					log.Println(http.ListenAndServe(viper.GetString("debug-address"), nil))
				}()
			}

			go func() {
				mux := http.NewServeMux()
				mux.Handle("/metrics", prometheus.Handler())
				log.WithFields(log.Fields{"endpoint": "api", "address": viper.GetString("listen-address")}).Info("Listening")
				srv := http.Server{Addr: viper.GetString("listen-address"), Handler: mux}
				log.Println(srv.ListenAndServe())
			}()
		},
	}
)

func init() {
	viper.SetEnvPrefix("wurzel")
	viper.AutomaticEnv()

	addStringFlag(RootCmd.PersistentFlags(), "listen-address", ":8080", "the address to listen on for API requests")
	addStringFlag(RootCmd.PersistentFlags(), "cgroups", "blkio,cpu,cpuacct,cpuset,devices,freezer,hugetlb,memory,net_cls,net_prio,perf_event", "enabled cgroups (comma-separated)")
	addStringFlag(RootCmd.PersistentFlags(), "debug-address", "localhost:6060", "the address to listen on for debug/profile requests")

	addBoolPFlag(RootCmd.PersistentFlags(), "verbose", "v", false, "verbose output")
	addBoolFlag(RootCmd.PersistentFlags(), "log-json", false, "log in JSON format")
	addBoolPFlag(RootCmd.PersistentFlags(), "debug", "d", false, "enable debug/profile endpoints")
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
