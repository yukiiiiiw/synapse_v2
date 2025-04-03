package repo

import (
	"gorm.io/gorm"
	entity "synapse/worker/repository/types"
)

type AgentNodeRepository struct {
	BaseRepo[*entity.AgentNodeInfo]
}

func NewAgentNodeRepository(db *gorm.DB) *AgentNodeRepository {
	return &AgentNodeRepository{BaseRepo[*entity.AgentNodeInfo]{DB: db}}
}

func (r *AgentNodeRepository) FindByAgentIDAndNodeID(agentID, nodeID string) (*entity.AgentNodeInfo, bool, error) {
	record := new(entity.AgentNodeInfo)
	err := r.DB.Where("agent_id = ? AND agent_node_id = ?", agentID, nodeID).First(record).Error
	return CheckFound(record, err)
}

func (r *AgentNodeRepository) SaveOrUpdate(node *entity.AgentNodeInfo) error {
	return r.DB.Save(node).Error
}

func (r *AgentNodeRepository) FindByAgentIDAndNodeIDs(agentID string, nodeIDs []string) ([]*entity.AgentNodeInfo, error) {
	var nodes []*entity.AgentNodeInfo
	err := r.DB.Where("agent_id = ? AND agent_node_id IN ?", agentID, nodeIDs).Find(&nodes).Error
	return nodes, err
}
