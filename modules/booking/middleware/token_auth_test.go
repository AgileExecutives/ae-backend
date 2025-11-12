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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the necessary tables
	err = db.AutoMigrate(&entities.BookingLink{}, &entities.TokenBlacklist{})
	require.NoError(t, err)

	return db
}

// setupTestServices creates middleware and service instances for testing
func setupTestServices(db *gorm.DB) *BookingTokenMiddleware {
	middleware := NewBookingTokenMiddleware(db, "test-secret-key-32-chars-long!!")
	return middleware
}

// generateToken creates a valid JWT token for testing
func generateToken(bookingLinkID uint, tenantID uint, secret string, expiry time.Duration) string {
	claims := BookingTokenClaims{
		BookingLinkID: bookingLinkID,
		TenantID:      tenantID,
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
	middleware := setupTestServices(db)

	// Create test booking link
	bookingLink := &entities.BookingLink{
		TenantID:  1,
		Token:     "test-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err := db.Create(bookingLink).Error
	require.NoError(t, err)

	// Generate valid token
	token := generateToken(bookingLink.ID, bookingLink.TenantID, "test-secret-key-32-chars-longrm middleware/token_auth_test.go", 24*time.Hour)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
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
	middleware := setupTestServices(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/booking/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestValidateBookingToken_InvalidFormat(t *testing.T) {
	db := setupTestDB(t)
	middleware := setupTestServices(db)

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
	assert.Contains(t, response["error"], "Invalid token format")
}

func TestValidateBookingToken_ExpiredToken(t *testing.T) {
	db := setupTestDB(t)
	middleware := setupTestServices(db)

	// Create test booking link
	bookingLink := &entities.BookingLink{
		TenantID:  1,
		Token:     "test-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err := db.Create(bookingLink).Error
	require.NoError(t, err)

	// Generate expired token
	token := generateToken(bookingLink.ID, bookingLink.TenantID, "test-secret-key-32-chars-longrm middleware/token_auth_test.go", -1*time.Hour)

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
	assert.Contains(t, response["error"], "Token has expired")
}

func TestValidateBookingToken_BookingLinkNotFound(t *testing.T) {
	db := setupTestDB(t)
	middleware := setupTestServices(db)

	// Generate token for non-existent booking link
	token := generateToken(999, 1, "test-secret-key-32-chars-longrm middleware/token_auth_test.go", 24*time.Hour)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/booking/:token", middleware.ValidateBookingToken(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/booking/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Booking link not found")
}

func TestValidateBookingToken_BlacklistedToken(t *testing.T) {
	db := setupTestDB(t)
	middleware := setupTestServices(db)

	// Create test booking link
	bookingLink := &entities.BookingLink{
		TenantID:  1,
		Token:     "test-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err := db.Create(bookingLink).Error
	require.NoError(t, err)

	// Generate valid token
	token := generateToken(bookingLink.ID, bookingLink.TenantID, "test-secret-key-32-chars-longrm middleware/token_auth_test.go", 24*time.Hour)

	// Blacklist the token
	err = middleware.BlacklistToken(token, time.Now().Add(24*time.Hour))
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
	assert.Contains(t, response["error"], "Token has been revoked")
}

func TestBlacklistToken_Success(t *testing.T) {
	db := setupTestDB(t)
	middleware := setupTestServices(db)

	token := "test-token-123"
	expiresAt := time.Now().Add(24 * time.Hour)

	err := middleware.BlacklistToken(token, expiresAt)
	require.NoError(t, err)

	// Verify token is blacklisted
	var blacklist entities.TokenBlacklist
	err = db.Where("token = ?", token).First(&blacklist).Error
	require.NoError(t, err)
	assert.Equal(t, token, blacklist.Token)
	assert.WithinDuration(t, expiresAt, blacklist.ExpiresAt, time.Second)
}

func TestGetBookingClaims_Success(t *testing.T) {
	// Create a gin context with claims
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	expectedClaims := &BookingTokenClaims{
		BookingLinkID: 123,
		TenantID:      456,
	}

	c.Set("booking_claims", expectedClaims)

	claims, err := GetBookingClaims(c)
	require.NoError(t, err)
	assert.Equal(t, expectedClaims.BookingLinkID, claims.BookingLinkID)
	assert.Equal(t, expectedClaims.TenantID, claims.TenantID)
}
