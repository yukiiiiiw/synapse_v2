package repo

import (
	"fmt"
	"gorm.io/gorm"
	"synapse/common/log"
	entity "synapse/worker/repository/types"
)

type AgentRepository struct {
	db *gorm.DB
}

func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) RegisterAgent(agent *entity.AgentInfo, node *entity.AgentNodeInfo, resource *entity.AgentNodeResource) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create repositories with transaction
		agentInfoRepo := NewAgentInfoRepository(tx)
		nodeRepo := NewAgentNodeRepository(tx)
		resourceRepo := NewAgentNodeResourceRepository(tx)

		// Save or update agent info
		if agent.ID == 0 {
			if err := agentInfoRepo.Save(agent); err != nil {
				log.Log.Errorw("Failed to save agent info",
					"error", err,
					"agent_id", agent.AgentID)
				return fmt.Errorf("failed to save agent info: %w", err)
			}
		} else {
			agent.Version += 1
			if err := agentInfoRepo.VersionSave(agent); err != nil {
				log.Log.Errorw("Failed to update agent info",
					"error", err,
					"agent_id", agent.AgentID)
				return fmt.Errorf("failed to update agent info: %w", err)
			}
		}

		// Save or update resource info
		if resource.ID == 0 {
			if err := resourceRepo.Save(resource); err != nil {
				log.Log.Errorw("Failed to save resource info",
					"error", err,
					"resource_md5", resource.ResourceMD5)
				return fmt.Errorf("failed to save resource info: %w", err)
			}
		} else {
			resource.Version += 1
			if err := resourceRepo.VersionSave(resource); err != nil {
				log.Log.Errorw("Failed to update resource info",
					"error", err,
					"resource_md5", resource.ResourceMD5)
				return fmt.Errorf("failed to update resource info: %w", err)
			}
		}

		// Set node resource ID and save or update node info
		node.NodeResourceID = resource.ID
		if node.ID == 0 {
			if err := nodeRepo.Save(node); err != nil {
				log.Log.Errorw("Failed to save node info",
					"error", err,
					"agent_id", node.AgentID,
					"node_id", node.AgentNodeID)
				return fmt.Errorf("failed to save node info: %w", err)
			}
		} else {
			node.Version += 1
			if err := nodeRepo.VersionSave(node); err != nil {
				log.Log.Errorw("Failed to update node info",
					"error", err,
					"agent_id", node.AgentID,
					"node_id", node.AgentNodeID)
				return fmt.Errorf("failed to update node info: %w", err)
			}
		}

		return nil
	})
}
