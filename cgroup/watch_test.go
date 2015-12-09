package cgroup

import (
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	w, err := NewWatcher("cpu")
	if err != nil {
		t.Errorf("%v", err)
	}
	err = w.Start()
	if err != nil {
		t.Errorf("%v", err)
	}
	time.Sleep(10 * time.Second)
}
