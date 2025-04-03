package repo

import (
	"gorm.io/gorm"
	entity "synapse/worker/repository/types"
)

type ServiceOperateLogRepository struct {
	BaseRepo[*entity.ServiceOperateLog]
}

func NewServiceOperateLogRepository(db *gorm.DB) *ServiceOperateLogRepository {
	return &ServiceOperateLogRepository{BaseRepo[*entity.ServiceOperateLog]{DB: db}}
}

func (r *ServiceOperateLogRepository) FindByID(id uint64) (*entity.ServiceOperateLog, bool, error) {
	record := new(entity.ServiceOperateLog)
	err := r.DB.Where("id = ?", id).First(record).Error
	return CheckFound(record, err)
}

func (r *ServiceOperateLogRepository) FindByAgentID(agentID int64) ([]*entity.ServiceOperateLog, error) {
	var records []*entity.ServiceOperateLog
	err := r.DB.Where("agent_id = ?", agentID).Find(&records).Error
	return records, err
}

func (r *ServiceOperateLogRepository) SaveOrUpdate(log *entity.ServiceOperateLog) error {
	return r.DB.Save(log).Error
}

func (r *ServiceOperateLogRepository) FindByServiceInfoID(serviceInfoID int64) (*entity.ServiceOperateLog, bool, error) {
	record := new(entity.ServiceOperateLog)
	err := r.DB.Where("service_info_id = ? AND operate_status = 0", serviceInfoID).First(record).Error
	return CheckFound(record, err)
}
