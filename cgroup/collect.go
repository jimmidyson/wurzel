package cgroup

import (
	"os"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fs"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/jimmidyson/wurzel/api/v1"
	"github.com/jimmidyson/wurzel/metrics"
)

var (
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

type collector interface {
	GetStats(path string, stats *cgroups.Stats) error
	Name() string
}

func subsystemCollector(subsystem string) collector {
	switch subsystem {
	case "blkio":
		return &fs.BlkioGroup{}
	case "cpu":
		return &fs.CpuGroup{}
	case "cpuacct":
		return &fs.CpuacctGroup{}
	case "cpuset":
		return &fs.CpusetGroup{}
	case "devices":
		return &fs.DevicesGroup{}
	case "freezer":
		return &fs.FreezerGroup{}
	case "hugetlb":
		return &fs.HugetlbGroup{}
	case "memory":
		return &fs.MemoryGroup{}
	case "net_cls":
		return &fs.NetClsGroup{}
	case "net_prio":
		return &fs.NetPrioGroup{}
	case "perf_event":
		return &fs.PerfEventGroup{}
	default:
		return nil
	}
}

func (w *watcher) startCollection() {
	ticker := time.NewTicker(w.collectionInterval).C
	go w.collectStats(time.Now())
	for {
		select {
		case t := <-ticker:
			go w.collectStats(t)
		case <-w.done:
			log.Debug("Stopping stats collection")
			return
		}
	}
}

var collecting uint32

func (w *watcher) collectStats(startTime time.Time) {
	if !atomic.CompareAndSwapUint32(&collecting, 0, 1) {
		log.Error("Time taken for collection is greater than the configured collection interval - skipping collection this time. You need to increase collection interval.")
		return
	}
	defer atomic.StoreUint32(&collecting, 0)

	w.cgroupMu.Lock()
	defer w.cgroupMu.Unlock()

	log.Debug("Collecting all cgroup stats")

	allStart := startTime

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
	cg.stats = cgroupStats(cg.path, c)

	for _, subCg := range cg.subcgroups {
		walkCgroup(subCg, c)
	}
}

func cgroupStats(path string, c collector) *v1.Stats {
	stats := cgroups.NewStats()

	err := c.GetStats(path, stats)
	if err != nil && !os.IsNotExist(err) {
		log.Error(err)
	}

	v1Stats := &v1.Stats{}
	switch c.Name() {
	case "blkio":
		v1Stats.BlkioStats = convertBlkio(stats.BlkioStats)
	case "cpu":
		v1Stats.CPUStats = &v1.CPUStats{ThrottlingData: convertCPUThrottlingData(stats.CpuStats.ThrottlingData)}
	case "cpuacct":
		v1Stats.CPUStats = &v1.CPUStats{CPUUsage: convertCPUUsage(stats.CpuStats.CpuUsage)}
	case "hugetlb":
		v1Stats.HugetlbStats = convertHugetlbStats(stats.HugetlbStats)
	case "memory":
		v1Stats.MemoryStats = convertMemory(stats.MemoryStats)
	}

	return v1Stats
}

func getPIDs(path string) ([]int32, error) {
	pids, err := cgroups.GetPids(path)
	if err != nil {
		return nil, err
	}
	toPids := make([]int32, 0, len(pids))
	for _, pid := range pids {
		toPids = append(toPids, int32(pid))
	}
	return toPids, nil
}
