package types

type RouteRule struct {
	ID          int64  `gorm:"primarykey" json:"id"`
	RuleCode    string `json:"ruleCode" gorm:"unique;not null"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled" gorm:"default:true"`
	Priority    int    `json:"priority" gorm:"default:0"`
	CreatedAt   int64  `json:"createdAt" gorm:"not null"`
	UpdatedAt   int64  `json:"updatedAt"`
	Version     int    `json:"version" gorm:"not null"`
}

func (*RouteRule) TableName() string {
	return "route_rule"
}
