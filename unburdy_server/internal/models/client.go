package models

import (
	"time"

	baseAPI "github.com/ae-saas-basic/ae-saas-basic/api"
	"gorm.io/gorm"
)

// Client represents a client entity extending the base system with multi-tenant support
// Uses ae-saas-basic models for User and Tenant relationships
type Client struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	FirstName     string         `gorm:"size:100;not null" json:"first_name" binding:"required" example:"John"`
	LastName      string         `gorm:"size:100;not null" json:"last_name" binding:"required" example:"Doe"`
	DateOfBirth   *time.Time     `gorm:"type:date" json:"date_of_birth,omitempty" example:"1990-01-15"`
	TenantID      uint           `gorm:"default:1" json:"tenant_id"`
	Tenant        baseAPI.Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	CreatedBy     uint           `gorm:"default:1" json:"created_by"`
	CreatedByUser baseAPI.User   `gorm:"foreignKey:CreatedBy" json:"created_by_user,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// CreateClientRequest represents the request payload for creating a client
type CreateClientRequest struct {
	FirstName   string     `json:"first_name" binding:"required" example:"John"`
	LastName    string     `json:"last_name" binding:"required" example:"Doe"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty" example:"1990-01-15"`
	// TenantID and CreatedBy will be set from the authenticated user context
}

// UpdateClientRequest represents the request payload for updating a client
type UpdateClientRequest struct {
	FirstName   *string    `json:"first_name,omitempty" example:"John"`
	LastName    *string    `json:"last_name,omitempty" example:"Doe"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty" example:"1990-01-15"`
}

// ClientResponse represents the response format for client data
type ClientResponse struct {
	ID            uint                   `json:"id"`
	FirstName     string                 `json:"first_name"`
	LastName      string                 `json:"last_name"`
	DateOfBirth   *time.Time             `json:"date_of_birth,omitempty"`
	TenantID      uint                   `json:"tenant_id"`
	Tenant        baseAPI.TenantResponse `json:"tenant,omitempty" swaggerignore:"true"`
	CreatedBy     uint                   `json:"created_by"`
	CreatedByUser baseAPI.UserResponse   `json:"created_by_user,omitempty" swaggerignore:"true"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ToResponse converts a Client model to ClientResponse
func (c *Client) ToResponse() ClientResponse {
	response := ClientResponse{
		ID:          c.ID,
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		DateOfBirth: c.DateOfBirth,
		TenantID:    c.TenantID,
		CreatedBy:   c.CreatedBy,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	// Include tenant information if loaded
	if c.Tenant.ID != 0 {
		response.Tenant = c.Tenant.ToResponse()
	}

	// Include creator information if loaded
	if c.CreatedByUser.ID != 0 {
		response.CreatedByUser = c.CreatedByUser.ToResponse()
	}

	return response
}
