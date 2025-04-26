package repositories

import (
	"context"

	"github.com/fasim/backend/internal/models"
)

// ItemRepository provides CRUD operations for items in the storage layer
type ItemRepository interface {
	Create(ctx context.Context, item *models.Item) error
	Get(ctx context.Context, id string) (*models.Item, error)
	List(ctx context.Context) ([]*models.Item, error)
	Update(ctx context.Context, item *models.Item) error
	Delete(ctx context.Context, id string) error
}

// FacilityRepository provides CRUD operations for facilities in the storage layer
type FacilityRepository interface {
	Create(ctx context.Context, facility *models.Facility) error
	Get(ctx context.Context, id string) (*models.Facility, error)
	List(ctx context.Context) ([]*models.Facility, error)
	Update(ctx context.Context, facility *models.Facility) error
	Delete(ctx context.Context, id string) error
}

// PipelineRepository provides CRUD operations for production pipelines in the storage layer
type PipelineRepository interface {
	Create(ctx context.Context, pipeline *models.Pipeline) error
	Get(ctx context.Context, id string) (*models.Pipeline, error)
	List(ctx context.Context) ([]*models.Pipeline, error)
	Update(ctx context.Context, pipeline *models.Pipeline) error
	Delete(ctx context.Context, id string) error
}

// Repositories provides access to all storage operations through a unified interface
type Repositories struct {
	Items      ItemRepository
	Facilities FacilityRepository
	Pipelines  PipelineRepository
}
