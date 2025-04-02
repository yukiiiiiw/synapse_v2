package constants

import "synapse/common/enum"

const (
	Restart   string = "restart"
	Pause     string = "pause"
	Terminate string = "terminate"
)

type OperateTypeDeploymentStatus struct {
	OperateType      enum.OperateType
	DeploymentStatus enum.DeploymentStatus
}

// PodActions defines the mapping of pod actions to their corresponding operate types.
var PodActionMapping = map[string]*OperateTypeDeploymentStatus{
	Restart: {
		OperateType:      enum.OperateType_Restart,
		DeploymentStatus: enum.Deployment_Status_Restarting,
	},
	Pause: {
		OperateType:      enum.OperateType_Pause,
		DeploymentStatus: enum.Deployment_Status_Pausing,
	},
	Terminate: {
		OperateType:      enum.OperateType_Delete,
		DeploymentStatus: enum.Deployment_Status_Deleting,
	},
}

func CheckAndGetOperateTypeDeploymentStatus(podAction string) (*OperateTypeDeploymentStatus, bool) {
	operateTypeDeploymentStatus, exists := PodActionMapping[podAction]
	return operateTypeDeploymentStatus, exists
}
