package entities

import (
	"time"

	"gorm.io/gorm"
)

// RegistrationToken represents a permanent token for client registration (organization-scoped)
type RegistrationToken struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	OrganizationID uint           `gorm:"not null;index" json:"organization_id"`
	TenantID       uint           `gorm:"not null;index" json:"tenant_id"`
	Token          string         `gorm:"size:500;not null;uniqueIndex" json:"token"`
	Email          string         `gorm:"size:255" json:"email,omitempty"` // Optional: pre-associate with email
	UsedCount      int            `gorm:"default:0" json:"used_count"`     // Track how many times token was used
	Blacklisted    bool           `gorm:"default:false" json:"blacklisted"`
	CreatedBy      uint           `gorm:"not null" json:"created_by"` // User ID who created the token
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// RegistrationTokenResponse represents the response for a registration token
type RegistrationTokenResponse struct {
	Token          string `json:"token"`
	Email          string `json:"email,omitempty"`
	OrganizationID uint   `json:"organization_id"`
}

// ClientRegistrationRequest represents the request for registering a new client via token
type ClientRegistrationRequest struct {
	FirstName        string       `json:"first_name" binding:"required" example:"John"`
	LastName         string       `json:"last_name" binding:"required" example:"Doe"`
	Email            string       `json:"email" binding:"required,email" example:"john.doe@example.com"`
	Phone            string       `json:"phone,omitempty" example:"+1234567890"`
	DateOfBirth      NullableDate `json:"date_of_birth,omitempty"`
	Gender           string       `json:"gender,omitempty" example:"male"`
	StreetAddress    string       `json:"street_address,omitempty" example:"123 Main Street"`
	Zip              string       `json:"zip,omitempty" example:"12345"`
	City             string       `json:"city,omitempty" example:"New York"`
	ContactFirstName string       `json:"contact_first_name,omitempty" example:"Jane"`
	ContactLastName  string       `json:"contact_last_name,omitempty" example:"Smith"`
	ContactEmail     string       `json:"contact_email,omitempty" example:"jane.smith@example.com"`
	ContactPhone     string       `json:"contact_phone,omitempty" example:"+1234567890"`
	Notes            string       `json:"notes,omitempty" example:"Referred by Dr. Smith"`
	Timezone         string       `json:"timezone,omitempty" example:"Europe/Berlin"`
}
