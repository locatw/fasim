package sqlite

import (
	"testing"

	"github.com/fasim/backend/internal/models"
	"github.com/fasim/backend/internal/repositories/entities"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type FacilityRepositoryTestSuite struct {
	BaseSQLiteTestSuite
	repo     *FacilityRepository
	itemRepo *ItemRepository
}

func TestFacilityRepositorySuite(t *testing.T) {
	suite.Run(t, new(FacilityRepositoryTestSuite))
}

func (s *FacilityRepositoryTestSuite) SetupSuite() {
	s.SetupDockerAndDB(
		&entities.ItemEntity{},
		&entities.FacilityEntity{},
		&entities.InputRequirementEntity{},
		&entities.OutputDefinitionEntity{},
	)
	s.repo = &FacilityRepository{db: s.db}
	s.itemRepo = &ItemRepository{db: s.db}
}

func (s *FacilityRepositoryTestSuite) TearDownSuite() {
	s.TearDownDocker()
}

func (s *FacilityRepositoryTestSuite) SetupTest() {
	s.NoError(s.db.Exec("DELETE FROM facilities").Error)
	s.NoError(s.db.Exec("DELETE FROM input_requirements").Error)
	s.NoError(s.db.Exec("DELETE FROM output_definitions").Error)
	s.NoError(s.db.Exec("DELETE FROM items").Error)
}

// createTestItem creates and persists a test item
func (s *FacilityRepositoryTestSuite) createTestItem(name string) *models.Item {
	item := models.NewItemFromParams(0, name, "Test Description for "+name)
	err := s.itemRepo.Create(s.T().Context(), item)
	s.NoError(err)
	return item
}

// createTestFacility creates and persists a test facility with the given name and optional items
func (s *FacilityRepositoryTestSuite) createTestFacility(name string, inputItems, outputItems []*models.Item) *models.Facility {
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

	err := s.repo.Create(s.T().Context(), facility)
	s.NoError(err)
	s.Greater(facility.ID(), 0)
	return facility
}

func (s *FacilityRepositoryTestSuite) TestCreate() {
	// Create test items
	inputItem := s.createTestItem("Input Item")
	outputItem := s.createTestItem("Output Item")

	// Create facility with relationships
	facility := models.NewFacility("Test Facility", 100)
	facility.AddInputRequirement(models.NewInputRequirement(inputItem, 2))
	facility.AddOutputDefinition(models.NewOutputDefinition(outputItem, 1))

	err := s.repo.Create(s.T().Context(), facility)
	s.NoError(err)
	s.Greater(facility.ID(), 0)

	// Verify the persistence
	var entity entities.FacilityEntity
	err = s.db.
		Preload("InputRequirements.Item").
		Preload("OutputDefinitions.Item").
		First(&entity, facility.ID()).Error
	s.NoError(err)

	// Verify facility data
	s.Equal(facility.ID(), int(entity.ID))
	s.Equal(facility.Name(), entity.Name)
	s.Equal(facility.Description(), entity.Description)
	s.Equal(facility.ProcessingTime(), entity.ProcessingTime)

	// Verify relationships
	s.Len(entity.InputRequirements, 1)
	s.Equal(inputItem.ID(), int(entity.InputRequirements[0].ItemID))
	s.Equal(2, entity.InputRequirements[0].Quantity)

	s.Len(entity.OutputDefinitions, 1)
	s.Equal(outputItem.ID(), int(entity.OutputDefinitions[0].ItemID))
	s.Equal(1, entity.OutputDefinitions[0].Quantity)
}

func (s *FacilityRepositoryTestSuite) TestGet() {
	// Create test items
	inputItem := s.createTestItem("Input Item")
	outputItem := s.createTestItem("Output Item")

	// Test cases
	testCases := []struct {
		name         string
		setupFunc    func() *models.Facility
		getID        func(*models.Facility) int
		expectErr    error
		expectExists bool
	}{
		{
			name: "successfully retrieves a facility when using valid ID",
			setupFunc: func() *models.Facility {
				return s.createTestFacility("Test Facility", []*models.Item{inputItem}, []*models.Item{outputItem})
			},
			getID: func(facility *models.Facility) int {
				return facility.ID()
			},
			expectErr:    nil,
			expectExists: true,
		},
		{
			name:         "returns nil when ID does not exist",
			setupFunc:    nil,
			getID:        func(facility *models.Facility) int { return 999 },
			expectErr:    nil,
			expectExists: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var setupFacility *models.Facility
			if tc.setupFunc != nil {
				setupFacility = tc.setupFunc()
			}

			inputID := tc.getID(setupFacility)
			result, err := s.repo.Get(s.T().Context(), inputID)

			if tc.expectErr != nil {
				s.Equal(tc.expectErr, err)
			} else {
				s.NoError(err)
			}

			if tc.expectExists {
				s.NotNil(result)
				s.Equal(setupFacility.ID(), result.ID())
				s.Equal(setupFacility.Name(), result.Name())
				s.Equal(setupFacility.Description(), result.Description())
				s.Equal(setupFacility.ProcessingTime(), result.ProcessingTime())

				// Verify relationships
				s.Len(result.InputRequirements(), 1)
				s.Equal(inputItem.ID(), result.InputRequirements()[0].Item().ID())
				s.Equal(1, result.InputRequirements()[0].Quantity())

				s.Len(result.OutputDefinitions(), 1)
				s.Equal(outputItem.ID(), result.OutputDefinitions()[0].Item().ID())
				s.Equal(2, result.OutputDefinitions()[0].Quantity())
			} else {
				s.Nil(result)
			}
		})
	}
}

func (s *FacilityRepositoryTestSuite) TestList() {
	// Create test items
	inputItem := s.createTestItem("Input Item")
	outputItem := s.createTestItem("Output Item")

	// Create multiple facilities
	facilities := []*models.Facility{
		s.createTestFacility("Facility 1", []*models.Item{inputItem}, []*models.Item{outputItem}),
		s.createTestFacility("Facility 2", []*models.Item{inputItem}, []*models.Item{outputItem}),
	}

	results, err := s.repo.List(s.T().Context())
	s.NoError(err)
	s.Len(results, len(facilities))

	// Verify each facility
	for i, result := range results {
		s.Equal(facilities[i].ID(), result.ID())
		s.Equal(facilities[i].Name(), result.Name())
		s.Equal(facilities[i].Description(), result.Description())
		s.Equal(facilities[i].ProcessingTime(), result.ProcessingTime())

		// Verify relationships
		s.Len(result.InputRequirements(), 1)
		s.Equal(inputItem.ID(), result.InputRequirements()[0].Item().ID())

		s.Len(result.OutputDefinitions(), 1)
		s.Equal(outputItem.ID(), result.OutputDefinitions()[0].Item().ID())
	}
}

func (s *FacilityRepositoryTestSuite) TestUpdate() {
	// Create test items
	inputItem1 := s.createTestItem("Input Item 1")
	inputItem2 := s.createTestItem("Input Item 2")
	outputItem1 := s.createTestItem("Output Item 1")
	outputItem2 := s.createTestItem("Output Item 2")

	// Create initial facility
	facility := s.createTestFacility("Original Facility", []*models.Item{inputItem1}, []*models.Item{outputItem1})

	// Create updated facility
	updatedFacility := models.NewFacilityFromParams(
		facility.ID(),
		"Updated Facility",
		"Updated Description",
		[]*models.InputRequirement{models.NewInputRequirement(inputItem2, 3)},
		[]*models.OutputDefinition{models.NewOutputDefinition(outputItem2, 4)},
		200,
	)

	err := s.repo.Update(s.T().Context(), updatedFacility)
	s.NoError(err)

	// Verify the update
	updated, err := s.repo.Get(s.T().Context(), facility.ID())
	s.NoError(err)
	s.NotNil(updated)

	// Verify basic fields
	s.Equal(updatedFacility.Name(), updated.Name())
	s.Equal(updatedFacility.Description(), updated.Description())
	s.Equal(updatedFacility.ProcessingTime(), updated.ProcessingTime())

	// Verify relationships
	s.Len(updated.InputRequirements(), 1)
	s.Equal(inputItem2.ID(), updated.InputRequirements()[0].Item().ID())
	s.Equal(3, updated.InputRequirements()[0].Quantity())

	s.Len(updated.OutputDefinitions(), 1)
	s.Equal(outputItem2.ID(), updated.OutputDefinitions()[0].Item().ID())
	s.Equal(4, updated.OutputDefinitions()[0].Quantity())

	// Test updating non-existent facility
	nonExistentFacility := models.NewFacilityFromParams(
		999,
		"Non-existent",
		"Non-existent",
		[]*models.InputRequirement{},
		[]*models.OutputDefinition{},
		100,
	)
	err = s.repo.Update(s.T().Context(), nonExistentFacility)
	s.Equal(gorm.ErrRecordNotFound, err)
}

func (s *FacilityRepositoryTestSuite) TestDelete() {
	// Create test items
	inputItem := s.createTestItem("Input Item")
	outputItem := s.createTestItem("Output Item")

	testCases := []struct {
		name      string
		setupFunc func() *models.Facility
		getID     func(*models.Facility) int
		expectErr error
	}{
		{
			name: "successfully deletes a facility when using valid ID",
			setupFunc: func() *models.Facility {
				return s.createTestFacility("Test Facility", []*models.Item{inputItem}, []*models.Item{outputItem})
			},
			getID: func(facility *models.Facility) int {
				return facility.ID()
			},
			expectErr: nil,
		},
		{
			name:      "returns error when ID does not exist",
			setupFunc: nil,
			getID:     func(facility *models.Facility) int { return 999 },
			expectErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var setupFacility *models.Facility
			if tc.setupFunc != nil {
				setupFacility = tc.setupFunc()
			}

			inputID := tc.getID(setupFacility)
			err := s.repo.Delete(s.T().Context(), inputID)

			if tc.expectErr != nil {
				s.Equal(tc.expectErr, err)
			} else {
				s.NoError(err)

				// Verify the facility was deleted
				var count int64
				s.NoError(s.db.Model(&entities.FacilityEntity{}).Where("id = ?", inputID).Count(&count).Error)
				s.Equal(int64(0), count)

				// Verify relationships were deleted
				var inputCount, outputCount int64
				s.NoError(s.db.Model(&entities.InputRequirementEntity{}).Where("facility_id = ?", inputID).Count(&inputCount).Error)
				s.NoError(s.db.Model(&entities.OutputDefinitionEntity{}).Where("facility_id = ?", inputID).Count(&outputCount).Error)
				s.Equal(int64(0), inputCount)
				s.Equal(int64(0), outputCount)
			}
		})
	}
}
