package node

import (
	"github.com/jimmidyson/wurzel/api/v1"
	"github.com/shirou/gopsutil/mem"
)

func Memory() (*v1.NodeMemory, error) {
	mi, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return &v1.NodeMemory{
		Total:       mi.Total,
		Available:   mi.Available,
		Used:        mi.Used,
		UsedPercent: mi.UsedPercent,
		Free:        mi.Free,
		Active:      mi.Active,
		Inactive:    mi.Inactive,
		Buffers:     mi.Buffers,
		Cached:      mi.Cached,
		Wired:       mi.Wired,
		Shared:      mi.Shared,
	}, nil
}

func Swap() (*v1.NodeSwap, error) {
	mi, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}

	return &v1.NodeSwap{
		Total:       mi.Total,
		Used:        mi.Used,
		Free:        mi.Free,
		UsedPercent: mi.UsedPercent,
		Sin:         mi.Sin,
		Sout:        mi.Sout,
	}, nil
}
