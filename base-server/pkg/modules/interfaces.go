package modules

import (
	"github.com/ae-saas-basic/ae-saas-basic/pkg/eventbus"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module represents a pluggable module in the application
type Module interface {
	// GetName returns the unique name of the module
	GetName() string

	// GetVersion returns the version of the module
	GetVersion() string

	// GetModels returns the GORM models that need to be migrated
	GetModels() []interface{}

	// GetEventHandlers returns event handlers to register with the event bus
	GetEventHandlers() []eventbus.EventHandler

	// RegisterRoutes registers HTTP routes for this module
	RegisterRoutes(router *gin.RouterGroup, db *gorm.DB)

	// Initialize performs any initialization required by the module
	Initialize(db *gorm.DB) error

	// Shutdown performs cleanup when the application is shutting down
	Shutdown() error

	// GetSwaggerInfo returns Swagger documentation info for this module
	GetSwaggerInfo() SwaggerModuleInfo
}

// SwaggerModuleInfo contains Swagger documentation information for a module
type SwaggerModuleInfo struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Tags        []SwaggerTag           `json:"tags"`
	Paths       map[string]interface{} `json:"paths,omitempty"`
	Definitions map[string]interface{} `json:"definitions,omitempty"`
}

// SwaggerTag represents a Swagger tag for grouping endpoints
type SwaggerTag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ModuleInfo contains metadata about a module
type ModuleInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Enabled     bool   `json:"enabled"`
}
