package rules

import (
	"context"
	"gorm.io/gorm"
	entity "synapse/worker/repository/types"
	"synapse/worker/types"
)

type ConfigMatchRule struct {
	db *gorm.DB
}

func NewConfigMatchRule(db *gorm.DB) *ConfigMatchRule {
	return &ConfigMatchRule{db: db}
}

func (r *ConfigMatchRule) GetName() string {
	return "config_match"
}

func (r *ConfigMatchRule) Handle(ctx context.Context, req *types.AgentFilterRequest, agents []*entity.AgentInfo) ([]*entity.AgentInfo, error) {
	// todo Filter agents based on configuration matching
	matchedAgents := make([]*entity.AgentInfo, 0)

	return matchedAgents, nil
}

func (r *ConfigMatchRule) Register(service RuleService) {
	service.RegisterRuleHandler(r)
}
