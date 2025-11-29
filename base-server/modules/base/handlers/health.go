package handlers

import (
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/ae-base-server/modules/base/models" // Import models for swagger
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandlers provides health check handlers
type HealthHandlers struct {
	db     *gorm.DB
	logger core.Logger
}

// NewHealthHandlers creates new health handlers
func NewHealthHandlers(db *gorm.DB, logger core.Logger) *HealthHandlers {
	return &HealthHandlers{
		db:     db,
		logger: logger,
	}
}

// getEnvAsBool gets environment variable as boolean with fallback to default value
func getEnvAsBool(key string, defaultVal bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultVal
	}
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultVal
}

// HealthCheck performs a health check
// @Summary Health check
// @ID healthCheck
// @Description Check the health status of the API and database
// @Tags health
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Failure 503 {object} models.ErrorResponse
// @Router /health [get]
func (h *HealthHandlers) HealthCheck(c *gin.Context) {
	response := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "2.0",
		"database":  "connected",
	}

	// Check database connectivity
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil || sqlDB.Ping() != nil {
			response["database"] = "disconnected"
			response["status"] = "unhealthy"
			c.JSON(http.StatusServiceUnavailable, response)
			return
		}
	}

	// Add environment configuration for test script
	environment := map[string]interface{}{
		"mock_email":         getEnvAsBool("MOCK_EMAIL", false),
		"rate_limit_enabled": getEnvAsBool("RATE_LIMIT_ENABLED", true),
		"email_verification": getEnvAsBool("FEATURE_EMAIL_VERIFICATION", true),
		"gin_mode":           os.Getenv("GIN_MODE"),
	}

	response["environment"] = environment

	c.JSON(http.StatusOK, response)
}

// Ping simple ping endpoint
// @Summary Ping check
// @ID ping
// @Description Simple ping endpoint
// @Tags health
// @Produce json
// @Success 200 {string} string "pong"
// @Router /ping [get]
func (h *HealthHandlers) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
