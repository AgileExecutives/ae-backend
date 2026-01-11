package entities

import "time"

// Setting represents a single configuration setting
type Setting struct {
	ID             uint      `json:"id" gorm:"primarykey"`
	TenantID       uint      `json:"tenant_id" gorm:"not null;index"`
	OrganizationID string    `json:"organization_id" gorm:"not null;index"`
	Domain         string    `json:"domain" gorm:"not null;index"`
	Key            string    `json:"key" gorm:"not null;index"`
	Value          string    `json:"value" gorm:"type:text"`
	Type           string    `json:"type" gorm:"not null"` // string, int, bool, float, json
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SettingRequest represents a request to create/update a setting
type SettingRequest struct {
	Domain string      `json:"domain" binding:"required" example:"company"`
	Key    string      `json:"key" binding:"required" example:"company_name"`
	Type   string      `json:"type" binding:"required" example:"string"`
	Value  interface{} `json:"value" binding:"required"`
}

// SettingResponse represents a setting response
type SettingResponse struct {
	ID             uint      `json:"id" example:"123"`
	TenantID       uint      `json:"tenant_id" example:"1"`
	OrganizationID string    `json:"organization_id" example:"org-123"`
	Domain         string    `json:"domain" example:"company"`
	Key            string    `json:"key" example:"company_name"`
	Value          string    `json:"value" example:"My Company"`
	Type           string    `json:"type" example:"string"`
	CreatedAt      time.Time `json:"created_at" example:"2025-01-09T10:00:00Z"`
	UpdatedAt      time.Time `json:"updated_at" example:"2025-01-09T10:00:00Z"`
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
