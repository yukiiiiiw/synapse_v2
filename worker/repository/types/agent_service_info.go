package types

type AgentServiceInfo struct {
	ID               int64       `gorm:"primarykey" json:"id"`
	SaasPodID        int64       `json:"saasPodId" gorm:"not null"`          // saas podId as k8s service name
	AgentID          string      `json:"agentId" gorm:"null"`                // k8s clusterId, strategy hit asynced
	ServiceName      string      `json:"serviceName" gorm:"null"`            // k8s service name
	ServiceInfo      interface{} `json:"serviceInfo" gorm:"type:jsonb;null"` // k8s service config (image,ip,port,gpu,cpu,disk...)
	MetricInfo       interface{} `json:"metricInfo" gorm:"type:jsonb;null"`  // pod realtime metric info (nodePort ingress pv http/tcp)
	Status           int         `json:"status" gorm:"not null"`             // -1:failure 0:deploying 10:updating 11:deleting 12:pausing 20:running 21:deleted 22:paused
	CreatedAt        int64       `json:"createdAt" gorm:"not null"`
	UpdatedAt        int64       `json:"updatedAt"`
	ServiceUpdatedAt int64       `json:"serviceUpdatedAt"`
	LastHeartBeat    int64       `json:"lastHeartBeat"`
	Version          int         `json:"version" gorm:"not null"` // Optimistic Locking
}

func (*AgentServiceInfo) TableName() string {
	return "agent_service_info"
}
