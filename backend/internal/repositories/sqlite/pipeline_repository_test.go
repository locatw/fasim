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
	item := models.NewItemFromParams(0, name, "Test Description for "+name)
	err := s.itemRepo.Create(s.T().Context(), item)
	s.NoError(err)
	return item
}

// createTestFacility creates and persists a test facility
func (s *PipelineRepositoryTestSuite) createTestFacility(name string, inputItems, outputItems []*models.Item) *models.Facility {
	facility := models.NewFacility(name, 100)

	// Add input requirements
	for i, item := range inputItems {
		req := models.NewInputRequirement(item, i+1)
		facility.AddInputRequirement(req)
	}

	// Add output definitions
	for i, item := range outputItems {
		def := models.NewOutputDefinition(item, i+2)
		facility.AddOutputDefinition(def)
	}

	err := s.facilityRepo.Create(s.T().Context(), facility)
	s.NoError(err)
	return facility
}

// createTestPipeline creates and persists a test pipeline with the given facilities
func (s *PipelineRepositoryTestSuite) createTestPipeline(name string, facilities []*models.Facility) *models.Pipeline {
	pipeline := models.NewPipeline(name)

	// Create nodes for each facility
	for i, facility := range facilities {
		node := models.NewPipelineNode(facility)
		// Connect nodes sequentially if there's a next facility
		if i < len(facilities)-1 {
			node.AddNextNodeID(i + 2) // Reference to next node's temporary ID
		}
		pipeline.AddNode(node)
	}

	err := s.repo.Create(s.T().Context(), pipeline)
	s.NoError(err)
	s.Greater(pipeline.ID(), 0)

	// Get the created pipeline to get actual node IDs
	created, err := s.repo.Get(s.T().Context(), pipeline.ID())
	s.NoError(err)
	s.NotNil(created)

	return created
}

func (s *PipelineRepositoryTestSuite) TestCreate() {
	testCases := []struct {
		name        string
		setup       func() (*models.Item, *models.Item, *models.Facility, *models.Facility)
		input       func(*models.Facility, *models.Facility) *models.Pipeline
		expectError bool
		errorMsg    string
	}{
		{
			name: "creates a pipeline with sequential nodes",
			setup: func() (*models.Item, *models.Item, *models.Facility, *models.Facility) {
				item1 := s.createTestItem("Item 1")
				item2 := s.createTestItem("Item 2")
				facility1 := s.createTestFacility("Facility 1", []*models.Item{item1}, []*models.Item{item2})
				facility2 := s.createTestFacility("Facility 2", []*models.Item{item2}, []*models.Item{item1})
				return item1, item2, facility1, facility2
			},
			input: func(facility1, facility2 *models.Facility) *models.Pipeline {
				pipeline := models.NewPipeline("Test Pipeline")
				node1 := models.NewPipelineNode(facility1)
				node1.AddNextNodeID(2)
				pipeline.AddNode(node1)
				node2 := models.NewPipelineNode(facility2)
				pipeline.AddNode(node2)
				return pipeline
			},
			expectError: false,
		},
		{
			name: "enforces unique name constraint",
			setup: func() (*models.Item, *models.Item, *models.Facility, *models.Facility) {
				item1 := s.createTestItem("Item 1")
				item2 := s.createTestItem("Item 2")
				facility1 := s.createTestFacility("Facility 1", []*models.Item{item1}, []*models.Item{item2})
				facility2 := s.createTestFacility("Facility 2", []*models.Item{item2}, []*models.Item{item1})

				existingPipeline := models.NewPipeline("Test Pipeline")
				node := models.NewPipelineNode(facility1)
				existingPipeline.AddNode(node)
				s.NoError(s.repo.Create(s.T().Context(), existingPipeline))

				return item1, item2, facility1, facility2
			},
			input: func(facility1, facility2 *models.Facility) *models.Pipeline {
				pipeline := models.NewPipeline("Test Pipeline")
				node1 := models.NewPipelineNode(facility1)
				node1.AddNextNodeID(2)
				pipeline.AddNode(node1)
				node2 := models.NewPipelineNode(facility2)
				pipeline.AddNode(node2)
				return pipeline
			},
			expectError: true,
			errorMsg:    "UNIQUE constraint failed",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			_, _, facility1, facility2 := tc.setup()
			pipeline := tc.input(facility1, facility2)
			err := s.repo.Create(s.T().Context(), pipeline)

			if tc.expectError {
				s.Error(err)
				s.Contains(err.Error(), tc.errorMsg)
			} else {
				s.NoError(err)
				s.Greater(pipeline.ID(), 0)

				result, err := s.repo.Get(s.T().Context(), pipeline.ID())
				s.NoError(err)
				s.NotNil(result)

				s.Equal(pipeline.ID(), result.ID())
				s.Equal(pipeline.Name(), result.Name())
				s.Equal(pipeline.Description(), result.Description())
				s.Len(result.Nodes(), len(pipeline.Nodes()))

				for _, node := range result.Nodes() {
					var originalNode *models.PipelineNode
					for _, n := range pipeline.Nodes() {
						if n.Facility().ID() == node.Facility().ID() {
							originalNode = n
							break
						}
					}
					s.Equal(originalNode.Facility().ID(), node.Facility().ID())

					if len(originalNode.NextNodeIDs()) > 0 {
						s.Len(node.NextNodeIDs(), len(originalNode.NextNodeIDs()))
						for _, targetID := range originalNode.NextNodeIDs() {
							targetFacilityID := pipeline.Nodes()[targetID].Facility().ID()
							found := false
							for _, resultTargetID := range node.NextNodeIDs() {
								if result.Nodes()[resultTargetID].Facility().ID() == targetFacilityID {
									found = true
									break
								}
							}
							s.True(found, "Connection to facility %d not found", targetFacilityID)
						}
					} else {
						s.Empty(node.NextNodeIDs())
					}
				}
			}
		})
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
				return pipeline.ID()
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
				s.Equal(setupPipeline.ID(), result.ID())
				s.Equal(setupPipeline.Name(), result.Name())
				s.Equal(setupPipeline.Description(), result.Description())
				s.Len(result.Nodes(), len(setupPipeline.Nodes()))

				// Verify nodes and their connections
				for _, node := range result.Nodes() {
					// Find original node by matching facility ID
					var originalNode *models.PipelineNode
					for _, n := range setupPipeline.Nodes() {
						if n.Facility().ID() == node.Facility().ID() {
							originalNode = n
							break
						}
					}
					s.Equal(originalNode.Facility().ID(), node.Facility().ID())

					// Verify connections by matching facility IDs
					if len(originalNode.NextNodeIDs()) > 0 {
						s.Len(node.NextNodeIDs(), len(originalNode.NextNodeIDs()))
						for _, targetID := range originalNode.NextNodeIDs() {
							targetFacilityID := setupPipeline.Nodes()[targetID].Facility().ID()
							found := false
							for _, resultTargetID := range node.NextNodeIDs() {
								if result.Nodes()[resultTargetID].Facility().ID() == targetFacilityID {
									found = true
									break
								}
							}
							s.True(found, "Connection to facility %d not found", targetFacilityID)
						}
					} else {
						s.Empty(node.NextNodeIDs())
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
		s.Equal(pipelines[i].ID(), result.ID())
		s.Equal(pipelines[i].Name(), result.Name())
		s.Equal(pipelines[i].Description(), result.Description())
		s.Len(result.Nodes(), len(pipelines[i].Nodes()))

		// Verify nodes and their connections
		for _, node := range result.Nodes() {
			// Find original node by matching facility ID
			var originalNode *models.PipelineNode
			for _, n := range pipelines[i].Nodes() {
				if n.Facility().ID() == node.Facility().ID() {
					originalNode = n
					break
				}
			}
			s.Equal(originalNode.Facility().ID(), node.Facility().ID())

			// Verify connections by matching facility IDs
			if len(originalNode.NextNodeIDs()) > 0 {
				s.Len(node.NextNodeIDs(), len(originalNode.NextNodeIDs()))
				for _, targetID := range originalNode.NextNodeIDs() {
					targetFacilityID := pipelines[i].Nodes()[targetID].Facility().ID()
					found := false
					for _, resultTargetID := range node.NextNodeIDs() {
						if result.Nodes()[resultTargetID].Facility().ID() == targetFacilityID {
							found = true
							break
						}
					}
					s.True(found, "Connection to facility %d not found", targetFacilityID)
				}
			} else {
				s.Empty(node.NextNodeIDs())
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

	// Create updated pipeline
	updatedPipeline := models.NewPipelineFromParams(
		pipeline.ID(),
		"Updated Pipeline",
		"Updated Description",
		make(map[int]*models.PipelineNode),
	)

	// Add first node with new facility
	node1 := models.NewPipelineNode(facilities[1])
	node1.AddNextNodeID(2) // Reference to second node's temporary ID
	updatedPipeline.AddNode(node1)

	// Add second node with new facility
	node2 := models.NewPipelineNode(facilities[2])
	updatedPipeline.AddNode(node2)

	err := s.repo.Update(s.T().Context(), updatedPipeline)
	s.NoError(err)

	// Verify the update
	updated, err := s.repo.Get(s.T().Context(), pipeline.ID())
	s.NoError(err)
	s.NotNil(updated)

	// Verify basic fields
	s.Equal(updatedPipeline.Name(), updated.Name())
	s.Equal(updatedPipeline.Description(), updated.Description())
	s.Len(updated.Nodes(), len(updatedPipeline.Nodes()))

	// Verify nodes and their connections
	for _, node := range updated.Nodes() {
		// Find original node by matching facility ID
		var originalNode *models.PipelineNode
		for _, n := range updatedPipeline.Nodes() {
			if n.Facility().ID() == node.Facility().ID() {
				originalNode = n
				break
			}
		}
		s.Equal(originalNode.Facility().ID(), node.Facility().ID())

		// Verify connections by matching facility IDs
		if len(originalNode.NextNodeIDs()) > 0 {
			s.Len(node.NextNodeIDs(), len(originalNode.NextNodeIDs()))
			for _, targetID := range originalNode.NextNodeIDs() {
				targetFacilityID := updatedPipeline.Nodes()[targetID].Facility().ID()
				found := false
				for _, resultTargetID := range node.NextNodeIDs() {
					if updated.Nodes()[resultTargetID].Facility().ID() == targetFacilityID {
						found = true
						break
					}
				}
				s.True(found, "Connection to facility %d not found", targetFacilityID)
			}
		} else {
			s.Empty(node.NextNodeIDs())
		}
	}

	// Test updating non-existent pipeline
	nonExistentPipeline := models.NewPipelineFromParams(
		999,
		"Non-existent",
		"Non-existent",
		make(map[int]*models.PipelineNode),
	)
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
				return pipeline.ID()
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
