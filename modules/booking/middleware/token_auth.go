package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	baseAPI "github.com/ae-base-server/api"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/booking-module/entities"
	"github.com/unburdy/booking-module/services"
	"gorm.io/gorm"
)

// BookingTokenMiddleware validates booking link tokens
type BookingTokenMiddleware struct {
	bookingLinkSvc *services.BookingLinkService
	db             *gorm.DB
}

// NewBookingTokenMiddleware creates a new booking token middleware
func NewBookingTokenMiddleware(svc *services.BookingLinkService, db *gorm.DB) *BookingTokenMiddleware {
	return &BookingTokenMiddleware{
		bookingLinkSvc: svc,
		db:             db,
	}
}

// ValidateBookingToken validates the booking token and injects claims into context
func (m *BookingTokenMiddleware) ValidateBookingToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from URL parameter
		token := c.Param("token")
		if token == "" {
			// Try query parameter as fallback
			token = c.Query("token")
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Missing token", "Booking token is required"))
			c.Abort()
			return
		}

		// Validate the token structure and signature
		claims, err := m.bookingLinkSvc.ValidateBookingLink(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Invalid token", err.Error()))
			c.Abort()
			return
		}

		// Check if token is blacklisted
		tokenID := generateTokenID(token)
		var blacklistedToken struct {
			ID        uint
			TokenID   string
			ExpiresAt time.Time
		}
		result := m.db.Table("token_blacklist").
			Select("id, token_id, expires_at").
			Where("token_id = ? AND deleted_at IS NULL", tokenID).
			First(&blacklistedToken)

		if result.Error == nil {
			// Token found in blacklist - check if still active
			if time.Now().Before(blacklistedToken.ExpiresAt) {
				c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Token revoked", "This booking link has been revoked"))
				c.Abort()
				return
			}
		}

		// Additional validation: Check if it's a valid booking purpose
		if claims.Purpose != entities.OneTimeBookingLink && claims.Purpose != entities.PermanentBookingLink {
			c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Invalid token", "Invalid token purpose"))
			c.Abort()
			return
		}

		// Inject claims into context for handlers to use
		c.Set("booking_claims", claims)
		c.Set("booking_token", token)
		c.Set("booking_token_id", tokenID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("calendar_id", claims.CalendarID)
		c.Set("template_id", claims.TemplateID)
		c.Set("client_id", claims.ClientID)

		c.Next()
	}
}

// generateTokenID creates a unique ID from the token for blacklist tracking
// Uses SHA256 hash of the token signature part
func generateTokenID(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GetBookingClaims retrieves booking claims from context
func GetBookingClaims(c *gin.Context) (*entities.BookingLinkClaims, bool) {
	claims, exists := c.Get("booking_claims")
	if !exists {
		return nil, false
	}
	bookingClaims, ok := claims.(*entities.BookingLinkClaims)
	return bookingClaims, ok
}

// BlacklistToken adds a token to the blacklist
func (m *BookingTokenMiddleware) BlacklistToken(token string, reason string, expiresAt time.Time) error {
	tokenID := generateTokenID(token)

	blacklistEntry := map[string]interface{}{
		"token_id":   tokenID,
		"user_id":    0, // Booking tokens don't have a user_id in the traditional sense
		"expires_at": expiresAt,
		"reason":     reason,
	}

	return m.db.Table("token_blacklist").Create(&blacklistEntry).Error
}
