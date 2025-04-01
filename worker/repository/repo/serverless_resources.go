package repo

import (
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"synapse/worker/repository/types"
)

type ServerlessResourceRepo struct {
	BaseRepo[*types.ServerlessResource]
}

func NewServerlessResourceRepository(db *gorm.DB) *ServerlessResourceRepo {
	return &ServerlessResourceRepo{BaseRepo[*types.ServerlessResource]{DB: db}}
}

func (repo *ServerlessResourceRepo) SaveOrUpdate(entity *types.ServerlessResource) (*types.ServerlessResource, error) {
	resource, exist, err := repo.FindByEndpointId(entity.EndpointId)
	if err != nil {
		return nil, err
	}
	if !exist {
		err := repo.DB.Create(entity).Error
		if err != nil {
			return nil, err
		}
		return entity, nil
	}

	err = copier.CopyWithOption(resource, entity, copier.Option{IgnoreEmpty: true})
	if err != nil {
		return nil, err
	}

	err = repo.DB.Select("Model", "Status", "UpdateAt").Where("id = ?", resource.ID).Updates(&types.ServerlessResource{
		Model:     entity.Model,
		Status:    entity.Status,
		UpdatedAt: entity.UpdatedAt,
	}).Error

	if err != nil {
		return nil, err
	}
	return resource, nil
}

func (repo *ServerlessResourceRepo) FindByEndpointId(id string) (*types.ServerlessResource, bool, error) {
	record := new(types.ServerlessResource)
	err := repo.DB.Where("endpoint_id = ?", id).First(record).Error
	return CheckFound(record, err)
}
