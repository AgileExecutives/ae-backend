package entities

import (
	"time"

	"gorm.io/datatypes"
)

// SettingDefinition represents a schema definition for a setting
type SettingDefinition struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	Domain    string         `json:"domain" gorm:"not null;uniqueIndex:idx_setting_def_domain_key"`
	Key       string         `json:"key" gorm:"not null;uniqueIndex:idx_setting_def_domain_key"`
	Version   int            `json:"version" gorm:"not null;default:1"`
	Schema    datatypes.JSON `json:"schema" gorm:"type:jsonb"`
	Data      datatypes.JSON `json:"data" gorm:"type:jsonb"` // Default/sample values
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// TableName specifies the table name for SettingDefinition
func (SettingDefinition) TableName() string {
	return "setting_definitions"
}

// Setting represents a tenant-specific configuration value
type Setting struct {
	ID                  uint               `json:"id" gorm:"primarykey"`
	TenantID            uint               `json:"tenant_id" gorm:"not null;uniqueIndex:idx_setting_tenant_domain_key"`
	Domain              string             `json:"domain" gorm:"not null;uniqueIndex:idx_setting_tenant_domain_key"`
	Key                 string             `json:"key" gorm:"not null;uniqueIndex:idx_setting_tenant_domain_key"`
	Version             int                `json:"version" gorm:"not null;default:1"`
	Data                datatypes.JSON     `json:"data" gorm:"type:jsonb"`
	SettingDefinitionID uint               `json:"setting_definition_id" gorm:"index"`
	SettingDefinition   *SettingDefinition `json:"setting_definition,omitempty" gorm:"foreignKey:SettingDefinitionID"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
}

// TableName specifies the table name for Setting
func (Setting) TableName() string {
	return "settings"
}

// SettingRegistration represents a request to register a setting definition
type SettingRegistration struct {
	Domain  string                 `json:"domain" binding:"required"`
	Key     string                 `json:"key" binding:"required"`
	Version int                    `json:"version"`
	Schema  map[string]interface{} `json:"schema" binding:"required"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// SettingRequest represents a request to create/update a setting value
type SettingRequest struct {
	Domain string                 `json:"domain" binding:"required" example:"organization"`
	Key    string                 `json:"key" binding:"required" example:"locale"`
	Data   map[string]interface{} `json:"data" binding:"required"`
}

// SettingResponse represents a setting response
type SettingResponse struct {
	ID        uint                   `json:"id" example:"123"`
	TenantID  uint                   `json:"tenant_id" example:"1"`
	Domain    string                 `json:"domain" example:"organization"`
	Key       string                 `json:"key" example:"locale"`
	Version   int                    `json:"version" example:"1"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at" example:"2025-01-09T10:00:00Z"`
	UpdatedAt time.Time              `json:"updated_at" example:"2025-01-09T10:00:00Z"`
}

// BulkSettingRequest represents a request to set multiple settings
type BulkSettingRequest struct {
	Settings []SettingRequest `json:"settings" binding:"required"`
}

// DomainSettingsRequest represents domain-specific settings request
type DomainSettingsRequest struct {
	Settings map[string]interface{} `json:"settings" binding:"required"`
}

// SettingsResponse represents grouped settings response
type SettingsResponse struct {
	Settings map[string]map[string]interface{} `json:"settings"`
}

// DomainResponse represents available domains
type DomainResponse struct {
	Domains []string `json:"domains"`
}

// ValidationRequest represents settings validation request
type ValidationRequest struct {
	Domain   string                 `json:"domain" binding:"required" example:"company"`
	Settings map[string]interface{} `json:"settings" binding:"required"`
}

// ValidationResponse represents validation results
type ValidationResponse struct {
	Valid  bool     `json:"valid" example:"true"`
	Errors []string `json:"errors,omitempty"`
}

// HealthResponse represents system health status
type HealthResponse struct {
	Status   string `json:"status" example:"ok"`
	Database string `json:"database" example:"connected"`
	Modules  int    `json:"modules" example:"7"`
	Version  string `json:"version" example:"1.0.0"`
}

// ModuleListResponse represents registered modules
type ModuleListResponse struct {
	Modules []string `json:"modules"`
}
