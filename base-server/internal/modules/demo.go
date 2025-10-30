package modules

import (
	"github.com/ae-saas-basic/ae-saas-basic/internal/config"
	"github.com/ae-saas-basic/ae-saas-basic/pkg/eventbus"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DemoModule is a simple demo module for testing
type DemoModule struct {
	name    string
	version string
}

// NewDemoModule creates a new demo module instance
func NewDemoModule() *DemoModule {
	return &DemoModule{
		name:    "demo",
		version: "1.0.0",
	}
}

// GetName returns the module name
func (m *DemoModule) GetName() string {
	return m.name
}

// GetVersion returns the module version
func (m *DemoModule) GetVersion() string {
	return m.version
}

// GetModels returns the models to migrate
func (m *DemoModule) GetModels() []interface{} {
	// No models for this demo module
	return []interface{}{}
}

// GetEventHandlers returns event handlers for this module
func (m *DemoModule) GetEventHandlers() []eventbus.EventHandler {
	// No event handlers for this demo module
	return []eventbus.EventHandler{}
}

// RegisterRoutes registers HTTP routes for this module
func (m *DemoModule) RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, cfg config.Config) {
	// For static Swagger generation, we need to import and use the handlers
	// that have the Swagger annotations. However, this creates a circular import
	// issue between modules and handlers.

	// The solution is to either:
	// 1. Move handler logic here with Swagger annotations, or
	// 2. Use a different approach for Swagger generation

	// For now, using inline handlers with Swagger comments

	// GetDemo returns demo information
	// @Summary Get demo information
	// @Description Get information about the demo module
	// @Tags demo
	// @Produce json
	// @Success 200 {object} map[string]interface{} "Demo information"
	// @Router /api/v1/modules/demo/demo [get]
	router.GET("/demo", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Demo module is working!", "module": m.name, "version": m.version})
	})
}

// Initialize performs module initialization
func (m *DemoModule) Initialize(db *gorm.DB, cfg config.Config) error {
	// Nothing to initialize for demo module
	return nil
}

// Shutdown performs cleanup
func (m *DemoModule) Shutdown() error {
	// Nothing to cleanup for demo module
	return nil
}
