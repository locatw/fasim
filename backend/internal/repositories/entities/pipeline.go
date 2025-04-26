package entities

import (
	"github.com/fasim/backend/internal/models"
	"gorm.io/gorm"
)

// PipelineNodeEntity represents a facility node within a production pipeline
type PipelineNodeEntity struct {
	gorm.Model
	ID           int `gorm:"primaryKey;autoIncrement"`
	PipelineID   int `gorm:"index:idx_pipeline_facility"`
	FacilityID   int `gorm:"index:idx_pipeline_facility"`
	Facility     FacilityEntity `gorm:"foreignKey:FacilityID"`
	Pipeline     *PipelineEntity `gorm:"foreignKey:PipelineID"`
	NextNodes    []PipelineNodeConnectionEntity `gorm:"foreignKey:SourceNodeID"`
}

func (PipelineNodeEntity) TableName() string {
	return "pipeline_nodes"
}

// PipelineNodeConnectionEntity represents a connection between two pipeline nodes
type PipelineNodeConnectionEntity struct {
	gorm.Model
	ID           int `gorm:"primaryKey;autoIncrement"`
	SourceNodeID int `gorm:"index"`
	TargetNodeID int `gorm:"index"`
	SourceNode   PipelineNodeEntity `gorm:"foreignKey:SourceNodeID"`
	TargetNode   PipelineNodeEntity `gorm:"foreignKey:TargetNodeID"`
}

func (PipelineNodeConnectionEntity) TableName() string {
	return "pipeline_node_connections"
}

func (e *PipelineNodeEntity) GetNextNodeIDs() []int {
	ids := make([]int, len(e.NextNodes))
	for i, conn := range e.NextNodes {
		ids[i] = conn.TargetNodeID
	}
	return ids
}

func (e *PipelineNodeEntity) ToModel() *models.PipelineNode {
	node := &models.PipelineNode{
		ID:          e.ID,
		Facility:    *e.Facility.ToModel(),
		NextNodeIDs: make([]int, len(e.NextNodes)),
	}
	// Add next node IDs from connections
	for i, conn := range e.NextNodes {
		node.NextNodeIDs[i] = conn.TargetNodeID
	}
	return node
}

// PipelineEntity represents a complete production line configuration,
// consisting of interconnected facility nodes
type PipelineEntity struct {
	gorm.Model
	ID          int `gorm:"primaryKey;autoIncrement"`
	Name        string `gorm:"not null;index"`
	Description string
	Nodes       []PipelineNodeEntity `gorm:"foreignKey:PipelineID"`
}

func (PipelineEntity) TableName() string {
	return "pipelines"
}

func (e *PipelineEntity) ToModel() *models.Pipeline {
	pipeline := &models.Pipeline{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		Nodes:       make(map[int]models.PipelineNode),
	}

	for _, node := range e.Nodes {
		nodeModel := node.ToModel()
		pipeline.Nodes[nodeModel.ID] = *nodeModel
	}

	return pipeline
}

// FromModel creates an entity from a domain model
func PipelineEntityFromModel(m *models.Pipeline) *PipelineEntity {
	pipeline := &PipelineEntity{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Nodes:       make([]PipelineNodeEntity, 0, len(m.Nodes)),
	}

	// Create nodes first
	nodeMap := make(map[int]*PipelineNodeEntity) // Map from model node ID to entity node
	for _, node := range m.Nodes {
		pipelineNode := &PipelineNodeEntity{
			PipelineID: m.ID,
			FacilityID: node.Facility.ID,
			NextNodes:  make([]PipelineNodeConnectionEntity, 0),
		}
		pipeline.Nodes = append(pipeline.Nodes, *pipelineNode)
		nodeMap[node.ID] = pipelineNode
	}

	// Create connections using the node map
	for _, node := range m.Nodes {
		sourceNode := nodeMap[node.ID]
		for _, targetID := range node.NextNodeIDs {
			targetNode := nodeMap[targetID]
			connection := PipelineNodeConnectionEntity{
				SourceNodeID: sourceNode.ID,
				TargetNodeID: targetNode.ID,
			}
			sourceNode.NextNodes = append(sourceNode.NextNodes, connection)
		}
	}

	// Update nodes in the pipeline with connections
	for i, node := range pipeline.Nodes {
		if pNode := nodeMap[node.ID]; pNode != nil {
			pipeline.Nodes[i].NextNodes = pNode.NextNodes
		}
	}

	return pipeline
}
