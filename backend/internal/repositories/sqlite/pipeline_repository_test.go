package sqlite

import (
	"testing"

	"github.com/fasim/backend/internal/models"
	"github.com/fasim/backend/internal/repositories/entities"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type PipelineRepositoryTestSuite struct {
	BaseSQLiteTestSuite
	repo         *PipelineRepository
	facilityRepo *FacilityRepository
	itemRepo     *ItemRepository
}

func TestPipelineRepositorySuite(t *testing.T) {
	suite.Run(t, new(PipelineRepositoryTestSuite))
}

func (s *PipelineRepositoryTestSuite) SetupSuite() {
	s.SetupDockerAndDB(
		&entities.ItemEntity{},
		&entities.FacilityEntity{},
		&entities.InputRequirementEntity{},
		&entities.OutputDefinitionEntity{},
		&entities.PipelineEntity{},
		&entities.PipelineNodeEntity{},
		&entities.PipelineNodeConnectionEntity{},
	)
	s.repo = &PipelineRepository{db: s.db}
	s.facilityRepo = &FacilityRepository{db: s.db}
	s.itemRepo = &ItemRepository{db: s.db}
}

func (s *PipelineRepositoryTestSuite) TearDownSuite() {
	s.TearDownDocker()
}

func (s *PipelineRepositoryTestSuite) SetupTest() {
	s.NoError(s.db.Exec("DELETE FROM pipeline_node_connections").Error)
	s.NoError(s.db.Exec("DELETE FROM pipeline_nodes").Error)
	s.NoError(s.db.Exec("DELETE FROM pipelines").Error)
	s.NoError(s.db.Exec("DELETE FROM facilities").Error)
	s.NoError(s.db.Exec("DELETE FROM input_requirements").Error)
	s.NoError(s.db.Exec("DELETE FROM output_definitions").Error)
	s.NoError(s.db.Exec("DELETE FROM items").Error)
}

// createTestItem creates and persists a test item
func (s *PipelineRepositoryTestSuite) createTestItem(name string) *models.Item {
	item := &models.Item{
		Name:        name,
		Description: "Test Description for " + name,
	}
	err := s.itemRepo.Create(s.T().Context(), item)
	s.NoError(err)
	return item
}

// createTestFacility creates and persists a test facility
func (s *PipelineRepositoryTestSuite) createTestFacility(name string, inputItems, outputItems []*models.Item) *models.Facility {
	facility := &models.Facility{
		Name:           name,
		Description:    "Test Description for " + name,
		ProcessingTime: 100,
	}

	facility.InputRequirements = make([]models.InputRequirement, len(inputItems))
	for i, item := range inputItems {
		facility.InputRequirements[i] = models.InputRequirement{
			Item:     *item,
			Quantity: i + 1,
		}
	}

	facility.OutputDefinitions = make([]models.OutputDefinition, len(outputItems))
	for i, item := range outputItems {
		facility.OutputDefinitions[i] = models.OutputDefinition{
			Item:     *item,
			Quantity: i + 2,
		}
	}

	err := s.facilityRepo.Create(s.T().Context(), facility)
	s.NoError(err)
	return facility
}

// createTestPipeline creates and persists a test pipeline with the given facilities
func (s *PipelineRepositoryTestSuite) createTestPipeline(name string, facilities []*models.Facility) *models.Pipeline {
	pipeline := &models.Pipeline{
		Name:        name,
		Description: "Test Description for " + name,
		Nodes:       make(map[int]models.PipelineNode),
	}

	// Create nodes for each facility
	for i, facility := range facilities {
		node := models.PipelineNode{
			ID:          i + 1, // Temporary ID for test setup
			Facility:    *facility,
			NextNodeIDs: []int{},
		}
		// Connect nodes sequentially if there's a next facility
		if i < len(facilities)-1 {
			node.NextNodeIDs = []int{i + 2} // Reference to next node's temporary ID
		}
		pipeline.Nodes[node.ID] = node
	}

	err := s.repo.Create(s.T().Context(), pipeline)
	s.NoError(err)
	s.Greater(pipeline.ID, 0)

	// Get the created pipeline to get actual node IDs
	created, err := s.repo.Get(s.T().Context(), pipeline.ID)
	s.NoError(err)
	s.NotNil(created)

	return created
}

func (s *PipelineRepositoryTestSuite) TestCreate() {
	// Create test items and facilities
	item1 := s.createTestItem("Item 1")
	item2 := s.createTestItem("Item 2")
	facility1 := s.createTestFacility("Facility 1", []*models.Item{item1}, []*models.Item{item2})
	facility2 := s.createTestFacility("Facility 2", []*models.Item{item2}, []*models.Item{item1})

	// Create pipeline with sequential nodes
	pipeline := &models.Pipeline{
		Name:        "Test Pipeline",
		Description: "Test Description",
		Nodes:       make(map[int]models.PipelineNode),
	}

	// Add first node
	node1 := models.PipelineNode{
		ID:          1, // Temporary ID
		Facility:    *facility1,
		NextNodeIDs: []int{2}, // Reference to second node's temporary ID
	}
	pipeline.Nodes[node1.ID] = node1

	// Add second node
	node2 := models.PipelineNode{
		ID:          2, // Temporary ID
		Facility:    *facility2,
		NextNodeIDs: []int{},
	}
	pipeline.Nodes[node2.ID] = node2

	err := s.repo.Create(s.T().Context(), pipeline)
	s.NoError(err)
	s.Greater(pipeline.ID, 0)

	// Verify the persistence by getting the complete pipeline
	result, err := s.repo.Get(s.T().Context(), pipeline.ID)
	s.NoError(err)
	s.NotNil(result)

	// Verify pipeline data
	s.Equal(pipeline.ID, result.ID)
	s.Equal(pipeline.Name, result.Name)
	s.Equal(pipeline.Description, result.Description)
	s.Len(result.Nodes, len(pipeline.Nodes))

	// Verify nodes and their connections
	for _, node := range result.Nodes {
		// Find original node by matching facility ID
		var originalNode models.PipelineNode
		for _, n := range pipeline.Nodes {
			if n.Facility.ID == node.Facility.ID {
				originalNode = n
				break
			}
		}
		s.Equal(originalNode.Facility.ID, node.Facility.ID)

		// Verify connections by matching facility IDs
		if len(originalNode.NextNodeIDs) > 0 {
			s.Len(node.NextNodeIDs, len(originalNode.NextNodeIDs))
			for _, targetID := range originalNode.NextNodeIDs {
				targetFacilityID := pipeline.Nodes[targetID].Facility.ID
				found := false
				for _, resultTargetID := range node.NextNodeIDs {
					if result.Nodes[resultTargetID].Facility.ID == targetFacilityID {
						found = true
						break
					}
				}
				s.True(found, "Connection to facility %d not found", targetFacilityID)
			}
		} else {
			s.Empty(node.NextNodeIDs)
		}
	}
}

func (s *PipelineRepositoryTestSuite) TestGet() {
	// Create test items and facilities
	item := s.createTestItem("Test Item")
	facilities := []*models.Facility{
		s.createTestFacility("Facility 1", []*models.Item{item}, []*models.Item{item}),
		s.createTestFacility("Facility 2", []*models.Item{item}, []*models.Item{item}),
	}

	testCases := []struct {
		name         string
		setupFunc    func() *models.Pipeline
		getID        func(*models.Pipeline) int
		expectErr    error
		expectExists bool
	}{
		{
			name: "successfully retrieves a pipeline when using valid ID",
			setupFunc: func() *models.Pipeline {
				return s.createTestPipeline("Test Pipeline", facilities)
			},
			getID: func(pipeline *models.Pipeline) int {
				return pipeline.ID
			},
			expectErr:    nil,
			expectExists: true,
		},
		{
			name:         "returns nil when ID does not exist",
			setupFunc:    nil,
			getID:        func(pipeline *models.Pipeline) int { return 999 },
			expectErr:    nil,
			expectExists: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var setupPipeline *models.Pipeline
			if tc.setupFunc != nil {
				setupPipeline = tc.setupFunc()
			}

			inputID := tc.getID(setupPipeline)
			result, err := s.repo.Get(s.T().Context(), inputID)

			if tc.expectErr != nil {
				s.Equal(tc.expectErr, err)
			} else {
				s.NoError(err)
			}

			if tc.expectExists {
				s.NotNil(result)
				s.Equal(setupPipeline.ID, result.ID)
				s.Equal(setupPipeline.Name, result.Name)
				s.Equal(setupPipeline.Description, result.Description)
				s.Len(result.Nodes, len(setupPipeline.Nodes))

				// Verify nodes and their connections
				for _, node := range result.Nodes {
					// Find original node by matching facility ID
					var originalNode models.PipelineNode
					for _, n := range setupPipeline.Nodes {
						if n.Facility.ID == node.Facility.ID {
							originalNode = n
							break
						}
					}
					s.Equal(originalNode.Facility.ID, node.Facility.ID)

					// Verify connections by matching facility IDs
					if len(originalNode.NextNodeIDs) > 0 {
						s.Len(node.NextNodeIDs, len(originalNode.NextNodeIDs))
						for _, targetID := range originalNode.NextNodeIDs {
							targetFacilityID := setupPipeline.Nodes[targetID].Facility.ID
							found := false
							for _, resultTargetID := range node.NextNodeIDs {
								if result.Nodes[resultTargetID].Facility.ID == targetFacilityID {
									found = true
									break
								}
							}
							s.True(found, "Connection to facility %d not found", targetFacilityID)
						}
					} else {
						s.Empty(node.NextNodeIDs)
					}
				}
			} else {
				s.Nil(result)
			}
		})
	}
}

func (s *PipelineRepositoryTestSuite) TestList() {
	// Create test items and facilities
	item := s.createTestItem("Test Item")
	facilities := []*models.Facility{
		s.createTestFacility("Facility 1", []*models.Item{item}, []*models.Item{item}),
		s.createTestFacility("Facility 2", []*models.Item{item}, []*models.Item{item}),
	}

	// Create multiple pipelines
	pipelines := []*models.Pipeline{
		s.createTestPipeline("Pipeline 1", facilities),
		s.createTestPipeline("Pipeline 2", facilities),
	}

	results, err := s.repo.List(s.T().Context())
	s.NoError(err)
	s.Len(results, len(pipelines))

	// Verify each pipeline
	for i, result := range results {
		s.Equal(pipelines[i].ID, result.ID)
		s.Equal(pipelines[i].Name, result.Name)
		s.Equal(pipelines[i].Description, result.Description)
		s.Len(result.Nodes, len(pipelines[i].Nodes))

		// Verify nodes and their connections
		for _, node := range result.Nodes {
			// Find original node by matching facility ID
			var originalNode models.PipelineNode
			for _, n := range pipelines[i].Nodes {
				if n.Facility.ID == node.Facility.ID {
					originalNode = n
					break
				}
			}
			s.Equal(originalNode.Facility.ID, node.Facility.ID)

			// Verify connections by matching facility IDs
			if len(originalNode.NextNodeIDs) > 0 {
				s.Len(node.NextNodeIDs, len(originalNode.NextNodeIDs))
				for _, targetID := range originalNode.NextNodeIDs {
					targetFacilityID := pipelines[i].Nodes[targetID].Facility.ID
					found := false
					for _, resultTargetID := range node.NextNodeIDs {
						if result.Nodes[resultTargetID].Facility.ID == targetFacilityID {
							found = true
							break
						}
					}
					s.True(found, "Connection to facility %d not found", targetFacilityID)
				}
			} else {
				s.Empty(node.NextNodeIDs)
			}
		}
	}
}

func (s *PipelineRepositoryTestSuite) TestUpdate() {
	// Create test items and facilities
	item := s.createTestItem("Test Item")
	facilities := []*models.Facility{
		s.createTestFacility("Facility 1", []*models.Item{item}, []*models.Item{item}),
		s.createTestFacility("Facility 2", []*models.Item{item}, []*models.Item{item}),
		s.createTestFacility("Facility 3", []*models.Item{item}, []*models.Item{item}),
	}

	// Create initial pipeline
	pipeline := s.createTestPipeline("Original Pipeline", facilities[:2])

	// Modify pipeline with new node configuration
	pipeline.Name = "Updated Pipeline"
	pipeline.Description = "Updated Description"
	pipeline.Nodes = make(map[int]models.PipelineNode)

	// Add first node with new facility
	node1 := models.PipelineNode{
		ID:          1, // Temporary ID
		Facility:    *facilities[1],
		NextNodeIDs: []int{2}, // Reference to second node's temporary ID
	}
	pipeline.Nodes[node1.ID] = node1

	// Add second node with new facility
	node2 := models.PipelineNode{
		ID:          2, // Temporary ID
		Facility:    *facilities[2],
		NextNodeIDs: []int{},
	}
	pipeline.Nodes[node2.ID] = node2

	err := s.repo.Update(s.T().Context(), pipeline)
	s.NoError(err)

	// Verify the update
	updated, err := s.repo.Get(s.T().Context(), pipeline.ID)
	s.NoError(err)
	s.NotNil(updated)

	// Verify basic fields
	s.Equal(pipeline.Name, updated.Name)
	s.Equal(pipeline.Description, updated.Description)
	s.Len(updated.Nodes, len(pipeline.Nodes))

	// Verify nodes and their connections
	for _, node := range updated.Nodes {
		// Find original node by matching facility ID
		var originalNode models.PipelineNode
		for _, n := range pipeline.Nodes {
			if n.Facility.ID == node.Facility.ID {
				originalNode = n
				break
			}
		}
		s.Equal(originalNode.Facility.ID, node.Facility.ID)

		// Verify connections by matching facility IDs
		if len(originalNode.NextNodeIDs) > 0 {
			s.Len(node.NextNodeIDs, len(originalNode.NextNodeIDs))
			for _, targetID := range originalNode.NextNodeIDs {
				targetFacilityID := pipeline.Nodes[targetID].Facility.ID
				found := false
				for _, resultTargetID := range node.NextNodeIDs {
					if updated.Nodes[resultTargetID].Facility.ID == targetFacilityID {
						found = true
						break
					}
				}
				s.True(found, "Connection to facility %d not found", targetFacilityID)
			}
		} else {
			s.Empty(node.NextNodeIDs)
		}
	}

	// Test updating non-existent pipeline
	nonExistentPipeline := &models.Pipeline{
		ID:          999,
		Name:        "Non-existent",
		Description: "Non-existent",
	}
	err = s.repo.Update(s.T().Context(), nonExistentPipeline)
	s.Equal(gorm.ErrRecordNotFound, err)
}

func (s *PipelineRepositoryTestSuite) TestDelete() {
	// Create test items and facilities
	item := s.createTestItem("Test Item")
	facilities := []*models.Facility{
		s.createTestFacility("Facility 1", []*models.Item{item}, []*models.Item{item}),
		s.createTestFacility("Facility 2", []*models.Item{item}, []*models.Item{item}),
	}

	testCases := []struct {
		name      string
		setupFunc func() *models.Pipeline
		getID     func(*models.Pipeline) int
		expectErr error
	}{
		{
			name: "successfully deletes a pipeline when using valid ID",
			setupFunc: func() *models.Pipeline {
				return s.createTestPipeline("Test Pipeline", facilities)
			},
			getID: func(pipeline *models.Pipeline) int {
				return pipeline.ID
			},
			expectErr: nil,
		},
		{
			name:      "returns error when ID does not exist",
			setupFunc: nil,
			getID:     func(pipeline *models.Pipeline) int { return 999 },
			expectErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var setupPipeline *models.Pipeline
			if tc.setupFunc != nil {
				setupPipeline = tc.setupFunc()
			}

			inputID := tc.getID(setupPipeline)
			err := s.repo.Delete(s.T().Context(), inputID)

			if tc.expectErr != nil {
				s.Equal(tc.expectErr, err)
			} else {
				s.NoError(err)

				// Verify the pipeline was deleted
				var count int64
				s.NoError(s.db.Model(&entities.PipelineEntity{}).Where("id = ?", inputID).Count(&count).Error)
				s.Equal(int64(0), count)

				// Verify nodes and connections were deleted
				var nodeCount, connectionCount int64
				s.NoError(s.db.Model(&entities.PipelineNodeEntity{}).Where("pipeline_id = ?", inputID).Count(&nodeCount).Error)
				s.NoError(s.db.Model(&entities.PipelineNodeConnectionEntity{}).Where("source_node_id IN (SELECT id FROM pipeline_nodes WHERE pipeline_id = ?)", inputID).Count(&connectionCount).Error)
				s.Equal(int64(0), nodeCount)
				s.Equal(int64(0), connectionCount)
			}
		})
	}
}
