package process

import (
	"fmt"

	"github.com/jimmidyson/wurzel/api/v1"
	"github.com/shirou/gopsutil/process"
)

func ProcessIDs() ([]int32, error) {
	return process.Pids()
}

func Processes() ([]v1.Process, error) {
	pids, err := ProcessIDs()
	if err != nil {
		return nil, err
	}

	processes := make([]v1.Process, len(pids))
	for i, pid := range pids {
		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}

		name, err := p.Name()
		if err != nil {
			return nil, fmt.Errorf("could not get process name for pid %d: %v", pid, err)
		}

		status, err := p.Status()
		if err != nil {
			return nil, fmt.Errorf("could not get process status for pid %d: %v", pid, err)
		}

		uids, err := p.Uids()
		if err != nil && !isNotImplementedError(err) {
			return nil, fmt.Errorf("could not get process uids for pid %d: %v", pid, err)
		}

		gids, err := p.Gids()
		if err != nil && !isNotImplementedError(err) {
			return nil, fmt.Errorf("could not get process for pid %d: %v", pid, err)
		}

		threads, err := p.NumThreads()
		if err != nil && !isNotImplementedError(err) {
			return nil, fmt.Errorf("could not get process number of threads for pid %d: %v", pid, err)
		}

		memoryInfo, err := p.MemoryInfo()
		if err != nil && !isNotImplementedError(err) {
			return nil, fmt.Errorf("could not get process memory info for pid %d: %v", pid, err)
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
			return nil, fmt.Errorf("could not get process memory extra info for pid %d: %v", pid, err)
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
			return nil, fmt.Errorf("could not get process cpu for pid %d: %v", pid, err)
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

		processes[i] = v1.Process{
			Name:     name,
			Status:   status,
			Uids:     uids,
			Gids:     gids,
			Threads:  threads,
			Memory:   memory,
			MemoryEx: memoryEx,
			CPUTime:  cpu,
		}
	}

	return processes, nil
}

func isNotImplementedError(err error) bool {
	return err != nil && err.Error() == "not implemented yet"
}
