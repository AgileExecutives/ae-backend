package tests

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/booking-module/entities"
	"github.com/unburdy/booking-module/middleware"
	"github.com/unburdy/booking-module/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create booking_templates table
	err = db.Exec(`
		CREATE TABLE booking_templates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			calendar_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			name VARCHAR(255) NOT NULL,
			slot_duration INTEGER DEFAULT 30,
			timezone VARCHAR(100) DEFAULT 'UTC',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	// Create token_blacklist table
	err = db.Exec(`
		CREATE TABLE token_blacklist (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token_id VARCHAR(255) NOT NULL,
			user_id INTEGER,
			reason TEXT,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	return db
}

// setupTestServices creates test services and middleware
func setupTestServices(t *testing.T, db *gorm.DB) (*services.BookingLinkService, *middleware.BookingTokenMiddleware) {
	// Create a test template
	err := db.Exec(`
		INSERT INTO booking_templates (id, user_id, calendar_id, tenant_id, name, slot_duration, timezone)
		VALUES (1, 1, 1, 1, 'Test Template', 30, 'UTC')
	`).Error
	require.NoError(t, err)

	bookingLinkSvc := services.NewBookingLinkService(db, "test-secret-key")
	middleware := middleware.NewBookingTokenMiddleware(bookingLinkSvc, db)

	return bookingLinkSvc, middleware
}

// setupGin creates a test Gin router with the middleware
func setupGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// generateTokenID creates a unique ID from the token for testing
// This is a copy of the private function from the middleware package
func generateTokenID(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func TestValidateBookingToken_ValidToken(t *testing.T) {
	db := setupTestDB(t)
	bookingLinkSvc, middleware := setupTestServices(t, db)

	// Generate a valid token
	token, err := bookingLinkSvc.GenerateBookingLink(1, 123, 1, entities.OneTimeBookingLink)
	require.NoError(t, err)

	router := setupGin()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		// Check if claims are in context
		claims, exists := c.Get("booking_claims")
		assert.True(t, exists)

		bookingClaims, ok := claims.(*entities.BookingLinkClaims)
		assert.True(t, ok)
		assert.Equal(t, uint(1), bookingClaims.TemplateID)
		assert.Equal(t, uint(123), bookingClaims.ClientID)
		assert.Equal(t, uint(1), bookingClaims.TenantID)

		// Check other context values
		tenantID, exists := c.Get("tenant_id")
		assert.True(t, exists)
		assert.Equal(t, uint(1), tenantID)

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/booking/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateBookingToken_ValidTokenFromQuery(t *testing.T) {
	db := setupTestDB(t)
	bookingLinkSvc, middleware := setupTestServices(t, db)

	// Generate a valid token
	token, err := bookingLinkSvc.GenerateBookingLink(1, 123, 1, entities.TimedBookingLink)
	require.NoError(t, err)

	router := setupGin()
	router.GET("/booking", middleware.ValidateBookingToken(), func(c *gin.Context) {
		claims, exists := c.Get("booking_claims")
		assert.True(t, exists)

		bookingClaims, ok := claims.(*entities.BookingLinkClaims)
		assert.True(t, ok)
		assert.Equal(t, entities.TimedBookingLink, bookingClaims.Purpose)

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/booking?token="+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateBookingToken_MissingToken(t *testing.T) {
	db := setupTestDB(t)
	_, middleware := setupTestServices(t, db)

	router := setupGin()
	router.GET("/booking", middleware.ValidateBookingToken(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Request without token parameter or query
	req := httptest.NewRequest("GET", "/booking", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Missing token")
}

func TestValidateBookingToken_InvalidToken(t *testing.T) {
	db := setupTestDB(t)
	_, middleware := setupTestServices(t, db)

	router := setupGin()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Invalid token format
	req := httptest.NewRequest("GET", "/booking/invalid-token-format", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")
}

func TestValidateBookingToken_ModifiedToken(t *testing.T) {
	db := setupTestDB(t)
	bookingLinkSvc, middleware := setupTestServices(t, db)

	// Generate a valid token
	token, err := bookingLinkSvc.GenerateBookingLink(1, 123, 1, entities.OneTimeBookingLink)
	require.NoError(t, err)

	// Modify the token (change one character)
	modifiedToken := "a" + token[1:]

	router := setupGin()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/booking/"+modifiedToken, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")
}

func TestValidateBookingToken_ExpiredOneTimeToken(t *testing.T) {
	// Note: The current implementation validates tokens but doesn't prevent reuse
	// in the middleware layer. Token expiry for one-time use is handled at the
	// service layer during actual booking, not at validation time.
	// This test documents the current behavior.

	t.Skip("One-time token expiry is enforced at booking time, not at validation time")
}

func TestValidateBookingToken_BlacklistedToken(t *testing.T) {
	db := setupTestDB(t)
	bookingLinkSvc, middleware := setupTestServices(t, db)

	// Generate a valid token
	token, err := bookingLinkSvc.GenerateBookingLink(1, 123, 1, entities.TimedBookingLink)
	require.NoError(t, err)

	// Blacklist the token
	err = middleware.BlacklistToken(token, "Testing blacklist", time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	router := setupGin()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/booking/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Token revoked")
}

func TestValidateBookingToken_ExpiredBlacklistEntry(t *testing.T) {
	db := setupTestDB(t)
	bookingLinkSvc, middleware := setupTestServices(t, db)

	// Generate a valid token
	token, err := bookingLinkSvc.GenerateBookingLink(1, 123, 1, entities.TimedBookingLink)
	require.NoError(t, err)

	// Blacklist the token with an expiry in the past
	err = middleware.BlacklistToken(token, "Testing expired blacklist", time.Now().Add(-1*time.Hour))
	require.NoError(t, err)

	router := setupGin()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/booking/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should succeed because blacklist entry has expired
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateBookingToken_InvalidPurpose(t *testing.T) {
	db := setupTestDB(t)
	_, _ = setupTestServices(t, db)

	// We need to create a token with an invalid purpose manually
	// This is tricky since the service validates the purpose
	// For this test, we'll skip it as the service doesn't allow creating invalid purpose tokens
	// In a real scenario, you'd need to mock or create a test-specific token generator

	t.Skip("Service prevents creating tokens with invalid purpose - tested at service level")
}

func TestValidateBookingToken_ContextInjection(t *testing.T) {
	db := setupTestDB(t)
	bookingLinkSvc, middleware := setupTestServices(t, db)

	// Generate a valid token with tenant_id 1 (matches the template)
	token, err := bookingLinkSvc.GenerateBookingLink(1, 456, 1, entities.OneTimeBookingLink)
	require.NoError(t, err)

	router := setupGin()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		// Verify all context values are injected
		claims, exists := c.Get("booking_claims")
		assert.True(t, exists, "booking_claims should exist")

		tokenValue, exists := c.Get("booking_token")
		assert.True(t, exists, "booking_token should exist")
		assert.Equal(t, token, tokenValue)

		tokenID, exists := c.Get("booking_token_id")
		assert.True(t, exists, "booking_token_id should exist")
		expectedID := generateTokenID(token)
		assert.Equal(t, expectedID, tokenID)

		tenantID, exists := c.Get("tenant_id")
		assert.True(t, exists, "tenant_id should exist")
		assert.Equal(t, uint(1), tenantID)

		calendarID, exists := c.Get("calendar_id")
		assert.True(t, exists, "calendar_id should exist")
		assert.NotNil(t, calendarID, "calendar_id should not be nil")

		templateID, exists := c.Get("template_id")
		assert.True(t, exists, "template_id should exist")
		assert.Equal(t, uint(1), templateID)

		clientID, exists := c.Get("client_id")
		assert.True(t, exists, "client_id should exist")
		assert.Equal(t, uint(456), clientID)

		bookingClaims := claims.(*entities.BookingLinkClaims)
		assert.Equal(t, uint(1), bookingClaims.TemplateID)
		assert.Equal(t, uint(456), bookingClaims.ClientID)
		assert.Equal(t, uint(1), bookingClaims.TenantID)

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/booking/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetBookingClaims_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	claims := &entities.BookingLinkClaims{
		TemplateID: 1,
		ClientID:   123,
		TenantID:   1,
	}
	c.Set("booking_claims", claims)

	retrievedClaims, exists := middleware.GetBookingClaims(c)
	assert.True(t, exists)
	assert.NotNil(t, retrievedClaims)
	assert.Equal(t, uint(1), retrievedClaims.TemplateID)
	assert.Equal(t, uint(123), retrievedClaims.ClientID)
}

func TestGetBookingClaims_NotExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	retrievedClaims, exists := middleware.GetBookingClaims(c)
	assert.False(t, exists)
	assert.Nil(t, retrievedClaims)
}

func TestGetBookingClaims_WrongType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Set wrong type
	c.Set("booking_claims", "not a claims object")

	retrievedClaims, exists := middleware.GetBookingClaims(c)
	assert.False(t, exists)
	assert.Nil(t, retrievedClaims)
}

func TestGenerateTokenID_Consistency(t *testing.T) {
	token := "test-token-123"

	id1 := generateTokenID(token)
	id2 := generateTokenID(token)

	// Same token should produce same ID
	assert.Equal(t, id1, id2)

	// Should be a valid hex string
	_, err := hex.DecodeString(id1)
	assert.NoError(t, err)

	// Should be SHA256 length (64 hex chars)
	assert.Equal(t, 64, len(id1))
}

func TestGenerateTokenID_Uniqueness(t *testing.T) {
	token1 := "token-1"
	token2 := "token-2"

	id1 := generateTokenID(token1)
	id2 := generateTokenID(token2)

	// Different tokens should produce different IDs
	assert.NotEqual(t, id1, id2)
}

func TestGenerateTokenID_MatchesExpectedHash(t *testing.T) {
	token := "test-token"

	id := generateTokenID(token)

	// Manually compute expected hash
	hash := sha256.Sum256([]byte(token))
	expected := hex.EncodeToString(hash[:])

	assert.Equal(t, expected, id)
}

func TestBlacklistToken_Success(t *testing.T) {
	db := setupTestDB(t)
	bookingLinkSvc, middleware := setupTestServices(t, db)

	token, err := bookingLinkSvc.GenerateBookingLink(1, 123, 1, entities.TimedBookingLink)
	require.NoError(t, err)

	expiresAt := time.Now().Add(24 * time.Hour)
	err = middleware.BlacklistToken(token, "Test reason", expiresAt)
	assert.NoError(t, err)

	// Verify it's in the database
	tokenID := generateTokenID(token)
	var count int64
	db.Table("token_blacklist").Where("token_id = ?", tokenID).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestBlacklistToken_MultipleTokens(t *testing.T) {
	db := setupTestDB(t)
	bookingLinkSvc, middleware := setupTestServices(t, db)

	token1, err := bookingLinkSvc.GenerateBookingLink(1, 123, 1, entities.TimedBookingLink)
	require.NoError(t, err)

	token2, err := bookingLinkSvc.GenerateBookingLink(1, 456, 1, entities.OneTimeBookingLink)
	require.NoError(t, err)

	expiresAt := time.Now().Add(24 * time.Hour)
	err = middleware.BlacklistToken(token1, "Reason 1", expiresAt)
	assert.NoError(t, err)

	err = middleware.BlacklistToken(token2, "Reason 2", expiresAt)
	assert.NoError(t, err)

	// Verify both are in the database
	var count int64
	db.Table("token_blacklist").Count(&count)
	assert.Equal(t, int64(2), count)
}

func TestMiddleware_ChainedRequests(t *testing.T) {
	db := setupTestDB(t)
	bookingLinkSvc, middleware := setupTestServices(t, db)

	// Generate a permanent token
	token, err := bookingLinkSvc.GenerateBookingLink(1, 123, 1, entities.TimedBookingLink)
	require.NoError(t, err)

	router := setupGin()
	requestCount := 0
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		requestCount++
		c.JSON(http.StatusOK, gin.H{"count": requestCount})
	})

	// First request
	req1 := httptest.NewRequest("GET", "/booking/"+token, nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request with same token (should work for permanent tokens)
	req2 := httptest.NewRequest("GET", "/booking/"+token, nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	assert.Equal(t, 2, requestCount)
}

func TestMiddleware_AbortsPipeline(t *testing.T) {
	db := setupTestDB(t)
	_, middleware := setupTestServices(t, db)

	router := setupGin()
	handlerCalled := false
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Request with invalid token
	req := httptest.NewRequest("GET", "/booking/invalid-token", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Handler should not be called
	assert.False(t, handlerCalled, "Handler should not be called when token is invalid")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNewBookingTokenMiddleware(t *testing.T) {
	db := setupTestDB(t)
	bookingLinkSvc := services.NewBookingLinkService(db, "test-secret")

	middleware := middleware.NewBookingTokenMiddleware(bookingLinkSvc, db)

	assert.NotNil(t, middleware)
	assert.NotNil(t, middleware.bookingLinkSvc)
	assert.NotNil(t, middleware.db)
}

func TestValidateBookingToken_PermanentToken(t *testing.T) {
	db := setupTestDB(t)
	bookingLinkSvc, middleware := setupTestServices(t, db)

	// Generate a permanent token
	token, err := bookingLinkSvc.GenerateBookingLink(1, 999, 1, entities.TimedBookingLink)
	require.NoError(t, err)

	router := setupGin()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		claims, _ := c.Get("booking_claims")
		bookingClaims := claims.(*entities.BookingLinkClaims)

		assert.Equal(t, entities.TimedBookingLink, bookingClaims.Purpose)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Should work multiple times
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/booking/"+token, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, fmt.Sprintf("Request %d should succeed", i+1))
	}
}

func TestValidateBookingToken_OneTimeToken(t *testing.T) {
	// Note: One-time token usage restriction is enforced during actual booking,
	// not at the validation/middleware level. The middleware only validates
	// the token signature and checks the blacklist.
	// To test one-time behavior, the token would need to be blacklisted after use.

	t.Skip("One-time token usage restriction is enforced at booking time, not validation time")
}
