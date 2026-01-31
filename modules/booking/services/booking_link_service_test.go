package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/unburdy/booking-module/entities"
)

// TestGenerateToken tests JWT token generation
func TestGenerateToken(t *testing.T) {
	secret := "test-secret-key"
	service := NewBookingLinkService(nil, secret)

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	claims := entities.BookingLinkClaims{
		TenantID:   1,
		UserID:     5,
		CalendarID: 10,
		TemplateID: 1,
		ClientID:   123,
		Purpose:    entities.OneTimeBookingLink,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token, err := service.generateToken(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Token should have 3 parts (header.payload.signature)
	parts := len(token)
	assert.Greater(t, parts, 50) // JWT tokens are typically longer than 50 chars
}

// TestValidateToken tests JWT token validation
func TestValidateToken(t *testing.T) {
	secret := "test-secret-key"
	service := NewBookingLinkService(nil, secret)

	// Generate a token
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	originalClaims := entities.BookingLinkClaims{
		TenantID:   1,
		UserID:     5,
		CalendarID: 10,
		TemplateID: 1,
		ClientID:   123,
		Purpose:    entities.OneTimeBookingLink,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token, err := service.generateToken(originalClaims)
	assert.NoError(t, err)

	// Validate the token
	claims, err := service.ValidateBookingLink(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	// Verify claims match
	assert.Equal(t, originalClaims.TenantID, claims.TenantID)
	assert.Equal(t, originalClaims.UserID, claims.UserID)
	assert.Equal(t, originalClaims.CalendarID, claims.CalendarID)
	assert.Equal(t, originalClaims.TemplateID, claims.TemplateID)
	assert.Equal(t, originalClaims.ClientID, claims.ClientID)
	assert.Equal(t, originalClaims.Purpose, claims.Purpose)
}

// TestValidateTokenWithInvalidSignature tests token with tampered signature
func TestValidateTokenWithInvalidSignature(t *testing.T) {
	secret := "test-secret-key"
	service := NewBookingLinkService(nil, secret)

	// Generate a token
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	claims := entities.BookingLinkClaims{
		TenantID:   1,
		UserID:     5,
		CalendarID: 10,
		TemplateID: 1,
		ClientID:   123,
		Purpose:    entities.OneTimeBookingLink,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token, err := service.generateToken(claims)
	assert.NoError(t, err)

	// Tamper with the token (change last character)
	tamperedToken := token[:len(token)-1] + "x"

	// Validation should fail
	_, err = service.ValidateBookingLink(tamperedToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token signature")
}

// TestValidateExpiredToken tests expired one-time token
func TestValidateExpiredToken(t *testing.T) {
	secret := "test-secret-key"
	service := NewBookingLinkService(nil, secret)

	// Generate a token that's already expired
	issuedAt := time.Now().Add(-48 * time.Hour)
	expiresAt := time.Now().Add(-24 * time.Hour)
	claims := entities.BookingLinkClaims{
		TenantID:   1,
		UserID:     5,
		CalendarID: 10,
		TemplateID: 1,
		ClientID:   123,
		Purpose:    entities.OneTimeBookingLink,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token, err := service.generateToken(claims)
	assert.NoError(t, err)

	// Validation should fail due to expiration
	_, err = service.ValidateBookingLink(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

// TestPermanentTokenNoExpiration tests permanent tokens don't expire
func TestPermanentTokenNoExpiration(t *testing.T) {
	secret := "test-secret-key"
	service := NewBookingLinkService(nil, secret)

	// Generate a permanent token (no expiration)
	issuedAt := time.Now().Add(-48 * time.Hour)
	claims := entities.BookingLinkClaims{
		TenantID:   1,
		UserID:     5,
		CalendarID: 10,
		TemplateID: 1,
		ClientID:   123,
		Purpose:    entities.TimedBookingLink,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(issuedAt),
			// No ExpiresAt for permanent tokens
		},
	}

	token, err := service.generateToken(claims)
	assert.NoError(t, err)

	// Validation should succeed even though issued 2 days ago
	validatedClaims, err := service.ValidateBookingLink(token)
	assert.NoError(t, err)
	assert.NotNil(t, validatedClaims)
	assert.Equal(t, entities.TimedBookingLink, validatedClaims.Purpose)
}

// TestInvalidTokenFormat tests validation of malformed tokens
func TestInvalidTokenFormat(t *testing.T) {
	service := NewBookingLinkService(nil, "test-secret")

	testCases := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"single part", "abc"},
		{"two parts", "abc.def"},
		{"four parts", "abc.def.ghi.jkl"},
		{"invalid base64", "abc.def.!!!"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.ValidateBookingLink(tc.token)
			assert.Error(t, err)
		})
	}
}
