package repo

import (
	"gorm.io/gorm"
	entity "synapse/worker/repository/types"
)

type AgentInfoRepository struct {
	BaseRepo[*entity.AgentInfo]
}

func NewAgentInfoRepository(db *gorm.DB) *AgentInfoRepository {
	return &AgentInfoRepository{BaseRepo[*entity.AgentInfo]{DB: db}}
}

func (r *AgentInfoRepository) FindByAgentID(agentID string) (*entity.AgentInfo, bool, error) {
	record := new(entity.AgentInfo)
	err := r.DB.Where("agent_id = ?", agentID).First(record).Error
	return CheckFound(record, err)
}

func (r *AgentInfoRepository) SaveOrUpdate(agent *entity.AgentInfo) error {
	return r.DB.Save(agent).Error
}
