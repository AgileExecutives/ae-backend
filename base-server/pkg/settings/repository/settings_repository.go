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

// GetSetting retrieves a specific setting
func (r *SettingsRepository) GetSetting(tenantID uint, organizationID, domain, key string) (*entities.Setting, error) {
	var setting entities.Setting
	err := r.db.Where("tenant_id = ? AND organization_id = ? AND domain = ? AND key = ?",
		tenantID, organizationID, domain, key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

// SetSetting creates or updates a setting
func (r *SettingsRepository) SetSetting(setting *entities.Setting) error {
	return r.db.Save(setting).Error
}

// GetDomainSettings retrieves all settings for a domain
func (r *SettingsRepository) GetDomainSettings(tenantID uint, organizationID, domain string) ([]entities.Setting, error) {
	var settings []entities.Setting
	err := r.db.Where("tenant_id = ? AND organization_id = ? AND domain = ?",
		tenantID, organizationID, domain).Find(&settings).Error
	return settings, err
}

// GetAllSettings retrieves all settings for an organization
func (r *SettingsRepository) GetAllSettings(tenantID uint, organizationID string) ([]entities.Setting, error) {
	var settings []entities.Setting
	err := r.db.Where("tenant_id = ? AND organization_id = ?",
		tenantID, organizationID).Find(&settings).Error
	return settings, err
}

// DeleteSetting removes a specific setting
func (r *SettingsRepository) DeleteSetting(tenantID uint, organizationID, domain, key string) error {
	result := r.db.Where("tenant_id = ? AND organization_id = ? AND domain = ? AND key = ?",
		tenantID, organizationID, domain, key).Delete(&entities.Setting{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("setting not found")
	}
	return nil
}

// DeleteDomainSettings removes all settings for a domain
func (r *SettingsRepository) DeleteDomainSettings(tenantID uint, organizationID, domain string) error {
	return r.db.Where("tenant_id = ? AND organization_id = ? AND domain = ?",
		tenantID, organizationID, domain).Delete(&entities.Setting{}).Error
}

// GetDomains returns available domains for an organization
func (r *SettingsRepository) GetDomains(tenantID uint, organizationID string) ([]string, error) {
	var domains []string
	err := r.db.Model(&entities.Setting{}).
		Where("tenant_id = ? AND organization_id = ?", tenantID, organizationID).
		Distinct("domain").
		Pluck("domain", &domains).Error
	return domains, err
}

// AutoMigrate creates the settings table
func (r *SettingsRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&entities.Setting{})
}

// HealthCheck verifies database connection
func (r *SettingsRepository) HealthCheck() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
