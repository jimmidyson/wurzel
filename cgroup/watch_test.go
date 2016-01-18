package cgroup

import (
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cgroup watch test")
	}

	w, err := NewWatcher(1*time.Second, "cpu")
	if err != nil {
		t.Errorf("%v", err)
	}
	err = w.Start()
	if err != nil {
		t.Errorf("%v", err)
	}
	time.Sleep(10 * time.Second)
}
