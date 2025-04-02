package repo

import (
	"gorm.io/gorm"
	entity "synapse/worker/repository/types"
)

type AgentNodeResourceRepository struct {
	BaseRepo[*entity.AgentNodeResource]
}

func NewAgentNodeResourceRepository(db *gorm.DB) *AgentNodeResourceRepository {
	return &AgentNodeResourceRepository{BaseRepo[*entity.AgentNodeResource]{DB: db}}
}

func (r *AgentNodeResourceRepository) FindByResourceMD5(resourceMD5 string) (*entity.AgentNodeResource, bool, error) {
	record := new(entity.AgentNodeResource)
	err := r.DB.Where("resource_md5 = ?", resourceMD5).First(record).Error
	return CheckFound(record, err)
}

func (r *AgentNodeResourceRepository) SaveOrUpdate(resource *entity.AgentNodeResource) error {
	return r.DB.Save(resource).Error
}

func (r *AgentNodeResourceRepository) List(pageMark int64, limit int, sort string) ([]entity.AgentNodeResource, bool, error) {
	var resources []entity.AgentNodeResource
	query := r.DB.Model(&entity.AgentNodeResource{})

	if pageMark > 0 {
		query = query.Where("id > ?", pageMark)
	}

	if sort == "desc" {
		query = query.Order("id desc")
	} else {
		query = query.Order("id asc")
	}

	err := query.Limit(limit + 1).Find(&resources).Error
	if err != nil {
		return nil, false, err
	}

	hasMore := false
	if len(resources) > limit {
		hasMore = true
		resources = resources[:limit]
	}

	return resources, hasMore, nil
}

// FindByResourceMD5s finds all resources by resource MD5s
func (r *AgentNodeResourceRepository) FindByResourceMD5s(resourceMD5s []string) ([]*entity.AgentNodeResource, error) {
	var resources []*entity.AgentNodeResource
	err := r.DB.Where("resource_md5 IN ?", resourceMD5s).Find(&resources).Error
	return resources, err
}
