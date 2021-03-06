package node

import (
	"github.com/jimmidyson/wurzel/api/v1"
)

// Info returns info about the node's CPU, memory, etc.
func Info() (*v1.Node, error) {
	cpu, err := CPUInfo()
	if err != nil {
		return nil, err
	}

	times, err := CPUTime()
	if err != nil {
		return nil, err
	}

	mem, err := Memory()
	if err != nil {
		return nil, err
	}

	swap, err := Swap()
	if err != nil {
		return nil, err
	}

	return &v1.Node{
		CPUInfo: cpu,
		CPUTime: times,
		Memory:  mem,
		Swap:    swap,
	}, nil
}
