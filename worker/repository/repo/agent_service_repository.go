package repo

import (
	"gorm.io/gorm"
	entity "synapse/worker/repository/types"
)

type AgentServiceRepository struct {
	BaseRepo[*entity.AgentServiceInfo]
}

func NewAgentServiceRepository(db *gorm.DB) *AgentServiceRepository {
	return &AgentServiceRepository{BaseRepo[*entity.AgentServiceInfo]{DB: db}}
}

func (r *AgentServiceRepository) FindBySaasPodID(saasPodID int64) (*entity.AgentServiceInfo, bool, error) {
	record := new(entity.AgentServiceInfo)
	err := r.DB.Where("saas_pod_id = ?", saasPodID).First(record).Error
	return CheckFound(record, err)
}

func (r *AgentServiceRepository) FindByAgentID(agentID string) ([]*entity.AgentServiceInfo, error) {
	var records []*entity.AgentServiceInfo
	err := r.DB.Where("agent_id = ?", agentID).Find(&records).Error
	return records, err
}

func (r *AgentServiceRepository) SaveOrUpdate(service *entity.AgentServiceInfo) error {
	return r.DB.Save(service).Error
}

func (r *AgentServiceRepository) UpdateAgentServiceInfoStatus(info *entity.AgentServiceInfo) (bool, error) {
	result := r.DB.Model(&entity.AgentServiceInfo{}).
		Where("id = ? AND version = ?", info.ID, info.Version).
		Updates(map[string]interface{}{
			"status":  info.Status,
			"version": info.Version + 1,
		})
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, ErrOptimisticLockConflict
	}
	return true, nil
}
