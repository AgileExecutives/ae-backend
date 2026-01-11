package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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
	setting, err := s.repo.GetSetting(tenantID, organizationID, domain, key)
	if err != nil {
		return nil, err
	}
	return s.parseSettingValue(setting.Value, setting.Type)
}

// SetSetting creates or updates a setting
func (s *SettingsService) SetSetting(tenantID uint, organizationID, domain, key string, value interface{}, valueType string) error {
	serializedValue, err := s.serializeValue(value, valueType)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	setting := &entities.Setting{
		TenantID:       tenantID,
		OrganizationID: organizationID,
		Domain:         domain,
		Key:            key,
		Value:          serializedValue,
		Type:           valueType,
		UpdatedAt:      time.Now(),
	}

	// Check if setting exists to set created time
	existingSetting, err := s.repo.GetSetting(tenantID, organizationID, domain, key)
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
	settings, err := s.repo.GetDomainSettings(tenantID, organizationID, domain)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, setting := range settings {
		value, err := s.parseSettingValue(setting.Value, setting.Type)
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
	settings, err := s.repo.GetAllSettings(tenantID, organizationID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]interface{})
	for _, setting := range settings {
		if result[setting.Domain] == nil {
			result[setting.Domain] = make(map[string]interface{})
		}

		value, err := s.parseSettingValue(setting.Value, setting.Type)
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
	return s.repo.DeleteSetting(tenantID, organizationID, domain, key)
}

// DeleteDomainSettings removes all settings for a domain
func (s *SettingsService) DeleteDomainSettings(tenantID uint, organizationID, domain string) error {
	return s.repo.DeleteDomainSettings(tenantID, organizationID, domain)
}

// GetDomains returns available domains for an organization
func (s *SettingsService) GetDomains(tenantID uint, organizationID string) ([]string, error) {
	return s.repo.GetDomains(tenantID, organizationID)
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

// parseSettingValue converts stored string value back to appropriate type
func (s *SettingsService) parseSettingValue(value, valueType string) (interface{}, error) {
	switch valueType {
	case "string":
		return value, nil
	case "int":
		return strconv.Atoi(value)
	case "bool":
		return strconv.ParseBool(value)
	case "float":
		return strconv.ParseFloat(value, 64)
	case "json":
		var result interface{}
		err := json.Unmarshal([]byte(value), &result)
		return result, err
	default:
		return value, nil
	}
}

// serializeValue converts value to string for storage
func (s *SettingsService) serializeValue(value interface{}, valueType string) (string, error) {
	switch valueType {
	case "string":
		if str, ok := value.(string); ok {
			return str, nil
		}
		return fmt.Sprintf("%v", value), nil
	case "int":
		return fmt.Sprintf("%v", value), nil
	case "bool":
		return fmt.Sprintf("%v", value), nil
	case "float":
		return fmt.Sprintf("%v", value), nil
	case "json":
		bytes, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	default:
		return fmt.Sprintf("%v", value), nil
	}
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
