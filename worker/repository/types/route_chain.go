package types

// RouteChain defines the structure of a route chain
type RouteChain struct {
	ID        string   `gorm:"primaryKey" json:"id"`
	ChainName string   `gorm:"uniqueIndex" json:"chain_name"`
	RuleCodes []string `gorm:"type:json" json:"rule_codes"`
	Enabled   bool     `gorm:"default:true" json:"enabled"`
}

func (*RouteChain) TableName() string {
	return "route_chain"
}
