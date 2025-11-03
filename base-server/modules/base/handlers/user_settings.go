package handlers

import (
	"net/http"

	"github.com/ae-base-server/internal/models"
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserSettingsHandlers struct {
	db     *gorm.DB
	logger core.Logger
}

func NewUserSettingsHandlers(db *gorm.DB, logger core.Logger) *UserSettingsHandlers {
	return &UserSettingsHandlers{
		db:     db,
		logger: logger,
	}
}

// @Summary Get user settings
// @ID getUserSettings
// @Description Get settings for the authenticated user (creates default if not found)
// @Tags user-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.UserSettings}
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/user-settings [get]
func (h *UserSettingsHandlers) GetUserSettings(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "User ID not found in context"))
		return
	}

	var settings models.UserSettings
	if err := h.db.Where("user_id = ?", userID).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default settings if not found
			settings = models.UserSettings{
				UserID:   userID.(uint),
				Language: "en",
				Timezone: "UTC",
				Theme:    "light",
				Settings: "{}",
			}
			if err := h.db.Create(&settings).Error; err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to create default settings", err.Error()))
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve settings", err.Error()))
			return
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse("User settings retrieved successfully", settings))
}

// @Summary Update user settings
// @ID updateUserSettings
// @Description Update settings for the authenticated user
// @Tags user-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param settings body models.UserSettingsUpdateRequest true "Updated settings"
// @Success 200 {object} models.APIResponse{data=models.UserSettings}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/user-settings [put]
func (h *UserSettingsHandlers) UpdateUserSettings(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "User ID not found in context"))
		return
	}

	var req models.UserSettingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	var settings models.UserSettings
	if err := h.db.Where("user_id = ?", userID).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new settings if not found
			settings = models.UserSettings{
				UserID: userID.(uint),
			}
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve settings", err.Error()))
			return
		}
	}

	// Update fields
	if req.Language != "" {
		settings.Language = req.Language
	}
	if req.Timezone != "" {
		settings.Timezone = req.Timezone
	}
	if req.Theme != "" {
		settings.Theme = req.Theme
	}
	if req.Settings != "" {
		settings.Settings = req.Settings
	}

	if err := h.db.Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to update settings", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("User settings updated successfully", settings))
}

// @Summary Reset user settings
// @ID resetUserSettings
// @Description Reset user settings to default values
// @Tags user-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.UserSettings}
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/user-settings/reset [post]
func (h *UserSettingsHandlers) ResetUserSettings(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "User ID not found in context"))
		return
	}

	var settings models.UserSettings
	if err := h.db.Where("user_id = ?", userID).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Settings not found", "No settings found for user"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve settings", err.Error()))
		return
	}

	// Reset to defaults
	settings.Language = "en"
	settings.Timezone = "UTC"
	settings.Theme = "light"
	settings.Settings = "{}"

	if err := h.db.Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to reset settings", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("User settings reset successfully", settings))
}
