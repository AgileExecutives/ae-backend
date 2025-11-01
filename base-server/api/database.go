// Package api provides public access to ae-base-server database functions
package api

import (
	"fmt"

	"github.com/ae-base-server/internal/database"
	"github.com/ae-base-server/modules/base"
	"github.com/ae-base-server/modules/customer"
	"github.com/ae-base-server/modules/email"
	"github.com/ae-base-server/modules/pdf"
	"github.com/ae-base-server/pkg/core"
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

// MigrateBaseEntities migrates all base-server entities (users, tenants, etc.)
// This function can be called by external applications that use base-server as a library
func MigrateBaseEntities(db *gorm.DB) error {
	// Create base-server modules that contain the entity definitions
	coreModules := []core.Module{
		base.NewBaseModule(),
		customer.NewCustomerModule(),
		email.NewEmailModule(),
		pdf.NewPDFModule(),
	}

	// Collect all entities from base-server modules
	allEntities := []interface{}{}

	// Extract entities from all modules
	for _, module := range coreModules {
		entities := module.Entities()
		for _, entity := range entities {
			allEntities = append(allEntities, entity.GetModel())
		}
	}

	// Run auto-migration for all base-server entities
	if len(allEntities) > 0 {
		if err := db.AutoMigrate(allEntities...); err != nil {
			return fmt.Errorf("failed to migrate base-server entities: %w", err)
		}
	}

	return nil
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
