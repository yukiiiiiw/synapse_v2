package constants

import "synapse/common/enum"

const (
	Restart   string = "restart"
	Pause     string = "pause"
	Terminate string = "terminate"
)

// PodActions defines the mapping of pod actions to their corresponding operate types.
var PodActions = map[string]enum.OperateType{
	Restart:   enum.OperateType_Restart,
	Pause:     enum.OperateType_Pause,
	Terminate: enum.OperateType_Delete,
}

func ValidPodAction(podAction string) (enum.OperateType, bool) {
	operateType, exists := PodActions[podAction]
	return operateType, exists
}
