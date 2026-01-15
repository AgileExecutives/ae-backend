package repository

import (
	"errors"

	"github.com/ae-base-server/pkg/settings/entities"
	"gorm.io/gorm"
)

// SettingsRepository handles database operations for settings
type SettingsRepository struct {
	db *gorm.DB
}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// --- Setting Definitions ---

// GetSettingDefinition retrieves a setting definition by domain and key
func (r *SettingsRepository) GetSettingDefinition(domain, key string) (*entities.SettingDefinition, error) {
	var definition entities.SettingDefinition
	err := r.db.Where("domain = ? AND key = ?", domain, key).First(&definition).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &definition, nil
}

// CreateSettingDefinition creates a new setting definition
func (r *SettingsRepository) CreateSettingDefinition(definition *entities.SettingDefinition) error {
	return r.db.Create(definition).Error
}

// UpdateSettingDefinition updates an existing setting definition
func (r *SettingsRepository) UpdateSettingDefinition(definition *entities.SettingDefinition) error {
	return r.db.Save(definition).Error
}

// GetAllSettingDefinitions retrieves all setting definitions
func (r *SettingsRepository) GetAllSettingDefinitions() ([]entities.SettingDefinition, error) {
	var definitions []entities.SettingDefinition
	err := r.db.Find(&definitions).Error
	return definitions, err
}

// GetSettingDefinitionsByDomain retrieves all setting definitions for a domain
func (r *SettingsRepository) GetSettingDefinitionsByDomain(domain string) ([]entities.SettingDefinition, error) {
	var definitions []entities.SettingDefinition
	err := r.db.Where("domain = ?", domain).Find(&definitions).Error
	return definitions, err
}

// --- Tenant Settings ---

// GetSetting retrieves a specific tenant setting
func (r *SettingsRepository) GetSetting(tenantID uint, domain, key string) (*entities.Setting, error) {
	var setting entities.Setting
	err := r.db.Where("tenant_id = ? AND domain = ? AND key = ?",
		tenantID, domain, key).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &setting, nil
}

// SetSetting creates or updates a tenant setting
func (r *SettingsRepository) SetSetting(setting *entities.Setting) error {
	return r.db.Save(setting).Error
}

// GetDomainSettings retrieves all tenant settings for a domain
func (r *SettingsRepository) GetDomainSettings(tenantID uint, domain string) ([]entities.Setting, error) {
	var settings []entities.Setting
	err := r.db.Where("tenant_id = ? AND domain = ?",
		tenantID, domain).Find(&settings).Error
	return settings, err
}

// GetAllSettings retrieves all tenant settings
func (r *SettingsRepository) GetAllSettings(tenantID uint) ([]entities.Setting, error) {
	var settings []entities.Setting
	err := r.db.Where("tenant_id = ?", tenantID).Find(&settings).Error
	return settings, err
}

// DeleteSetting removes a specific tenant setting
func (r *SettingsRepository) DeleteSetting(tenantID uint, domain, key string) error {
	result := r.db.Where("tenant_id = ? AND domain = ? AND key = ?",
		tenantID, domain, key).Delete(&entities.Setting{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("setting not found")
	}
	return nil
}

// GetDomains returns available domains for a tenant
func (r *SettingsRepository) GetDomains(tenantID uint) ([]string, error) {
	var domains []string
	err := r.db.Model(&entities.Setting{}).
		Where("tenant_id = ?", tenantID).
		Distinct("domain").
		Pluck("domain", &domains).Error
	return domains, err
}

// AutoMigrate creates the settings tables
func (r *SettingsRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&entities.SettingDefinition{}, &entities.Setting{})
}

// HealthCheck verifies database connection
func (r *SettingsRepository) HealthCheck() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
