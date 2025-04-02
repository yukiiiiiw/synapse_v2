package types

import (
	"synapse/common/enum"
	"time"
)

type PageResponse[T any] struct {
	Items    []T   `json:"records"`
	PageSize int   `json:"pageSize"`
	PageMark int64 `json:"pageMark"`
	HasMore  bool  `json:"hasMore"`
}

type GpuResourceItem struct {
	ID                     int64          `json:"id"`
	CloudType              enum.CloudType `json:"cloudType" binding:"required"`
	GpuType                string         `json:"gpuType"`
	DriverVersion          string         `json:"driverVersion"`
	GpuCardMaxAvailableNum int            `json:"gpuCardMaxAvailableNum"`
	SingleCardVram         int            `json:"singleCardVram"` //single card video ram in GB
	SingleCardRam          int            `json:"singleCardRam"`  //single card ram in GB
	SingleCardVcpu         int            `json:"singleCardVcpu"` //single card ram GPU
	SingleCardPrice        float64        `json:"singleCardPrice"`
	PersistentVolumeLimit  int            `json:"singleStorageLimit"`
	ContainerVolumeLimit   int            `json:"containerVolumeLimit"`
	PersistentVolumePrice  float64        `json:"persistentVolumePrice"`
	ContainerVolumePrice   float64        `json:"containerVolumePrice"`
	RamType                string         `json:"ramType"`
	Region                 string         `json:"region"`
	Status                 int            `json:"status"`
}

type ApplyGpuResourceRequest struct {
	SassPodID             int64                  `json:"podId" binding:"required"`
	Image                 string                 `json:"image" binding:"required"`
	Region                string                 `json:"region" binding:"required"`
	CloudType             enum.CloudType         `json:"cloudType" binding:"required"`
	ResourceType          enum.ResourceType      `json:"resourceType" binding:"required"`
	GpuType               string                 `json:"gpuType" binding:"required"`
	GpuCount              int                    `json:"gpuCount" binding:"required"`
	SingleCardVram        int                    `json:"singleCardVram" binding:"required"`
	SingleCardRam         int                    `json:"singleCardRam" binding:"required"`
	SingleCardVcpu        int                    `json:"singleCardVcpu" binding:"required"`
	ContainerVolume       int                    `json:"containerVolume" binding:"required"`
	PersistentVolume      int                    `json:"persistentVolume"`
	PersistentVolumeId    int                    `json:"persistentVolumeId"`
	PersistentMountPath   string                 `json:"persistentMountPath"`
	SingleCardPrice       float64                `json:"singleCardPrice"`
	PersistentVolumePrice float64                `json:"persistentVolumePrice"`
	ContainerVolumePrice  float64                `json:"containerVolumePrice"`
	InitCommand           string                 `json:"initializationCommand"`
	EnvVars               []KeyValuePair[string] `json:"environmentVars" binding:"omitempty,dive"`
	Expose                []PortMapping          `json:"expose" binding:"omitempty,dive"`
	SshUser               string                 `json:"sshUser"`
	SshPublicKey          string                 `json:"sshPublicKey"`
}
type EditGpuPodRequest struct {
	SassPodID           int64                  `json:"podId" binding:"required"`                  // Pod ID, used to identify the Pod to be edited
	Image               string                 `json:"image" binding:"required"`                  // Container image
	ContainerVolume     int                    `json:"containerVolume" binding:"required"`        // Container volume size GB
	PersistentVolume    int                    `json:"persistentVolume"`                          // Persistent volume size GB
	PersistentVolumeId  int                    `json:"persistentVolumeId"`                        // Persistent volume ID
	PersistentMountPath string                 `json:"persistentMountPath"`                       // Persistent volume mount path
	InitCommand         string                 `json:"initializationCommand"`                     // Initialization command
	EnvVars             []KeyValuePair[string] `json:"environmentVars"  binding:"omitempty,dive"` // Environment variables
	Expose              []PortMapping          `json:"expose" binding:"omitempty,dive"`           // Port exposure configuration
	SshUser             string                 `json:"sshUser"`                                   // SSH username
	SshPublicKey        string                 `json:"sshPublicKey"`                              // SSH public key
}
type GpuStatusResponse struct {
	Image                 string                 `json:"image"`
	CloudType             enum.CloudType         `json:"cloudType" binding:"required"`
	ResourceType          enum.ResourceType      `json:"resourceType" binding:"required"`
	GpuType               string                 `json:"gpuType"`
	GpuCount              int                    `json:"gpuCount"`
	SingleCardVram        int                    `json:"singleCardVram"`
	SingleCardRam         int                    `json:"singleCardRam"`
	SingleCardVcpu        int                    `json:"singleCardVcpu"`
	Location              string                 `json:"location"`
	Region                string                 `json:"region"`
	ContainerVolume       int                    `json:"containerVolume"`
	PersistentVolume      int                    `json:"persistentVolume"`
	PersistentVolumeId    int                    `json:"persistentVolumeId"`
	PersistentMountPath   string                 `json:"persistentMountPath"`
	NetworkUpload         int                    `json:"networkUpload"`
	NetworkDownload       int                    `json:"networkDownload"`
	DiskReadSpeed         int                    `json:"diskReadSpeed"`
	DiskWriteSpeed        int                    `json:"diskWriteSpeed"`
	SingleCardPrice       float64                `json:"singleCardPrice"`
	PersistentVolumePrice float64                `json:"persistentVolumePrice"`
	ContainerVolumePrice  float64                `json:"containerVolumePrice"`
	InitCommand           string                 `json:"initializationCommand"`
	EnvVars               []KeyValuePair[string] `json:"environmentVars"`
	Expose                []ExposedPort          `json:"expose" `
	SshUser               string                 `json:"sshUser"`
	SshHost               string                 `json:"sshHost"`
	SshPort               string                 `json:"sshPort"`
	Status                int                    `json:"status"`
}

type ExposedPort struct {
	PortMapping
	Host      string `json:"host"`
	ProxyPort int    `json:"proxyPort"`
}

type PortMapping struct {
	Port     int    `json:"port" binding:"required,min=1"`
	Protocol string `json:"protocol" binding:"required,oneof=tcp udp http"`
}

type KeyValuePair[T any] struct {
	Key   string `json:"key" binding:"required"`
	Value T      `json:"value" binding:"required"`
}

// -------------------------------------------------------------------------------------

type CreateServerlessResourceRequest struct {
	EndpointId string `json:"endpointId"    binding:"required"`
	Model      string `json:"model"   binding:"required"`
}

type ServerlessResourceResponse struct {
	ID         int64     `json:"id"`
	EndpointId string    `json:"endpointId"`
	Model      string    `json:"model"`
	Status     int       `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// Inference types
type InferenceMessageRequest struct {
	Temperature       float64       `json:"temperature" binding:"required"`
	TopP              float64       `json:"top_p" binding:"required"`
	MaxTokens         int32         `json:"max_tokens" binding:"required"`
	FrequencyPenalty  float64       `json:"frequency_penalty" binding:"required"`
	PresencePenalty   float64       `json:"presence_penalty" binding:"required"`
	RepetitionPenalty float64       `json:"repetition_penalty" binding:"required"`
	Model             string        `json:"model" binding:"required"`
	Messages          []Message     `json:"messages" binding:"required"`
	Stream            bool          `json:"stream"`
	StreamOptions     StreamOptions `json:"stream_options"`
}

type InferenceMessage struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type ChatCompletion struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int       `json:"index"`
	Message      Message   `json:"message"`
	Logprobs     *Logprobs `json:"logprobs"`
	FinishReason string    `json:"finish_reason"`
	StopReason   *string   `json:"stop_reason"`
}

type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

type ToolCall struct {
	// Define fields for ToolCall if needed
}

type Logprobs struct {
	// Define fields for Logprobs if needed
}

type Usage struct {
	PromptTokens     int       `json:"prompt_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	PromptLogprobs   *Logprobs `json:"prompt_logprobs"`
}

// TextToImage types
type TextToImageRequest struct {
	Model             string  `json:"model" binding:"required"`
	Prompt            string  `json:"prompt"`
	NumInferenceSteps int32   `json:"num_inference_steps" binding:"required"`
	GuidanceScale     float64 `json:"guidance_scale"`
	LoraWeight        float64 `json:"lora_weight"`
	Seed              int32   `json:"seed"`
	Width             int32   `json:"width"`
	Height            int32   `json:"height"`
	PagScale          float64 `json:"pag_scale"`
}

type TextToImageResponse struct {
	Created int64          `json:"created"`
	Data    []*ImageResult `json:"data"`
}

type StatusResponse struct {
	Resources Resources `json:"resources"`
	Models    Models    `json:"models"`
}

type Resources struct {
	TotalNodes int `json:"totalNodes"`
}

type Models struct {
	List []*ModelInfo `json:"list"`
}

type ModelInfo struct {
	Model string `json:"model"`
	Count int    `json:"count"`
	TPM   int    `json:"tpm"`
}

type ImageResult struct {
	Url          string  `json:"url"`
	Latency      float64 `json:"latency"`
	IsSafePrompt bool    `json:"is_safe_prompt"`
}
