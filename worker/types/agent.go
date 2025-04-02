package types

type AgentRegisterRequest struct {
	Agent Agent `json:"agent"`
}

type AgentReportRequest struct {
	Agent Agent `json:"agent"`
}

type Agent struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Location        string `json:"location"`
	Region          string `json:"region"`
	CloudType       string `json:"cloud_type"`
	MetricTimestamp string `json:"metric_timestamp"`
	Nodes           []Node `json:"nodes"`
}

type Node struct {
	ID                        string    `json:"id"`
	Name                      string    `json:"name"`
	ResourceType              string    `json:"resource_type"`
	IP                        string    `json:"ip"`
	DriverVersion             string    `json:"driver_version"`
	NetworkUpload             string    `json:"network_upload"`
	NetworkDownload           string    `json:"network_download"`
	DiskReadSpeed             string    `json:"disk_read_speed"`
	DiskWriteSpeed            string    `json:"disk_write_speed"`
	RandomType                string    `json:"random_type"`
	StorageType               string    `json:"storage_type"`
	GPUType                   string    `json:"gpu_type"`
	TotalGPUCount             string    `json:"total_gpu_count"`
	TotalVCPU                 string    `json:"total_vcpu"`
	TotalVRAM                 string    `json:"total_vram"`
	TotalRAM                  string    `json:"total_ram"`
	TotalCPU                  string    `json:"total_cpu"`
	TotalMemory               string    `json:"total_memory"`
	TotalStorage              string    `json:"total_storage"`
	TotalPersistentVolume     string    `json:"total_persistent_volume"`
	AllocatedGPUCount         string    `json:"allocated_gpu_count"`
	AllocatedVCPU             string    `json:"allocated_vcpu"`
	AllocatedVRAM             string    `json:"allocated_vram"`
	AllocatedRAM              string    `json:"allocated_ram"`
	AllocatedCPU              string    `json:"allocated_cpu"`
	AllocatedMemory           string    `json:"allocated_memory"`
	AllocatedStorage          string    `json:"allocated_storage"`
	AllocatedPersistentVolume string    `json:"allocated_persistent_volume"`
	Services                  []Service `json:"services"`
}

type Service struct {
	Name                      string  `json:"name"`
	Namespace                 string  `json:"namespace"`
	Image                     string  `json:"image"`
	AllocatedGPUCount         string  `json:"allocated_gpu_count"`
	AllocatedVCPU             string  `json:"allocated_vcpu"`
	AllocatedVRAM             string  `json:"allocated_vram"`
	AllocatedRAM              string  `json:"allocated_ram"`
	AllocatedCPU              string  `json:"allocated_cpu"`
	AllocatedMemory           string  `json:"allocated_memory"`
	AllocatedStorage          string  `json:"allocated_storage"`
	AllocatedPersistentVolume string  `json:"allocated_persistent_volume"`
	ClusterIP                 string  `json:"cluster_ip"`
	ExternalIP                string  `json:"external_ip"`
	VPC                       string  `json:"vpc"`
	Ingress                   Ingress `json:"ingress"`
	Ports                     []Port  `json:"ports"`
	Status                    string  `json:"status"`
	StartAt                   string  `json:"start_at"`
}

type Ingress struct {
	Host        string            `json:"host"`
	TLS         bool              `json:"tls"`
	Path        string            `json:"path"`
	Annotations map[string]string `json:"annotations"`
}

type Port struct {
	Protocol   string `json:"protocol"`
	Port       int    `json:"port"`
	TargetPort int    `json:"target_port"`
	NodePort   int    `json:"node_port"`
}

type AgentNodeResourceResponse struct {
	ID                  string `json:"id"`
	ResourceMD5         string `json:"resource_md5"`
	ResourceType        string `json:"resource_type"`
	Location            string `json:"location"`
	Region              string `json:"region"`
	GPUType             string `json:"gpu_type"`
	DriverVersion       string `json:"driver_version"`
	MaxGPUCount         string `json:"max_gpu_count"`
	MaxVCPU             string `json:"max_vcpu"`
	MaxVRAM             string `json:"max_vram"`
	MaxRAM              string `json:"max_ram"`
	MaxPersistentVolume string `json:"max_persistent_volume"`
	NetworkUpload       string `json:"network_upload"`
	NetworkDownload     string `json:"network_download"`
	DiskReadSpeed       string `json:"disk_read_speed"`
	RandomType          string `json:"random_type"`
	StorageType         string `json:"storage_type"`
}

type AgentNodeResourceListResponse struct {
	Items    []AgentNodeResourceResponse `json:"items"`
	PageMark string                      `json:"page_mark"`
	HasMore  bool                        `json:"has_more"`
	Total    int64                       `json:"total"`
}
