package entities

import (
	"time"
)

// TokenPurpose represents the purpose of a booking link token
type TokenPurpose string

const (
	OneTimeBookingLink   TokenPurpose = "one-time-booking-link"
	PermanentBookingLink TokenPurpose = "permanent-booking-link"
)

// BookingLinkClaims represents the JWT claims for a booking link token
type BookingLinkClaims struct {
	TenantID   uint         `json:"tenant_id"`
	UserID     uint         `json:"user_id"`
	CalendarID uint         `json:"calendar_id"`
	TemplateID uint         `json:"template_id"`
	ClientID   uint         `json:"client_id"`
	Purpose    TokenPurpose `json:"purpose"`
	IssuedAt   int64        `json:"iat"`
	ExpiresAt  int64        `json:"exp,omitempty"` // Only for one-time links
}

// CreateBookingLinkRequest represents the request to create a booking link
type CreateBookingLinkRequest struct {
	TemplateID uint         `json:"template_id" binding:"required"`
	ClientID   uint         `json:"client_id" binding:"required"`
	Purpose    TokenPurpose `json:"token_purpose" binding:"required,oneof=one-time-booking-link permanent-booking-link"`
}

// BookingLinkResponse represents the response with the booking link
type BookingLinkResponse struct {
	Token     string       `json:"token"`
	URL       string       `json:"url"`
	Purpose   TokenPurpose `json:"purpose"`
	ExpiresAt *time.Time   `json:"expires_at,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}
