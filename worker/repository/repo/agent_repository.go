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

// NodeResourceGroup represents a group of nodes sharing the same resource configuration
type NodeResourceGroup struct {
	Nodes    []*entity.AgentNodeInfo
	Resource *entity.AgentNodeResource
}

// RegisterAgent registers or updates agent information in a transaction
func (r *AgentRepository) RegisterAgent(agentInfo *entity.AgentInfo, nodeResourceGroups map[string]*NodeResourceGroup) error {
	// Process each resource group
	resources := make([]*entity.AgentNodeResource, 0)
	nodes := make([]*entity.AgentNodeInfo, 0)

	for _, group := range nodeResourceGroups {
		resource := group.Resource
		resources = append(resources, resource)

		groupNodes := group.Nodes
		for _, node := range groupNodes {
			node.NodeResourceID = resource.ID
			nodes = append(nodes, node)
		}
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create repositories with transaction
		agentInfoRepo := NewAgentInfoRepository(tx)
		nodeRepo := NewAgentNodeRepository(tx)
		resourceRepo := NewAgentNodeResourceRepository(tx)

		// Save or update agent info
		if err := agentInfoRepo.VersionSave(agentInfo); err != nil {
			log.Log.Errorw("Failed to save or update agent info",
				"error", err,
				"agent_id", agentInfo.AgentID)
			return fmt.Errorf("failed to save or update agent info: %w", err)
		}

		// Save or update resource info
		if len(resources) > 0 {
			if err := resourceRepo.BatchVersionSave(resources, "agent_node_resource"); err != nil {
				log.Log.Errorw("Failed to save or update resource info",
					"error", err,
					"agent_id", agentInfo.AgentID)
				return fmt.Errorf("failed to update resource info: %w", err)
			}
		}

		// Update or save nodes
		if len(nodes) > 0 {
			if err := nodeRepo.BatchVersionSave(nodes, "agent_node_info"); err != nil {
				log.Log.Errorw("Failed to save or update node info",
					"error", err,
					"agent_id", agentInfo.AgentID)
				return fmt.Errorf("failed to update node info: %w", err)
			}
		}

		return nil
	})
}
