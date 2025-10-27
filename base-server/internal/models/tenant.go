package models

import (
	"time"

	"gorm.io/gorm"
)

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name      string         `gorm:"not null;unique" json:"name" binding:"required"`
	Slug      string         `gorm:"not null;unique" json:"slug" binding:"required"`
}

// TableName specifies the table name for Tenant
func (Tenant) TableName() string {
	return "tenants"
}

// TenantResponse represents the API response structure for Tenant
type TenantResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts Tenant to TenantResponse
func (t *Tenant) ToResponse() TenantResponse {
	return TenantResponse{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt,
	}
}

// TenantCreateRequest represents the request structure for creating a tenant
type TenantCreateRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

// TenantUpdateRequest represents the request structure for updating a tenant
type TenantUpdateRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}
