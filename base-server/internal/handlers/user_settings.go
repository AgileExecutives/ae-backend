package handlers

import (
	"net/http"

	"github.com/ae-base-server/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserSettingsHandler struct {
	db *gorm.DB
}

// NewUserSettingsHandler creates a new user settings handler
func NewUserSettingsHandler(db *gorm.DB) *UserSettingsHandler {
	return &UserSettingsHandler{db: db}
}

// GetUserSettings retrieves the current user's settings
// DISABLED-SWAGGER: @Summary Get user settings
// DISABLED-SWAGGER: @Description Get the authenticated user's settings
// DISABLED-SWAGGER: @Tags user-settings
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Security BearerAuth
// DISABLED-SWAGGER: @Success 200 {object} models.APIResponse{data=models.UserSettingsResponse}
// DISABLED-SWAGGER: @Failure 401 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Failure 404 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Router /user-settings [get]
func (h *UserSettingsHandler) GetUserSettings(c *gin.Context) {
	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("User not found", "User not authenticated"))
		return
	}
	user := userInterface.(*models.User)

	var userSettings models.UserSettings
	if err := h.db.Where("user_id = ?", user.ID).First(&userSettings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default settings if not found
			userSettings = models.UserSettings{
				UserID:   user.ID,
				Language: "en",
				Timezone: "UTC",
				Theme:    "light",
				Settings: "{}",
			}

			if err := h.db.Create(&userSettings).Error; err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to create user settings", err.Error()))
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve user settings", err.Error()))
			return
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse("User settings retrieved successfully", userSettings.ToResponse()))
}

// UpdateUserSettings updates the current user's settings
// DISABLED-SWAGGER: @Summary Update user settings
// DISABLED-SWAGGER: @Description Update the authenticated user's settings
// DISABLED-SWAGGER: @Tags user-settings
// DISABLED-SWAGGER: @Accept json
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Security BearerAuth
// DISABLED-SWAGGER: @Param request body models.UserSettingsUpdateRequest true "User settings update data"
// DISABLED-SWAGGER: @Success 200 {object} models.APIResponse{data=models.UserSettingsResponse}
// DISABLED-SWAGGER: @Failure 400 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Failure 401 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Router /user-settings [put]
func (h *UserSettingsHandler) UpdateUserSettings(c *gin.Context) {
	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("User not found", "User not authenticated"))
		return
	}
	user := userInterface.(*models.User)

	var req models.UserSettingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	var userSettings models.UserSettings
	if err := h.db.Where("user_id = ?", user.ID).First(&userSettings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new settings if not found
			userSettings = models.UserSettings{
				UserID:   user.ID,
				Language: "en",
				Timezone: "UTC",
				Theme:    "light",
				Settings: "{}",
			}
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve user settings", err.Error()))
			return
		}
	}

	// Update fields if provided
	if req.Language != "" {
		userSettings.Language = req.Language
	}
	if req.Timezone != "" {
		userSettings.Timezone = req.Timezone
	}
	if req.Theme != "" {
		userSettings.Theme = req.Theme
	}
	if req.Settings != "" {
		userSettings.Settings = req.Settings
	}

	if err := h.db.Save(&userSettings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to update user settings", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("User settings updated successfully", userSettings.ToResponse()))
}

// ResetUserSettings resets the current user's settings to defaults
// DISABLED-SWAGGER: @Summary Reset user settings
// DISABLED-SWAGGER: @Description Reset the authenticated user's settings to default values
// DISABLED-SWAGGER: @Tags user-settings
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Security BearerAuth
// DISABLED-SWAGGER: @Success 200 {object} models.APIResponse{data=models.UserSettingsResponse}
// DISABLED-SWAGGER: @Failure 401 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Router /user-settings/reset [post]
func (h *UserSettingsHandler) ResetUserSettings(c *gin.Context) {
	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("User not found", "User not authenticated"))
		return
	}
	user := userInterface.(*models.User)

	var userSettings models.UserSettings
	if err := h.db.Where("user_id = ?", user.ID).First(&userSettings).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve user settings", err.Error()))
			return
		}
	}

	// Reset to default values
	userSettings.UserID = user.ID
	userSettings.Language = "en"
	userSettings.Timezone = "UTC"
	userSettings.Theme = "light"
	userSettings.Settings = "{}"

	if err := h.db.Save(&userSettings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to reset user settings", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("User settings reset to defaults", userSettings.ToResponse()))
}
