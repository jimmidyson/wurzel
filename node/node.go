package node

import (
	"github.com/jimmidyson/wurzel/api/v1"
)

func Node() (*v1.Node, error) {
	cpu, err := CPUInfo()
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
		Memory:  mem,
		Swap:    swap,
	}, nil
}
