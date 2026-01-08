package entities

import (
	"time"

	"gorm.io/gorm"
)

// CostProvider represents a cost provider entity for clients
type CostProvider struct {
	ID            uint   `gorm:"primarykey" json:"id"`
	TenantID      uint   `gorm:"not null;index" json:"tenant_id"`
	Organization  string `gorm:"size:255;not null" json:"organization" binding:"required" example:"Health Insurance Corp"`
	Department    string `gorm:"size:255" json:"department,omitempty" example:"Mental Health Division"`
	ContactName   string `gorm:"size:255" json:"contact_name,omitempty" example:"Jane Smith"`
	StreetAddress string `gorm:"size:500" json:"street_address,omitempty" example:"456 Insurance Blvd"`
	Zip           string `gorm:"size:20" json:"zip,omitempty" example:"12345"`
	City          string `gorm:"size:100" json:"city,omitempty" example:"New York"`

	// Government customer fields (for XRechnung/E-Rechnung)
	IsGovernmentCustomer bool   `gorm:"not null;default:false" json:"is_government_customer"`
	LeitwegID            string `gorm:"size:100" json:"leitweg_id,omitempty"` // German government routing ID
	AuthorityName        string `gorm:"size:255" json:"authority_name,omitempty"`
	ReferenceNumber      string `gorm:"size:100" json:"reference_number,omitempty"` // Cost center or reference

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// CreateCostProviderRequest represents the request payload for creating a cost provider
type CreateCostProviderRequest struct {
	Organization         string `json:"organization" binding:"required" example:"Health Insurance Corp"`
	Department           string `json:"department,omitempty" example:"Mental Health Division"`
	ContactName          string `json:"contact_name,omitempty" example:"Jane Smith"`
	StreetAddress        string `json:"street_address,omitempty" example:"456 Insurance Blvd"`
	Zip                  string `json:"zip,omitempty" example:"12345"`
	City                 string `json:"city,omitempty" example:"New York"`
	IsGovernmentCustomer bool   `json:"is_government_customer,omitempty" example:"false"`
	LeitwegID            string `json:"leitweg_id,omitempty" example:"99-12345-67"`
	AuthorityName        string `json:"authority_name,omitempty" example:"Jugendamt Berlin"`
	ReferenceNumber      string `json:"reference_number,omitempty" example:"KST-2024-001"`
}

// UpdateCostProviderRequest represents the request payload for updating a cost provider
type UpdateCostProviderRequest struct {
	Organization         *string `json:"organization,omitempty" example:"Health Insurance Corp"`
	Department           *string `json:"department,omitempty" example:"Mental Health Division"`
	ContactName          *string `json:"contact_name,omitempty" example:"Jane Smith"`
	StreetAddress        *string `json:"street_address,omitempty" example:"456 Insurance Blvd"`
	Zip                  *string `json:"zip,omitempty" example:"12345"`
	City                 *string `json:"city,omitempty" example:"New York"`
	IsGovernmentCustomer *bool   `json:"is_government_customer,omitempty" example:"false"`
	LeitwegID            *string `json:"leitweg_id,omitempty" example:"99-12345-67"`
	AuthorityName        *string `json:"authority_name,omitempty" example:"Jugendamt Berlin"`
	ReferenceNumber      *string `json:"reference_number,omitempty" example:"KST-2024-001"`
}

// CostProviderResponse represents the response format for cost provider data
type CostProviderResponse struct {
	ID                   uint      `json:"id"`
	TenantID             uint      `json:"tenant_id"`
	Organization         string    `json:"organization"`
	Department           string    `json:"department,omitempty"`
	ContactName          string    `json:"contact_name,omitempty"`
	StreetAddress        string    `json:"street_address,omitempty"`
	Zip                  string    `json:"zip,omitempty"`
	City                 string    `json:"city,omitempty"`
	IsGovernmentCustomer bool      `json:"is_government_customer"`
	LeitwegID            string    `json:"leitweg_id,omitempty"`
	AuthorityName        string    `json:"authority_name,omitempty"`
	ReferenceNumber      string    `json:"reference_number,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// ToResponse converts a CostProvider model to CostProviderResponse
func (cp *CostProvider) ToResponse() CostProviderResponse {
	return CostProviderResponse{
		ID:                   cp.ID,
		TenantID:             cp.TenantID,
		Organization:         cp.Organization,
		Department:           cp.Department,
		ContactName:          cp.ContactName,
		StreetAddress:        cp.StreetAddress,
		Zip:                  cp.Zip,
		City:                 cp.City,
		IsGovernmentCustomer: cp.IsGovernmentCustomer,
		LeitwegID:            cp.LeitwegID,
		AuthorityName:        cp.AuthorityName,
		ReferenceNumber:      cp.ReferenceNumber,
		CreatedAt:            cp.CreatedAt,
		UpdatedAt:            cp.UpdatedAt,
	}
}
