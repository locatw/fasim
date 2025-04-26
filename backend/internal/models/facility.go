package models

// InputRequirement defines the quantity of a specific item required for processing
type InputRequirement struct {
	item     *Item
	quantity int
}

func NewInputRequirement(item *Item, quantity int) *InputRequirement {
	return &InputRequirement{
		item:     item,
		quantity: quantity,
	}
}

func (i *InputRequirement) Item() *Item {
	return i.item
}

func (i *InputRequirement) Quantity() int {
	return i.quantity
}

// OutputDefinition specifies the quantity of a specific item produced by processing
type OutputDefinition struct {
	item     *Item
	quantity int
}

func NewOutputDefinition(item *Item, quantity int) *OutputDefinition {
	return &OutputDefinition{
		item:     item,
		quantity: quantity,
	}
}

func (o *OutputDefinition) Item() *Item {
	return o.item
}

func (o *OutputDefinition) Quantity() int {
	return o.quantity
}

// Facility represents a manufacturing unit that transforms input materials into output products
// through a time-based production process
type Facility struct {
	id                int
	name              string
	description       string
	inputRequirements []*InputRequirement
	outputDefinitions []*OutputDefinition
	processingTime    int64
}

// NewFacility creates a new facility with empty input/output requirements
func NewFacility(name string, description string, processingTime int64) *Facility {
	return &Facility{
		name:              name,
		description:       description,
		processingTime:    processingTime,
		inputRequirements: make([]*InputRequirement, 0),
		outputDefinitions: make([]*OutputDefinition, 0),
	}
}

// NewFacilityFromParams creates a facility with all parameters specified.
// Use this function only when creating objects from persisted data, and use NewFacility() for other purposes.
func NewFacilityFromParams(id int, name string, description string, inputReqs []*InputRequirement, outputDefs []*OutputDefinition, processingTime int64) *Facility {
	return &Facility{
		id:                id,
		name:              name,
		description:       description,
		inputRequirements: inputReqs,
		outputDefinitions: outputDefs,
		processingTime:    processingTime,
	}
}

func (f *Facility) ID() int {
	return f.id
}

func (f *Facility) Name() string {
	return f.name
}

func (f *Facility) Description() string {
	return f.description
}

func (f *Facility) InputRequirements() []*InputRequirement {
	return f.inputRequirements
}

func (f *Facility) OutputDefinitions() []*OutputDefinition {
	return f.outputDefinitions
}

func (f *Facility) ProcessingTime() int64 {
	return f.processingTime
}

func (f *Facility) AddInputRequirement(req *InputRequirement) {
	f.inputRequirements = append(f.inputRequirements, req)
}

func (f *Facility) AddOutputDefinition(def *OutputDefinition) {
	f.outputDefinitions = append(f.outputDefinitions, def)
}
