package sqlite

import (
	"context"

	"github.com/fasim/backend/internal/models"
	"github.com/fasim/backend/internal/repositories"
	"github.com/fasim/backend/internal/repositories/db"
	"github.com/fasim/backend/internal/repositories/entities"
	"gorm.io/gorm"
)

// ItemRepository implements the ItemRepository interface using SQLite with GORM
type ItemRepository struct {
	db *db.DB
}

// NewItemRepository creates a new SQLite-backed item repository
func NewItemRepository(db *db.DB) repositories.ItemRepository {
	return &ItemRepository{db: db}
}

// Create stores a new item
func (r *ItemRepository) Create(ctx context.Context, item *models.Item) error {
	entity := entities.ItemEntityFromModel(item)
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return err
	}
	item.ID = int(entity.ID)
	return nil
}

// Get retrieves an item by ID
func (r *ItemRepository) Get(ctx context.Context, id int) (*models.Item, error) {
	var entity entities.ItemEntity
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return entity.ToModel(), nil
}

// List retrieves all items
func (r *ItemRepository) List(ctx context.Context) ([]*models.Item, error) {
	var entities []entities.ItemEntity
	if err := r.db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}

	items := make([]*models.Item, len(entities))
	for i, entity := range entities {
		items[i] = entity.ToModel()
	}
	return items, nil
}

// Update updates an existing item
func (r *ItemRepository) Update(ctx context.Context, item *models.Item) error {
	entity := entities.ItemEntityFromModel(item)
	result := r.db.WithContext(ctx).Model(&entities.ItemEntity{}).
		Where("id = ?", item.ID).
		Updates(map[string]interface{}{
			"name":        entity.Name,
			"description": entity.Description,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete removes an item by ID
func (r *ItemRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&entities.ItemEntity{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
