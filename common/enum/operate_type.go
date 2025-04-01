package enum

type OperateType int

const (
	OperateType_Deploy OperateType = iota
	OperateType_Update
	OperateType_Delete
	OperateType_Pause
	OperateType_Restart
	OperateType_Resume
)
