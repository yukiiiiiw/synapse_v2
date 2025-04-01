package types

type AgentNodeResource struct {
	ID                  int64   `gorm:"primarykey" json:"id"`
	Status              int     `json:"status" gorm:"not null"`
	ResourceType        int     `json:"resourceType" gorm:"not null"`
	Location            string  `json:"location" gorm:"not null"`
	Region              string  `json:"region" gorm:"not null"`
	GPUType             string  `json:"gpuType"`
	DriverVersion       string  `json:"driverVersion"`
	RandomType          string  `json:"randomType"`
	StorageType         string  `json:"storageType"`
	CloudType           string  `json:"cloudType"`
	MaxGPUCount         int     `json:"maxGpuCount" gorm:"default:0"`
	MaxVCPU             int     `json:"maxVcpu" gorm:"column:max_vcpu;default:0"`
	MaxVRAM             int     `json:"maxVram" gorm:"column:max_vram;default:0"`
	MaxRAM              int     `json:"maxRam" gorm:"default:0"`
	MaxPersistentVolume int     `json:"maxPersistentVolume" gorm:"not null"`
	NetworkUpload       float64 `json:"networkUpload" gorm:"type:numeric(38,18);default:0.0"`
	NetworkDownload     float64 `json:"networkDownload" gorm:"type:numeric(38,18);default:0.0"`
	DiskReadSpeed       float64 `json:"diskReadSpeed" gorm:"type:numeric(38,18);default:0.0"`
	DiskWriteSpeed      float64 `json:"diskWriteSpeed" gorm:"type:numeric(38,18);default:0.0"`
	ResourceMD5         string  `json:"resourceMd5" gorm:"unique;not null"`
	CreatedAt           int64   `json:"createdAt" gorm:"not null"`
	UpdatedAt           int64   `json:"updatedAt"`
	Version             int     `json:"version" gorm:"not null"`
}

func (*AgentNodeResource) TableName() string {
	return "agent_node_resource"
}
