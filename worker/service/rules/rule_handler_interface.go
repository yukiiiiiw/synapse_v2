package rules

import (
	"context"
	entity "synapse/worker/repository/types"
	"synapse/worker/types"
)

// RuleHandler defines the interface for rule processors
type RuleHandler interface {
	// GetName returns the name of the rule
	GetName() string
	// Handle processes the rule with given agents
	Handle(ctx context.Context, req *types.AgentFilterRequest, agents []*entity.AgentInfo) ([]*entity.AgentInfo, error)
	// Register registers the rule to the handler
	Register(service RuleService)
}
