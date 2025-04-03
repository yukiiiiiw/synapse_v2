package repo

import (
	"gorm.io/gorm"
	entity "synapse/worker/repository/types"
)

type RouteRuleRepository struct {
	BaseRepo[*entity.RouteRule]
}

func NewRouteRuleRepository(db *gorm.DB) *RouteRuleRepository {
	return &RouteRuleRepository{BaseRepo[*entity.RouteRule]{DB: db}}
}

func (r *RouteRuleRepository) FindByRuleCodes(ruleCodes []string) ([]*entity.RouteRule, error) {
	var rules []*entity.RouteRule
	err := r.DB.Where("rule_code IN ? AND enabled = true", ruleCodes).Find(&rules).Error
	return rules, err
}
