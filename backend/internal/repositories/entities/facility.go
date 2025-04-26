package entities

import (
	"github.com/fasim/backend/internal/models"
	"gorm.io/gorm"
)

// InputRequirementEntity maps the relationship between facilities and their required input items
type InputRequirementEntity struct {
	gorm.Model
	ID         int `gorm:"primaryKey;autoIncrement"`
	FacilityID int `gorm:"index:idx_facility_item"`
	ItemID     int `gorm:"index:idx_facility_item"`
	Quantity   int
	Item       ItemEntity `gorm:"foreignKey:ItemID"`
	Facility   *FacilityEntity `gorm:"foreignKey:FacilityID"`
}

func (InputRequirementEntity) TableName() string {
	return "input_requirements"
}

// OutputDefinitionEntity maps the relationship between facilities and their produced output items
type OutputDefinitionEntity struct {
	gorm.Model
	ID         int `gorm:"primaryKey;autoIncrement"`
	FacilityID int `gorm:"index:idx_facility_item_out"`
	ItemID     int `gorm:"index:idx_facility_item_out"`
	Quantity   int
	Item       ItemEntity `gorm:"foreignKey:ItemID"`
	Facility   *FacilityEntity `gorm:"foreignKey:FacilityID"`
}

func (OutputDefinitionEntity) TableName() string {
	return "output_definitions"
}

// FacilityEntity represents a production facility and its input/output relationships
type FacilityEntity struct {
	gorm.Model
	ID               int `gorm:"primaryKey;autoIncrement"`
	Name             string `gorm:"not null;index"`
	Description      string
	ProcessingTime   int64
	InputRequirements []InputRequirementEntity `gorm:"foreignKey:FacilityID"`
	OutputDefinitions []OutputDefinitionEntity `gorm:"foreignKey:FacilityID"`
}

func (FacilityEntity) TableName() string {
	return "facilities"
}

func (e *FacilityEntity) ToModel() *models.Facility {
	// Convert input requirements
	inputReqs := make([]*models.InputRequirement, len(e.InputRequirements))
	for i, input := range e.InputRequirements {
		inputReqs[i] = models.NewInputRequirement(input.Item.ToModel(), input.Quantity)
	}

	// Convert output definitions
	outputDefs := make([]*models.OutputDefinition, len(e.OutputDefinitions))
	for i, output := range e.OutputDefinitions {
		outputDefs[i] = models.NewOutputDefinition(output.Item.ToModel(), output.Quantity)
	}

	return models.NewFacilityFromParams(
		e.ID,
		e.Name,
		e.Description,
		inputReqs,
		outputDefs,
		e.ProcessingTime,
	)
}

// FromModel creates an entity from a domain model
func FacilityEntityFromModel(m *models.Facility) *FacilityEntity {
	facility := &FacilityEntity{
		ID:             m.ID(),
		Name:           m.Name(),
		Description:    m.Description(),
		ProcessingTime: m.ProcessingTime(),
	}

	// Convert input requirements
	facility.InputRequirements = make([]InputRequirementEntity, len(m.InputRequirements()))
	for i, input := range m.InputRequirements() {
		facility.InputRequirements[i] = InputRequirementEntity{
			FacilityID: m.ID(),
			ItemID:     input.Item().ID(),
			Quantity:   input.Quantity(),
		}
	}

	// Convert output definitions
	facility.OutputDefinitions = make([]OutputDefinitionEntity, len(m.OutputDefinitions()))
	for i, output := range m.OutputDefinitions() {
		facility.OutputDefinitions[i] = OutputDefinitionEntity{
			FacilityID: m.ID(),
			ItemID:     output.Item().ID(),
			Quantity:   output.Quantity(),
		}
	}

	return facility
}
