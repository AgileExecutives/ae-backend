package settingsapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ae-base-server/pkg/settings/entities"
	"github.com/ae-base-server/pkg/settings/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SettingsHandler handles settings API requests
type SettingsHandler struct {
	repo *repository.SettingsRepository
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{
		repo: repository.NewSettingsRepository(db),
	}
}

// GetDomainSettings returns all settings for a domain
func (h *SettingsHandler) GetDomainSettings(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	domain := c.Param("domain")

	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	settings, err := h.repo.GetDomainSettings(uint(tenantID), domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to map for easier frontend consumption
	result := make(map[string]interface{})
	for _, setting := range settings {
		var data map[string]interface{}
		if err := json.Unmarshal(setting.Data, &data); err == nil {
			// Flatten the data into the result
			for k, v := range data {
				result[k] = v
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"domain":   domain,
		"settings": result,
	})
}

// GetSetting returns a specific setting
func (h *SettingsHandler) GetSetting(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	domain := c.Param("domain")
	key := c.Param("key")

	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	setting, err := h.repo.GetSetting(uint(tenantID), domain, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if setting == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(setting.Data, &data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse setting data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domain": domain,
		"key":    key,
		"data":   data,
	})
}

// UpdateSetting updates a setting
func (h *SettingsHandler) UpdateSetting(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	domain := c.Param("domain")
	key := c.Param("key")

	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get the setting definition to get the version
	settingDef, err := h.repo.GetSettingDefinition(domain, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if settingDef == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting definition not found"})
		return
	}

	// Get or create the setting
	setting, err := h.repo.GetSetting(uint(tenantID), domain, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Marshal the request data to JSON
	dataJSON, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to encode data"})
		return
	}

	if setting == nil {
		// Create new setting
		setting = &entities.Setting{
			TenantID:            uint(tenantID),
			Domain:              domain,
			Key:                 key,
			Version:             settingDef.Version,
			Data:                dataJSON,
			SettingDefinitionID: settingDef.ID,
		}
	} else {
		// Update existing setting
		setting.Data = dataJSON
		setting.Version = settingDef.Version
	}

	if err := h.repo.SetSetting(setting); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Setting updated successfully",
		"domain":  domain,
		"key":     key,
	})
}

// UpdateDomainSettings updates multiple settings for a domain at once
func (h *SettingsHandler) UpdateDomainSettings(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	domain := c.Param("domain")

	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// For each key in the request, update the corresponding setting
	updatedKeys := []string{}
	for key, value := range req {
		// Get the setting definition
		settingDef, err := h.repo.GetSettingDefinition(domain, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get setting definition for " + key})
			return
		}

		if settingDef == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Setting definition not found for " + key})
			return
		}

		// Get or create the setting
		setting, err := h.repo.GetSetting(uint(tenantID), domain, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get setting for " + key})
			return
		}

		// Marshal the value to JSON
		dataJSON, err := json.Marshal(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to encode data for " + key})
			return
		}

		if setting == nil {
			// Create new setting
			setting = &entities.Setting{
				TenantID:            uint(tenantID),
				Domain:              domain,
				Key:                 key,
				Version:             settingDef.Version,
				Data:                dataJSON,
				SettingDefinitionID: settingDef.ID,
			}
		} else {
			// Update existing setting
			setting.Data = dataJSON
			setting.Version = settingDef.Version
		}

		if err := h.repo.SetSetting(setting); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save setting for " + key})
			return
		}

		updatedKeys = append(updatedKeys, key)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Settings updated successfully",
		"domain":       domain,
		"updated_keys": updatedKeys,
	})
}

// RegisterRoutes registers the settings API routes
func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	handler := NewSettingsHandler(db)

	settingsGroup := router.Group("/api/v1/settings")
	{
		// Tenant-based routes (organization_id is actually tenant_id)
		tenantGroup := settingsGroup.Group("/organizations/:tenant_id")
		{
			tenantGroup.GET("/domains/:domain", handler.GetDomainSettings)
			tenantGroup.POST("/domains/:domain", handler.UpdateDomainSettings)
			tenantGroup.PUT("/domains/:domain", handler.UpdateDomainSettings)
			tenantGroup.GET("/domains/:domain/:key", handler.GetSetting)
			tenantGroup.POST("/domains/:domain/:key", handler.UpdateSetting)
			tenantGroup.PUT("/domains/:domain/:key", handler.UpdateSetting)
		}
	}
}
