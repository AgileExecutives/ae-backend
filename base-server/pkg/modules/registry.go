package modules

import (
	"fmt"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Registry manages all registered modules
type Registry struct {
	mu      sync.RWMutex
	modules map[string]Module
	enabled map[string]bool
	db      *gorm.DB
}

// NewRegistry creates a new module registry
func NewRegistry(db *gorm.DB) *Registry {
	return &Registry{
		modules: make(map[string]Module),
		enabled: make(map[string]bool),
		db:      db,
	}
}

// RegisterModule registers a new module
func (r *Registry) RegisterModule(module Module) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := module.GetName()
	if _, exists := r.modules[name]; exists {
		return fmt.Errorf("module %s is already registered", name)
	}

	log.Printf("📦 Registering module: %s v%s", name, module.GetVersion())

	r.modules[name] = module
	r.enabled[name] = true // Enable by default

	// Initialize the module
	if err := module.Initialize(r.db); err != nil {
		delete(r.modules, name)
		delete(r.enabled, name)
		return fmt.Errorf("failed to initialize module %s: %w", name, err)
	}

	log.Printf("✅ Module %s registered and initialized successfully", name)
	return nil
}

// GetModule returns a module by name
func (r *Registry) GetModule(name string) (Module, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	module, exists := r.modules[name]
	return module, exists
}

// GetEnabledModules returns only enabled modules
func (r *Registry) GetEnabledModules() []Module {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var modules []Module
	for name, module := range r.modules {
		if r.enabled[name] {
			modules = append(modules, module)
		}
	}
	return modules
}

// MigrateModels runs migrations for all enabled modules
func (r *Registry) MigrateModels() error {
	log.Println("🗃️  Running module migrations...")

	enabledModules := r.GetEnabledModules()
	for _, module := range enabledModules {
		models := module.GetModels()
		if len(models) == 0 {
			log.Printf("📝 Module %s: No models to migrate", module.GetName())
			continue
		}

		log.Printf("📝 Module %s: Migrating %d models", module.GetName(), len(models))
		for i, model := range models {
			log.Printf("   Migrating model %d/%d: %T", i+1, len(models), model)
			if err := r.db.AutoMigrate(model); err != nil {
				return fmt.Errorf("failed to migrate model %T for module %s: %w", model, module.GetName(), err)
			}
		}
		log.Printf("✅ Module %s: All models migrated successfully", module.GetName())
	}

	log.Println("✅ All module migrations completed")
	return nil
}

// RegisterModuleRoutes registers HTTP routes for all enabled modules
func (r *Registry) RegisterModuleRoutes(router *gin.Engine) {
	log.Println("🛣️  Registering module routes...")

	enabledModules := r.GetEnabledModules()
	for _, module := range enabledModules {
		log.Printf("🛣️  Module %s: Registering routes", module.GetName())

		// Create a route group for the module
		moduleGroup := router.Group("/api/v1/modules/" + module.GetName())
		module.RegisterRoutes(moduleGroup, r.db)

		log.Printf("✅ Module %s: Routes registered at /api/v1/modules/%s", module.GetName(), module.GetName())
	}

	log.Println("✅ All module routes registered")
}

// GetModuleInfo returns information about all registered modules
func (r *Registry) GetModuleInfo() []ModuleInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var info []ModuleInfo
	for name, module := range r.modules {
		info = append(info, ModuleInfo{
			Name:        name,
			Version:     module.GetVersion(),
			Description: fmt.Sprintf("Module: %s", name),
			Author:      "Base Server",
			Enabled:     r.enabled[name],
		})
	}
	return info
}

// Shutdown gracefully shuts down all modules
func (r *Registry) Shutdown() error {
	log.Println("🛑 Shutting down modules...")

	var errors []error
	for name, module := range r.modules {
		log.Printf("🛑 Shutting down module: %s", name)
		if err := module.Shutdown(); err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown module %s: %w", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("module shutdown errors: %v", errors)
	}

	log.Println("✅ All modules shut down successfully")
	return nil
}

// GetCombinedSwaggerInfo combines Swagger documentation from all enabled modules
func (r *Registry) GetCombinedSwaggerInfo() map[string]SwaggerModuleInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	combined := make(map[string]SwaggerModuleInfo)

	for name, module := range r.modules {
		if r.enabled[name] {
			swaggerInfo := module.GetSwaggerInfo()
			combined[name] = swaggerInfo
			log.Printf("📋 Added Swagger documentation for module: %s", name)
		}
	}

	return combined
}

// GenerateSwaggerTags returns all Swagger tags from enabled modules
func (r *Registry) GenerateSwaggerTags() []SwaggerTag {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allTags []SwaggerTag

	for name, module := range r.modules {
		if r.enabled[name] {
			swaggerInfo := module.GetSwaggerInfo()
			allTags = append(allTags, swaggerInfo.Tags...)
		}
	}

	return allTags
}
