package cgroup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/docker/libcontainer/cgroups"
	"gopkg.in/fsnotify.v1"
)

var stopChan chan struct{}

type Watcher interface {
	Start() error
	Stop() error
}

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
			w.Stop()
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
						fmt.Printf("Watching %s\n", path)
					}
					return nil
				})
				return err
			}
		}
	}
	return fmt.Errorf("cannot find cgroup mount for %s", cgroup)
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
					fmt.Println("Start watching", event.Name)
					err := w.fsnotifyWatcher.Add(event.Name)
					if err != nil {
						log.Fatal(err)
					}
				}()
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				go func() {
					fmt.Println("Stop watching", event.Name)
					err := w.fsnotifyWatcher.Remove(event.Name)
					if err != nil {
						log.Fatal(err)
					}
				}()
			}
		case err := <-w.fsnotifyWatcher.Errors:
			log.Println("error:", err)
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
