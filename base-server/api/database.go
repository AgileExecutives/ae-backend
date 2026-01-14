// Package api provides public access to ae-base-server database functions
package api

import (
	"fmt"

	internalDB "github.com/ae-base-server/internal/database"
	internalServices "github.com/ae-base-server/internal/services"
	"github.com/ae-base-server/modules/base"
	"github.com/ae-base-server/modules/customer"
	"github.com/ae-base-server/modules/email"
	"github.com/ae-base-server/modules/pdf"
	"github.com/ae-base-server/modules/templates/services/storage"
	"github.com/ae-base-server/pkg/core"
	"github.com/ae-base-server/pkg/database"
	pkgServices "github.com/ae-base-server/pkg/services"
	"gorm.io/gorm"
)

// DatabaseConfig exports the database configuration type
type DatabaseConfig = database.Config

// ConnectDatabase connects to the database using the provided configuration
func ConnectDatabase(config DatabaseConfig) (*gorm.DB, error) {
	return database.Connect(config)
}

// ConnectDatabaseWithAutoCreate connects and creates the database if needed
func ConnectDatabaseWithAutoCreate(config DatabaseConfig) (*gorm.DB, error) {
	return database.ConnectWithAutoCreate(config)
}

// SeedBaseData loads and seeds data from seed-data.json with MinIO bucket creation
func SeedBaseData(db *gorm.DB) error {
	// Initialize MinIO storage for tenant buckets
	minioConfig := storage.MinIOConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin123",
		UseSSL:          false,
		Region:          "us-east-1",
	}
	minioStorage, err := storage.NewMinIOStorage(minioConfig)
	if err != nil {
		// Log error but don't fail entirely - some features may not work
		fmt.Printf("⚠️ Warning: Failed to initialize MinIO storage for seeding: %v\n", err)
	}

	// Create services for tenant bucket management
	tenantBucketService := pkgServices.NewTenantBucketService(minioStorage)
	tenantService := internalServices.NewTenantService(db, tenantBucketService)
	
	return internalDB.Seed(db, tenantService)
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
