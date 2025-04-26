package entities

import (
	"github.com/fasim/backend/internal/models"
	"gorm.io/gorm"
)

// ItemEntity represents a material or product that can be used in production processes
type ItemEntity struct {
	gorm.Model
	ID          int    `gorm:"primaryKey;autoIncrement"`
	Name        string `gorm:"not null;index"`
	Description string
}

func (ItemEntity) TableName() string {
	return "items"
}

func (e *ItemEntity) ToModel() *models.Item {
	return &models.Item{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
	}
}

// FromModel creates an entity from a domain model
func ItemEntityFromModel(m *models.Item) *ItemEntity {
	return &ItemEntity{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
	}
}
