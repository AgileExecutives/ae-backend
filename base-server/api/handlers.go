// Package api provides public access to ae-base-server handlers
// This allows external modules to use the base authentication and user management
package api

import (
	"log"

	"github.com/ae-base-server/internal/handlers"
	"github.com/ae-base-server/internal/services"
	"github.com/ae-base-server/modules/templates/services/storage"
	"github.com/ae-base-server/pkg/config"
	pkgServices "github.com/ae-base-server/pkg/services"
	"gorm.io/gorm"
)

// NewAuthHandler creates a new auth handler instance with tenant service
// Returns the internal handler which has all auth methods (Login, Register, Logout, etc.)
func NewAuthHandler(db *gorm.DB) *handlers.AuthHandler {
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
		log.Printf("⚠️ Warning: Failed to initialize MinIO storage for auth handler: %v", err)
		// For now, creating a nil service to avoid breaking existing functionality
		return handlers.NewAuthHandler(db, nil)
	}

	// Create services for tenant bucket management
	tenantBucketService := pkgServices.NewTenantBucketService(minioStorage)
	tenantService := services.NewTenantService(db, tenantBucketService)

	return handlers.NewAuthHandler(db, tenantService)
}

// NewHealthHandler creates a new health handler instance
// Returns the internal handler which has Health method
func NewHealthHandler(db *gorm.DB, cfg interface{}) *handlers.HealthHandler {
	// Accept config as interface{} to avoid circular imports
	// The calling code should pass config.Config
	return handlers.NewHealthHandler(db, cfg.(config.Config))
}

// NewHealthHandlerWithConfig creates a health handler with loaded config
func NewHealthHandlerWithConfig(db *gorm.DB) *handlers.HealthHandler {
	// Load config and create handler with proper configuration
	cfg := config.Load()
	return handlers.NewHealthHandler(db, cfg)
}
