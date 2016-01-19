package cgroup

import (
	"github.com/jimmidyson/wurzel/api/v1"
	"github.com/opencontainers/runc/libcontainer/cgroups"
)

func convertBlkio(from cgroups.BlkioStats) *v1.BlkioStats {
	to := &v1.BlkioStats{
		IoServiceBytesRecursive: convertBlkioStatEntries(from.IoServiceBytesRecursive),
		IoServicedRecursive:     convertBlkioStatEntries(from.IoServicedRecursive),
		IoQueuedRecursive:       convertBlkioStatEntries(from.IoQueuedRecursive),
		IoServiceTimeRecursive:  convertBlkioStatEntries(from.IoServiceTimeRecursive),
		IoWaitTimeRecursive:     convertBlkioStatEntries(from.IoWaitTimeRecursive),
		IoMergedRecursive:       convertBlkioStatEntries(from.IoMergedRecursive),
		IoTimeRecursive:         convertBlkioStatEntries(from.IoTimeRecursive),
		SectorsRecursive:        convertBlkioStatEntries(from.SectorsRecursive),
	}

	return to
}

func convertBlkioStatEntries(from []cgroups.BlkioStatEntry) []v1.BlkioStatEntry {
	to := make([]v1.BlkioStatEntry, 0, len(from))
	for _, entry := range from {
		to = append(to, v1.BlkioStatEntry{
			Major: entry.Major,
			Minor: entry.Minor,
			Op:    entry.Op,
			Value: entry.Value,
		})
	}
	return to
}

func convertCPUThrottlingData(from cgroups.ThrottlingData) *v1.ThrottlingData {
	return &v1.ThrottlingData{
		Periods:          from.Periods,
		ThrottledPeriods: from.ThrottledPeriods,
		ThrottledTime:    from.ThrottledTime,
	}
}

func convertCPUUsage(from cgroups.CpuUsage) *v1.CPUUsage {
	return &v1.CPUUsage{
		TotalUsage:        from.TotalUsage,
		PerCPUUsage:       from.PercpuUsage,
		UsageInKernelmode: from.UsageInKernelmode,
		UsageInUsermode:   from.UsageInUsermode,
	}
}

func convertHugetlbStats(from map[string]cgroups.HugetlbStats) map[string]v1.HugetlbStats {
	to := make(map[string]v1.HugetlbStats, len(from))
	for k, v := range from {
		to[k] = v1.HugetlbStats{
			Usage:    v.Usage,
			MaxUsage: v.MaxUsage,
			Failcnt:  v.Failcnt,
		}
	}
	return to
}

func convertMemory(from cgroups.MemoryStats) *v1.MemoryStats {
	return &v1.MemoryStats{
		Cache:       from.Cache,
		Usage:       convertMemoryData(from.Usage),
		SwapUsage:   convertMemoryData(from.SwapUsage),
		KernelUsage: convertMemoryData(from.KernelUsage),
		Stats:       from.Stats,
	}
}

func convertMemoryData(from cgroups.MemoryData) v1.MemoryData {
	return v1.MemoryData{
		Usage:    from.Usage,
		MaxUsage: from.MaxUsage,
		Failcnt:  from.Failcnt,
	}
}
