package sqlite

import (
	"testing"

	"github.com/fasim/backend/internal/models"
	"github.com/fasim/backend/internal/repositories/entities"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ItemRepositoryTestSuite struct {
	BaseSQLiteTestSuite
	repo *ItemRepository
}

func TestItemRepositorySuite(t *testing.T) {
	suite.Run(t, new(ItemRepositoryTestSuite))
}

func (s *ItemRepositoryTestSuite) SetupSuite() {
	s.SetupDockerAndDB(&entities.ItemEntity{})
	s.repo = &ItemRepository{db: s.db}
}

func (s *ItemRepositoryTestSuite) TearDownSuite() {
	s.TearDownDocker()
}

func (s *ItemRepositoryTestSuite) SetupTest() {
	// Ensures each test starts with a clean state by removing all existing items
	s.NoError(s.db.Exec("DELETE FROM items").Error)
}

// createTestItem creates and persists a test item with the given name and description.
// Returns the created item with its auto-generated ID.
func (s *ItemRepositoryTestSuite) createTestItem(name, description string) *models.Item {
	item := &models.Item{
		Name:        name,
		Description: description,
	}
	err := s.repo.Create(s.T().Context(), item)
	s.NoError(err)
	s.Greater(item.ID, 0)
	return item
}

func (s *ItemRepositoryTestSuite) TestCreate() {
	// Validates that Create method properly persists an item and assigns an ID
	item := &models.Item{
		Name:        "Test Item",
		Description: "Test Description",
	}

	err := s.repo.Create(s.T().Context(), item)
	s.NoError(err)
	s.Greater(item.ID, 0)

	// Verifies the persistence by retrieving the item directly from the database
	var entity entities.ItemEntity
	err = s.db.First(&entity, item.ID).Error
	s.NoError(err)
	s.Equal(item.ID, int(entity.ID))
	s.Equal(item.Name, entity.Name)
	s.Equal(item.Description, entity.Description)
}

func (s *ItemRepositoryTestSuite) TestGet() {
	// Test cases for validating Get method behavior with different input scenarios
	testCases := []struct {
		name         string // Test case identifier
		setupFunc    func() *models.Item
		getID        func(*models.Item) int // Dynamically determines the ID to test with
		expectErr    error
		expectExists bool
	}{
		{
			name: "successfully retrieves an item when using valid ID",
			setupFunc: func() *models.Item {
				return s.createTestItem("Test Item", "Test Description")
			},
			getID: func(item *models.Item) int {
				return item.ID
			},
			expectErr:    nil,
			expectExists: true,
		},
		{
			name:         "returns nil when ID does not exist",
			setupFunc:    nil,
			getID:        func(item *models.Item) int { return 999 },
			expectErr:    nil,
			expectExists: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data if required by the test case
			var setupItem *models.Item
			if tc.setupFunc != nil {
				setupItem = tc.setupFunc()
			}

			inputID := tc.getID(setupItem)

			result, err := s.repo.Get(s.T().Context(), inputID)

			if tc.expectErr != nil {
				s.Equal(tc.expectErr, err)
			} else {
				s.NoError(err)
			}

			if tc.expectExists {
				s.NotNil(result)
				s.Equal(setupItem.ID, result.ID)
				s.Equal(setupItem.Name, result.Name)
				s.Equal(setupItem.Description, result.Description)
			} else {
				s.Nil(result)
			}
		})
	}
}

func (s *ItemRepositoryTestSuite) TestList() {
	// Creates multiple items to validate List method's behavior with multiple records
	items := []*models.Item{
		s.createTestItem("Item 1", "Description 1"),
		s.createTestItem("Item 2", "Description 2"),
	}

	results, err := s.repo.List(s.T().Context())
	s.NoError(err)
	s.Len(results, len(items))

	// Validates that all items are retrieved with correct data
	for i, result := range results {
		s.Equal(items[i].ID, result.ID)
		s.Equal(items[i].Name, result.Name)
		s.Equal(items[i].Description, result.Description)
	}
}

func (s *ItemRepositoryTestSuite) TestUpdate() {
	// Creates an initial item to test the update functionality
	item := s.createTestItem("Original Name", "Original Description")

	// Modifies the item's properties to test update operation
	item.Name = "Updated Name"
	item.Description = "Updated Description"
	err := s.repo.Update(s.T().Context(), item)
	s.NoError(err)

	// Validates the persistence of updates by retrieving the item directly
	var updatedEntity entities.ItemEntity
	err = s.db.First(&updatedEntity, item.ID).Error
	s.NoError(err)
	s.Equal(item.Name, updatedEntity.Name)
	s.Equal(item.Description, updatedEntity.Description)

	// Validates error handling when updating non-existent items
	nonExistentItem := &models.Item{
		ID:          999,
		Name:        "Non-existent",
		Description: "Non-existent",
	}
	err = s.repo.Update(s.T().Context(), nonExistentItem)
	s.Equal(gorm.ErrRecordNotFound, err)
}

func (s *ItemRepositoryTestSuite) TestDelete() {
	// Test cases for validating Delete method behavior with different input scenarios
	testCases := []struct {
		name      string // Test case identifier
		setupFunc func() *models.Item
		getID     func(*models.Item) int // Dynamically determines the ID to test with
		expectErr error
	}{
		{
			name: "successfully deletes an item when using valid ID",
			setupFunc: func() *models.Item {
				return s.createTestItem("Test Item", "Test Description")
			},
			getID: func(item *models.Item) int {
				return item.ID
			},
			expectErr: nil,
		},
		{
			name:      "returns error when ID does not exist",
			setupFunc: nil,
			getID:     func(item *models.Item) int { return 999 },
			expectErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data if required by the test case
			var setupItem *models.Item
			if tc.setupFunc != nil {
				setupItem = tc.setupFunc()
			}

			// Execute deletion with the test ID
			inputID := tc.getID(setupItem)
			err := s.repo.Delete(s.T().Context(), inputID)

			// Validate error response
			if tc.expectErr != nil {
				s.Equal(tc.expectErr, err)
			} else {
				s.NoError(err)

				// Verify the item was actually deleted
				var count int64
				s.NoError(s.db.Model(&entities.ItemEntity{}).Where("id = ?", inputID).Count(&count).Error)
				s.Equal(int64(0), count)
			}
		})
	}
}
