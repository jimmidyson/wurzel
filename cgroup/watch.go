package cgroup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fs"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/fsnotify.v1"

	"github.com/jimmidyson/wurzel/metrics"
)

const (
	// MetricsSubsystem is the metrics subsystem for cgroups.
	MetricsSubsystem = "cgroups"
)

var (
	allSubsystems = []string{"blkio", "cpu", "cpuacct", "cpuset", "devices", "freezer", "hugetlb", "memory", "net_cls", "net_prio", "perf_event"}

	inotifyCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: MetricsSubsystem,
			Name:      "fsnotify_count_current",
			Help:      "The current number of fs notifies labeled by subsystem.",
		},
		[]string{"subsystem"},
	)

	statsCollectionSummary = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: metrics.Namespace,
			Subsystem: MetricsSubsystem,
			Name:      "stats_collection_duration_microseconds",
			Help:      "The time taken to collect cgroup stats.",
		},
	)

	subsystemStatsCollectionSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: metrics.Namespace,
			Subsystem: MetricsSubsystem,
			Name:      "subsystem_stats_collection_duration_microseconds",
			Help:      "The time taken to collect cgroup stats, labeled by subsystem.",
		},
		[]string{"subsystem"},
	)
)

func init() {
	prometheus.MustRegister(inotifyCount)
	prometheus.MustRegister(statsCollectionSummary)
	prometheus.MustRegister(subsystemStatsCollectionSummary)
}

// Watcher interface is implemented by anything watching cgroups.
type Watcher interface {
	Start() error
	Stop() error
}

type collector interface {
	GetStats(path string, stats *cgroups.Stats) error
}

type watcher struct {
	cgroups            map[string]*cgroup
	subsystems         map[string]collector
	fsnotifyWatcher    *fsnotify.Watcher
	done               chan struct{}
	collectionInterval time.Duration
	wg                 sync.WaitGroup
	cgroupMu           sync.RWMutex
}

type cgroup struct {
	name       string
	path       string
	stats      *cgroups.Stats
	subcgroups map[string]*cgroup
}

// NewWatcher is a factory method for a new watcher for a number of cgroups.
func NewWatcher(statsInterval time.Duration, subsystems ...string) (Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &watcher{
		subsystems:         make(map[string]collector),
		fsnotifyWatcher:    fsWatcher,
		done:               make(chan struct{}),
		cgroups:            make(map[string]*cgroup, len(subsystems)),
		collectionInterval: statsInterval,
	}

	mounts, err := cgroups.GetCgroupMounts()
	if err != nil {
		return nil, err
	}

	for _, subsystem := range subsystems {
		err := w.watchSubsystem(subsystem, mounts)
		if err != nil {
			return nil, err
		}
	}

	for _, subsystem := range allSubsystems {
		_, ok := w.subsystems[subsystem]
		registerSubsystemMetrics(subsystem, ok)
	}

	return w, nil
}

func registerSubsystemMetrics(subsystem string, enabled bool) {
	value := 0.0
	if enabled {
		value = 1.0
	}
	m := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   MetricsSubsystem,
			Name:        "subsystem_enabled",
			Help:        "A metric with a constant '0' for disabled or '1' for enabled labeled by subsystem.",
			ConstLabels: prometheus.Labels{"subsystem": subsystem},
		},
	)
	m.Set(value)
	prometheus.MustRegister(m)
}

func (w *watcher) Start() error {
	go w.handleEvents()

	watchedDirs := map[string]struct{}{}
	for _, cg := range w.cgroups {
		err := filepath.Walk(cg.path, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				if _, ok := watchedDirs[path]; !ok {
					err := w.watch(path)
					if err != nil {
						return err
					}
					watchedDirs[path] = struct{}{}
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	go w.startCollection()

	return nil
}

func (w *watcher) watchSubsystem(subsystem string, mounts []cgroups.Mount) error {
	for _, mount := range mounts {
		for _, mountedSubsystem := range mount.Subsystems {
			if mountedSubsystem == subsystem {
				w.initializeSubsystem(subsystem, mount)
				return nil
			}
		}
	}
	return fmt.Errorf("cannot find subsystem mount for %s. Discovered subsystem mounts: %#v", subsystem, mounts)
}

func (w *watcher) initializeSubsystem(subsystem string, mount cgroups.Mount) {
	// Sometimes cgroups share mount points, e.g. cpu,cpuacct.
	var subcgroups map[string]*cgroup
	for _, existingCg := range w.cgroups {
		if existingCg.path == mount.Mountpoint {
			subcgroups = existingCg.subcgroups
		}
	}

	var sys collector
	switch subsystem {
	case "blkio":
		sys = &fs.BlkioGroup{}
	case "cpu":
		sys = &fs.CpuGroup{}
	case "cpuacct":
		sys = &fs.CpuacctGroup{}
	case "cpuset":
		sys = &fs.CpusetGroup{}
	case "devices":
		sys = &fs.DevicesGroup{}
	case "freezer":
		sys = &fs.FreezerGroup{}
	case "hugetlb":
		sys = &fs.HugetlbGroup{}
	case "memory":
		sys = &fs.MemoryGroup{}
	case "net_cls":
		sys = &fs.NetClsGroup{}
	case "net_prio":
		sys = &fs.NetPrioGroup{}
	case "perf_event":
		sys = &fs.PerfEventGroup{}
	}

	if sys != nil {
		w.subsystems[subsystem] = sys
	}

	if subcgroups == nil {
		subcgroups = make(map[string]*cgroup)
	}

	w.cgroups[subsystem] = &cgroup{
		name:       subsystem,
		path:       mount.Mountpoint,
		subcgroups: subcgroups,
	}

	log.WithFields(log.Fields{"subsystem": subsystem, "path": mount.Mountpoint}).Info("Initialized subsystem")
}

func (w *watcher) findCgroupMountpoint(path string) (string, string) {
	for _, cg := range w.cgroups {
		if strings.HasPrefix(path, cg.path) {
			return cg.name, cg.path
		}
	}

	return "", ""
}

func (w *watcher) watch(path string) error {
	w.cgroupMu.Lock()
	defer w.cgroupMu.Unlock()

	log.WithField("target", path).Debug("Adding watch")

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	subsystem, cgroupMountPoint := w.findCgroupMountpoint(absPath)
	if cgroupMountPoint == "" {
		return fmt.Errorf("Cannot find cgroup mount point for %s", absPath)
	}

	if cgroupMountPoint != absPath {
		rel, relErr := filepath.Rel(cgroupMountPoint, absPath)
		if relErr != nil {
			return relErr
		}

		parentCgroup := w.cgroups[subsystem]

		spl := filepath.SplitList(filepath.Dir(rel))
		for _, s := range spl {
			if s != "." {
				parentCgroup = parentCgroup.subcgroups[s]
			}
		}

		name := filepath.Base(rel)
		parentCgroup.subcgroups[name] = &cgroup{
			name:       name,
			path:       absPath,
			subcgroups: make(map[string]*cgroup),
		}
	}

	err = w.fsnotifyWatcher.Add(path)
	if err != nil {
		return err
	}
	log.WithField("target", absPath).Debug("Started watching cgroup dir")
	inotifyCount.WithLabelValues(subsystem).Inc()

	return nil
}

func (w *watcher) unwatch(path string) error {
	w.cgroupMu.Lock()
	defer w.cgroupMu.Unlock()
	log.WithField("target", path).Debug("Removing watch")

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	subsystem, cgroupMountPoint := w.findCgroupMountpoint(absPath)
	if cgroupMountPoint == "" {
		return fmt.Errorf("Cannot find cgroup mount point for %s", absPath)
	}

	log.WithField("target", absPath).Debug("Stopping watching cgroup dir")
	err = w.fsnotifyWatcher.Remove(absPath)
	if err != nil {
		return err
	}

	inotifyCount.WithLabelValues(subsystem).Dec()

	rel, err := filepath.Rel(cgroupMountPoint, absPath)
	if err != nil {
		return err
	}

	spl := filepath.SplitList(filepath.Dir(rel))
	parentCgroup := w.cgroups[subsystem]
	for _, s := range spl {
		if s != "." {
			parentCgroup = parentCgroup.subcgroups[s]
		}
	}
	name := filepath.Base(rel)
	delete(parentCgroup.subcgroups, name)

	log.WithField("target", absPath).Debug("Stopped watching cgroup dir")

	return nil
}

func (w *watcher) handleEvents() {
	w.wg.Add(1)
	defer w.wg.Done()
	for {
		select {
		case event := <-w.fsnotifyWatcher.Events:
			switch {
			case event.Op&fsnotify.Create == fsnotify.Create:
				go func() {
					log.WithField("target", event.Name).Debug("Received create event")
					fi, err := os.Lstat(event.Name)
					if err != nil {
						log.WithFields(log.Fields{
							"target": event.Name,
							"error":  err,
						}).Error("Failed to get lstat")
						return
					}

					if !fi.IsDir() {
						log.WithField("target", event.Name).Error("Ignoring create event - not dir")
						return
					}

					err = w.watch(event.Name)
					if err != nil {
						log.WithFields(log.Fields{
							"target": event.Name,
							"error":  err,
						}).Error("Failed to add watch")
					}
				}()
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				go func() {
					log.WithField("target", event.Name).Debug("Received remove event")
					err := w.unwatch(event.Name)
					if err != nil {
						log.WithFields(log.Fields{
							"target": event.Name,
							"error":  err,
						}).Error("Failed to remove watch")
					}
				}()
			}
		case err := <-w.fsnotifyWatcher.Errors:
			log.WithField("error", err).Error("Received notify error")
		case <-w.done:
			return
		}
	}
}

func (w *watcher) Stop() error {
	log.Debug("Stopping cgroup watcher")
	close(w.done)
	w.wg.Wait()
	return w.fsnotifyWatcher.Close()
}

func (w *watcher) startCollection() {
	for {
		select {
		case <-time.After(w.collectionInterval):
			w.collectStats()
		case <-w.done:
			log.Debug("Stopping stats collection")
			return
		}
	}
}

func (w *watcher) collectStats() {
	w.cgroupMu.Lock()
	defer w.cgroupMu.Unlock()

	log.Debug("Collecting all cgroup stats")

	allStart := time.Now()

	for name, rootCgroup := range w.cgroups {
		c := w.subsystems[name]
		if c == nil {
			log.WithField("subsystem", name).Debug("No collector for subsystem")
			continue
		}
		log.WithField("subsystem", name).Debug("Collecting cgroup stats")
		subsystemStart := time.Now()
		walkCgroup(rootCgroup, c)
		subsystemElapsed := float64(time.Since(subsystemStart)) / float64(time.Microsecond)
		subsystemStatsCollectionSummary.WithLabelValues(name).Observe(subsystemElapsed)

		log.WithFields(log.Fields{"subsystem": name, "duration": time.Duration(subsystemElapsed) * time.Microsecond}).Debug("Finished collecting cgroup stats")
	}

	allElapsed := float64(time.Since(allStart)) / float64(time.Microsecond)
	statsCollectionSummary.Observe(allElapsed)

	log.WithField("duration", time.Duration(allElapsed)*time.Microsecond).Debug("Finished collecting all cgroup stats")
}

func walkCgroup(cg *cgroup, c collector) {
	if cg.stats == nil {
		cg.stats = cgroups.NewStats()
	}
	stats := cg.stats
	err := c.GetStats(cg.path, stats)
	if err != nil {
		log.Error(err)
	}
	cg.stats = stats

	for _, subCg := range cg.subcgroups {
		walkCgroup(subCg, c)
	}
}
