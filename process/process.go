package process

import (
	"github.com/jimmidyson/wurzel/api/v1"
	"github.com/shirou/gopsutil/process"
)

// IDs returns all the current running process IDs.
func IDs() ([]int32, error) {
	return process.Pids()
}

// List returns information about all the processes running on the node.
func List() ([]v1.Process, error) {
	pids, err := IDs()
	if err != nil {
		return nil, err
	}

	processes := make([]v1.Process, 0, len(pids))
	for _, pid := range pids {
		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}

		name, err := p.Name()
		if err != nil {
			continue
		}

		status, err := p.Status()
		if err != nil {
			continue
		}

		uids, err := p.Uids()
		if err != nil && !isNotImplementedError(err) {
			continue
		}

		gids, err := p.Gids()
		if err != nil && !isNotImplementedError(err) {
			continue
		}

		threads, err := p.NumThreads()
		if err != nil && !isNotImplementedError(err) {
			continue
		}

		memoryInfo, err := p.MemoryInfo()
		if err != nil && !isNotImplementedError(err) {
			continue
		}
		var memory *v1.ProcessMemory
		if memoryInfo != nil {
			memory = &v1.ProcessMemory{
				RSS:  memoryInfo.RSS,
				VMS:  memoryInfo.VMS,
				Swap: memoryInfo.Swap,
			}
		}

		memoryInfoEx, err := p.MemoryInfoEx()
		if err != nil && !isNotImplementedError(err) {
			continue
		}
		var memoryEx *v1.ProcessMemoryEx
		if memoryInfoEx != nil {
			memoryEx = &v1.ProcessMemoryEx{
				RSS:    memoryInfoEx.RSS,
				VMS:    memoryInfoEx.VMS,
				Shared: memoryInfoEx.Shared,
				Text:   memoryInfoEx.Text,
				Lib:    memoryInfoEx.Lib,
				Data:   memoryInfoEx.Data,
				Dirty:  memoryInfoEx.Dirty,
			}
		}

		cpuTime, err := p.CPUTimes()
		if err != nil && !isNotImplementedError(err) {
			continue
		}
		var cpu *v1.CPUTime
		if cpuTime != nil {
			cpu = &v1.CPUTime{
				CPU:       cpuTime.CPU,
				User:      cpuTime.User,
				System:    cpuTime.System,
				Idle:      cpuTime.Idle,
				Nice:      cpuTime.Nice,
				Iowait:    cpuTime.Iowait,
				Irq:       cpuTime.Irq,
				Softirq:   cpuTime.Softirq,
				Steal:     cpuTime.Steal,
				Guest:     cpuTime.Guest,
				GuestNice: cpuTime.GuestNice,
				Stolen:    cpuTime.Stolen,
			}
		}

		processes = append(processes, v1.Process{
			Pid:      p.Pid,
			Name:     name,
			Status:   status,
			Uids:     uids,
			Gids:     gids,
			Threads:  threads,
			Memory:   memory,
			MemoryEx: memoryEx,
			CPUTime:  cpu,
		})
	}

	return processes, nil
}

func isNotImplementedError(err error) bool {
	return err != nil && err.Error() == "not implemented yet"
}
