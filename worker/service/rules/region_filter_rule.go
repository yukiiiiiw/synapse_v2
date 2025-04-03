package rules

import (
	"context"
	"gorm.io/gorm"
	entity "synapse/worker/repository/types"
	"synapse/worker/types"
)

// RegionFilterRule filters agents based on their region
type RegionFilterRule struct {
	db     *gorm.DB
	region string
}

func NewRegionFilterRule(db *gorm.DB, region string) *RegionFilterRule {
	return &RegionFilterRule{db: db, region: region}
}

func (r *RegionFilterRule) GetName() string {
	return "region_filter"
}

func (r *RegionFilterRule) Handle(ctx context.Context, req *types.AgentFilterRequest, agents []*entity.AgentInfo) ([]*entity.AgentInfo, error) {
	if len(agents) == 0 {
		return agents, nil
	}

	// Filter agents by region
	filteredAgents := make([]*entity.AgentInfo, 0)
	for _, agent := range agents {
		if agent.Region == r.region {
			filteredAgents = append(filteredAgents, agent)
		}
	}

	return filteredAgents, nil
}

func (r *RegionFilterRule) Register(service RuleService) {
	service.RegisterRuleHandler(r)
}
