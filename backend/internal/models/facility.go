package models

// InputRequirement specifies the type and amount of an item required for processing
type InputRequirement struct {
	Item     Item `json:"item"`
	Quantity int  `json:"quantity"`
}

// OutputDefinition specifies the type and amount of an item produced by processing
type OutputDefinition struct {
	Item     Item `json:"item"`
	Quantity int  `json:"quantity"`
}

// Facility represents a production unit that transforms input items into output items
// through a manufacturing process that takes a specified amount of time
type Facility struct {
	ID                int                `json:"id"`
	Name              string             `json:"name"`
	Description       string             `json:"description,omitempty"`
	InputRequirements []InputRequirement `json:"inputRequirements"`
	OutputDefinitions []OutputDefinition `json:"outputDefinitions"`
	ProcessingTime    int64              `json:"processingTime"`
}
