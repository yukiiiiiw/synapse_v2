package types

type AgentInfo struct {
	ID               int64       `gorm:"primarykey" json:"id"`
	AgentID          string      `json:"agentId" gorm:"unique;not null"`
	Status           int         `json:"status" gorm:"not null"`
	AgentStatus      int         `json:"agentStatus" gorm:"not null"`
	CloudType        int         `json:"cloudType" gorm:"not null"`
	MetricInfo       interface{} `json:"metricInfo" gorm:"type:jsonb;not null"`
	ResourceFreeInfo interface{} `json:"resourceFreeInfo" gorm:"type:jsonb;not null"`
	BaseImage        string      `json:"baseImage"`
	Location         string      `json:"location" gorm:"not null"`
	Region           string      `json:"region" gorm:"not null"`
	CreatedAt        int64       `json:"createdAt" gorm:"not null"`
	UpdatedAt        int64       `json:"updatedAt"`
	LastHeartBeat    int64       `json:"lastHeartBeat"`
	Version          int         `json:"version" gorm:"not null"`
}

func (*AgentInfo) TableName() string {
	return "agent_info"
}
