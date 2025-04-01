package types

type RouteChain struct {
	ID           int64       `gorm:"primarykey" json:"id"`
	ChainName    string      `json:"chainName" gorm:"unique;not null"`
	ResourceType int         `json:"resourceType" gorm:"not null"`
	Rules        interface{} `json:"rules" gorm:"type:jsonb"`
	Enabled      bool        `json:"enabled" gorm:"default:true"`
	CreatedAt    int64       `json:"createdAt" gorm:"not null"`
	UpdatedAt    int64       `json:"updatedAt"`
	Version      int         `json:"version" gorm:"not null"`
}

func (*RouteChain) TableName() string {
	return "route_chain"
}
