package v1

// Node holds the overall node information.
type Node struct {
	CPUInfo []NodeCPUInfo `json:"cpuinfo"`
	CPUTime []CPUTime     `json:"cputime"`
	Memory  *NodeMemory   `json:"memory"`
	Swap    *NodeSwap     `json:"memory"`
}

// NodeCPUInfo holds info about the node's CPUs.
type NodeCPUInfo struct {
	CPU        int32    `json:"cpu"`
	CacheSize  int32    `json:"cache_size"`
	Cores      int32    `json:"cores"`
	Stepping   int32    `json:"stepping"`
	VendorID   string   `json:"vendor_id"`
	Family     string   `json:"family"`
	Model      string   `json:"model"`
	PhysicalID string   `json:"physical_id"`
	CoreID     string   `json:"core_id"`
	ModelName  string   `json:"model_name"`
	Mhz        float64  `json:"mhz"`
	Flags      []string `json:"flags"`
}

// CPUTime holds CPU time information for either a node or a cgroup.
type CPUTime struct {
	CPU       string  `json:"cpu"`
	User      float64 `json:"user"`
	System    float64 `json:"system"`
	Idle      float64 `json:"idle"`
	Nice      float64 `json:"nice"`
	Iowait    float64 `json:"iowait"`
	Irq       float64 `json:"irq"`
	Softirq   float64 `json:"softirq"`
	Steal     float64 `json:"steal"`
	Guest     float64 `json:"guest"`
	GuestNice float64 `json:"guest_nice"`
	Stolen    float64 `json:"stolen"`
}

// NodeMemory holds info on the current state of the node's memory.
type NodeMemory struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
	Free        uint64  `json:"free"`
	Active      uint64  `json:"active"`
	Inactive    uint64  `json:"inactive"`
	Buffers     uint64  `json:"buffers"`
	Cached      uint64  `json:"cached"`
	Wired       uint64  `json:"wired"`
	Shared      uint64  `json:"shared"`
}

// NodeSwap holds info on the current state of the node's swap.
type NodeSwap struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
	Sin         uint64  `json:"sin"`
	Sout        uint64  `json:"sout"`
}

// Process holds info related to a single process.
type Process struct {
	Pid      int32            `json:"pid"`
	Threads  int32            `json:"threads"`
	Name     string           `json:"name"`
	Status   string           `json:"status"`
	Uids     []int32          `json:"uids"`
	Gids     []int32          `json:"gids"`
	Memory   *ProcessMemory   `json:"memory"`
	MemoryEx *ProcessMemoryEx `json:"memoryex,omitempty"`
	CPUTime  *CPUTime         `json:"cputime"`
	Created  int64            `json:"created"`
}

// ProcessMemory holds memory info related to a single process.
type ProcessMemory struct {
	RSS  uint64 `json:"rss"`
	VMS  uint64 `json:"vms"`
	Swap uint64 `json:"swap"`
}

// ProcessMemoryEx holds extra memory info related to a single process, if available.
type ProcessMemoryEx struct {
	RSS    uint64 `json:"rss"`
	VMS    uint64 `json:"vms"`
	Shared uint64 `json:"shared"`
	Text   uint64 `json:"text"`
	Lib    uint64 `json:"lib"`
	Data   uint64 `json:"data"`
	Dirty  uint64 `json:"dirty"`
}
