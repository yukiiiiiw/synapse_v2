package repo

import (
	"gorm.io/gorm"
	"synapse/worker/repository/types"
)

// RouteChainRepository handles database operations for route chains
type RouteChainRepository struct {
	DB *gorm.DB
}

// NewRouteChainRepository creates a new route chain repository
func NewRouteChainRepository(db *gorm.DB) *RouteChainRepository {
	return &RouteChainRepository{DB: db}
}

// FindByChainName finds a route chain by its name
func (r *RouteChainRepository) FindByChainName(chainName string) (*types.RouteChain, error) {
	var chain types.RouteChain
	err := r.DB.Where("chain_name = ? AND enabled = true", chainName).First(&chain).Error
	if err != nil {
		return nil, err
	}
	return &chain, nil
}
