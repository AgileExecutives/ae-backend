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

	"github.com/unburdy/booking-module/entities"
	"gorm.io/gorm"
)

// BookingLinkService handles booking link generation and validation with JWT tokens
type BookingLinkService struct {
	db        *gorm.DB
	secretKey []byte
}

// NewBookingLinkService creates a new booking link service
func NewBookingLinkService(db *gorm.DB, secretKey string) *BookingLinkService {
	return &BookingLinkService{
		db:        db,
		secretKey: []byte(secretKey),
	}
}

// GenerateBookingLink creates a self-contained JWT token for a booking link
func (s *BookingLinkService) GenerateBookingLink(templateID, clientID, tenantID uint, tokenPurpose entities.TokenPurpose) (string, error) {
	// Fetch the template to get user_id and calendar_id
	var template entities.BookingTemplate
	if err := s.db.Where("id = ? AND tenant_id = ?", templateID, tenantID).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("booking template not found")
		}
		return "", fmt.Errorf("failed to retrieve booking template: %w", err)
	}

	// Create JWT claims
	claims := entities.BookingLinkClaims{
		TenantID:   tenantID,
		UserID:     template.UserID,
		CalendarID: template.CalendarID,
		TemplateID: templateID,
		ClientID:   clientID,
		IssuedAt:   time.Now().Unix(),
		Purpose:    tokenPurpose,
	}

	// Set expiry for one-time links (24 hours)
	if tokenPurpose == entities.OneTimeBookingLink {
		claims.ExpiresAt = time.Now().Add(24 * time.Hour).Unix()
	}

	// Generate JWT token
	token, err := s.generateToken(claims)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// ValidateBookingLink validates and decodes a booking link token
func (s *BookingLinkService) ValidateBookingLink(token string) (*entities.BookingLinkClaims, error) {
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

	// Check expiration for one-time links
	if claims.Purpose == entities.OneTimeBookingLink && claims.ExpiresAt > 0 {
		if time.Now().Unix() > claims.ExpiresAt {
			return nil, errors.New("token has expired")
		}
	}

	return &claims, nil
}

// generateToken creates a self-contained JWT token with HMAC-SHA256 signature
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
