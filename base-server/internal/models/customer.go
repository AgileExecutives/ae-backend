package models

import (
	"time"

	"gorm.io/gorm"
)

// Customer represents a billing customer in the system
type Customer struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name      string         `gorm:"not null" json:"name" binding:"required"`
	Email     string         `gorm:"not null" json:"email" binding:"required,email"`
	Phone     string         `json:"phone"`
	Street    string         `json:"street"`
	Zip       string         `json:"zip"`
	City      string         `json:"city"`
	Country   string         `json:"country"`
	TaxID     string         `json:"tax_id"`
	VAT       string         `json:"vat"`
	PlanID    uint           `gorm:"not null" json:"plan_id" binding:"required"`
	// Plan          Plan           `gorm:"foreignKey:PlanID" json:"plan,omitempty"` // Disabled for migration
	TenantID uint `gorm:"not null" json:"tenant_id"`
	// Tenant        Tenant         `gorm:"foreignKey:TenantID" json:"tenant,omitempty"` // Disabled for migration
	Status        string `gorm:"default:'active'" json:"status"`
	PaymentMethod string `json:"payment_method"`
	Active        bool   `gorm:"default:true" json:"active"`
}

// TableName specifies the table name for Customer
func (Customer) TableName() string {
	return "customers"
}

// CustomerResponse represents the API response structure for Customer
type CustomerResponse struct {
	ID            uint           `json:"id"`
	Name          string         `json:"name"`
	Email         string         `json:"email"`
	Phone         string         `json:"phone"`
	Street        string         `json:"street"`
	Zip           string         `json:"zip"`
	City          string         `json:"city"`
	Country       string         `json:"country"`
	TaxID         string         `json:"tax_id"`
	VAT           string         `json:"vat"`
	PlanID        uint           `json:"plan_id"`
	Plan          PlanResponse   `json:"plan,omitempty"`
	TenantID      uint           `json:"tenant_id"`
	Tenant        TenantResponse `json:"tenant,omitempty"`
	Status        string         `json:"status"`
	PaymentMethod string         `json:"payment_method"`
	Active        bool           `json:"active"`
	CreatedAt     time.Time      `json:"created_at"`
}

// ToResponse converts Customer to CustomerResponse
func (c *Customer) ToResponse() CustomerResponse {
	response := CustomerResponse{
		ID:            c.ID,
		Name:          c.Name,
		Email:         c.Email,
		Phone:         c.Phone,
		Street:        c.Street,
		Zip:           c.Zip,
		City:          c.City,
		Country:       c.Country,
		TaxID:         c.TaxID,
		VAT:           c.VAT,
		PlanID:        c.PlanID,
		TenantID:      c.TenantID,
		Status:        c.Status,
		PaymentMethod: c.PaymentMethod,
		Active:        c.Active,
		CreatedAt:     c.CreatedAt,
	}

	// Temporarily disabled for migration
	// if c.Plan.ID != 0 {
	// 	response.Plan = c.Plan.ToResponse()
	// }

	// if c.Tenant.ID != 0 {
	// 	response.Tenant = c.Tenant.ToResponse()
	// }

	return response
}

// CustomerCreateRequest represents the request structure for creating a customer
type CustomerCreateRequest struct {
	Name          string `json:"name" binding:"required"`
	Email         string `json:"email" binding:"required,email"`
	Phone         string `json:"phone"`
	Street        string `json:"street"`
	Zip           string `json:"zip"`
	City          string `json:"city"`
	Country       string `json:"country"`
	TaxID         string `json:"tax_id"`
	VAT           string `json:"vat"`
	PlanID        uint   `json:"plan_id" binding:"required"`
	TenantID      uint   `json:"tenant_id" binding:"required"`
	PaymentMethod string `json:"payment_method"`
}

// CustomerUpdateRequest represents the request structure for updating a customer
type CustomerUpdateRequest struct {
	Name          string `json:"name"`
	Email         string `json:"email" binding:"omitempty,email"`
	Phone         string `json:"phone"`
	Street        string `json:"street"`
	Zip           string `json:"zip"`
	City          string `json:"city"`
	Country       string `json:"country"`
	TaxID         string `json:"tax_id"`
	VAT           string `json:"vat"`
	PlanID        *uint  `json:"plan_id"`
	Status        string `json:"status"`
	PaymentMethod string `json:"payment_method"`
	Active        *bool  `json:"active"`
}
