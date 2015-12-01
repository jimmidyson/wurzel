package node

import "github.com/jimmidyson/wurzel/api/v1"
import "github.com/shirou/gopsutil/cpu"

func CPUInfo() ([]v1.CPUInfo, error) {
	cpus, err := cpu.CPUInfo()
	if err != nil {
		return nil, err
	}

	info := make([]v1.CPUInfo, len(cpus))

	for i, cpu := range cpus {
		info[i] = v1.CPUInfo{
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

	return info, nil
}
