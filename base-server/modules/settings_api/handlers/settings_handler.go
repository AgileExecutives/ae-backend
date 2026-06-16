package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ae/base-server/pkg/settings/entities"
	"github.com/ae/base-server/pkg/settings/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DomainSettingsResponse is the API response for a full settings domain.
type DomainSettingsResponse struct {
	Domain   string                 `json:"domain" example:"invoice"`
	Settings map[string]interface{} `json:"settings" swaggertype:"object"`
}

// SettingResponse is the API response for a single setting.
type SettingResponse struct {
	Domain string                 `json:"domain" example:"invoice"`
	Key    string                 `json:"key" example:"invoice_prefix"`
	Data   map[string]interface{} `json:"data" swaggertype:"object"`
}

// UpdateKeysResponse is the API response for bulk domain updates.
type UpdateKeysResponse struct {
	Message     string   `json:"message" example:"Settings updated successfully"`
	Domain      string   `json:"domain" example:"invoice"`
	UpdatedKeys []string `json:"updated_keys" example:"invoice_prefix,next_invoice_number"`
}

// SimpleMessageResponse is a basic message response.
type SimpleMessageResponse struct {
	Message string `json:"message" example:"Setting updated successfully"`
	Domain  string `json:"domain" example:"invoice"`
	Key     string `json:"key" example:"invoice_prefix"`
}

// ErrorResponse is a minimal error response used by this module.
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid tenant ID"`
}

// SettingsHandler handles settings API requests.
//
// NOTE: The routes use "/settings/organizations/:tenant_id" for backward compatibility
// with Unburdy. Here, "organization" is effectively the tenant scope.
type SettingsHandler struct {
	repo *repository.SettingsRepository
}

// NewSettingsHandler creates a new settings handler.
func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{repo: repository.NewSettingsRepository(db)}
}

// GetDomainSettings returns all settings for a domain.
// @Summary Get domain settings
// @Description Get all settings for a domain for a tenant
// @Tags settings
// @Produce json
// @Param tenant_id path int true "Tenant ID"
// @Param domain path string true "Domain"
// @ID getDomainSettings
// @Success 200 {object} DomainSettingsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /settings/organizations/{tenant_id}/domains/{domain} [get]
// @Security BearerAuth
func (h *SettingsHandler) GetDomainSettings(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	domain := c.Param("domain")

	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid tenant ID"})
		return
	}

	settings, err := h.repo.GetDomainSettings(uint(tenantID), domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	result := make(map[string]interface{})
	for _, setting := range settings {
		var data map[string]interface{}
		if err := json.Unmarshal(setting.Data, &data); err == nil {
			for k, v := range data {
				result[k] = v
			}
		}
	}

	c.JSON(http.StatusOK, DomainSettingsResponse{Domain: domain, Settings: result})
}

// GetSetting returns a specific setting.
// @Summary Get a setting
// @Description Get a specific setting for a tenant, domain and key
// @Tags settings
// @Produce json
// @Param tenant_id path int true "Tenant ID"
// @Param domain path string true "Domain"
// @Param key path string true "Key"
// @ID getSetting
// @Success 200 {object} SettingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /settings/organizations/{tenant_id}/domains/{domain}/{key} [get]
// @Security BearerAuth
func (h *SettingsHandler) GetSetting(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	domain := c.Param("domain")
	key := c.Param("key")

	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid tenant ID"})
		return
	}

	setting, err := h.repo.GetSetting(uint(tenantID), domain, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	if setting == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Setting not found"})
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(setting.Data, &data); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to parse setting data"})
		return
	}

	c.JSON(http.StatusOK, SettingResponse{Domain: domain, Key: key, Data: data})
}

// UpdateSetting updates a setting.
// @Summary Update a setting
// @Description Create or update a specific setting for a tenant (value is an arbitrary JSON object)
// @Tags settings
// @Accept json
// @Produce json
// @Param tenant_id path int true "Tenant ID"
// @Param domain path string true "Domain"
// @Param key path string true "Key"
// @Param request body object true "Setting payload (arbitrary JSON)"
// @ID updateSetting
// @Success 200 {object} SimpleMessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /settings/organizations/{tenant_id}/domains/{domain}/{key} [put]
// @Router /settings/organizations/{tenant_id}/domains/{domain}/{key} [post]
// @Security BearerAuth
func (h *SettingsHandler) UpdateSetting(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	domain := c.Param("domain")
	key := c.Param("key")

	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid tenant ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	settingDef, err := h.repo.GetSettingDefinition(domain, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	if settingDef == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Setting definition not found"})
		return
	}

	setting, err := h.repo.GetSetting(uint(tenantID), domain, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	dataJSON, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to encode data"})
		return
	}

	if setting == nil {
		setting = &entities.Setting{
			TenantID:            uint(tenantID),
			Domain:              domain,
			Key:                 key,
			Version:             settingDef.Version,
			Data:                dataJSON,
			SettingDefinitionID: settingDef.ID,
		}
	} else {
		setting.Data = dataJSON
		setting.Version = settingDef.Version
	}

	if err := h.repo.SetSetting(setting); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SimpleMessageResponse{Message: "Setting updated successfully", Domain: domain, Key: key})
}

// UpdateDomainSettings updates multiple settings for a domain at once.
// @Summary Update domain settings
// @Description Create or update multiple settings for a domain for a tenant (payload is an arbitrary JSON object of key->value)
// @Tags settings
// @Accept json
// @Produce json
// @Param tenant_id path int true "Tenant ID"
// @Param domain path string true "Domain"
// @Param request body object true "Domain payload (arbitrary JSON of key->value)"
// @ID updateDomainSettings
// @Success 200 {object} UpdateKeysResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /settings/organizations/{tenant_id}/domains/{domain} [post]
// @Router /settings/organizations/{tenant_id}/domains/{domain} [put]
// @Security BearerAuth
func (h *SettingsHandler) UpdateDomainSettings(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	domain := c.Param("domain")

	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid tenant ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	updatedKeys := make([]string, 0, len(req))
	for key, value := range req {
		settingDef, err := h.repo.GetSettingDefinition(domain, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get setting definition for " + key})
			return
		}
		if settingDef == nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Setting definition not found for " + key})
			return
		}

		setting, err := h.repo.GetSetting(uint(tenantID), domain, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get setting for " + key})
			return
		}

		dataJSON, err := json.Marshal(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to encode data for " + key})
			return
		}

		if setting == nil {
			setting = &entities.Setting{
				TenantID:            uint(tenantID),
				Domain:              domain,
				Key:                 key,
				Version:             settingDef.Version,
				Data:                dataJSON,
				SettingDefinitionID: settingDef.ID,
			}
		} else {
			setting.Data = dataJSON
			setting.Version = settingDef.Version
		}

		if err := h.repo.SetSetting(setting); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to save setting for " + key})
			return
		}

		updatedKeys = append(updatedKeys, key)
	}

	c.JSON(http.StatusOK, UpdateKeysResponse{Message: "Settings updated successfully", Domain: domain, UpdatedKeys: updatedKeys})
}
