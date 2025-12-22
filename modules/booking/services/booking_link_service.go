package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/unburdy/booking-module/entities"
	"gorm.io/gorm"
)

// TokenServiceInterface defines the interface for token operations
// This allows us to use either the base TokenService or fall back to legacy implementation
type TokenServiceInterface interface {
	GenerateToken(claims jwt.Claims) (string, error)
	ValidateToken(tokenString string, claims jwt.Claims) error
}

// BookingLinkService handles booking link generation and validation with JWT tokens
type BookingLinkService struct {
	db           *gorm.DB
	secretKey    []byte
	tokenService TokenServiceInterface
}

// NewBookingLinkService creates a new booking link service
func NewBookingLinkService(db *gorm.DB, secretKey string) *BookingLinkService {
	return &BookingLinkService{
		db:           db,
		secretKey:    []byte(secretKey),
		tokenService: nil, // Will use legacy implementation
	}
}

// NewBookingLinkServiceWithTokenService creates a service using the unified token service
func NewBookingLinkServiceWithTokenService(db *gorm.DB, tokenService TokenServiceInterface) *BookingLinkService {
	return &BookingLinkService{
		db:           db,
		secretKey:    nil,
		tokenService: tokenService,
	}
}

// GenerateBookingLink creates a self-contained JWT token for a booking link
func (s *BookingLinkService) GenerateBookingLink(templateID, clientID, tenantID uint, tokenPurpose entities.TokenPurpose) (string, error) {
	return s.GenerateBookingLinkWithOptions(templateID, clientID, tenantID, tokenPurpose, 0, 0)
}

// GenerateBookingLinkWithOptions creates a self-contained JWT token with custom options
func (s *BookingLinkService) GenerateBookingLinkWithOptions(templateID, clientID, tenantID uint, tokenPurpose entities.TokenPurpose, maxUseCount, validityDays int) (string, error) {
	// Fetch the template to get user_id and calendar_id
	var template entities.BookingTemplate
	if err := s.db.Where("id = ? AND tenant_id = ?", templateID, tenantID).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("booking template not found")
		}
		return "", fmt.Errorf("failed to retrieve booking template: %w", err)
	}

	// Create JWT claims
	now := time.Now()
	claims := entities.BookingLinkClaims{
		TenantID:     tenantID,
		UserID:       template.UserID,
		CalendarID:   template.CalendarID,
		TemplateID:   templateID,
		ClientID:     clientID,
		Purpose:      tokenPurpose,
		MaxUseCount:  maxUseCount,
		ValidityDays: validityDays,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        fmt.Sprintf("booking_%d_%d_%d", templateID, clientID, now.Unix()),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	// Set expiry based on purpose and validity days
	if tokenPurpose == entities.OneTimeBookingLink {
		// One-time links expire in 24 hours by default, or custom validity
		if validityDays > 0 {
			claims.ExpiresAt = jwt.NewNumericDate(now.AddDate(0, 0, validityDays))
		} else {
			claims.ExpiresAt = jwt.NewNumericDate(now.Add(24 * time.Hour))
		}
	} else if validityDays > 0 {
		// Permanent links can have custom validity period
		claims.ExpiresAt = jwt.NewNumericDate(now.AddDate(0, 0, validityDays))
	}

	// Use unified token service if available, otherwise fall back to legacy
	if s.tokenService != nil {
		return s.tokenService.GenerateToken(&claims)
	}

	// Legacy implementation for backward compatibility
	return s.generateToken(claims)
}

// ValidateBookingLink validates and decodes a booking link token
func (s *BookingLinkService) ValidateBookingLink(token string) (*entities.BookingLinkClaims, error) {
	// Try unified token service first if available
	if s.tokenService != nil {
		claims := &entities.BookingLinkClaims{}
		err := s.tokenService.ValidateToken(token, claims)
		if err != nil {
			// Fall back to legacy validation for backward compatibility
			return s.validateLegacyToken(token)
		}
		return claims, nil
	}

	// Legacy validation
	return s.validateLegacyToken(token)
}

// validateLegacyToken validates tokens created with the old custom implementation
func (s *BookingLinkService) validateLegacyToken(token string) (*entities.BookingLinkClaims, error) {
	// Split token into parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	headerEncoded := parts[0]
	payloadEncoded := parts[1]
	signatureEncoded := parts[2]

	// Verify signature
	message := headerEncoded + "." + payloadEncoded
	expectedSignature := s.createSignature(message)
	expectedSignatureEncoded := base64.RawURLEncoding.EncodeToString(expectedSignature)

	if signatureEncoded != expectedSignatureEncoded {
		return nil, errors.New("invalid token signature")
	}

	// Decode payload
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	var claims entities.BookingLinkClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	// Check token expiration - handle both old and new format
	if claims.ExpiresAt != nil && !claims.ExpiresAt.IsZero() {
		if time.Now().After(claims.ExpiresAt.Time) {
			return nil, errors.New("token has expired")
		}
	}

	return &claims, nil
}

// InvalidateOneTimeToken invalidates a one-time booking token by adding it to the blacklist
func (s *BookingLinkService) InvalidateOneTimeToken(token string, claims *entities.BookingLinkClaims) error {
	// Only invalidate if it's a one-time token or has max_use_count = 1
	if claims.Purpose != entities.OneTimeBookingLink && claims.MaxUseCount != 1 {
		return nil // Not a one-time token, no need to invalidate
	}

	// Generate token ID
	hash := sha256.Sum256([]byte(token))
	tokenID := fmt.Sprintf("%x", hash[:])

	// Add to blacklist with expiration matching the token's expiration
	expiresAt := time.Now().Add(24 * time.Hour) // Default to 24 hours
	if claims.ExpiresAt != nil && !claims.ExpiresAt.IsZero() {
		expiresAt = claims.ExpiresAt.Time
	}

	blacklistEntry := map[string]interface{}{
		"token_id":   tokenID,
		"user_id":    claims.UserID,
		"expires_at": expiresAt,
		"reason":     fmt.Sprintf("One-time token exhausted after booking (client_id: %d)", claims.ClientID),
	}

	if err := s.db.Table("token_blacklist").Create(&blacklistEntry).Error; err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// generateToken creates a self-contained JWT token with HMAC-SHA256 signature (legacy)
func (s *BookingLinkService) generateToken(claims entities.BookingLinkClaims) (string, error) {
	// Create header
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}

	// Encode header
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Encode payload (claims)
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature
	message := headerEncoded + "." + payloadEncoded
	signature := s.createSignature(message)
	signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)

	// Combine to create JWT
	token := message + "." + signatureEncoded

	return token, nil
}

// createSignature creates HMAC-SHA256 signature
func (s *BookingLinkService) createSignature(message string) []byte {
	h := hmac.New(sha256.New, s.secretKey)
	h.Write([]byte(message))
	return h.Sum(nil)
}
