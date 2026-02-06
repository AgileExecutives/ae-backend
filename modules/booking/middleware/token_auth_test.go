package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/booking-module/entities"
	"github.com/unburdy/booking-module/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the necessary tables
	err = db.AutoMigrate(&entities.BookingTemplate{})
	require.NoError(t, err)

	// Create token_blacklist table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS token_blacklist (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token_id TEXT NOT NULL,
			user_id INTEGER,
			expires_at DATETIME NOT NULL,
			reason TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	return db
}

// setupTestServices creates middleware and service instances for testing
func setupTestServices(db *gorm.DB) (*BookingTokenMiddleware, *services.BookingLinkService) {
	secret := "test-secret-key-32-chars-long!!"
	bookingLinkSvc := services.NewBookingLinkService(db, secret)
	middleware := NewBookingTokenMiddleware(bookingLinkSvc, db)
	return middleware, bookingLinkSvc
}

// generateToken creates a valid JWT token for testing using BookingLinkClaims
func generateToken(templateID, clientID, tenantID, calendarID, userID uint, purpose entities.TokenPurpose, secret string, expiry time.Duration) string {
	claims := entities.BookingLinkClaims{
		TenantID:   tenantID,
		UserID:     userID,
		CalendarID: calendarID,
		TemplateID: templateID,
		ClientID:   clientID,
		Purpose:    purpose,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestValidateBookingToken_Success(t *testing.T) {
	db := setupTestDB(t)
	middleware, bookingLinkSvc := setupTestServices(db)

	// Create test template for token generation
	template := &entities.BookingTemplate{
		TenantID:   1,
		UserID:     1,
		CalendarID: 1,
		Name:       "Test Template",
	}
	err := db.Create(template).Error
	require.NoError(t, err)

	// Generate valid token using BookingLinkService
	token, err := bookingLinkSvc.GenerateBookingLink(template.ID, 1, 1, entities.OneTimeBookingLink)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		// Verify claims are set in context
		claims, exists := GetBookingClaims(c)
		assert.True(t, exists)
		assert.NotNil(t, claims)
		assert.Equal(t, uint(1), claims.TenantID)
		assert.Equal(t, template.ID, claims.TemplateID)

		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Make request
	req := httptest.NewRequest("GET", "/booking/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
}

func TestValidateBookingToken_MissingToken(t *testing.T) {
	db := setupTestDB(t)
	middleware, _ := setupTestServices(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/booking/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Route doesn't match, returns 404
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestValidateBookingToken_InvalidFormat(t *testing.T) {
	db := setupTestDB(t)
	middleware, _ := setupTestServices(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/booking/invalid-token", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestValidateBookingToken_ExpiredToken(t *testing.T) {
	db := setupTestDB(t)
	middleware, _ := setupTestServices(db)

	// Generate expired token
	secret := "test-secret-key-32-chars-long!!"
	token := generateToken(1, 1, 1, 1, 1, entities.OneTimeBookingLink, secret, -1*time.Hour)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/booking/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestValidateBookingToken_InvalidPurpose(t *testing.T) {
	db := setupTestDB(t)
	middleware, _ := setupTestServices(db)

	// Generate token with invalid purpose
	secret := "test-secret-key-32-chars-long!!"
	claims := entities.BookingLinkClaims{
		TenantID:   1,
		UserID:     1,
		CalendarID: 1,
		TemplateID: 1,
		ClientID:   1,
		Purpose:    entities.TokenPurpose("invalid-purpose"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, _ := jwtToken.SignedString([]byte(secret))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/booking/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestValidateBookingToken_BlacklistedToken(t *testing.T) {
	db := setupTestDB(t)
	middleware, bookingLinkSvc := setupTestServices(db)

	// Create test template for token generation
	template := &entities.BookingTemplate{
		TenantID:   1,
		UserID:     1,
		CalendarID: 1,
		Name:       "Test Template",
	}
	err := db.Create(template).Error
	require.NoError(t, err)

	// Generate valid token
	token, err := bookingLinkSvc.GenerateBookingLink(template.ID, 1, 1, entities.OneTimeBookingLink)
	require.NoError(t, err)

	// Blacklist the token
	err = middleware.BlacklistToken(token, "Test revocation", time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/booking/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestBlacklistToken_Success(t *testing.T) {
	db := setupTestDB(t)
	middleware, _ := setupTestServices(db)

	token := "test-token-123"
	reason := "Test reason"
	expiresAt := time.Now().Add(24 * time.Hour)

	err := middleware.BlacklistToken(token, reason, expiresAt)
	require.NoError(t, err)

	// Verify token is blacklisted
	tokenID := generateTokenID(token)
	var blacklist struct {
		TokenID   string
		Reason    string
		ExpiresAt time.Time
	}
	err = db.Table("token_blacklist").
		Select("token_id, reason, expires_at").
		Where("token_id = ?", tokenID).
		First(&blacklist).Error
	require.NoError(t, err)
	assert.Equal(t, tokenID, blacklist.TokenID)
	assert.Equal(t, reason, blacklist.Reason)
	assert.WithinDuration(t, expiresAt, blacklist.ExpiresAt, time.Second)
}

func TestGetBookingClaims_Success(t *testing.T) {
	// Create a gin context with claims
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	expectedClaims := &entities.BookingLinkClaims{
		TenantID:   123,
		TemplateID: 456,
		ClientID:   789,
		Purpose:    entities.OneTimeBookingLink,
	}

	c.Set("booking_claims", expectedClaims)

	claims, exists := GetBookingClaims(c)
	require.True(t, exists)
	assert.NotNil(t, claims)
	assert.Equal(t, expectedClaims.TenantID, claims.TenantID)
	assert.Equal(t, expectedClaims.TemplateID, claims.TemplateID)
	assert.Equal(t, expectedClaims.ClientID, claims.ClientID)
}

func TestGetBookingClaims_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	claims, exists := GetBookingClaims(c)
	assert.False(t, exists)
	assert.Nil(t, claims)
}
