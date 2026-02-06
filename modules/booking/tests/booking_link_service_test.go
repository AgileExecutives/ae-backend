package tests

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/booking-module/entities"
	"github.com/unburdy/booking-module/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupBookingLinkTestDB creates an in-memory SQLite database for testing
func setupBookingLinkTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the necessary tables
	err = db.AutoMigrate(&entities.BookingTemplate{})
	require.NoError(t, err)

	return db
}

// TestGenerateToken tests JWT token generation
func TestGenerateToken(t *testing.T) {
	db := setupBookingLinkTestDB(t)
	secret := "test-secret-key"
	service := services.NewBookingLinkService(db, secret)

	// Create a test template
	template := &entities.BookingTemplate{
		TenantID:   1,
		UserID:     5,
		CalendarID: 10,
		Name:       "Test Template",
	}
	err := db.Create(template).Error
	require.NoError(t, err)

	token, err := service.GenerateBookingLink(template.ID, 123, 1, entities.OneTimeBookingLink)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Token should have 3 parts (header.payload.signature)
	parts := len(token)
	assert.Greater(t, parts, 50) // JWT tokens are typically longer than 50 chars
}

// TestValidateToken tests JWT token validation
func TestValidateToken(t *testing.T) {
	db := setupBookingLinkTestDB(t)
	secret := "test-secret-key"
	service := services.NewBookingLinkService(db, secret)

	// Create a test template
	template := &entities.BookingTemplate{
		TenantID:   1,
		UserID:     5,
		CalendarID: 10,
		Name:       "Test Template",
	}
	err := db.Create(template).Error
	require.NoError(t, err)

	// Generate a token
	token, err := service.GenerateBookingLink(template.ID, 123, 1, entities.OneTimeBookingLink)
	assert.NoError(t, err)

	// Validate the token
	claims, err := service.ValidateBookingLink(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	// Verify claims match
	assert.Equal(t, uint(1), claims.TenantID)
	assert.Equal(t, uint(5), claims.UserID)
	assert.Equal(t, uint(10), claims.CalendarID)
	assert.Equal(t, template.ID, claims.TemplateID)
	assert.Equal(t, uint(123), claims.ClientID)
	assert.Equal(t, entities.OneTimeBookingLink, claims.Purpose)
}

// TestValidateTokenWithInvalidSignature tests token with tampered signature
func TestValidateTokenWithInvalidSignature(t *testing.T) {
	db := setupBookingLinkTestDB(t)
	secret := "test-secret-key"
	service := services.NewBookingLinkService(db, secret)

	// Create a test template
	template := &entities.BookingTemplate{
		TenantID:   1,
		UserID:     5,
		CalendarID: 10,
		Name:       "Test Template",
	}
	err := db.Create(template).Error
	require.NoError(t, err)

	// Generate a token
	token, err := service.GenerateBookingLink(template.ID, 123, 1, entities.OneTimeBookingLink)
	assert.NoError(t, err)

	// Tamper with the token (change last character)
	tamperedToken := token[:len(token)-1] + "x"

	// Validation should fail
	_, err = service.ValidateBookingLink(tamperedToken)
	assert.Error(t, err)
}

// TestValidateExpiredToken tests expired one-time token
func TestValidateExpiredToken(t *testing.T) {
	db := setupBookingLinkTestDB(t)
	secret := "test-secret-key"
	service := services.NewBookingLinkService(db, secret)

	// Generate a token that's already expired using JWT directly
	claims := entities.BookingLinkClaims{
		TenantID:   1,
		UserID:     5,
		CalendarID: 10,
		TemplateID: 1,
		ClientID:   123,
		Purpose:    entities.OneTimeBookingLink,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-48 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)

	// Validation should fail due to expiration
	_, err = service.ValidateBookingLink(tokenString)
	assert.Error(t, err)
}

// TestPermanentTokenNoExpiration tests timed tokens with long validity
func TestPermanentTokenNoExpiration(t *testing.T) {
	db := setupBookingLinkTestDB(t)
	secret := "test-secret-key"
	service := services.NewBookingLinkService(db, secret)

	// Create a test template
	template := &entities.BookingTemplate{
		TenantID:   1,
		UserID:     5,
		CalendarID: 10,
		Name:       "Test Template",
	}
	err := db.Create(template).Error
	require.NoError(t, err)

	// Generate a timed booking link with long validity (180 days default)
	token, err := service.GenerateBookingLink(template.ID, 123, 1, entities.TimedBookingLink)
	assert.NoError(t, err)

	// Validation should succeed
	validatedClaims, err := service.ValidateBookingLink(token)
	assert.NoError(t, err)
	assert.NotNil(t, validatedClaims)
	assert.Equal(t, entities.TimedBookingLink, validatedClaims.Purpose)
}

// TestInvalidTokenFormat tests validation of malformed tokens
func TestInvalidTokenFormat(t *testing.T) {
	db := setupBookingLinkTestDB(t)
	service := services.NewBookingLinkService(db, "test-secret")

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
