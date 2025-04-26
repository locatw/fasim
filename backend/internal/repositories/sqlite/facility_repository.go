package sqlite

import (
	"context"

	"github.com/fasim/backend/internal/models"
	"github.com/fasim/backend/internal/repositories"
	"github.com/fasim/backend/internal/repositories/db"
	"github.com/fasim/backend/internal/repositories/entities"
	"gorm.io/gorm"
)

// FacilityRepository implements the FacilityRepository interface using SQLite with GORM
type FacilityRepository struct {
	db *db.DB
}

// NewFacilityRepository creates a new SQLite-backed facility repository
func NewFacilityRepository(db *db.DB) repositories.FacilityRepository {
	return &FacilityRepository{db: db}
}

// Create stores a new facility
func (r *FacilityRepository) Create(ctx context.Context, facility *models.Facility) error {
	entity := entities.FacilityEntityFromModel(facility)
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return err
	}
	facility.ID = entity.ID
	return nil
}

// Get retrieves a facility by ID
func (r *FacilityRepository) Get(ctx context.Context, id int) (*models.Facility, error) {
	var entity entities.FacilityEntity
	if err := r.db.WithContext(ctx).
		Preload("InputRequirements.Item").
		Preload("OutputDefinitions.Item").
		First(&entity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return entity.ToModel(), nil
}

// List retrieves all facilities
func (r *FacilityRepository) List(ctx context.Context) ([]*models.Facility, error) {
	var entities []entities.FacilityEntity
	if err := r.db.WithContext(ctx).
		Preload("InputRequirements.Item").
		Preload("OutputDefinitions.Item").
		Find(&entities).Error; err != nil {
		return nil, err
	}

	facilities := make([]*models.Facility, len(entities))
	for i, entity := range entities {
		facilities[i] = entity.ToModel()
	}
	return facilities, nil
}

// Update updates an existing facility
func (r *FacilityRepository) Update(ctx context.Context, facility *models.Facility) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if facility exists
		var count int64
		if err := tx.Model(&entities.FacilityEntity{}).Where("id = ?", facility.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}

		// Delete existing relationships
		if err := tx.Where("facility_id = ?", facility.ID).Delete(&entities.InputRequirementEntity{}).Error; err != nil {
			return err
		}
		if err := tx.Where("facility_id = ?", facility.ID).Delete(&entities.OutputDefinitionEntity{}).Error; err != nil {
			return err
		}

		// Create new entity with relationships
		entity := entities.FacilityEntityFromModel(facility)

		// Update facility
		if err := tx.Model(&entities.FacilityEntity{}).
			Where("id = ?", facility.ID).
			Updates(map[string]interface{}{
				"name":            entity.Name,
				"description":     entity.Description,
				"processing_time": entity.ProcessingTime,
			}).Error; err != nil {
			return err
		}

		// Create new relationships
		if err := tx.Create(&entity.InputRequirements).Error; err != nil {
			return err
		}
		if err := tx.Create(&entity.OutputDefinitions).Error; err != nil {
			return err
		}

		return nil
	})
}

// Delete removes a facility by ID
func (r *FacilityRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if facility exists
		var count int64
		if err := tx.Model(&entities.FacilityEntity{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}

		// Delete relationships first
		if err := tx.Where("facility_id = ?", id).Delete(&entities.InputRequirementEntity{}).Error; err != nil {
			return err
		}
		if err := tx.Where("facility_id = ?", id).Delete(&entities.OutputDefinitionEntity{}).Error; err != nil {
			return err
		}

		// Delete facility
		return tx.Delete(&entities.FacilityEntity{}, id).Error
	})
}
