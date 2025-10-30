package models

import (
	"time"

	"gorm.io/gorm"
)

// CostProvider represents a cost provider entity for clients
type CostProvider struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	TenantID      uint           `gorm:"not null;index" json:"tenant_id"`
	Organization  string         `gorm:"size:255;not null" json:"organization" binding:"required" example:"Health Insurance Corp"`
	Department    string         `gorm:"size:255" json:"department,omitempty" example:"Mental Health Division"`
	ContactName   string         `gorm:"size:255" json:"contact_name,omitempty" example:"Jane Smith"`
	StreetAddress string         `gorm:"size:500" json:"street_address,omitempty" example:"456 Insurance Blvd"`
	Zip           string         `gorm:"size:20" json:"zip,omitempty" example:"12345"`
	City          string         `gorm:"size:100" json:"city,omitempty" example:"New York"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// CreateCostProviderRequest represents the request payload for creating a cost provider
type CreateCostProviderRequest struct {
	Organization  string `json:"organization" binding:"required" example:"Health Insurance Corp"`
	Department    string `json:"department,omitempty" example:"Mental Health Division"`
	ContactName   string `json:"contact_name,omitempty" example:"Jane Smith"`
	StreetAddress string `json:"street_address,omitempty" example:"456 Insurance Blvd"`
	Zip           string `json:"zip,omitempty" example:"12345"`
	City          string `json:"city,omitempty" example:"New York"`
}

// UpdateCostProviderRequest represents the request payload for updating a cost provider
type UpdateCostProviderRequest struct {
	Organization  *string `json:"organization,omitempty" example:"Health Insurance Corp"`
	Department    *string `json:"department,omitempty" example:"Mental Health Division"`
	ContactName   *string `json:"contact_name,omitempty" example:"Jane Smith"`
	StreetAddress *string `json:"street_address,omitempty" example:"456 Insurance Blvd"`
	Zip           *string `json:"zip,omitempty" example:"12345"`
	City          *string `json:"city,omitempty" example:"New York"`
}

// CostProviderResponse represents the response format for cost provider data
type CostProviderResponse struct {
	ID            uint      `json:"id"`
	TenantID      uint      `json:"tenant_id"`
	Organization  string    `json:"organization"`
	Department    string    `json:"department,omitempty"`
	ContactName   string    `json:"contact_name,omitempty"`
	StreetAddress string    `json:"street_address,omitempty"`
	Zip           string    `json:"zip,omitempty"`
	City          string    `json:"city,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ToResponse converts a CostProvider model to CostProviderResponse
func (cp *CostProvider) ToResponse() CostProviderResponse {
	return CostProviderResponse{
		ID:            cp.ID,
		TenantID:      cp.TenantID,
		Organization:  cp.Organization,
		Department:    cp.Department,
		ContactName:   cp.ContactName,
		StreetAddress: cp.StreetAddress,
		Zip:           cp.Zip,
		City:          cp.City,
		CreatedAt:     cp.CreatedAt,
		UpdatedAt:     cp.UpdatedAt,
	}
}
