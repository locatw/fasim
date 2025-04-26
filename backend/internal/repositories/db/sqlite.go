package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB represents a database connection
type DB struct {
	*gorm.DB
}

// New creates a new database connection
func New(dataSourceName string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

// RunMigrations runs all database migrations
func (db *DB) RunMigrations(models ...interface{}) error {
	return db.AutoMigrate(models...)
}
