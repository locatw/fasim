package entities

// GetModels returns all entity models for auto-migration
func GetModels() []interface{} {
	return []interface{}{
		&ItemEntity{},
		&FacilityEntity{},
		&InputRequirementEntity{},
		&OutputDefinitionEntity{},
		&PipelineEntity{},
		&PipelineNodeEntity{},
		&PipelineNodeConnectionEntity{},
	}
}
