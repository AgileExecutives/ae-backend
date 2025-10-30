// Package api provides public access to ae-base-server handlers
// This allows external modules to use the base authentication and user management
package api

import (
	"github.com/ae-base-server/internal/config"
	"github.com/ae-base-server/internal/handlers"
	"gorm.io/gorm"
)

// NewAuthHandler creates a new auth handler instance
// Returns the internal handler which has all auth methods (Login, Register, Logout, etc.)
func NewAuthHandler(db *gorm.DB) *handlers.AuthHandler {
	return handlers.NewAuthHandler(db)
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
