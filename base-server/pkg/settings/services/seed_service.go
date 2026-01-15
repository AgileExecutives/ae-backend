package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ae-base-server/pkg/settings/entities"
	"github.com/ae-base-server/pkg/settings/repository"
	"gorm.io/datatypes"
)

// SeedService handles seeding of setting definitions
type SeedService struct {
	repo *repository.SettingsRepository
}

// NewSeedService creates a new seed service
func NewSeedService(repo *repository.SettingsRepository) *SeedService {
	return &SeedService{repo: repo}
}

// RegisterSettingDefinition registers or updates a setting definition
func (s *SeedService) RegisterSettingDefinition(reg entities.SettingRegistration) error {
	// Check if definition already exists
	existing, err := s.repo.GetSettingDefinition(reg.Domain, reg.Key)

	// Marshal schema and data
	schemaJSON, err := json.Marshal(reg.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	var dataJSON datatypes.JSON
	if reg.Data != nil {
		dataBytes, err := json.Marshal(reg.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		dataJSON = dataBytes
	}

	if existing != nil {
		// Update if version is newer
		if reg.Version > existing.Version {
			existing.Version = reg.Version
			existing.Schema = schemaJSON
			existing.Data = dataJSON
			return s.repo.UpdateSettingDefinition(existing)
		}
		log.Printf("⏭️  Setting definition %s.%s already exists with version %d, skipping", reg.Domain, reg.Key, existing.Version)
		return nil
	}

	// Create new definition
	version := reg.Version
	if version == 0 {
		version = 1
	}

	definition := &entities.SettingDefinition{
		Domain:  reg.Domain,
		Key:     reg.Key,
		Version: version,
		Schema:  schemaJSON,
		Data:    dataJSON,
	}

	err = s.repo.CreateSettingDefinition(definition)
	if err != nil {
		return fmt.Errorf("failed to create setting definition: %w", err)
	}

	log.Printf("✅ Registered setting definition: %s.%s (version %d)", reg.Domain, reg.Key, version)
	return nil
}

// SeedBaseServerSettings seeds core settings from base-server
func (s *SeedService) SeedBaseServerSettings() error {
	log.Println("📦 Seeding base-server settings definitions...")

	// Organization domain - locale
	localeSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"locale": map[string]interface{}{
				"type":    "string",
				"default": "de-DE",
			},
			"format": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"date": map[string]interface{}{
						"type":    "string",
						"default": "02.01.2006",
					},
					"time": map[string]interface{}{
						"type":    "string",
						"default": "15:04",
					},
					"amount": map[string]interface{}{
						"type":    "string",
						"default": "de",
					},
				},
			},
		},
	}

	localeData := map[string]interface{}{
		"locale": "de-DE",
		"format": map[string]interface{}{
			"date":   "02.01.2006",
			"time":   "15:04",
			"amount": "de",
		},
	}

	err := s.RegisterSettingDefinition(entities.SettingRegistration{
		Domain:  "organization",
		Key:     "locale",
		Version: 1,
		Schema:  localeSchema,
		Data:    localeData,
	})
	if err != nil {
		return fmt.Errorf("failed to register organization.locale: %w", err)
	}

	log.Println("✅ Base-server settings definitions seeded")
	return nil
}
