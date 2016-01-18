package version

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/jimmidyson/wurzel/metrics"
)

// Build information. Populated at build-time.
var (
	Version   string
	Revision  string
	Branch    string
	BuildUser string
	BuildDate string
	GoVersion string
)

// Map provides the iterable version information.
var Map = map[string]string{
	"version":   Version,
	"revision":  Revision,
	"branch":    Branch,
	"buildUser": BuildUser,
	"buildDate": BuildDate,
	"goVersion": GoVersion,
}

func init() {
	buildInfo := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   "build",
			Name:        "info",
			Help:        "A metric with a constant '1' value labeled by version, revision, and branch from which Wurzel was built.",
			ConstLabels: prometheus.Labels{"version": Version, "revision": Revision, "branch": Branch},
		},
	)
	buildInfo.Set(1)
	prometheus.MustRegister(buildInfo)
}
