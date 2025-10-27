package models

import (
	"time"

	"gorm.io/gorm"
)

// Newsletter represents a newsletter subscription
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

// NewsletterSubscribeRequest represents the request for newsletter subscription
type NewsletterSubscribeRequest struct {
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Source    string `json:"source,omitempty"`
	Tags      string `json:"tags,omitempty"`
}

// NewsletterUnsubscribeRequest represents the request for newsletter unsubscription
type NewsletterUnsubscribeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// NewsletterResponse represents the API response structure for Newsletter
type NewsletterResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Status    string    `json:"status"`
	Source    string    `json:"source"`
	Tags      string    `json:"tags"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// ContactFormRequest represents the contact form submission data
// @Description Contact form submission request
type ContactFormRequest struct {
	Name       string `json:"name" binding:"required" example:"John Doe"`
	Email      string `json:"email" binding:"required,email" example:"john.doe@example.com"`
	Subject    string `json:"subject" binding:"required" example:"Inquiry about therapy services"`
	Message    string `json:"message" binding:"required" example:"I am interested in learning more about your therapy services."`
	Newsletter bool   `json:"newsletter" example:"true"`
	Timestamp  string `json:"timestamp" example:"2025-08-03T10:00:00Z"`
	Source     string `json:"source" binding:"required" example:"website"`
}

// ContactFormResponse represents the response after contact form submission
// @Description Contact form submission response
type ContactFormResponse struct {
	Message           string `json:"message" example:"Contact form submitted successfully"`
	NewsletterAdded   bool   `json:"newsletterAdded,omitempty" example:"true"`
	NewsletterMessage string `json:"newsletterMessage,omitempty" example:"Successfully subscribed to newsletter"`
}
