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

	"github.com/jimmidyson/wurzel/api/v1"
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
	stats      *v1.Stats
	subcgroups map[string]*cgroup
	pids       []int32
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
	w.cgroupMu.Lock()
	defer w.cgroupMu.Unlock()

	go w.handleEvents()

	watched := map[string]struct{}{}
	for _, cg := range w.cgroups {
		err := filepath.Walk(cg.path, func(path string, info os.FileInfo, err error) error {
			if _, ok := watched[path]; ok {
				return nil
			}

			if info.IsDir() {
				err := w.watch(path)
				if err != nil {
					return err
				}
				watched[path] = struct{}{}
			} else if filepath.Base(path) == fs.CgroupProcesses {
				err := w.watch(path)
				if err != nil {
					return err
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

	sys := subsystemCollector(subsystem)

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

func (w *watcher) findCgroup(subsystem, relPath string) *cgroup {
	cg := w.cgroups[subsystem]

	spl := strings.Split(filepath.Dir(relPath), string(os.PathSeparator))
	for _, s := range spl {
		if s != "." {
			cg = cg.subcgroups[s]
		}
	}

	return cg
}

func (w *watcher) watch(path string) error {
	log.WithField("target", path).Debug("Adding watch")

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	subsystem, cgroupMountPoint := w.findCgroupMountpoint(absPath)
	if cgroupMountPoint == "" {
		return fmt.Errorf("Cannot find cgroup mount point for %s", absPath)
	}

	err = w.fsnotifyWatcher.Add(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.WithField("target", path).Debug("Target no longer exists - ignoring")
			return nil
		}
		return err
	}

	rel, relErr := filepath.Rel(cgroupMountPoint, absPath)
	if relErr != nil {
		return relErr
	}

	parentCgroup := w.findCgroup(subsystem, rel)

	name := filepath.Base(absPath)

	if name == fs.CgroupProcesses {
		err := w.updatePIDs(filepath.Dir(absPath), parentCgroup)
		if err != nil {
			return err
		}
	} else if absPath != parentCgroup.path {
		parentCgroup.subcgroups[name] = &cgroup{
			name:       name,
			path:       absPath,
			subcgroups: make(map[string]*cgroup),
		}
		log.WithField("target", absPath).Debug("Started watching cgroup dir")
		inotifyCount.WithLabelValues(subsystem).Inc()
	}

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

	log.WithField("target", absPath).Debug("Stopping watch")
	err = w.fsnotifyWatcher.Remove(absPath)
	if err != nil {
		return err
	}

	inotifyCount.WithLabelValues(subsystem).Dec()

	rel, err := filepath.Rel(cgroupMountPoint, absPath)
	if err != nil {
		return err
	}

	cg := w.findCgroup(subsystem, filepath.Dir(rel))

	if filepath.Base(absPath) == fs.CgroupProcesses {
		cg.pids = nil
	} else {
		name := filepath.Base(rel)
		delete(cg.subcgroups, name)
	}

	log.WithField("target", absPath).Debug("Stopped watch")

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

					if !fi.IsDir() && fi.Name() != fs.CgroupProcesses {
						log.WithField("target", event.Name).Error("Ignoring create event - not dir or ", fs.CgroupProcesses)
						return
					}

					w.cgroupMu.Lock()
					defer w.cgroupMu.Unlock()

					err = w.watch(event.Name)
					if err != nil {
						log.WithFields(log.Fields{
							"target": event.Name,
							"error":  err,
						}).Error("Failed to add watch")
					}

					if fi.IsDir() {
						cgProcs := filepath.Join(event.Name, fs.CgroupProcesses)
						_, err = os.Stat(cgProcs)
						if err == nil {
							err = w.watch(cgProcs)
							if err != nil {
								log.WithFields(log.Fields{
									"target": cgProcs,
									"error":  err,
								}).Error("Failed to add watch")
							}
						}
					}
				}()
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				go func() {
					log.WithField("target", event.Name).Debug("Received remove event")

					w.cgroupMu.Lock()
					defer w.cgroupMu.Unlock()

					err := w.unwatch(event.Name)
					if err != nil {
						log.WithFields(log.Fields{
							"target": event.Name,
							"error":  err,
						}).Error("Failed to remove watch")
					}
				}()
			case event.Op&fsnotify.Write == fsnotify.Write && filepath.Base(event.Name) == fs.CgroupProcesses:
				go func() {
					log.WithField("target", event.Name).Debug("Received write event")

					w.cgroupMu.Lock()
					defer w.cgroupMu.Unlock()

					absPath, err := filepath.Abs(event.Name)
					if err != nil {
						log.WithFields(log.Fields{"target": event.Name, "error": err}).Error("Cannot find absolute dir")
						return
					}

					subsystem, cgroupMountPoint := w.findCgroupMountpoint(absPath)
					if cgroupMountPoint == "" {
						log.WithField("target", absPath).Error("Cannot find cgroup mount point")
						return
					}

					rel, relErr := filepath.Rel(cgroupMountPoint, absPath)
					if relErr != nil {
						log.WithFields(log.Fields{"error": err, "target": absPath}).Error("Cannot find relative path")
						return
					}
					err = w.updatePIDs(filepath.Dir(event.Name), w.findCgroup(subsystem, rel))

					if err != nil {
						log.WithFields(log.Fields{
							"target": event.Name,
							"error":  err,
						}).Error("Failed to update pids")
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

func (w *watcher) updatePIDs(path string, cg *cgroup) error {
	pids, err := getPIDs(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.WithField("target", path).Debug("Target cgroup.procs no longer exists - ignoring")
			cg.pids = pids
			return nil
		}
		log.WithFields(log.Fields{"target": path, "error": err}).Debug("Cannot get cgroup PIDS")
		cg.pids = pids
		return err
	}
	cg.pids = pids
	return nil
}

func (w *watcher) Stop() error {
	log.Debug("Stopping cgroup watcher")
	close(w.done)
	w.wg.Wait()
	return w.fsnotifyWatcher.Close()
}
