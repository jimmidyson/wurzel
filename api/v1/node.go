package v1

type Node struct {
	CPUInfo []CPUInfo `json:"cpuinfo"`
	Memory  *Memory   `json:"memory"`
	Swap    *Swap     `json:"memory"`
}

type CPUInfo struct {
	CPU        int32    `json:"cpu"`
	VendorID   string   `json:"vendor_id"`
	Family     string   `json:"family"`
	Model      string   `json:"model"`
	Stepping   int32    `json:"stepping"`
	PhysicalID string   `json:"physical_id"`
	CoreID     string   `json:"core_id"`
	Cores      int32    `json:"cores"`
	ModelName  string   `json:"model_name"`
	Mhz        float64  `json:"mhz"`
	CacheSize  int32    `json:"cache_size"`
	Flags      []string `json:"flags"`
}

type Memory struct {
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

type Swap struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
	Sin         uint64  `json:"sin"`
	Sout        uint64  `json:"sout"`
}
