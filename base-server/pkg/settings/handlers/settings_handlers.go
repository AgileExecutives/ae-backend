package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ae-base-server/pkg/settings/entities"
	"github.com/ae-base-server/pkg/settings/services"
	"github.com/gin-gonic/gin"
)

// SettingsHandler handles HTTP requests for the settings system
type SettingsHandler struct {
	service *services.SettingsService
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(service *services.SettingsService) *SettingsHandler {
	return &SettingsHandler{service: service}
}

// @Summary Settings system health check
// @ID settingsHealthCheck
// @Description Check the health status of the settings system
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} entities.HealthResponse
// @Failure 500 {object} map[string]string
// @Router /settings/health [get]
func (h *SettingsHandler) HealthCheck(c *gin.Context) {
	health, err := h.service.HealthCheck()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Health check failed"})
		return
	}
	c.JSON(http.StatusOK, health)
}

// @Summary Get registered settings modules
// @ID getRegisteredModules
// @Description Get list of all registered settings modules
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} entities.ModuleListResponse
// @Failure 500 {object} map[string]string
// @Router /settings/modules [get]
func (h *SettingsHandler) GetRegisteredModules(c *gin.Context) {
	modules := h.service.GetModules()
	c.JSON(http.StatusOK, entities.ModuleListResponse{Modules: modules})
}

// @Summary Get settings system version
// @ID getSettingsVersion
// @Description Get settings system version information
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /settings/version [get]
func (h *SettingsHandler) GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": "1.0.0",
		"system":  "settings",
	})
}

// @Summary Get organization settings
// @ID getOrganizationSettings
// @Description Get all settings for an organization grouped by domain
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path string true "Organization ID"
// @Success 200 {object} entities.SettingsResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/organizations/{organization_id} [get]
func (h *SettingsHandler) GetOrganizationSettings(c *gin.Context) {
	tenantID, organizationID, err := h.extractTenantAndOrg(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.service.GetAllSettings(tenantID, organizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve settings"})
		return
	}

	c.JSON(http.StatusOK, entities.SettingsResponse{Settings: settings})
}

// @Summary Set organization setting
// @ID setOrganizationSetting
// @Description Set a single setting for an organization
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path string true "Organization ID"
// @Param setting body entities.SettingRequest true "Setting data"
// @Success 201 {object} entities.SettingResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/organizations/{organization_id} [post]
func (h *SettingsHandler) SetOrganizationSetting(c *gin.Context) {
	tenantID, organizationID, err := h.extractTenantAndOrg(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req entities.SettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.service.SetSetting(tenantID, organizationID, req.Domain, req.Key, req.Value, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set setting"})
		return
	}

	// Return the created setting
	setting := entities.SettingResponse{
		TenantID:       tenantID,
		OrganizationID: organizationID,
		Domain:         req.Domain,
		Key:            req.Key,
		Value:          req.Value.(string),
		Type:           req.Type,
	}

	c.JSON(http.StatusCreated, setting)
}

// @Summary Update organization setting
// @ID updateOrganizationSetting
// @Description Update a specific setting for an organization
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path string true "Organization ID"
// @Param domain path string true "Settings domain"
// @Param key path string true "Setting key"
// @Param setting body map[string]interface{} true "Updated setting data"
// @Success 200 {object} entities.SettingResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/organizations/{organization_id}/{domain}/{key} [put]
func (h *SettingsHandler) UpdateOrganizationSetting(c *gin.Context) {
	tenantID, organizationID, err := h.extractTenantAndOrg(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domain := c.Param("domain")
	key := c.Param("key")

	var req struct {
		Type  string      `json:"type" binding:"required"`
		Value interface{} `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.service.SetSetting(tenantID, organizationID, domain, key, req.Value, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting"})
		return
	}

	setting := entities.SettingResponse{
		TenantID:       tenantID,
		OrganizationID: organizationID,
		Domain:         domain,
		Key:            key,
		Value:          req.Value.(string),
		Type:           req.Type,
	}

	c.JSON(http.StatusOK, setting)
}

// @Summary Delete organization setting
// @ID deleteOrganizationSetting
// @Description Delete a specific setting for an organization
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path string true "Organization ID"
// @Param domain path string true "Settings domain"
// @Param key path string true "Setting key"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/organizations/{organization_id}/{domain}/{key} [delete]
func (h *SettingsHandler) DeleteOrganizationSetting(c *gin.Context) {
	tenantID, organizationID, err := h.extractTenantAndOrg(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domain := c.Param("domain")
	key := c.Param("key")

	err = h.service.DeleteSetting(tenantID, organizationID, domain, key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Setting deleted successfully"})
}

// @Summary Bulk set organization settings
// @ID bulkSetOrganizationSettings
// @Description Set multiple settings for an organization
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path string true "Organization ID"
// @Param settings body entities.BulkSettingRequest true "Multiple settings data"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/organizations/{organization_id}/bulk [post]
func (h *SettingsHandler) BulkSetOrganizationSettings(c *gin.Context) {
	tenantID, organizationID, err := h.extractTenantAndOrg(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req entities.BulkSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	created := 0
	var errors []string

	for _, setting := range req.Settings {
		err = h.service.SetSetting(tenantID, organizationID, setting.Domain, setting.Key, setting.Value, setting.Type)
		if err != nil {
			errors = append(errors, err.Error())
		} else {
			created++
		}
	}

	response := gin.H{
		"success": created > 0,
		"created": created,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	c.JSON(http.StatusCreated, response)
}

// @Summary Get organization domains
// @ID getOrganizationDomains
// @Description Get list of available settings domains for an organization
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path string true "Organization ID"
// @Success 200 {object} entities.DomainResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/organizations/{organization_id}/domains [get]
func (h *SettingsHandler) GetOrganizationDomains(c *gin.Context) {
	tenantID, organizationID, err := h.extractTenantAndOrg(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domains, err := h.service.GetDomains(tenantID, organizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve domains"})
		return
	}

	c.JSON(http.StatusOK, entities.DomainResponse{Domains: domains})
}

// @Summary Get domain settings
// @ID getOrganizationDomainSettings
// @Description Get all settings for a specific domain
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path string true "Organization ID"
// @Param domain path string true "Settings domain"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/organizations/{organization_id}/domains/{domain} [get]
func (h *SettingsHandler) GetOrganizationDomainSettings(c *gin.Context) {
	tenantID, organizationID, err := h.extractTenantAndOrg(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domain := c.Param("domain")
	settings, err := h.service.GetDomainSettings(tenantID, organizationID, domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve domain settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

// @Summary Set domain settings
// @ID setOrganizationDomainSettings
// @Description Set multiple settings for a specific domain
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path string true "Organization ID"
// @Param domain path string true "Settings domain"
// @Param settings body entities.DomainSettingsRequest true "Domain settings"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/organizations/{organization_id}/domains/{domain} [post]
func (h *SettingsHandler) SetOrganizationDomainSettings(c *gin.Context) {
	tenantID, organizationID, err := h.extractTenantAndOrg(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domain := c.Param("domain")
	var req entities.DomainSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	created := 0
	var errors []string

	for key, value := range req.Settings {
		// Determine type from value
		valueType := "string"
		switch value.(type) {
		case bool:
			valueType = "bool"
		case int, int64, float64:
			valueType = "int"
		case map[string]interface{}, []interface{}:
			valueType = "json"
		}

		err = h.service.SetSetting(tenantID, organizationID, domain, key, value, valueType)
		if err != nil {
			errors = append(errors, err.Error())
		} else {
			created++
		}
	}

	response := gin.H{
		"success": created > 0,
		"created": created,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	c.JSON(http.StatusCreated, response)
}

// @Summary Delete domain settings
// @ID deleteOrganizationDomainSettings
// @Description Delete all settings for a specific domain
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path string true "Organization ID"
// @Param domain path string true "Settings domain"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/organizations/{organization_id}/domains/{domain} [delete]
func (h *SettingsHandler) DeleteOrganizationDomainSettings(c *gin.Context) {
	tenantID, organizationID, err := h.extractTenantAndOrg(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domain := c.Param("domain")
	err = h.service.DeleteDomainSettings(tenantID, organizationID, domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete domain settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Domain settings deleted successfully"})
}

// @Summary Validate settings
// @ID validateOrganizationSettings
// @Description Validate settings against their schema definitions
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path string true "Organization ID"
// @Param request body entities.ValidationRequest true "Settings to validate"
// @Success 200 {object} entities.ValidationResponse
// @Failure 400 {object} entities.ValidationResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/organizations/{organization_id}/validate [post]
func (h *SettingsHandler) ValidateOrganizationSettings(c *gin.Context) {
	var req entities.ValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	valid, errors := h.service.ValidateSettings(req.Domain, req.Settings)

	response := entities.ValidationResponse{
		Valid:  valid,
		Errors: errors,
	}

	if valid {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusBadRequest, response)
	}
}

// @Summary Export organization settings
// @ID exportOrganizationSettings
// @Description Export all settings for an organization
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path string true "Organization ID"
// @Param format query string false "Export format" Enums(json, yaml)
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/organizations/{organization_id}/export [get]
func (h *SettingsHandler) ExportOrganizationSettings(c *gin.Context) {
	tenantID, organizationID, err := h.extractTenantAndOrg(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.service.GetAllSettings(tenantID, organizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export settings"})
		return
	}

	export := gin.H{
		"organization_id": organizationID,
		"export_date":     "2025-01-09T10:00:00Z",
		"version":         "1.0.0",
		"settings":        settings,
	}

	c.JSON(http.StatusOK, export)
}

// @Summary Import organization settings
// @ID importOrganizationSettings
// @Description Import settings for an organization
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path string true "Organization ID"
// @Param import body map[string]interface{} true "Settings to import"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/organizations/{organization_id}/import [post]
func (h *SettingsHandler) ImportOrganizationSettings(c *gin.Context) {
	tenantID, organizationID, err := h.extractTenantAndOrg(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Settings map[string]map[string]interface{} `json:"settings" binding:"required"`
		Merge    bool                              `json:"merge"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	imported := 0
	skipped := 0

	for domain, domainSettings := range req.Settings {
		for key, value := range domainSettings {
			// Determine type from value
			valueType := "string"
			switch value.(type) {
			case bool:
				valueType = "bool"
			case int, int64, float64:
				valueType = "int"
			case map[string]interface{}, []interface{}:
				valueType = "json"
			}

			err = h.service.SetSetting(tenantID, organizationID, domain, key, value, valueType)
			if err != nil {
				skipped++
			} else {
				imported++
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":  imported > 0,
		"imported": imported,
		"skipped":  skipped,
	})
}

// Helper method to extract tenant and organization IDs
func (h *SettingsHandler) extractTenantAndOrg(c *gin.Context) (uint, string, error) {
	organizationID := c.Param("organization_id")
	if organizationID == "" {
		return 0, "", errors.New("organization_id is required")
	}

	// For now, use a default tenant ID - in production this should come from authentication context
	tenantID := uint(1)

	return tenantID, organizationID, nil
}
