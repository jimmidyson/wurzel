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
	Created  int64            `json:"created"`
	Memory   *ProcessMemory   `json:"memory"`
	MemoryEx *ProcessMemoryEx `json:"memoryex,omitempty"`
	CPUTime  *CPUTime         `json:"cputime"`
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

// ThrottlingData holds data on CPU throttling.
type ThrottlingData struct {
	// Number of periods with throttling active
	Periods uint64 `json:"periods,omitempty"`
	// Number of periods when the container hit its throttling limit.
	ThrottledPeriods uint64 `json:"throttled_periods,omitempty"`
	// Aggregate time the container was throttled for in nanoseconds.
	ThrottledTime uint64 `json:"throttled_time,omitempty"`
}

// CPUUsage holds all CPU stats, aggregated since cgroup inception.
type CPUUsage struct {
	// Total CPU time consumed.
	// Units: nanoseconds.
	TotalUsage uint64 `json:"total_usage,omitempty"`
	// Total CPU time consumed per core.
	// Units: nanoseconds.
	PerCPUUsage []uint64 `json:"percpu_usage,omitempty"`
	// Time spent by tasks of the cgroup in kernel mode.
	// Units: nanoseconds.
	UsageInKernelmode uint64 `json:"usage_in_kernelmode"`
	// Time spent by tasks of the cgroup in user mode.
	// Units: nanoseconds.
	UsageInUsermode uint64 `json:"usage_in_usermode"`
}

// CPUStats holds stats on CPU usage.
type CPUStats struct {
	CPUUsage       *CPUUsage       `json:"cpu_usage,omitempty"`
	ThrottlingData *ThrottlingData `json:"throttling_data,omitempty"`
}

// MemoryData holds stats on memory usage.
type MemoryData struct {
	Usage    uint64 `json:"usage,omitempty"`
	MaxUsage uint64 `json:"max_usage,omitempty"`
	Failcnt  uint64 `json:"failcnt"`
}

// MemoryStats holds stats on memory.
type MemoryStats struct {
	// memory used for cache
	Cache uint64 `json:"cache,omitempty"`
	// usage of memory
	Usage MemoryData `json:"usage,omitempty"`
	// usage of memory + swap
	SwapUsage MemoryData `json:"swap_usage,omitempty"`
	// usafe of kernel memory
	KernelUsage MemoryData        `json:"kernel_usage,omitempty"`
	Stats       map[string]uint64 `json:"stats,omitempty"`
}

// PIDsStats holds stats of process IDs.
type PIDsStats struct {
	// number of pids in the cgroup
	Current uint64 `json:"current,omitempty"`
}

// BlkioStatEntry holds stats on single blkio.
type BlkioStatEntry struct {
	Major uint64 `json:"major,omitempty"`
	Minor uint64 `json:"minor,omitempty"`
	Op    string `json:"op,omitempty"`
	Value uint64 `json:"value,omitempty"`
}

// BlkioStats holds block device stats.
type BlkioStats struct {
	// number of bytes tranferred to and from the block device
	IoServiceBytesRecursive []BlkioStatEntry `json:"io_service_bytes_recursive,omitempty"`
	IoServicedRecursive     []BlkioStatEntry `json:"io_serviced_recursive,omitempty"`
	IoQueuedRecursive       []BlkioStatEntry `json:"io_queue_recursive,omitempty"`
	IoServiceTimeRecursive  []BlkioStatEntry `json:"io_service_time_recursive,omitempty"`
	IoWaitTimeRecursive     []BlkioStatEntry `json:"io_wait_time_recursive,omitempty"`
	IoMergedRecursive       []BlkioStatEntry `json:"io_merged_recursive,omitempty"`
	IoTimeRecursive         []BlkioStatEntry `json:"io_time_recursive,omitempty"`
	SectorsRecursive        []BlkioStatEntry `json:"sectors_recursive,omitempty"`
}

// HugetlbStats holds stats on hugetlb.
type HugetlbStats struct {
	// current res_counter usage for hugetlb
	Usage uint64 `json:"usage,omitempty"`
	// maximum usage ever recorded.
	MaxUsage uint64 `json:"max_usage,omitempty"`
	// number of times htgetlb usage allocation failure.
	Failcnt uint64 `json:"failcnt"`
}

// Stats holds cgroup stats.
type Stats struct {
	CPUStats    *CPUStats    `json:"cpu_stats,omitempty"`
	MemoryStats *MemoryStats `json:"memory_stats,omitempty"`
	BlkioStats  *BlkioStats  `json:"blkio_stats,omitempty"`
	// the map is in the format "size of hugepage: stats of the hugepage"
	HugetlbStats map[string]HugetlbStats `json:"hugetlb_stats,omitempty"`
}
