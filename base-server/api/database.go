// Package api provides public access to ae-base-server database functions
package api

import (
	"github.com/ae-base-server/internal/database"
	pkgDB "github.com/ae-base-server/pkg/database"
	"gorm.io/gorm"
)

// DatabaseConfig exports the database configuration type
type DatabaseConfig = pkgDB.Config

// ConnectDatabase connects to the database using the provided configuration
func ConnectDatabase(config DatabaseConfig) (*gorm.DB, error) {
	return pkgDB.Connect(config)
}

// ConnectDatabaseWithAutoCreate connects and creates the database if needed
func ConnectDatabaseWithAutoCreate(config DatabaseConfig) (*gorm.DB, error) {
	return pkgDB.ConnectWithAutoCreate(config)
}

// SeedBaseData loads and seeds data from seed-data.json
func SeedBaseData(db *gorm.DB) error {
	return database.Seed(db)
}

// SetupDatabase is deprecated - use bootstrap system for complete database setup
// This function only connects and seeds data. Migrations are handled by the bootstrap system.
func SetupDatabase(config DatabaseConfig) (*gorm.DB, error) {
	// Connect with auto-create
	db, err := pkgDB.ConnectWithAutoCreate(config)
	if err != nil {
		return nil, err
	}

	// Note: Migrations are now handled by the bootstrap system when modules are registered
	// This function only handles seeding for backward compatibility

	// Seed data
	if err := database.Seed(db); err != nil {
		return nil, err
	}

	return db, nil
}
