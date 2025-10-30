package modules

import (
"log"

"github.com/ae-saas-basic/ae-saas-basic/internal/config"
"github.com/ae-saas-basic/ae-saas-basic/internal/eventbus"
"github.com/gin-gonic/gin"
"gorm.io/gorm"
)

// Manager handles module lifecycle and integration
type Manager struct {
	registry *DefaultModuleRegistry
	db       *gorm.DB
	cfg      config.Config
}

// NewManager creates a new module manager
func NewManager(db *gorm.DB, cfg config.Config) *Manager {
	return &Manager{
		registry: NewModuleRegistry(db, cfg),
		db:       db,
		cfg:      cfg,
	}
}

// RegisterModule registers a module with the system
func (m *Manager) RegisterModule(module Module) error {
	return m.registry.RegisterModule(module)
}

// GetRegistry returns the module registry
func (m *Manager) GetRegistry() *DefaultModuleRegistry {
	return m.registry
}

// InitializeModules performs full module initialization
func (m *Manager) InitializeModules() error {
	log.Println("🚀 Initializing module system...")
	
	// Run module migrations
	if err := m.registry.MigrateModels(); err != nil {
		return err
	}
	
	// Register event handlers with the global event bus
	if err := m.registerEventHandlers(); err != nil {
		return err
	}
	
	log.Println("✅ Module system initialized successfully")
	return nil
}

// registerEventHandlers registers all module event handlers with the event bus
func (m *Manager) registerEventHandlers() error {
	log.Println("🎯 Registering module event handlers with event bus...")
	
	enabledModules := m.registry.GetEnabledModules()
	for _, module := range enabledModules {
		handlers := module.GetEventHandlers()
		if len(handlers) == 0 {
			log.Printf("📡 Module %s: No event handlers to register", module.GetName())
			continue
		}
		
		log.Printf("📡 Module %s: Registering %d event handlers", module.GetName(), len(handlers))
		for _, handler := range handlers {
			if err := eventbus.Subscribe(handler); err != nil {
				return err
			}
			log.Printf("   ✅ Registered handler: %s for events: %v", handler.GetName(), handler.GetEventTypes())
		}
		log.Printf("✅ Module %s: All event handlers registered with event bus", module.GetName())
	}
	
	log.Println("✅ All module event handlers registered with event bus")
	return nil
}

// RegisterModuleRoutes registers HTTP routes for all enabled modules
func (m *Manager) RegisterModuleRoutes(router *gin.Engine) {
	log.Println("🛣️  Registering module routes...")
	
	enabledModules := m.registry.GetEnabledModules()
	for _, module := range enabledModules {
		log.Printf("🛣️  Module %s: Registering routes", module.GetName())
		
		// Create a route group for the module
		moduleGroup := router.Group("/api/v1/modules/" + module.GetName())
		module.RegisterRoutes(moduleGroup, m.db, m.cfg)
		
		log.Printf("✅ Module %s: Routes registered at /api/v1/modules/%s", module.GetName(), module.GetName())
	}
	
	log.Println("✅ All module routes registered")
}

// GetModuleInfo returns information about all registered modules
func (m *Manager) GetModuleInfo() []ModuleInfo {
	return m.registry.GetModuleInfo()
}

// Shutdown gracefully shuts down all modules
func (m *Manager) Shutdown() error {
	return m.registry.Shutdown()
}
