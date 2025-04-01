package types

type ServiceOperateLog struct {
	ID            int64       `gorm:"primarykey" json:"id"`
	ServiceInfoID int64       `json:"serviceInfoId" gorm:"not null"`
	SaasPodID     string      `json:"saasPodId" gorm:"not null"`
	ScheduleInfo  interface{} `json:"scheduleInfo" gorm:"type:jsonb;not null"` // saas required resources
	OperateType   int         `json:"operateType" gorm:"not null"`             // operator type 0:deploy 1:update 2:delete 3:pause 4:resume 5:restart
	Status        int         `json:"status" gorm:"not null"`                  // operator status 0:init 1:completed
	CreatedAt     int64       `json:"createdAt" gorm:"not null"`
	UpdatedAt     int64       `json:"updatedAt"`
	CompletedAt   int64       `json:"completedAt"`
	Version       int         `json:"version" gorm:"not null"` // Optimistic Locking
}

func (*ServiceOperateLog) TableName() string {
	return "service_operate_log"
}
