package enum

// ResourceStatus the resource usage status.
type ResourceStatus int

const (
	Resource_Status_Offline   ResourceStatus = -1 // Offline
	Resource_Status_Available ResourceStatus = 0  // Available
	Resource_Status_Busy      ResourceStatus = 1  // Busy
)
