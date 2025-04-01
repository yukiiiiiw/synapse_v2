package types

// ApplyRequest represents the request for applying an agent service
type ApplyRequest struct {
	SaasPodID   string                 `json:"saas_pod_id"`
	ServiceName string                 `json:"service_name"`
	ServiceInfo map[string]interface{} `json:"service_info"`
	OperateInfo map[string]interface{} `json:"operate_info"`
}

func (r *ApplyRequest) GetServiceName() string {
	return r.ServiceName
}

func (r *ApplyRequest) GetServiceInfo() map[string]interface{} {
	return r.ServiceInfo
}

// UpdateAgentServiceRequest represents the request for updating an agent service
type UpdateAgentServiceRequest struct {
	SaasPodID   string                 `json:"saas_pod_id"`
	ServiceName string                 `json:"service_name"`
	ServiceInfo map[string]interface{} `json:"service_info"`
	OperateInfo map[string]interface{} `json:"operate_info"`
}

func (r *UpdateAgentServiceRequest) GetServiceName() string {
	return r.ServiceName
}

func (r *UpdateAgentServiceRequest) GetServiceInfo() map[string]interface{} {
	return r.ServiceInfo
}
