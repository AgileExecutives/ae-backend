package entities

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Organization represents an organization entity
type Organization struct {
	ID                       uint           `gorm:"primaryKey" json:"id"`
	TenantID                 uint           `gorm:"not null;index:idx_organization_tenant" json:"tenant_id"`
	UserID                   uint           `gorm:"not null;index:idx_organization_user" json:"user_id"`
	Name                     string         `gorm:"size:255;not null" json:"name"`
	OwnerName                string         `gorm:"size:255" json:"owner_name"`
	OwnerTitle               string         `gorm:"size:100" json:"owner_title"`
	StreetAddress            string         `gorm:"size:255" json:"street_address"`
	Zip                      string         `gorm:"size:20" json:"zip"`
	City                     string         `gorm:"size:100" json:"city"`
	Email                    string         `gorm:"size:255" json:"email"`
	Phone                    string         `gorm:"size:50" json:"phone"`
	TaxID                    string         `gorm:"size:100" json:"tax_id"`
	TaxRate                  *float64       `gorm:"type:decimal(5,2)" json:"tax_rate"`
	TaxUstID                 string         `gorm:"size:100" json:"tax_ustid"`
	UnitPrice                *float64       `gorm:"type:decimal(10,2)" json:"unit_price"`
	BankAccountOwner         string         `gorm:"size:255" json:"bankaccount_owner"`
	BankAccountBank          string         `gorm:"size:255" json:"bankaccount_bank"`
	BankAccountBIC           string         `gorm:"size:50" json:"bankaccount_bic"`
	BankAccountIBAN          string         `gorm:"size:100" json:"bankaccount_iban"`
	AdditionalPaymentMethods datatypes.JSON `gorm:"type:jsonb" json:"additional_payment_methods"`
	InvoiceContent           datatypes.JSON `gorm:"type:jsonb" json:"invoice_content"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for the Organization model
func (Organization) TableName() string {
	return "organizations"
}

// CreateOrganizationRequest represents the request payload for creating an organization
type CreateOrganizationRequest struct {
	Name                     string         `json:"name" binding:"required" example:"Acme Corporation"`
	OwnerName                string         `json:"owner_name,omitempty" example:"John Doe"`
	OwnerTitle               string         `json:"owner_title,omitempty" example:"CEO"`
	StreetAddress            string         `json:"street_address,omitempty" example:"123 Business St"`
	Zip                      string         `json:"zip,omitempty" example:"12345"`
	City                     string         `json:"city,omitempty" example:"New York"`
	Email                    string         `json:"email,omitempty" example:"info@acme.com"`
	Phone                    string         `json:"phone,omitempty" example:"+1-555-0123"`
	TaxID                    string         `json:"tax_id,omitempty" example:"TAX123456"`
	TaxRate                  *float64       `json:"tax_rate,omitempty" example:"19.00"`
	TaxUstID                 string         `json:"tax_ustid,omitempty" example:"DE123456789"`
	UnitPrice                *float64       `json:"unit_price,omitempty" example:"150.00"`
	BankAccountOwner         string         `json:"bankaccount_owner,omitempty" example:"Acme Corporation"`
	BankAccountBank          string         `json:"bankaccount_bank,omitempty" example:"Deutsche Bank"`
	BankAccountBIC           string         `json:"bankaccount_bic,omitempty" example:"DEUTDEFF"`
	BankAccountIBAN          string         `json:"bankaccount_iban,omitempty" example:"DE89370400440532013000"`
	AdditionalPaymentMethods datatypes.JSON `json:"additional_payment_methods,omitempty" swaggertype:"object"`
	InvoiceContent           datatypes.JSON `json:"invoice_content,omitempty" swaggertype:"object"`
}

// UpdateOrganizationRequest represents the request payload for updating an organization
type UpdateOrganizationRequest struct {
	Name                     *string        `json:"name,omitempty" example:"Acme Corporation"`
	OwnerName                *string        `json:"owner_name,omitempty" example:"John Doe"`
	OwnerTitle               *string        `json:"owner_title,omitempty" example:"CEO"`
	StreetAddress            *string        `json:"street_address,omitempty" example:"123 Business St"`
	Zip                      *string        `json:"zip,omitempty" example:"12345"`
	City                     *string        `json:"city,omitempty" example:"New York"`
	Email                    *string        `json:"email,omitempty" example:"info@acme.com"`
	Phone                    *string        `json:"phone,omitempty" example:"+1-555-0123"`
	TaxID                    *string        `json:"tax_id,omitempty" example:"TAX123456"`
	TaxRate                  *float64       `json:"tax_rate,omitempty" example:"19.00"`
	TaxUstID                 *string        `json:"tax_ustid,omitempty" example:"DE123456789"`
	UnitPrice                *float64       `json:"unit_price,omitempty" example:"150.00"`
	BankAccountOwner         *string        `json:"bankaccount_owner,omitempty" example:"Acme Corporation"`
	BankAccountBank          *string        `json:"bankaccount_bank,omitempty" example:"Deutsche Bank"`
	BankAccountBIC           *string        `json:"bankaccount_bic,omitempty" example:"DEUTDEFF"`
	BankAccountIBAN          *string        `json:"bankaccount_iban,omitempty" example:"DE89370400440532013000"`
	AdditionalPaymentMethods datatypes.JSON `json:"additional_payment_methods,omitempty" swaggertype:"object"`
	InvoiceContent           datatypes.JSON `json:"invoice_content,omitempty" swaggertype:"object"`
}

// OrganizationResponse represents the response format for organization data
type OrganizationResponse struct {
	ID                       uint           `json:"id"`
	TenantID                 uint           `json:"tenant_id"`
	UserID                   uint           `json:"user_id"`
	Name                     string         `json:"name"`
	OwnerName                string         `json:"owner_name"`
	OwnerTitle               string         `json:"owner_title"`
	StreetAddress            string         `json:"street_address"`
	Zip                      string         `json:"zip"`
	City                     string         `json:"city"`
	Email                    string         `json:"email"`
	Phone                    string         `json:"phone"`
	TaxID                    string         `json:"tax_id"`
	TaxRate                  *float64       `json:"tax_rate"`
	TaxUstID                 string         `json:"tax_ustid"`
	UnitPrice                *float64       `json:"unit_price"`
	BankAccountOwner         string         `json:"bankaccount_owner"`
	BankAccountBank          string         `json:"bankaccount_bank"`
	BankAccountBIC           string         `json:"bankaccount_bic"`
	BankAccountIBAN          string         `json:"bankaccount_iban"`
	AdditionalPaymentMethods datatypes.JSON `json:"additional_payment_methods"`
	InvoiceContent           datatypes.JSON `json:"invoice_content"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
}

// OrganizationAPIResponse represents the API response for a single organization
type OrganizationAPIResponse struct {
	Success bool                 `json:"success" example:"true"`
	Message string               `json:"message" example:"Organization retrieved successfully"`
	Data    OrganizationResponse `json:"data"`
}

// OrganizationListAPIResponse represents the API response for organization list
type OrganizationListAPIResponse struct {
	Success bool                   `json:"success" example:"true"`
	Message string                 `json:"message" example:"Organizations retrieved successfully"`
	Data    []OrganizationResponse `json:"data"`
	Page    int                    `json:"page" example:"1"`
	Limit   int                    `json:"limit" example:"10"`
	Total   int                    `json:"total" example:"100"`
}

// OrganizationDeleteResponse represents the API response for organization deletion
type OrganizationDeleteResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Organization deleted successfully"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Error message"`
	Error   string `json:"error,omitempty" example:"Detailed error information"`
}

// ToResponse converts an Organization to OrganizationResponse
func (o *Organization) ToResponse() OrganizationResponse {
	return OrganizationResponse{
		ID:                       o.ID,
		TenantID:                 o.TenantID,
		UserID:                   o.UserID,
		Name:                     o.Name,
		OwnerName:                o.OwnerName,
		OwnerTitle:               o.OwnerTitle,
		StreetAddress:            o.StreetAddress,
		Zip:                      o.Zip,
		City:                     o.City,
		Email:                    o.Email,
		Phone:                    o.Phone,
		TaxID:                    o.TaxID,
		TaxRate:                  o.TaxRate,
		TaxUstID:                 o.TaxUstID,
		UnitPrice:                o.UnitPrice,
		BankAccountOwner:         o.BankAccountOwner,
		BankAccountBank:          o.BankAccountBank,
		BankAccountBIC:           o.BankAccountBIC,
		BankAccountIBAN:          o.BankAccountIBAN,
		AdditionalPaymentMethods: o.AdditionalPaymentMethods,
		InvoiceContent:           o.InvoiceContent,
		CreatedAt:                o.CreatedAt,
		UpdatedAt:                o.UpdatedAt,
	}
}
