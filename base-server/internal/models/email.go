package models

import (
	"time"

	"gorm.io/gorm"
)

// Email represents an email record in the system
type Email struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty" swaggerignore:"true"`
	To           string         `gorm:"column:to;not null" json:"to" binding:"required,email"`
	From         string         `gorm:"column:from;not null" json:"from" binding:"required,email"`
	Subject      string         `gorm:"not null" json:"subject" binding:"required"`
	Body         string         `gorm:"type:text;not null" json:"body" binding:"required"`
	HTMLBody     string         `gorm:"type:text" json:"html_body"`
	Status       string         `gorm:"default:'pending'" json:"status"`
	SentAt       *time.Time     `json:"sent_at"`
	DeliveredAt  *time.Time     `json:"delivered_at"`
	ErrorMessage string         `json:"error_message"`
	// Metadata field removed due to PostgreSQL JSON parsing issues
	// Metadata     *string        `gorm:"type:json;default:'{}'" json:"metadata,omitempty"`
}

// TableName specifies the table name for Email
func (Email) TableName() string {
	return "emails"
}

// EmailResponse represents the API response structure for Email
type EmailResponse struct {
	ID           uint       `json:"id"`
	To           string     `json:"to"`
	From         string     `json:"from"`
	Subject      string     `json:"subject"`
	Status       string     `json:"status"`
	SentAt       *time.Time `json:"sent_at"`
	DeliveredAt  *time.Time `json:"delivered_at"`
	ErrorMessage string     `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ToResponse converts Email to EmailResponse
func (e *Email) ToResponse() EmailResponse {
	return EmailResponse{
		ID:           e.ID,
		To:           e.To,
		From:         e.From,
		Subject:      e.Subject,
		Status:       e.Status,
		SentAt:       e.SentAt,
		DeliveredAt:  e.DeliveredAt,
		ErrorMessage: e.ErrorMessage,
		CreatedAt:    e.CreatedAt,
	}
}

// EmailSendRequest represents the request structure for sending an email
type EmailSendRequest struct {
	To       string `json:"to" binding:"required,email"`
	From     string `json:"from" binding:"required,email"`
	Subject  string `json:"subject" binding:"required"`
	Body     string `json:"body" binding:"required"`
	HTMLBody string `json:"html_body"`
}
