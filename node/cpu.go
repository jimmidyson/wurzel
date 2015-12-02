package node

import "github.com/jimmidyson/wurzel/api/v1"
import "github.com/shirou/gopsutil/cpu"

func CPUInfo() ([]v1.NodeCPUInfo, error) {
	cpus, err := cpu.CPUInfo()
	if err != nil {
		return nil, err
	}

	ret := make([]v1.NodeCPUInfo, len(cpus))

	for i, cpu := range cpus {
		ret[i] = v1.NodeCPUInfo{
			CPU:        cpu.CPU,
			VendorID:   cpu.VendorID,
			Family:     cpu.Family,
			Model:      cpu.Model,
			Stepping:   cpu.Stepping,
			PhysicalID: cpu.PhysicalID,
			CoreID:     cpu.CoreID,
			Cores:      cpu.Cores,
			ModelName:  cpu.ModelName,
			Mhz:        cpu.Mhz,
			CacheSize:  cpu.CacheSize,
			Flags:      cpu.Flags,
		}
	}

	return ret, nil
}

func CPUTime() ([]v1.NodeCPUTime, error) {
	times, err := cpu.CPUTimes(true)
	if err != nil {
		return nil, err
	}

	ret := make([]v1.NodeCPUTime, len(times))

	for i, time := range times {
		ret[i] = v1.NodeCPUTime{
			CPU:       time.CPU,
			User:      time.User,
			System:    time.System,
			Idle:      time.Idle,
			Nice:      time.Nice,
			Iowait:    time.Iowait,
			Irq:       time.Irq,
			Softirq:   time.Softirq,
			Steal:     time.Steal,
			Guest:     time.Guest,
			GuestNice: time.GuestNice,
			Stolen:    time.Stolen,
		}
	}

	return ret, nil
}
