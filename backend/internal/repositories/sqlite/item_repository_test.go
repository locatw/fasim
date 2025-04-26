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
	s.NoError(s.db.Exec("DELETE FROM items").Error)
}

// createTestItem creates and persists a test item with the given name and description.
// Returns the created item with its auto-generated ID.
func (s *ItemRepositoryTestSuite) createTestItem(name, description string) *models.Item {
	item := models.NewItemFromParams(0, name, description)
	err := s.repo.Create(s.T().Context(), item)
	s.NoError(err)
	s.Greater(item.ID(), 0)
	return item
}

func (s *ItemRepositoryTestSuite) TestCreate() {
	item := models.NewItemFromParams(0, "Test Item", "Test Description")

	err := s.repo.Create(s.T().Context(), item)
	s.NoError(err)
	s.Greater(item.ID(), 0)

	var entity entities.ItemEntity
	err = s.db.First(&entity, item.ID()).Error
	s.NoError(err)
	s.Equal(item.ID(), int(entity.ID))
	s.Equal(item.Name(), entity.Name)
	s.Equal(item.Description(), entity.Description)
}

func (s *ItemRepositoryTestSuite) TestGet() {
	testCases := []struct {
		name         string
		setupFunc    func() *models.Item
		getID        func(*models.Item) int
		expectErr    error
		expectExists bool
	}{
		{
			name: "successfully retrieves an item when using valid ID",
			setupFunc: func() *models.Item {
				return s.createTestItem("Test Item", "Test Description")
			},
			getID: func(item *models.Item) int {
				return item.ID()
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
				s.Equal(setupItem.ID(), result.ID())
				s.Equal(setupItem.Name(), result.Name())
				s.Equal(setupItem.Description(), result.Description())
			} else {
				s.Nil(result)
			}
		})
	}
}

func (s *ItemRepositoryTestSuite) TestList() {
	items := []*models.Item{
		s.createTestItem("Item 1", "Description 1"),
		s.createTestItem("Item 2", "Description 2"),
	}

	results, err := s.repo.List(s.T().Context())
	s.NoError(err)
	s.Len(results, len(items))

	for i, result := range results {
		s.Equal(items[i].ID(), result.ID())
		s.Equal(items[i].Name(), result.Name())
		s.Equal(items[i].Description(), result.Description())
	}
}

func (s *ItemRepositoryTestSuite) TestUpdate() {
	item := s.createTestItem("Original Name", "Original Description")

	updatedItem := models.NewItemFromParams(item.ID(), "Updated Name", "Updated Description")
	err := s.repo.Update(s.T().Context(), updatedItem)
	s.NoError(err)

	var updatedEntity entities.ItemEntity
	err = s.db.First(&updatedEntity, item.ID()).Error
	s.NoError(err)
	s.Equal(updatedItem.Name(), updatedEntity.Name)
	s.Equal(updatedItem.Description(), updatedEntity.Description)

	nonExistentItem := models.NewItemFromParams(999, "Non-existent", "Non-existent")
	err = s.repo.Update(s.T().Context(), nonExistentItem)
	s.Equal(gorm.ErrRecordNotFound, err)
}

func (s *ItemRepositoryTestSuite) TestDelete() {
	testCases := []struct {
		name      string
		setupFunc func() *models.Item
		getID     func(*models.Item) int
		expectErr error
	}{
		{
			name: "successfully deletes an item when using valid ID",
			setupFunc: func() *models.Item {
				return s.createTestItem("Test Item", "Test Description")
			},
			getID: func(item *models.Item) int {
				return item.ID()
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
			var setupItem *models.Item
			if tc.setupFunc != nil {
				setupItem = tc.setupFunc()
			}

			inputID := tc.getID(setupItem)
			err := s.repo.Delete(s.T().Context(), inputID)

			if tc.expectErr != nil {
				s.Equal(tc.expectErr, err)
			} else {
				s.NoError(err)

				var count int64
				s.NoError(s.db.Model(&entities.ItemEntity{}).Where("id = ?", inputID).Count(&count).Error)
				s.Equal(int64(0), count)
			}
		})
	}
}
