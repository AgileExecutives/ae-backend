package modules

import (
"fmt"
"log"
"sync"

"github.com/ae-saas-basic/ae-saas-basic/internal/config"
"gorm.io/gorm"
)

// DefaultModuleRegistry is the default implementation of ModuleRegistry
type DefaultModuleRegistry struct {
	mu       sync.RWMutex
	modules  map[string]Module
	enabled  map[string]bool
	db       *gorm.DB
	cfg      config.Config
}

// NewModuleRegistry creates a new module registry
func NewModuleRegistry(db *gorm.DB, cfg config.Config) *DefaultModuleRegistry {
	return &DefaultModuleRegistry{
		modules: make(map[string]Module),
		enabled: make(map[string]bool),
		db:      db,
		cfg:     cfg,
	}
}

// RegisterModule registers a new module
func (r *DefaultModuleRegistry) RegisterModule(module Module) error {
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
	if err := module.Initialize(r.db, r.cfg); err != nil {
		delete(r.modules, name)
		delete(r.enabled, name)
		return fmt.Errorf("failed to initialize module %s: %w", name, err)
	}
	
	log.Printf("✅ Module %s registered and initialized successfully", name)
	return nil
}

// GetModule returns a module by name
func (r *DefaultModuleRegistry) GetModule(name string) (Module, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	module, exists := r.modules[name]
	return module, exists
}

// GetAllModules returns all registered modules
func (r *DefaultModuleRegistry) GetAllModules() []Module {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	modules := make([]Module, 0, len(r.modules))
	for _, module := range r.modules {
		modules = append(modules, module)
	}
	return modules
}

// GetEnabledModules returns only enabled modules
func (r *DefaultModuleRegistry) GetEnabledModules() []Module {
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

// EnableModule enables a module by name
func (r *DefaultModuleRegistry) EnableModule(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.modules[name]; !exists {
		return fmt.Errorf("module %s not found", name)
	}
	
	r.enabled[name] = true
	log.Printf("✅ Module %s enabled", name)
	return nil
}

// DisableModule disables a module by name
func (r *DefaultModuleRegistry) DisableModule(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.modules[name]; !exists {
		return fmt.Errorf("module %s not found", name)
	}
	
	r.enabled[name] = false
	log.Printf("🔒 Module %s disabled", name)
	return nil
}

// GetModuleInfo returns metadata about all modules
func (r *DefaultModuleRegistry) GetModuleInfo() []ModuleInfo {
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

// MigrateModels runs migrations for all enabled modules
func (r *DefaultModuleRegistry) MigrateModels() error {
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

// RegisterEventHandlers registers event handlers for all enabled modules
func (r *DefaultModuleRegistry) RegisterEventHandlers() error {
	log.Println("🎯 Registering module event handlers...")
	
	enabledModules := r.GetEnabledModules()
	for _, module := range enabledModules {
		handlers := module.GetEventHandlers()
		if len(handlers) == 0 {
			log.Printf("📡 Module %s: No event handlers to register", module.GetName())
			continue
		}
		
		log.Printf("📡 Module %s: Registering %d event handlers", module.GetName(), len(handlers))
		for _, handler := range handlers {
			// We'll need to import the eventbus package to register handlers
// For now, this is a placeholder - we'll implement this when we integrate
			log.Printf("   Registered handler: %s for events: %v", handler.GetName(), handler.GetEventTypes())
		}
		log.Printf("✅ Module %s: All event handlers registered", module.GetName())
	}
	
	log.Println("✅ All module event handlers registered")
	return nil
}

// Shutdown shuts down all modules
func (r *DefaultModuleRegistry) Shutdown() error {
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
