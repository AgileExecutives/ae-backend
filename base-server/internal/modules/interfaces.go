package modules

import (
	"github.com/ae-base-server/internal/config"
	"github.com/ae-base-server/pkg/eventbus"
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
	RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, cfg config.Config)

	// Initialize performs any initialization required by the module
	Initialize(db *gorm.DB, cfg config.Config) error

	// Shutdown performs cleanup when the application is shutting down
	Shutdown() error
}

// ModuleInfo contains metadata about a module
type ModuleInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Enabled     bool   `json:"enabled"`
}

// ModuleRegistry manages all registered modules
type ModuleRegistry interface {
	// RegisterModule registers a new module
	RegisterModule(module Module) error

	// GetModule returns a module by name
	GetModule(name string) (Module, bool)

	// GetAllModules returns all registered modules
	GetAllModules() []Module

	// GetEnabledModules returns only enabled modules
	GetEnabledModules() []Module

	// EnableModule enables a module by name
	EnableModule(name string) error

	// DisableModule disables a module by name
	DisableModule(name string) error

	// GetModuleInfo returns metadata about all modules
	GetModuleInfo() []ModuleInfo
}
