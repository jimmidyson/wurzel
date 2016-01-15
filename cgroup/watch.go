package cgroup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"gopkg.in/fsnotify.v1"
)

// Watcher interface is implemented by anything watching cgroups.
type Watcher interface {
	Start() error
	Stop() error
}

// NewWatcher is a factory method for a new watcher for a number of cgroups.
func NewWatcher(cgs ...string) (Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &watcher{
		watchedCgroups:  cgs,
		fsnotifyWatcher: fsWatcher,
		done:            make(chan struct{}),
		cgroups:         make(map[string]*cgroup, len(cgs)),
	}, nil
}

type watcher struct {
	cgroups         map[string]*cgroup
	watchedCgroups  []string
	fsnotifyWatcher *fsnotify.Watcher
	done            chan struct{}
	wg              sync.WaitGroup
	cgroupMu        sync.RWMutex
}

type cgroup struct {
	name       string
	path       string
	subcgroups map[string]*cgroup
}

func (w *watcher) Start() error {
	mounts, err := cgroups.GetCgroupMounts()
	if err != nil {
		return err
	}

	go w.handleEvents()

	for _, cg := range w.watchedCgroups {
		err := w.watchCgroup(cg, mounts)
		if err != nil {
			stopErr := w.Stop()
			if stopErr != nil {
				log.WithFields(log.Fields{"error": stopErr}).Debug("Cannot stop watch")
			}
			return err
		}
	}
	return nil
}

func (w *watcher) watchCgroup(cg string, mounts []cgroups.Mount) error {
	for _, mount := range mounts {
		for _, subsystem := range mount.Subsystems {
			if subsystem == cg {
				w.cgroupMu.Lock()

				// Sometimes cgroups share mount points, e.g. cpu,cpuacct.
				var subcgroups map[string]*cgroup
				for _, existingCg := range w.cgroups {
					if existingCg.path == mount.Mountpoint {
						subcgroups = existingCg.subcgroups
					}
				}
				if subcgroups == nil {
					subcgroups = make(map[string]*cgroup)
				}

				w.cgroups[subsystem] = &cgroup{
					name:       cg,
					path:       mount.Mountpoint,
					subcgroups: subcgroups,
				}

				w.cgroupMu.Unlock()

				err := filepath.Walk(mount.Mountpoint, func(path string, info os.FileInfo, err error) error {
					if info.IsDir() {
						err := w.watch(path)
						if err != nil {
							return err
						}
					}
					return nil
				})
				return err
			}
		}
	}
	return fmt.Errorf("cannot find cgroup mount for %s. Discovered cgroup mounts: %#v", cg, mounts)
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

	log.WithFields(log.Fields{"target": path}).Debug("Adding watch")

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
	log.WithFields(log.Fields{"target": absPath}).Debug("Started watching cgroup dir")

	return nil
}

func (w *watcher) unwatch(path string) error {
	w.cgroupMu.Lock()
	defer w.cgroupMu.Unlock()
	log.WithFields(log.Fields{"target": path}).Debug("Removing watch")

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"target": absPath}).Debug("Stopping watching cgroup dir")
	err = w.fsnotifyWatcher.Remove(absPath)
	if err != nil {
		return err
	}

	subsystem, cgroupMountPoint := w.findCgroupMountpoint(absPath)
	if cgroupMountPoint == "" {
		return fmt.Errorf("Cannot find cgroup mount point for %s", absPath)
	}

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

	log.WithFields(log.Fields{"target": absPath}).Debug("Stopped watching cgroup dir")

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
					log.WithFields(log.Fields{"target": event.Name}).Debug("Received create event")
					fi, err := os.Lstat(event.Name)
					if err != nil {
						log.WithFields(log.Fields{
							"target": event.Name,
							"error":  err,
						}).Error("Failed to get lstat")
						return
					}

					if !fi.IsDir() {
						log.WithFields(log.Fields{"target": event.Name}).Error("Ignoring create event - not dir")
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
					log.WithFields(log.Fields{"target": event.Name}).Debug("Received remove event")
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
			log.WithFields(log.Fields{"error": err}).Error("Received notify error")
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
