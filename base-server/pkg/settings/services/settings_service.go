package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ae-base-server/pkg/settings/entities"
	"github.com/ae-base-server/pkg/settings/repository"
)

// SettingsService provides business logic for settings management
type SettingsService struct {
	repo *repository.SettingsRepository
}

// NewSettingsService creates a new settings service
func NewSettingsService(repo *repository.SettingsRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

// GetSetting retrieves and parses a setting value
func (s *SettingsService) GetSetting(tenantID uint, organizationID, domain, key string) (interface{}, error) {
	// Note: organizationID is kept in signature for backward compatibility but not used in current schema
	setting, err := s.repo.GetSetting(tenantID, domain, key)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, nil
	}
	return s.parseSettingData(setting.Data)
}

// SetSetting creates or updates a setting
func (s *SettingsService) SetSetting(tenantID uint, organizationID, domain, key string, value interface{}, valueType string) error {
	// Note: organizationID and valueType are kept in signature for backward compatibility but not used in current schema
	serializedData, err := s.serializeData(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	setting := &entities.Setting{
		TenantID:  tenantID,
		Domain:    domain,
		Key:       key,
		Data:      serializedData,
		UpdatedAt: time.Now(),
	}

	// Check if setting exists to set created time
	existingSetting, err := s.repo.GetSetting(tenantID, domain, key)
	if err != nil && !errors.Is(err, errors.New("record not found")) {
		return err
	}

	if existingSetting != nil {
		setting.ID = existingSetting.ID
		setting.CreatedAt = existingSetting.CreatedAt
	} else {
		setting.CreatedAt = time.Now()
	}

	return s.repo.SetSetting(setting)
}

// GetDomainSettings retrieves all settings for a domain
func (s *SettingsService) GetDomainSettings(tenantID uint, organizationID, domain string) (map[string]interface{}, error) {
	// Note: organizationID is kept in signature for backward compatibility but not used in current schema
	settings, err := s.repo.GetDomainSettings(tenantID, domain)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, setting := range settings {
		value, err := s.parseSettingData(setting.Data)
		if err != nil {
			// Skip invalid settings
			continue
		}
		result[setting.Key] = value
	}

	return result, nil
}

// GetAllSettings retrieves all settings grouped by domain
func (s *SettingsService) GetAllSettings(tenantID uint, organizationID string) (map[string]map[string]interface{}, error) {
	// Note: organizationID is kept in signature for backward compatibility but not used in current schema
	settings, err := s.repo.GetAllSettings(tenantID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]interface{})
	for _, setting := range settings {
		if result[setting.Domain] == nil {
			result[setting.Domain] = make(map[string]interface{})
		}

		value, err := s.parseSettingData(setting.Data)
		if err != nil {
			// Skip invalid settings
			continue
		}
		result[setting.Domain][setting.Key] = value
	}

	return result, nil
}

// DeleteSetting removes a specific setting
func (s *SettingsService) DeleteSetting(tenantID uint, organizationID, domain, key string) error {
	// Note: organizationID is kept in signature for backward compatibility but not used in current schema
	return s.repo.DeleteSetting(tenantID, domain, key)
}

// DeleteDomainSettings removes all settings for a domain
func (s *SettingsService) DeleteDomainSettings(tenantID uint, organizationID, domain string) error {
	// Note: organizationID is kept in signature for backward compatibility but not used in current schema
	// DeleteDomainSettings doesn't exist in repo - need to delete individually
	settings, err := s.repo.GetDomainSettings(tenantID, domain)
	if err != nil {
		return err
	}
	for _, setting := range settings {
		if err := s.repo.DeleteSetting(tenantID, setting.Domain, setting.Key); err != nil {
			return err
		}
	}
	return nil
}

// GetDomains returns available domains for an organization
func (s *SettingsService) GetDomains(tenantID uint, organizationID string) ([]string, error) {
	// Note: organizationID is kept in signature for backward compatibility but not used in current schema
	return s.repo.GetDomains(tenantID)
}

// ValidateSettings validates settings against basic rules
func (s *SettingsService) ValidateSettings(domain string, settings map[string]interface{}) (bool, []string) {
	var errors []string

	// Basic validation rules
	for key, value := range settings {
		if value == nil {
			errors = append(errors, fmt.Sprintf("%s cannot be null", key))
			continue
		}

		// Domain-specific validation
		switch domain {
		case "company":
			if key == "company_email" {
				if str, ok := value.(string); ok {
					if !isValidEmail(str) {
						errors = append(errors, "company_email must be a valid email address")
					}
				}
			}
		case "invoice":
			if key == "invoice_prefix" {
				if str, ok := value.(string); ok {
					if len(str) == 0 {
						errors = append(errors, "invoice_prefix cannot be empty")
					}
				}
			}
		}
	}

	return len(errors) == 0, errors
}

// GetModules returns list of available modules
func (s *SettingsService) GetModules() []string {
	return []string{"company", "invoice", "billing", "localization", "booking", "notification", "integration"}
}

// HealthCheck performs system health check
func (s *SettingsService) HealthCheck() (*entities.HealthResponse, error) {
	err := s.repo.HealthCheck()
	status := "ok"
	dbStatus := "connected"

	if err != nil {
		status = "error"
		dbStatus = "disconnected"
	}

	return &entities.HealthResponse{
		Status:   status,
		Database: dbStatus,
		Modules:  len(s.GetModules()),
		Version:  "1.0.0",
	}, nil
}

// parseSettingData parses JSONB data field to interface{}
func (s *SettingsService) parseSettingData(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var result interface{}
	err := json.Unmarshal(data, &result)
	return result, err
}

// serializeData converts value to JSONB for storage
func (s *SettingsService) serializeData(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	// Simple email validation - in production use a proper regex or library
	return len(email) > 0 &&
		len(email) < 254 &&
		containsAt(email) &&
		containsDot(email)
}

func containsAt(s string) bool {
	for _, r := range s {
		if r == '@' {
			return true
		}
	}
	return false
}

func containsDot(s string) bool {
	for _, r := range s {
		if r == '.' {
			return true
		}
	}
	return false
}
