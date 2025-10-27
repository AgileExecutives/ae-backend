package models

import (
	"time"

	"gorm.io/gorm"
)

// Newsletter represents a newsletter subscription
// Public interface for external modules to use
type Newsletter struct {
	ID          uint           `json:"id" gorm:"primaryKey" example:"1"`
	Name        string         `json:"name" gorm:"not null" example:"John Doe"`
	Email       string         `json:"email" gorm:"not null;index" example:"john.doe@example.com"`
	Interest    string         `json:"interest" gorm:"default:'general'" example:"mental_health"`
	Source      string         `json:"source" gorm:"not null" example:"website"`
	LastContact time.Time      `json:"lastContact" gorm:"autoUpdateTime" example:"2025-08-03T10:00:00Z"`
	CreatedAt   time.Time      `json:"createdAt" example:"2025-08-03T10:00:00Z"`
	UpdatedAt   time.Time      `json:"updatedAt" example:"2025-08-03T10:00:00Z"`
	DeletedAt   gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index" swaggerignore:"true"`
}

// TableName specifies the table name for Newsletter
func (Newsletter) TableName() string {
	return "newsletters"
}

// ContactFormRequest represents the request for contact form submission
type ContactFormRequest struct {
	Name       string `json:"name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Subject    string `json:"subject" binding:"required"`
	Message    string `json:"message" binding:"required"`
	Newsletter bool   `json:"newsletter"`
	Source     string `json:"source"`
	Timestamp  string `json:"timestamp"`
}

// ContactFormResponse represents the response for contact form submission
type ContactFormResponse struct {
	Success           bool   `json:"success"`
	Message           string `json:"message"`
	NewsletterAdded   bool   `json:"newsletter_added,omitempty"`
	NewsletterMessage string `json:"newsletter_message,omitempty"`
}

// SharedTokenBlacklist represents a blacklisted JWT token (public interface)
// This provides a common interface that can be used across modules
type SharedTokenBlacklist struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	TokenID   string    `json:"token_id" gorm:"not null;uniqueIndex"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Reason    string    `json:"reason"`
}

// TableName specifies the table name for SharedTokenBlacklist
func (SharedTokenBlacklist) TableName() string {
	return "shared_token_blacklist"
}
