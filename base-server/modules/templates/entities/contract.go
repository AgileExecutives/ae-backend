package entities

import (
	"time"

	baseCore "github.com/ae-base-server/pkg/core"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TemplateContract represents a module-owned template definition
// This defines WHAT can be rendered (purpose, variables, channels) but not HOW
type TemplateContract struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Module      string `gorm:"size:100;not null;index:idx_contract_unique,priority:1" json:"module"`
	TemplateKey string `gorm:"size:100;not null;index:idx_contract_unique,priority:2" json:"template_key"`
	Description string `gorm:"type:text" json:"description,omitempty"`

	// Supported output channels (EMAIL, DOCUMENT)
	SupportedChannels datatypes.JSON `gorm:"type:jsonb;not null" json:"supported_channels"`

	// Variable schema - defines expected variables and their types
	// Stored as JSON Schema or similar structure
	VariableSchema datatypes.JSON `gorm:"type:jsonb" json:"variable_schema,omitempty"`

	// Default sample data for previews
	DefaultSampleData datatypes.JSON `gorm:"type:jsonb" json:"default_sample_data,omitempty"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for the TemplateContract model
func (TemplateContract) TableName() string {
	return "template_contracts"
}

// ContractResponse represents the API response format
type ContractResponse struct {
	ID                uint                   `json:"id"`
	Module            string                 `json:"module"`
	TemplateKey       string                 `json:"template_key"`
	Description       string                 `json:"description,omitempty"`
	SupportedChannels []string               `json:"supported_channels"`
	VariableSchema    map[string]interface{} `json:"variable_schema,omitempty"`
	DefaultSampleData map[string]interface{} `json:"default_sample_data,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// ToResponse converts TemplateContract to ContractResponse
func (c *TemplateContract) ToResponse() ContractResponse {
	resp := ContractResponse{
		ID:          c.ID,
		Module:      c.Module,
		TemplateKey: c.TemplateKey,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	// Parse supported channels
	if len(c.SupportedChannels) > 0 {
		var channels []string
		_ = c.SupportedChannels.Scan(&channels)
		resp.SupportedChannels = channels
	}

	// Parse variable schema
	if len(c.VariableSchema) > 0 {
		var schema map[string]interface{}
		_ = c.VariableSchema.Scan(&schema)
		resp.VariableSchema = schema
	}

	// Parse default sample data
	if len(c.DefaultSampleData) > 0 {
		var sampleData map[string]interface{}
		_ = c.DefaultSampleData.Scan(&sampleData)
		resp.DefaultSampleData = sampleData
	}

	return resp
}

// SupportsChannel checks if a channel is supported by this contract
func (c *TemplateContract) SupportsChannel(channel string) bool {
	var channels []string
	if err := c.SupportedChannels.Scan(&channels); err != nil {
		return false
	}

	for _, ch := range channels {
		if ch == channel {
			return true
		}
	}
	return false
}

// RegisterContractRequest represents a contract registration request
type RegisterContractRequest struct {
	Module            string                 `json:"module" binding:"required"`
	TemplateKey       string                 `json:"template_key" binding:"required"`
	Description       string                 `json:"description"`
	SupportedChannels []string               `json:"supported_channels" binding:"required,min=1"`
	VariableSchema    map[string]interface{} `json:"variable_schema"`
	DefaultSampleData map[string]interface{} `json:"default_sample_data"`
}

// UpdateContractRequest represents a contract update request
type UpdateContractRequest struct {
	Description       *string                 `json:"description,omitempty"`
	SupportedChannels *[]string               `json:"supported_channels,omitempty"`
	VariableSchema    *map[string]interface{} `json:"variable_schema,omitempty"`
	DefaultSampleData *map[string]interface{} `json:"default_sample_data,omitempty"`
}

// NewTemplateContractEntity creates a new entity instance for migrations
func NewTemplateContractEntity() *TemplateContractEntity {
	return &TemplateContractEntity{}
}

// TemplateContractEntity implements core.Entity for TemplateContract model
type TemplateContractEntity struct{}

func (e *TemplateContractEntity) TableName() string {
	return "template_contracts"
}

func (e *TemplateContractEntity) GetModel() interface{} {
	return &TemplateContract{}
}

func (e *TemplateContractEntity) GetMigrations() []baseCore.Migration {
	return []baseCore.Migration{}
}

// ContractRegistration represents the registration of a contract by a module
type ContractRegistration struct {
	ModuleName  string                 `json:"module_name"`
	TemplateKey string                 `json:"template_key"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema"`
	Version     string                 `json:"version"`
}
