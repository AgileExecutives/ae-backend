package modules

import (
	"log"

	"github.com/ae-base-server/internal/config"
	"gorm.io/gorm"
)

// RegisterAllModules registers all available modules with the manager
func RegisterAllModules(manager *Manager) error {
	log.Println("📦 Registering available modules...")

	// Register demo module for testing
	demo := NewDemoModule()
	if err := manager.RegisterModule(demo); err != nil {
		return err
	}

	// Register calendar module
	// Commented out until we have proper import path
	// calendar := calendar.NewCalendarModule()
	// if err := manager.RegisterModule(calendar); err != nil {
	//     return fmt.Errorf("failed to register calendar module: %w", err)
	// }

	// Add more modules here as they become available

	log.Println("✅ All modules registered successfully")
	return nil
}

// CreateModuleManager creates and initializes a module manager with all modules
func CreateModuleManager(db *gorm.DB, cfg config.Config) (*Manager, error) {
	manager := NewManager(db, cfg)

	// Register all available modules
	if err := RegisterAllModules(manager); err != nil {
		return nil, err
	}

	return manager, nil
}
