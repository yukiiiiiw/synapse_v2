package rules

// RuleService defines the interface for rule service
type RuleService interface {
	// RegisterRuleHandler registers a rule handler to rule chain service
	RegisterRuleHandler(handler RuleHandler)
}
