package entities

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Channel represents template output channel
type Channel string

const (
	ChannelEmail    Channel = "EMAIL"
	ChannelDocument Channel = "DOCUMENT"
)

// Template represents a document template (invoice, contract, etc.)
type Template struct {
	ID             uint  `gorm:"primaryKey" json:"id"`
	TenantID       uint  `gorm:"not null;index:idx_tenant_template" json:"tenant_id"`
	OrganizationID *uint `gorm:"index:idx_org_template" json:"organization_id,omitempty"` // NULL = system default

	// Contract binding - links to TemplateContract
	Module      string  `gorm:"size:100;index:idx_template_contract,priority:1" json:"module,omitempty"`
	TemplateKey string  `gorm:"size:100;index:idx_template_contract,priority:2" json:"template_key,omitempty"`
	Channel     Channel `gorm:"size:20;index:idx_template_contract,priority:3" json:"channel,omitempty"`

	// Channel-specific fields
	Subject *string `gorm:"type:text" json:"subject,omitempty"` // Required for EMAIL, null for DOCUMENT

	// Legacy field (deprecated, keep for backward compatibility)
	TemplateType string `gorm:"size:50;index" json:"template_type"` // "invoice", "contract", "report"

	Name        string `gorm:"size:255;not null" json:"name"`
	Description string `gorm:"type:text" json:"description,omitempty"`

	// Version control
	Version   int  `gorm:"not null;default:1" json:"version"`
	IsActive  bool `gorm:"default:true;index" json:"is_active"`
	IsDefault bool `gorm:"default:false" json:"is_default"`

	// Storage
	StorageKey string `gorm:"size:500;not null;uniqueIndex" json:"storage_key"`
	PreviewKey string `gorm:"size:500" json:"preview_key,omitempty"` // Preview image storage key

	// Template metadata
	Variables  datatypes.JSON `gorm:"type:jsonb" json:"variables,omitempty"`   // Expected template variables
	SampleData datatypes.JSON `gorm:"type:jsonb" json:"sample_data,omitempty"` // Sample data for preview

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for the Template model
func (Template) TableName() string {
	return "templates"
}

// TemplateResponse represents the API response format
type TemplateResponse struct {
	ID             uint                   `json:"id"`
	TenantID       uint                   `json:"tenant_id"`
	OrganizationID *uint                  `json:"organization_id,omitempty"`
	TemplateType   string                 `json:"template_type"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Version        int                    `json:"version"`
	IsActive       bool                   `json:"is_active"`
	IsDefault      bool                   `json:"is_default"`
	PreviewURL     string                 `json:"preview_url,omitempty"`
	Variables      []string               `json:"variables"`
	SampleData     map[string]interface{} `json:"sample_data"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// ToResponse converts Template to TemplateResponse
func (t *Template) ToResponse() TemplateResponse {
	resp := TemplateResponse{
		ID:             t.ID,
		TenantID:       t.TenantID,
		OrganizationID: t.OrganizationID,
		TemplateType:   t.TemplateType,
		Name:           t.Name,
		Description:    t.Description,
		Version:        t.Version,
		IsActive:       t.IsActive,
		IsDefault:      t.IsDefault,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}

	// Always initialize Variables and SampleData slices (even if empty)
	resp.Variables = []string{}
	resp.SampleData = make(map[string]interface{})

	// Parse variables if present
	if len(t.Variables) > 0 {
		var variables []string
		if err := json.Unmarshal(t.Variables, &variables); err == nil {
			resp.Variables = variables
		}
	}

	// Parse sample data if present
	if len(t.SampleData) > 0 {
		var sampleData map[string]interface{}
		if err := json.Unmarshal(t.SampleData, &sampleData); err == nil {
			resp.SampleData = sampleData
		}
	}

	return resp
}

// UploadTemplateRequest represents a template upload request
type UploadTemplateRequest struct {
	OrganizationID *uint    `json:"organization_id,omitempty"`
	TemplateType   string   `json:"template_type" binding:"required"`
	Name           string   `json:"name" binding:"required"`
	Description    string   `json:"description,omitempty"`
	Content        []byte   `json:"-"` // HTML content
	Version        int      `json:"version"`
	IsDefault      bool     `json:"is_default"`
	Variables      []string `json:"variables,omitempty"`
}

// TemplateListResponse represents paginated template list
type TemplateListResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Data    []TemplateResponse `json:"data"`
	Page    int                `json:"page"`
	Limit   int                `json:"limit"`
	Total   int64              `json:"total"`
}

// TemplateAPIResponse represents single template response
type TemplateAPIResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    TemplateResponse `json:"data"`
}

// DuplicateTemplateRequest represents a request to duplicate a template
type DuplicateTemplateRequest struct {
	Name string `json:"name" binding:"required" example:"Copy of Welcome Email Template"`
}

// RenderTemplateRequest represents a request to render a template with data
type RenderTemplateRequest struct {
	FirstName        string `json:"FirstName,omitempty" example:"Alice"`
	LastName         string `json:"LastName,omitempty" example:"Johnson"`
	OrganizationName string `json:"OrganizationName,omitempty" example:"Tech Innovators Inc"`
	Email            string `json:"Email,omitempty" example:"alice.johnson@techinnovators.com"`
}
