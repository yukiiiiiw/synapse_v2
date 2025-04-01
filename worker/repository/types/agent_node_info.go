package types

type AgentNodeInfo struct {
	ID               int64       `gorm:"primarykey" json:"id"`
	AgentID          string      `json:"agentId" gorm:"not null"`
	AgentNodeID      string      `json:"agentNodeId" gorm:"not null"`
	NodeStatus       int         `json:"nodeStatus" gorm:"not null"`
	Status           int         `json:"status" gorm:"not null"`
	ResourceType     int         `json:"resourceType" gorm:"not null"`
	MetricInfo       interface{} `json:"metricInfo" gorm:"type:jsonb;not null"`
	ResourceFreeInfo interface{} `json:"resourceFreeInfo" gorm:"type:jsonb;not null"`
	BaseImage        string      `json:"baseImage"`
	Location         string      `json:"location" gorm:"not null"`
	Region           string      `json:"region" gorm:"not null"`
	GPUType          string      `json:"gpuType" gorm:"not null"`
	TotalGPUCount    int         `json:"totalGpuCount" gorm:"default:0"`
	TotalVRAM        int         `json:"totalVram" gorm:"column:total_vram;default:0"`
	TotalRAM         int         `json:"totalRam" gorm:"default:0"`
	TotalVCPU        int         `json:"totalVcpu" gorm:"column:total_vcpu;default:0"`
	NodeResourceID   int64       `json:"nodeResourceId" gorm:"not null"`
	CreatedAt        int64       `json:"createdAt" gorm:"not null"`
	UpdatedAt        int64       `json:"updatedAt"`
	LastHeartBeat    int64       `json:"lastHeartBeat"`
	Version          int         `json:"version" gorm:"not null"`
}

func (*AgentNodeInfo) TableName() string {
	return "agent_node_info"
}
