package constants

const (
	workerLockKeyPrefix = "worker:lock:"
	GpuResourceLockKey  = workerLockKeyPrefix + "gpu_resource:%d"
)
