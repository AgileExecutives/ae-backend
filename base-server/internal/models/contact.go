package models

import (
	"time"

	"gorm.io/gorm"
)

// Contact represents a contact in the system
type Contact struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	FirstName string         `gorm:"not null" json:"first_name" binding:"required"`
	LastName  string         `gorm:"not null" json:"last_name" binding:"required"`
	Email     string         `json:"email" binding:"omitempty,email"`
	Phone     string         `json:"phone"`
	Mobile    string         `json:"mobile"`
	Street    string         `json:"street"`
	Zip       string         `json:"zip"`
	City      string         `json:"city"`
	Country   string         `json:"country"`
	Type      string         `gorm:"default:'contact'" json:"type"`
	Notes     string         `gorm:"type:text" json:"notes"`
	Active    bool           `gorm:"default:true" json:"active"`
}

// TableName specifies the table name for Contact
func (Contact) TableName() string {
	return "contacts"
}

// ContactResponse represents the API response structure for Contact
type ContactResponse struct {
	ID        uint      `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Mobile    string    `json:"mobile"`
	Street    string    `json:"street"`
	Zip       string    `json:"zip"`
	City      string    `json:"city"`
	Country   string    `json:"country"`
	Type      string    `json:"type"`
	Notes     string    `json:"notes"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts Contact to ContactResponse
func (c *Contact) ToResponse() ContactResponse {
	return ContactResponse{
		ID:        c.ID,
		FirstName: c.FirstName,
		LastName:  c.LastName,
		Email:     c.Email,
		Phone:     c.Phone,
		Mobile:    c.Mobile,
		Street:    c.Street,
		Zip:       c.Zip,
		City:      c.City,
		Country:   c.Country,
		Type:      c.Type,
		Notes:     c.Notes,
		Active:    c.Active,
		CreatedAt: c.CreatedAt,
	}
}

// ContactCreateRequest represents the request structure for creating a contact
type ContactCreateRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"omitempty,email"`
	Phone     string `json:"phone"`
	Mobile    string `json:"mobile"`
	Street    string `json:"street"`
	Zip       string `json:"zip"`
	City      string `json:"city"`
	Country   string `json:"country"`
	Type      string `json:"type"`
	Notes     string `json:"notes"`
}

// ContactUpdateRequest represents the request structure for updating a contact
type ContactUpdateRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email" binding:"omitempty,email"`
	Phone     string `json:"phone"`
	Mobile    string `json:"mobile"`
	Street    string `json:"street"`
	Zip       string `json:"zip"`
	City      string `json:"city"`
	Country   string `json:"country"`
	Type      string `json:"type"`
	Notes     string `json:"notes"`
	Active    *bool  `json:"active"`
}
