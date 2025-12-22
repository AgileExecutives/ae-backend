package entities

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenPurpose represents the purpose of a booking link token
type TokenPurpose string

const (
	OneTimeBookingLink TokenPurpose = "one-time-booking-link"
	TimedBookingLink   TokenPurpose = "timed-booking-link"
)

// BookingLinkClaims represents the JWT claims for a booking link token
// Implements jwt.Claims interface for compatibility with base token service
type BookingLinkClaims struct {
	TenantID     uint         `json:"tenant_id"`
	UserID       uint         `json:"user_id"`
	CalendarID   uint         `json:"calendar_id"`
	TemplateID   uint         `json:"template_id"`
	ClientID     uint         `json:"client_id"`
	Purpose      TokenPurpose `json:"purpose"`
	MaxUseCount  int          `json:"max_use,omitempty"`  // Maximum number of times token can be used (0 = unlimited)
	ValidityDays int          `json:"validity,omitempty"` // Number of days token is valid for
	jwt.RegisteredClaims
}

// CreateBookingLinkRequest represents the request to create a booking link
type CreateBookingLinkRequest struct {
	TemplateID   uint         `json:"template_id" binding:"required"`
	ClientID     uint         `json:"client_id" binding:"required"`
	Purpose      TokenPurpose `json:"token_purpose" binding:"required,oneof=one-time-booking-link timed-booking-link"`
	MaxUseCount  int          `json:"max_use_count,omitempty"` // Maximum number of uses (0 = unlimited, 1 = one-time)
	ValidityDays int          `json:"validity_days,omitempty"` // Number of days token is valid (default: 180 for timed links, 1 for one-time)
}

// BookingLinkResponse represents the response with the booking link
type BookingLinkResponse struct {
	Token     string       `json:"token"`
	URL       string       `json:"url"`
	Purpose   TokenPurpose `json:"purpose"`
	ExpiresAt *time.Time   `json:"expires_at,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}
