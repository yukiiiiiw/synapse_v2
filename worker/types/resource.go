package types

// ResourceFreeInfo represents the available resource information of a node
type ResourceFreeInfo struct {
	GPUCount float64 `json:"gpu_count"` // Available GPU count
	VRAM     float64 `json:"vram"`      // Available video memory
	RAM      float64 `json:"ram"`       // Available system memory
	Disk     float64 `json:"disk"`      // Available disk space
}
