package service

import (
	"context"
	"errors"
	"sort"
	"synapse/worker/repository/repo"
	entity "synapse/worker/repository/types"
	"synapse/worker/service/rules"
	"synapse/worker/types"
)

// RuleChainService manages the execution of rule chains
type RuleChainService struct {
	routeRuleRepo  *repo.RouteRuleRepository
	routeChainRepo *repo.RouteChainRepository
	ruleHandlers   map[string]rules.RuleHandler
}

// NewRuleChainService creates a new rule chain service
func NewRuleChainService(
	routeRuleRepo *repo.RouteRuleRepository,
	routeChainRepo *repo.RouteChainRepository,
) *RuleChainService {
	return &RuleChainService{
		routeRuleRepo:  routeRuleRepo,
		routeChainRepo: routeChainRepo,
		ruleHandlers:   make(map[string]rules.RuleHandler),
	}
}

// RegisterRuleHandler implements interface
func (s *RuleChainService) RegisterRuleHandler(handler rules.RuleHandler) {
	s.ruleHandlers[handler.GetName()] = handler
}

// ExecuteChain executes a rule chain with the given agents
func (s *RuleChainService) ExecuteChain(ctx context.Context, chainName string, req *types.AgentFilterRequest) ([]*entity.AgentInfo, error) {
	// Get rule chain configuration
	chain, err := s.routeChainRepo.FindByChainName(chainName)
	if err != nil {
		return nil, err
	}

	// Get rule list
	rules, err := s.routeRuleRepo.FindByRuleCodes(chain.RuleCodes)
	if err != nil {
		return nil, err
	}

	// Sort rules by priority
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

	// Execute rule chain
	result := make([]*entity.AgentInfo, 0)
	for _, rule := range rules {
		handler, ok := s.ruleHandlers[rule.RuleCode]
		if !ok {
			return nil, errors.New("rule handler not found: " + rule.RuleCode)
		}

		result, err = handler.Handle(ctx, req, result)
		if err != nil {
			return nil, err
		}

		if len(result) == 0 {
			break
		}
	}

	return result, nil
}
