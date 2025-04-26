package models

// PipelineNode represents a facility within a production line, defining its connections
// to downstream facilities to establish material flow
type PipelineNode struct {
	id          int
	facility    *Facility
	nextNodeIDs []int
}

// NewPipelineNode creates a node with no downstream connections
func NewPipelineNode(facility *Facility) *PipelineNode {
	return &PipelineNode{
		facility:    facility,
		nextNodeIDs: make([]int, 0),
	}
}

// NewPipelineNodeFromParams creates a node with all parameters specified, typically used when
// reconstructing a node from persistent storage
func NewPipelineNodeFromParams(id int, facility *Facility, nextNodeIDs []int) *PipelineNode {
	return &PipelineNode{
		id:          id,
		facility:    facility,
		nextNodeIDs: nextNodeIDs,
	}
}

func (n *PipelineNode) ID() int {
	return n.id
}

func (n *PipelineNode) Facility() *Facility {
	return n.facility
}

func (n *PipelineNode) NextNodeIDs() []int {
	return n.nextNodeIDs
}

func (n *PipelineNode) AddNextNodeID(nodeID int) {
	n.nextNodeIDs = append(n.nextNodeIDs, nodeID)
}

// Pipeline represents a manufacturing line that connects multiple facilities
// to create a complete production process with defined material flows
type Pipeline struct {
	id          int
	name        string
	description string
	nodes       map[int]*PipelineNode
}

// NewPipeline creates an empty pipeline with no nodes
func NewPipeline(name string) *Pipeline {
	return &Pipeline{
		name:  name,
		nodes: make(map[int]*PipelineNode),
	}
}

// NewPipelineFromParams creates a pipeline with all parameters specified, typically used when
// reconstructing a pipeline from persistent storage
func NewPipelineFromParams(id int, name string, description string, nodes map[int]*PipelineNode) *Pipeline {
	return &Pipeline{
		id:          id,
		name:        name,
		description: description,
		nodes:       nodes,
	}
}

func (p *Pipeline) ID() int {
	return p.id
}

func (p *Pipeline) Name() string {
	return p.name
}

func (p *Pipeline) Description() string {
	return p.description
}

func (p *Pipeline) Nodes() map[int]*PipelineNode {
	return p.nodes
}

func (p *Pipeline) AddNode(node *PipelineNode) {
	p.nodes[node.id] = node
}
