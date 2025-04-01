package enum

// OperateStatus the operation completion status.
type OperateStatus int

const (
	Operate_Status_Init      OperateStatus = 0 // Init
	Operate_Status_Completed OperateStatus = 1 // Completed
)
