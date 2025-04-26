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
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create pipeline first
		pipelineEntity := &entities.PipelineEntity{
			Name:        pipeline.Name,
			Description: pipeline.Description,
		}
		if err := tx.Create(pipelineEntity).Error; err != nil {
			return err
		}
		pipeline.ID = pipelineEntity.ID

		// Create nodes with auto-generated IDs
		nodeMap := make(map[int]*entities.PipelineNodeEntity)
		nodeIDMap := make(map[int]int) // Map from temporary ID to actual ID
		for _, node := range pipeline.Nodes {
			nodeEntity := &entities.PipelineNodeEntity{
				PipelineID: pipelineEntity.ID,
				FacilityID: node.Facility.ID,
			}
			if err := tx.Create(nodeEntity).Error; err != nil {
				return err
			}
			nodeMap[node.ID] = nodeEntity
			nodeIDMap[node.ID] = nodeEntity.ID
		}

		// Create node connections using actual IDs
		for _, node := range pipeline.Nodes {
			for _, targetID := range node.NextNodeIDs {
				conn := &entities.PipelineNodeConnectionEntity{
					SourceNodeID: nodeIDMap[node.ID],
					TargetNodeID: nodeIDMap[targetID],
				}
				if err := tx.Create(conn).Error; err != nil {
					return err
				}
			}
		}

		// Get the complete pipeline with all relationships
		var entity entities.PipelineEntity
		if err := tx.Preload("Nodes.Facility").
			Preload("Nodes.NextNodes").
			Preload("Nodes.NextNodes.TargetNode").
			Preload("Nodes.Facility.InputRequirements.Item").
			Preload("Nodes.Facility.OutputDefinitions.Item").
			First(&entity, pipelineEntity.ID).Error; err != nil {
			return err
		}

		// Update the model with actual node IDs
		result := entity.ToModel()
		pipeline.Nodes = result.Nodes

		return nil
	})
}

// Get retrieves a pipeline by ID
func (r *PipelineRepository) Get(ctx context.Context, id int) (*models.Pipeline, error) {
	var entity entities.PipelineEntity
	if err := r.db.WithContext(ctx).
		Preload("Nodes.Facility.InputRequirements.Item").
		Preload("Nodes.Facility.OutputDefinitions.Item").
		Preload("Nodes.NextNodes").
		Preload("Nodes.NextNodes.TargetNode").
		First(&entity, id).Error; err != nil {
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
		Preload("Nodes.NextNodes").
		Preload("Nodes.NextNodes.TargetNode").
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
		// Check if pipeline exists
		var count int64
		if err := tx.Model(&entities.PipelineEntity{}).Where("id = ?", pipeline.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}

		// Delete existing nodes and their connections
		if err := tx.Where("source_node_id IN (SELECT id FROM pipeline_nodes WHERE pipeline_id = ?)", pipeline.ID).
			Delete(&entities.PipelineNodeConnectionEntity{}).Error; err != nil {
			return err
		}
		if err := tx.Where("pipeline_id = ?", pipeline.ID).Delete(&entities.PipelineNodeEntity{}).Error; err != nil {
			return err
		}

		// Update pipeline
		if err := tx.Model(&entities.PipelineEntity{}).
			Where("id = ?", pipeline.ID).
			Updates(map[string]interface{}{
				"name":        pipeline.Name,
				"description": pipeline.Description,
			}).Error; err != nil {
			return err
		}

		// Create nodes with auto-generated IDs
		nodeMap := make(map[int]*entities.PipelineNodeEntity)
		nodeIDMap := make(map[int]int) // Map from temporary ID to actual ID
		for _, node := range pipeline.Nodes {
			nodeEntity := &entities.PipelineNodeEntity{
				PipelineID: pipeline.ID,
				FacilityID: node.Facility.ID,
			}
			if err := tx.Create(nodeEntity).Error; err != nil {
				return err
			}
			nodeMap[node.ID] = nodeEntity
			nodeIDMap[node.ID] = nodeEntity.ID
		}

		// Create node connections using actual IDs
		for _, node := range pipeline.Nodes {
			for _, targetID := range node.NextNodeIDs {
				conn := &entities.PipelineNodeConnectionEntity{
					SourceNodeID: nodeIDMap[node.ID],
					TargetNodeID: nodeIDMap[targetID],
				}
				if err := tx.Create(conn).Error; err != nil {
					return err
				}
			}
		}

		// Get the complete pipeline with all relationships
		var entity entities.PipelineEntity
		if err := tx.Preload("Nodes.Facility").
			Preload("Nodes.NextNodes").
			Preload("Nodes.NextNodes.TargetNode").
			Preload("Nodes.Facility.InputRequirements.Item").
			Preload("Nodes.Facility.OutputDefinitions.Item").
			First(&entity, pipeline.ID).Error; err != nil {
			return err
		}

		// Update the model with actual node IDs
		result := entity.ToModel()
		pipeline.Nodes = result.Nodes

		return nil
	})
}

// Delete removes a pipeline by ID
func (r *PipelineRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if pipeline exists
		var count int64
		if err := tx.Model(&entities.PipelineEntity{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}

		// Delete node connections first
		if err := tx.Where("source_node_id IN (SELECT id FROM pipeline_nodes WHERE pipeline_id = ?)", id).
			Delete(&entities.PipelineNodeConnectionEntity{}).Error; err != nil {
			return err
		}

		// Delete nodes
		if err := tx.Where("pipeline_id = ?", id).Delete(&entities.PipelineNodeEntity{}).Error; err != nil {
			return err
		}

		// Delete pipeline
		return tx.Delete(&entities.PipelineEntity{}, id).Error
	})
}
