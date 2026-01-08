package models

import (
	"time"

	"gorm.io/gorm"
)

// Tenant represents a tenant in the multi-tenant system
// A customer can have multiple tenants
type Tenant struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty" swaggerignore:"true"`
	CustomerID uint           `gorm:"not null;index:idx_tenant_customer" json:"customer_id"`
	Name       string         `gorm:"not null" json:"name" binding:"required"`
	Slug       string         `gorm:"not null;uniqueIndex:idx_tenant_slug" json:"slug" binding:"required"`
}

// TableName specifies the table name for Tenant
func (Tenant) TableName() string {
	return "tenants"
}

// TenantResponse represents the API response structure for Tenant
type TenantResponse struct {
	ID         uint      `json:"id"`
	CustomerID uint      `json:"customer_id"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	CreatedAt  time.Time `json:"created_at"`
}

// ToResponse converts Tenant to TenantResponse
func (t *Tenant) ToResponse() TenantResponse {
	return TenantResponse{
		ID:         t.ID,
		CustomerID: t.CustomerID,
		Name:       t.Name,
		Slug:       t.Slug,
		CreatedAt:  t.CreatedAt,
	}
}

// TenantCreateRequest represents the request structure for creating a tenant
type TenantCreateRequest struct {
	CustomerID uint   `json:"customer_id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Slug       string `json:"slug" binding:"required"`
}

// TenantUpdateRequest represents the request structure for updating a tenant
type TenantUpdateRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}
