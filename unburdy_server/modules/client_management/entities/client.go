package entities

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

// NullableDate is a custom type that handles empty strings by converting them to nil
type NullableDate struct {
	*time.Time
}

// UnmarshalJSON implements custom JSON unmarshaling for date fields
// Empty strings are converted to nil, otherwise it attempts to parse as a date
func (nd *NullableDate) UnmarshalJSON(data []byte) error {
	// Remove quotes from the JSON string
	str := strings.Trim(string(data), `"`)

	// If empty string or "null", set to nil
	if str == "" || str == "null" {
		nd.Time = nil
		return nil
	}

	// Try to parse as RFC3339 date format (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SSZ)
	if t, err := time.Parse("2006-01-02", str); err == nil {
		nd.Time = &t
		return nil
	}

	// Try to parse as full RFC3339 format
	if t, err := time.Parse(time.RFC3339, str); err == nil {
		nd.Time = &t
		return nil
	}

	// If neither format works, set to nil
	nd.Time = nil
	return nil
}

// MarshalJSON implements custom JSON marshaling for date fields
func (nd NullableDate) MarshalJSON() ([]byte, error) {
	if nd.Time == nil {
		return []byte("null"), nil
	}
	return json.Marshal(nd.Time.Format("2006-01-02"))
}

// Client represents a client entity with comprehensive information
type Client struct {
	ID                   uint           `gorm:"primarykey" json:"id"`
	TenantID             uint           `gorm:"not null;index" json:"tenant_id"`
	CostProviderID       *uint          `gorm:"index" json:"cost_provider_id,omitempty"`
	CostProvider         *CostProvider  `gorm:"foreignKey:CostProviderID" json:"cost_provider,omitempty"`
	FirstName            string         `gorm:"size:100;not null" json:"first_name" binding:"required" example:"John"`
	LastName             string         `gorm:"size:100;not null" json:"last_name" binding:"required" example:"Doe"`
	DateOfBirth          *time.Time     `gorm:"type:date" json:"date_of_birth,omitempty" example:"1990-01-15"`
	Gender               string         `gorm:"size:20;default:'undisclosed'" json:"gender" example:"male"` // male, female, undisclosed
	PrimaryLanguage      string         `gorm:"size:50" json:"primary_language,omitempty" example:"English"`
	ContactFirstName     string         `gorm:"size:100" json:"contact_first_name,omitempty" example:"Jane"`
	ContactLastName      string         `gorm:"size:100" json:"contact_last_name,omitempty" example:"Smith"`
	ContactEmail         string         `gorm:"size:255" json:"contact_email,omitempty" example:"jane.smith@example.com"`
	ContactPhone         string         `gorm:"size:50" json:"contact_phone,omitempty" example:"+1234567890"`
	AlternativeFirstName string         `gorm:"size:100" json:"alternative_first_name,omitempty" example:"Johnny"`
	AlternativeLastName  string         `gorm:"size:100" json:"alternative_last_name,omitempty" example:"D"`
	AlternativePhone     string         `gorm:"size:50" json:"alternative_phone,omitempty" example:"+0987654321"`
	AlternativeEmail     string         `gorm:"size:255" json:"alternative_email,omitempty" example:"johnny.d@example.com"`
	StreetAddress        string         `gorm:"size:255" json:"street_address,omitempty" example:"123 Main Street"`
	Zip                  string         `gorm:"size:20" json:"zip,omitempty" example:"12345"`
	City                 string         `gorm:"size:100" json:"city,omitempty" example:"New York"`
	Email                string         `gorm:"size:255" json:"email,omitempty" example:"john.doe@example.com"`
	Phone                string         `gorm:"size:50" json:"phone,omitempty" example:"+1234567890"`
	InvoicedIndividually bool           `gorm:"default:false" json:"invoiced_individually" example:"false"`
	TherapyTitle         string         `gorm:"size:255" json:"therapy_title,omitempty" example:"Cognitive Behavioral Therapy"`
	ProviderApprovalCode string         `gorm:"size:100" json:"provider_approval_code,omitempty" example:"PROV123456"`
	ProviderApprovalDate *time.Time     `gorm:"type:date" json:"provider_approval_date,omitempty" example:"2025-01-15"`
	UnitPrice            *float64       `gorm:"type:decimal(10,2)" json:"unit_price,omitempty" example:"150.00"`
	Status               string         `gorm:"size:20;default:'waiting'" json:"status" example:"active"` // waiting, active, archived
	AdmissionDate        *time.Time     `gorm:"type:date" json:"admission_date,omitempty" example:"2025-01-01"`
	ReferralSource       string         `gorm:"size:255" json:"referral_source,omitempty" example:"Doctor Smith"`
	Notes                string         `gorm:"type:text" json:"notes,omitempty" example:"Additional notes about the client"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

// CreateClientRequest represents the request payload for creating a client
type CreateClientRequest struct {
	CostProviderID       *uint        `json:"cost_provider_id,omitempty" example:"1"`
	FirstName            string       `json:"first_name" binding:"required" example:"John"`
	LastName             string       `json:"last_name" binding:"required" example:"Doe"`
	DateOfBirth          NullableDate `json:"date_of_birth,omitempty" example:"1990-01-15"`
	Gender               string       `json:"gender,omitempty" example:"male"`
	PrimaryLanguage      string       `json:"primary_language,omitempty" example:"English"`
	ContactFirstName     string       `json:"contact_first_name,omitempty" example:"Jane"`
	ContactLastName      string       `json:"contact_last_name,omitempty" example:"Smith"`
	ContactEmail         string       `json:"contact_email,omitempty" example:"jane.smith@example.com"`
	ContactPhone         string       `json:"contact_phone,omitempty" example:"+1234567890"`
	AlternativeFirstName string       `json:"alternative_first_name,omitempty" example:"Johnny"`
	AlternativeLastName  string       `json:"alternative_last_name,omitempty" example:"D"`
	AlternativePhone     string       `json:"alternative_phone,omitempty" example:"+0987654321"`
	AlternativeEmail     string       `json:"alternative_email,omitempty" example:"johnny.d@example.com"`
	StreetAddress        string       `json:"street_address,omitempty" example:"123 Main Street"`
	Zip                  string       `json:"zip,omitempty" example:"12345"`
	City                 string       `json:"city,omitempty" example:"New York"`
	Email                string       `json:"email,omitempty" example:"john.doe@example.com"`
	Phone                string       `json:"phone,omitempty" example:"+1234567890"`
	InvoicedIndividually *bool        `json:"invoiced_individually,omitempty" example:"false"`
	TherapyTitle         string       `json:"therapy_title,omitempty" example:"Cognitive Behavioral Therapy"`
	ProviderApprovalCode string       `json:"provider_approval_code,omitempty" example:"PROV123456"`
	ProviderApprovalDate NullableDate `json:"provider_approval_date,omitempty" example:"2025-01-15"`
	UnitPrice            *float64     `json:"unit_price,omitempty" example:"150.00"`
	Status               string       `json:"status,omitempty" example:"waiting"`
	AdmissionDate        NullableDate `json:"admission_date,omitempty" example:"2025-01-01"`
	ReferralSource       string       `json:"referral_source,omitempty" example:"Doctor Smith"`
	Notes                string       `json:"notes,omitempty" example:"Additional notes about the client"`
}

// UpdateClientRequest represents the request payload for updating a client
type UpdateClientRequest struct {
	CostProviderID       *uint         `json:"cost_provider_id,omitempty" example:"1"`
	FirstName            *string       `json:"first_name,omitempty" example:"John"`
	LastName             *string       `json:"last_name,omitempty" example:"Doe"`
	DateOfBirth          *NullableDate `json:"date_of_birth,omitempty" example:"1990-01-15"`
	Gender               *string       `json:"gender,omitempty" example:"male"`
	PrimaryLanguage      *string       `json:"primary_language,omitempty" example:"English"`
	ContactFirstName     *string       `json:"contact_first_name,omitempty" example:"Jane"`
	ContactLastName      *string       `json:"contact_last_name,omitempty" example:"Smith"`
	ContactEmail         *string       `json:"contact_email,omitempty" example:"jane.smith@example.com"`
	ContactPhone         *string       `json:"contact_phone,omitempty" example:"+1234567890"`
	AlternativeFirstName *string       `json:"alternative_first_name,omitempty" example:"Johnny"`
	AlternativeLastName  *string       `json:"alternative_last_name,omitempty" example:"D"`
	AlternativePhone     *string       `json:"alternative_phone,omitempty" example:"+0987654321"`
	AlternativeEmail     *string       `json:"alternative_email,omitempty" example:"johnny.d@example.com"`
	StreetAddress        *string       `json:"street_address,omitempty" example:"123 Main Street"`
	Zip                  *string       `json:"zip,omitempty" example:"12345"`
	City                 *string       `json:"city,omitempty" example:"New York"`
	Email                *string       `json:"email,omitempty" example:"john.doe@example.com"`
	Phone                *string       `json:"phone,omitempty" example:"+1234567890"`
	InvoicedIndividually *bool         `json:"invoiced_individually,omitempty" example:"false"`
	TherapyTitle         *string       `json:"therapy_title,omitempty" example:"Cognitive Behavioral Therapy"`
	ProviderApprovalCode *string       `json:"provider_approval_code,omitempty" example:"PROV123456"`
	ProviderApprovalDate *NullableDate `json:"provider_approval_date,omitempty" example:"2025-01-15"`
	UnitPrice            *float64      `json:"unit_price,omitempty" example:"150.00"`
	Status               *string       `json:"status,omitempty" example:"active"`
	AdmissionDate        *NullableDate `json:"admission_date,omitempty" example:"2025-01-01"`
	ReferralSource       *string       `json:"referral_source,omitempty" example:"Doctor Smith"`
	Notes                *string       `json:"notes,omitempty" example:"Additional notes about the client"`
}

// ClientResponse represents the response format for client data
type ClientResponse struct {
	ID                   uint                  `json:"id"`
	TenantID             uint                  `json:"tenant_id"`
	CostProviderID       *uint                 `json:"cost_provider_id,omitempty"`
	CostProvider         *CostProviderResponse `json:"cost_provider,omitempty"`
	FirstName            string                `json:"first_name"`
	LastName             string                `json:"last_name"`
	DateOfBirth          *time.Time            `json:"date_of_birth,omitempty"`
	Gender               string                `json:"gender"`
	PrimaryLanguage      string                `json:"primary_language,omitempty"`
	ContactFirstName     string                `json:"contact_first_name,omitempty"`
	ContactLastName      string                `json:"contact_last_name,omitempty"`
	ContactEmail         string                `json:"contact_email,omitempty"`
	ContactPhone         string                `json:"contact_phone,omitempty"`
	AlternativeFirstName string                `json:"alternative_first_name,omitempty"`
	AlternativeLastName  string                `json:"alternative_last_name,omitempty"`
	AlternativePhone     string                `json:"alternative_phone,omitempty"`
	AlternativeEmail     string                `json:"alternative_email,omitempty"`
	StreetAddress        string                `json:"street_address,omitempty"`
	Zip                  string                `json:"zip,omitempty"`
	City                 string                `json:"city,omitempty"`
	Email                string                `json:"email,omitempty"`
	Phone                string                `json:"phone,omitempty"`
	InvoicedIndividually bool                  `json:"invoiced_individually"`
	TherapyTitle         string                `json:"therapy_title,omitempty"`
	ProviderApprovalCode string                `json:"provider_approval_code,omitempty"`
	ProviderApprovalDate *time.Time            `json:"provider_approval_date,omitempty"`
	UnitPrice            *float64              `json:"unit_price,omitempty"`
	Status               string                `json:"status"`
	AdmissionDate        *time.Time            `json:"admission_date,omitempty"`
	ReferralSource       string                `json:"referral_source,omitempty"`
	Notes                string                `json:"notes,omitempty"`
	CreatedAt            time.Time             `json:"created_at"`
	UpdatedAt            time.Time             `json:"updated_at"`
}

// ToResponse converts a Client model to ClientResponse
func (c *Client) ToResponse() ClientResponse {
	var costProvider *CostProviderResponse
	if c.CostProvider != nil {
		response := c.CostProvider.ToResponse()
		costProvider = &response
	}

	return ClientResponse{
		ID:                   c.ID,
		TenantID:             c.TenantID,
		CostProviderID:       c.CostProviderID,
		CostProvider:         costProvider,
		FirstName:            c.FirstName,
		LastName:             c.LastName,
		DateOfBirth:          c.DateOfBirth,
		Gender:               c.Gender,
		PrimaryLanguage:      c.PrimaryLanguage,
		ContactFirstName:     c.ContactFirstName,
		ContactLastName:      c.ContactLastName,
		ContactEmail:         c.ContactEmail,
		ContactPhone:         c.ContactPhone,
		AlternativeFirstName: c.AlternativeFirstName,
		AlternativeLastName:  c.AlternativeLastName,
		AlternativePhone:     c.AlternativePhone,
		AlternativeEmail:     c.AlternativeEmail,
		StreetAddress:        c.StreetAddress,
		Zip:                  c.Zip,
		City:                 c.City,
		Email:                c.Email,
		Phone:                c.Phone,
		InvoicedIndividually: c.InvoicedIndividually,
		TherapyTitle:         c.TherapyTitle,
		ProviderApprovalCode: c.ProviderApprovalCode,
		ProviderApprovalDate: c.ProviderApprovalDate,
		UnitPrice:            c.UnitPrice,
		Status:               c.Status,
		AdmissionDate:        c.AdmissionDate,
		ReferralSource:       c.ReferralSource,
		Notes:                c.Notes,
		CreatedAt:            c.CreatedAt,
		UpdatedAt:            c.UpdatedAt,
	}
}
