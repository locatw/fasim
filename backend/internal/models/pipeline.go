package models

// PipelineNode represents a facility within a production line, connected to other facilities
// to form a directed flow of materials and products
type PipelineNode struct {
	ID          int      `json:"id"`
	Facility    Facility `json:"facility"`
	NextNodeIDs []int    `json:"nextNodeIds"`
}

// Pipeline represents a manufacturing line that connects multiple facilities together
// to create a complete production process
type Pipeline struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Nodes       map[int]PipelineNode   `json:"nodes"`
}
