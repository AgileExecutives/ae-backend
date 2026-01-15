package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ae-base-server/pkg/settings/entities"
	"github.com/ae-base-server/pkg/settings/repository"
	"github.com/unburdy/unburdy-server-api/modules/client_management/settings"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SettingsSeedService handles seeding of billing-related settings
type SettingsSeedService struct {
	db *gorm.DB
}

// NewSettingsSeedService creates a new settings seed service
func NewSettingsSeedService(db *gorm.DB) *SettingsSeedService {
	return &SettingsSeedService{db: db}
}

// RegisterSettingDefinition registers or updates a setting definition
func (s *SettingsSeedService) RegisterSettingDefinition(domain, key string, version int, schema, data map[string]interface{}) error {
	log.Printf("🔧 Registering setting definition: %s.%s (version %d)", domain, key, version)
	repo := repository.NewSettingsRepository(s.db)

	// Convert maps to JSON for storage
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Check if definition exists
	existing, err := repo.GetSettingDefinition(domain, key)
	if err != nil {
		return fmt.Errorf("failed to check existing definition: %w", err)
	}

	if existing != nil {
		// Update if version is newer
		if version > existing.Version {
			existing.Version = version
			existing.Schema = datatypes.JSON(schemaJSON)
			existing.Data = datatypes.JSON(dataJSON)
			if err := repo.UpdateSettingDefinition(existing); err != nil {
				return fmt.Errorf("failed to update setting definition: %w", err)
			}
			log.Printf("✅ Updated setting definition: %s.%s (version %d)", domain, key, version)
		} else {
			log.Printf("⏭️  Setting definition %s.%s already exists with version %d, skipping", domain, key, existing.Version)
		}
		return nil
	}

	// Create new definition
	if version == 0 {
		version = 1
	}

	definition := &entities.SettingDefinition{
		Domain:  domain,
		Key:     key,
		Version: version,
		Schema:  datatypes.JSON(schemaJSON),
		Data:    datatypes.JSON(dataJSON),
	}

	if err := repo.CreateSettingDefinition(definition); err != nil {
		return fmt.Errorf("failed to create setting definition: %w", err)
	}

	log.Printf("✅ Registered setting definition: %s.%s (version %d)", domain, key, version)
	return nil
}

// SeedBillingSettings seeds billing-related settings definitions and tenant defaults
func (s *SettingsSeedService) SeedBillingSettings() error {
	log.Println("📦 Seeding billing settings definitions...")
	fmt.Println("📦📦📦 IN SeedBillingSettings() 📦📦📦")

	// Get all billing settings definitions from the settings package
	definitions := settings.GetBillingSettingsDefinitions()
	fmt.Printf("📋 Got %d billing settings definitions\n", len(definitions))

	// Register each definition
	for _, def := range definitions {
		fmt.Printf("🔧 Registering: %s.%s\n", def.Domain, def.Key)
		err := s.RegisterSettingDefinition(def.Domain, def.Key, def.Version, def.Schema, def.Data)
		if err != nil {
			fmt.Printf("❌ Failed to register %s.%s: %v\n", def.Domain, def.Key, err)
			return fmt.Errorf("failed to register %s.%s: %w", def.Domain, def.Key, err)
		}
		fmt.Printf("✅ Registered: %s.%s\n", def.Domain, def.Key)
	}

	log.Println("✅ Billing settings definitions seeded")

	// Seed default settings for each tenant/organization
	if err := s.SeedTenantDefaults(); err != nil {
		return fmt.Errorf("failed to seed tenant defaults: %w", err)
	}

	return nil
}

// SeedTenantDefaults creates default settings for all tenants that don't have them
func (s *SettingsSeedService) SeedTenantDefaults() error {
	log.Println("📦 Seeding default settings for tenants...")

	// Get all tenants (query directly without importing internal models)
	type TenantBasic struct {
		ID   uint
		Name string
	}
	var tenants []TenantBasic
	if err := s.db.Table("tenants").Select("id, name").Find(&tenants).Error; err != nil {
		return fmt.Errorf("failed to fetch tenants: %w", err)
	}

	if len(tenants) == 0 {
		log.Println("⏭️  No tenants found, skipping tenant settings seeding")
		return nil
	}

	repo := repository.NewSettingsRepository(s.db)
	definitions := settings.GetBillingSettingsDefinitions()

	// For each tenant, seed default settings
	for _, tenant := range tenants {
		log.Printf("📦 Seeding settings for tenant %d (%s)...", tenant.ID, tenant.Name)

		for _, def := range definitions {
			// Check if setting already exists for this tenant
			existing, err := repo.GetSetting(tenant.ID, def.Domain, def.Key)
			if err != nil {
				return fmt.Errorf("failed to check existing setting for tenant %d: %w", tenant.ID, err)
			}

			if existing != nil {
				log.Printf("⏭️  Setting %s.%s already exists for tenant %d, skipping", def.Domain, def.Key, tenant.ID)
				continue
			}

			// Get the setting definition to get the default data
			settingDef, err := repo.GetSettingDefinition(def.Domain, def.Key)
			if err != nil {
				return fmt.Errorf("failed to get setting definition %s.%s: %w", def.Domain, def.Key, err)
			}

			if settingDef == nil {
				log.Printf("⚠️  Setting definition %s.%s not found, skipping", def.Domain, def.Key)
				continue
			}

			// Create setting with default data from definition
			setting := &entities.Setting{
				TenantID:            tenant.ID,
				Domain:              def.Domain,
				Key:                 def.Key,
				Version:             def.Version,
				Data:                settingDef.Data, // Use default data from definition
				SettingDefinitionID: settingDef.ID,
			}

			if err := repo.SetSetting(setting); err != nil {
				return fmt.Errorf("failed to create setting %s.%s for tenant %d: %w", def.Domain, def.Key, tenant.ID, err)
			}

			log.Printf("✅ Created default setting %s.%s for tenant %d", def.Domain, def.Key, tenant.ID)
		}
	}

	log.Printf("✅ Default settings seeded for %d tenant(s)", len(tenants))
	return nil
}
