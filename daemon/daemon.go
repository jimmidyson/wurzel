package daemon

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/jimmidyson/wurzel/cgroup"
)

// Run starts the daemon.
func Run(cgroups []string, statsInterval time.Duration) {
	log.WithFields(log.Fields{"cgroups": cgroups}).Debug("Enabled cgroups")
	w, err := cgroup.NewWatcher(statsInterval, cgroups...)
	if err != nil {
		log.Fatal(err)
	}
	err = w.Start()
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	// Block until a signal is received.
	<-c
	err = w.Stop()
	if err != nil {
		log.Fatal(err)
	}
}
