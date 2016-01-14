package cgroup

import (
	"fmt"
	"os"
	"path/filepath"
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
func NewWatcher(cgroup ...string) (Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &watcher{
		cgroups:         cgroup,
		fsnotifyWatcher: fsWatcher,
		done:            make(chan struct{}),
	}, nil
}

type watcher struct {
	cgroups         []string
	fsnotifyWatcher *fsnotify.Watcher
	done            chan struct{}
	wg              sync.WaitGroup
}

func (w *watcher) Start() error {
	mounts, err := cgroups.GetCgroupMounts()
	if err != nil {
		return err
	}

	go w.handleEvents()

	for _, cgroup := range w.cgroups {
		err := w.watchCgroup(cgroup, mounts)
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

func (w *watcher) watchCgroup(cgroup string, mounts []cgroups.Mount) error {
	for _, mount := range mounts {
		for _, subsystem := range mount.Subsystems {
			if subsystem == cgroup {
				err := filepath.Walk(mount.Mountpoint, func(path string, info os.FileInfo, err error) error {
					if info.IsDir() {
						err := w.fsnotifyWatcher.Add(path)
						if err != nil {
							return err
						}
						log.WithFields(log.Fields{"target": path}).Debug("Start watching cgroup dir")
					}
					return nil
				})
				return err
			}
		}
	}
	return fmt.Errorf("cannot find cgroup mount for %s. Discovered cgroup mounts: %#v", cgroup, mounts)
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
					log.WithFields(log.Fields{"target": event.Name}).Debug("Start watching cgroup dir")
					err := w.fsnotifyWatcher.Add(event.Name)
					if err != nil {
						log.WithFields(log.Fields{
							"target": event.Name,
							"error":  err,
						}).Error("Failed to add watch")
					}
				}()
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				go func() {
					log.WithFields(log.Fields{"target": event.Name}).Debug("Stop watching cgroup dir")
					err := w.fsnotifyWatcher.Remove(event.Name)
					if err != nil {
						log.WithFields(log.Fields{
							"target": event.Name,
							"error":  err,
						}).Error("Failed to remove watch")
					}
				}()
			}
		case err := <-w.fsnotifyWatcher.Errors:
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Received notify error")
		case <-w.done:
			return
		}
	}
}

func (w *watcher) Stop() error {
	w.done <- struct{}{}
	w.wg.Wait()
	return w.fsnotifyWatcher.Close()
}
