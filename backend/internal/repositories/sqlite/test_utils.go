package sqlite

import (
	"github.com/fasim/backend/internal/repositories/db"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// BaseSQLiteTestSuite provides common setup for SQLite-based repository tests
type BaseSQLiteTestSuite struct {
	suite.Suite
	pool     *dockertest.Pool
	resource *dockertest.Resource
	db       *db.DB
}

// SetupDockerAndDB initializes docker and database for testing
func (s *BaseSQLiteTestSuite) SetupDockerAndDB(entities ...interface{}) {
	var err error
	s.pool, err = dockertest.NewPool("")
	require.NoError(s.T(), err, "Could not connect to docker")

	s.resource, err = s.pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "keinos/sqlite3",
		Tag:        "latest",
	})
	require.NoError(s.T(), err, "Could not start resource")

	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(s.T(), err, "Failed to open database")

	// Auto migrate all required schemas
	err = gormDB.AutoMigrate(entities...)
	require.NoError(s.T(), err, "Failed to migrate database")

	s.db = &db.DB{DB: gormDB}
}

// TearDownDocker cleans up docker resources
func (s *BaseSQLiteTestSuite) TearDownDocker() {
	if s.resource != nil {
		s.NoError(s.pool.Purge(s.resource), "Failed to purge resource")
	}
}
