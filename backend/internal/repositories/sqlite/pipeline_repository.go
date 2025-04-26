package sqlite

import (
	"context"

	"github.com/fasim/backend/internal/models"
	"github.com/fasim/backend/internal/repositories"
	"github.com/fasim/backend/internal/repositories/db"
	"github.com/fasim/backend/internal/repositories/entities"
	"gorm.io/gorm"
)

// PipelineRepository implements the PipelineRepository interface using SQLite with GORM
type PipelineRepository struct {
	db *db.DB
}

// NewPipelineRepository creates a new SQLite-backed pipeline repository
func NewPipelineRepository(db *db.DB) repositories.PipelineRepository {
	return &PipelineRepository{db: db}
}

// Create stores a new pipeline
func (r *PipelineRepository) Create(ctx context.Context, pipeline *models.Pipeline) error {
	entity := entities.PipelineEntityFromModel(pipeline)
	return r.db.WithContext(ctx).Create(entity).Error
}

// Get retrieves a pipeline by ID
func (r *PipelineRepository) Get(ctx context.Context, id string) (*models.Pipeline, error) {
	var entity entities.PipelineEntity
	if err := r.db.WithContext(ctx).
		Preload("Nodes.Facility.InputRequirements.Item").
		Preload("Nodes.Facility.OutputDefinitions.Item").
		First(&entity, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return entity.ToModel(), nil
}

// List retrieves all pipelines
func (r *PipelineRepository) List(ctx context.Context) ([]*models.Pipeline, error) {
	var entities []entities.PipelineEntity
	if err := r.db.WithContext(ctx).
		Preload("Nodes.Facility.InputRequirements.Item").
		Preload("Nodes.Facility.OutputDefinitions.Item").
		Find(&entities).Error; err != nil {
		return nil, err
	}

	pipelines := make([]*models.Pipeline, len(entities))
	for i, entity := range entities {
		pipelines[i] = entity.ToModel()
	}
	return pipelines, nil
}

// Update updates an existing pipeline
func (r *PipelineRepository) Update(ctx context.Context, pipeline *models.Pipeline) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing nodes
		if err := tx.Where("pipeline_id = ?", pipeline.ID).Delete(&entities.PipelineNodeEntity{}).Error; err != nil {
			return err
		}

		// Create new entity with nodes
		entity := entities.PipelineEntityFromModel(pipeline)

		// Update pipeline
		if err := tx.Model(&entities.PipelineEntity{}).
			Where("id = ?", pipeline.ID).
			Updates(map[string]interface{}{
				"name":        entity.Name,
				"description": entity.Description,
			}).Error; err != nil {
			return err
		}

		// Create new nodes
		if err := tx.Create(&entity.Nodes).Error; err != nil {
			return err
		}

		return nil
	})
}

// Delete removes a pipeline by ID
func (r *PipelineRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&entities.PipelineEntity{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
