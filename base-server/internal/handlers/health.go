package handlers

import (
	"net/http"
	"time"

	"github.com/ae-base-server/internal/config"
	"github.com/ae-base-server/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db  *gorm.DB
	cfg config.Config
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *gorm.DB, cfg config.Config) *HealthHandler {
	return &HealthHandler{
		db:  db,
		cfg: cfg,
	}
}

// Health performs a health check
// DISABLED-SWAGGER: @Summary Health check
// DISABLED-SWAGGER: @Description Check the health status of the API and database
// DISABLED-SWAGGER: @Tags health
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Success 200 {object} models.HealthResponse
// DISABLED-SWAGGER: @Failure 503 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	// Check database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "Database connection error",
			Details: err.Error(),
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "Database ping failed",
			Details: err.Error(),
		})
		return
	}

	response := models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0", // TODO: Get from config
		Database:  "connected",
		Settings: models.HealthSettingsResponse{
			MockEmail:        h.cfg.Email.MockEmail,
			RateLimitEnabled: h.cfg.RateLimit.Enabled,
		},
	}

	c.JSON(http.StatusOK, response)
}

// Ping performs a simple ping check
// DISABLED-SWAGGER: @Summary Ping check
// DISABLED-SWAGGER: @Description Simple ping endpoint
// DISABLED-SWAGGER: @Tags health
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Success 200 {string} string "pong"
// DISABLED-SWAGGER: @Router /api/v1/ping [get]
func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
