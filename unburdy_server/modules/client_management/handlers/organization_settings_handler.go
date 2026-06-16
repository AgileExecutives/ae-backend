package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	baseAPI "github.com/ae/base-server/api"
	settingsEntities "github.com/ae/base-server/pkg/settings/entities"
	settingsRepo "github.com/ae/base-server/pkg/settings/repository"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/unburdy-server-api/internal/models"
	"github.com/unburdy/unburdy-server-api/modules/client_management/services"
	cmsettings "github.com/unburdy/unburdy-server-api/modules/client_management/settings"
	"gorm.io/gorm"
)

// OrganizationSettingsHandler exposes compatibility endpoints used by the
// Unburdy frontend for organization-scoped settings reads.
type OrganizationSettingsHandler struct {
	repo          *settingsRepo.SettingsRepository
	clientService *services.ClientService
}

// RegistrationSettingsData describes the public registration page configuration.
type RegistrationSettingsData struct {
	RequiredFields           []string `json:"required_fields"`
	OptionalFields           []string `json:"optional_fields"`
	EmailVerificationEnabled bool     `json:"email_verification_enabled"`
	CostProvidersEnabled     bool     `json:"cost_providers_enabled"`
	RegistrationHeadline     string   `json:"registration_headline,omitempty"`
	RegistrationIntroText    string   `json:"registration_intro_text,omitempty"`
}

// RegistrationSettingsEnvelope wraps the registration settings with a domain name.
type RegistrationSettingsEnvelope struct {
	Domain   string                   `json:"domain"`
	Settings RegistrationSettingsData `json:"settings"`
}

// UpdateRegistrationSettingsRequest is the request payload for updating registration settings.
type UpdateRegistrationSettingsRequest struct {
	Settings RegistrationSettingsData `json:"settings"`
}

const registrationSettingsDomain = "registration"
const registrationSettingsKey = "config"

// NewOrganizationSettingsHandler creates a new compatibility settings handler.
func NewOrganizationSettingsHandler(db *gorm.DB, clientService *services.ClientService) *OrganizationSettingsHandler {
	return &OrganizationSettingsHandler{
		repo:          settingsRepo.NewSettingsRepository(db),
		clientService: clientService,
	}
}

// GetBillingSettings returns merged billing settings for the authenticated tenant.
func (h *OrganizationSettingsHandler) GetBillingSettings(c *gin.Context) {
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		if _, err := strconv.ParseUint(orgIDStr, 10, 32); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Bad request", "invalid organization_id"))
			return
		}
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	settingsData, err := h.getMergedBillingSettings(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Billing settings retrieved successfully", gin.H{
		"domain":   cmsettings.DomainBilling,
		"settings": settingsData,
	}))
}

// GetRegistrationSettings returns a stable registration configuration payload.
// @Summary Get registration settings
// @Description Retrieve registration page settings for the authenticated tenant.
// @Tags clients
// @ID getRegistrationSettings
// @Produce json
// @Success 200 {object} models.APIResponse{data=RegistrationSettingsEnvelope} "Registration settings retrieved successfully"
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /organization/settings/registration [get]
func (h *OrganizationSettingsHandler) GetRegistrationSettings(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	settingsData, err := h.getRegistrationSettings(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Registration settings retrieved successfully", gin.H{
		"domain":   registrationSettingsDomain,
		"settings": settingsData,
	}))
}

// GetRegistrationSettingsWithToken returns registration settings for the tenant
// associated with a valid registration token. No bearer auth is required.
// @Summary Get registration settings with token
// @Description Retrieve registration page settings for the tenant associated with the registration token. No Bearer auth required.
// @Tags clients
// @ID getRegistrationSettingsWithToken
// @Produce json
// @Param token path string true "Registration token"
// @Success 200 {object} models.APIResponse{data=RegistrationSettingsData} "Registration settings retrieved successfully"
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /clients/registration-settings/{token} [get]
func (h *OrganizationSettingsHandler) GetRegistrationSettingsWithToken(c *gin.Context) {
	token := c.Param("token")

	regToken, err := h.clientService.ValidateRegistrationToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Access denied", err.Error()))
		return
	}

	settingsData, err := h.getRegistrationSettings(regToken.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Registration settings retrieved successfully", settingsData))
}

// UpdateRegistrationSettings stores registration settings for the authenticated tenant.
// @Summary Update registration settings
// @Description Update registration page settings for the authenticated tenant.
// @Tags clients
// @ID updateRegistrationSettings
// @Accept json
// @Produce json
// @Param settings body UpdateRegistrationSettingsRequest true "Registration settings payload"
// @Success 200 {object} models.APIResponse{data=RegistrationSettingsEnvelope} "Registration settings updated successfully"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /organization/settings/registration [put]
func (h *OrganizationSettingsHandler) UpdateRegistrationSettings(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Bad request", "invalid request body"))
		return
	}

	payload := normalizeRegistrationSettingsPayload(req)
	if payload == nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Bad request", "invalid registration settings payload"))
		return
	}

	settingsData, err := h.getRegistrationSettings(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	mergeSettings(settingsData, payload)

	dataJSON, err := json.Marshal(settingsData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	settingDef, err := h.ensureRegistrationSettingDefinition()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	setting, err := h.repo.GetSetting(tenantID, registrationSettingsDomain, registrationSettingsKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	if setting == nil {
		setting = &settingsEntities.Setting{
			TenantID:            tenantID,
			Domain:              registrationSettingsDomain,
			Key:                 registrationSettingsKey,
			Version:             settingDef.Version,
			Data:                dataJSON,
			SettingDefinitionID: settingDef.ID,
		}
	} else {
		setting.Data = dataJSON
		setting.Version = settingDef.Version
		setting.SettingDefinitionID = settingDef.ID
	}

	if err := h.repo.SetSetting(setting); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Registration settings updated successfully", gin.H{
		"domain":   registrationSettingsDomain,
		"settings": settingsData,
	}))
}

func (h *OrganizationSettingsHandler) getMergedBillingSettings(tenantID uint) (map[string]interface{}, error) {
	merged := make(map[string]interface{})

	for _, def := range cmsettings.GetBillingSettingsDefinitions() {
		data := def.Data

		setting, err := h.repo.GetSetting(tenantID, def.Domain, def.Key)
		if err != nil {
			return nil, err
		}

		if setting != nil {
			var stored map[string]interface{}
			if err := json.Unmarshal(setting.Data, &stored); err != nil {
				return nil, err
			}
			data = stored
		}

		for key, value := range data {
			merged[key] = value
		}
	}

	return merged, nil
}

func (h *OrganizationSettingsHandler) getRegistrationSettings(tenantID uint) (map[string]interface{}, error) {
	settingsData := defaultRegistrationSettings()

	setting, err := h.repo.GetSetting(tenantID, registrationSettingsDomain, registrationSettingsKey)
	if err != nil {
		return nil, err
	}

	if setting == nil {
		return settingsData, nil
	}

	var stored map[string]interface{}
	if err := json.Unmarshal(setting.Data, &stored); err != nil {
		return nil, err
	}

	mergeSettings(settingsData, stored)
	return settingsData, nil
}

func defaultRegistrationSettings() map[string]interface{} {
	return map[string]interface{}{
		"required_fields": []interface{}{"first_name", "last_name", "email"},
		"optional_fields": []interface{}{
			"phone",
			"date_of_birth",
			"gender",
			"street_address",
			"zip",
			"city",
			"contact_first_name",
			"contact_last_name",
			"contact_email",
			"contact_phone",
			"notes",
			"timezone",
		},
		"email_verification_enabled": true,
		"cost_providers_enabled":      true,
	}
}

func mergeSettings(base map[string]interface{}, overrides map[string]interface{}) {
	for key, value := range overrides {
		base[key] = value
	}
}

func normalizeRegistrationSettingsPayload(req map[string]interface{}) map[string]interface{} {
	if settings, ok := asStringMap(req["settings"]); ok {
		return settings
	}

	if data, ok := asStringMap(req["data"]); ok {
		if settings, ok := asStringMap(data["settings"]); ok {
			return settings
		}
		return data
	}

	return req
}

func asStringMap(value interface{}) (map[string]interface{}, bool) {
	settings, ok := value.(map[string]interface{})
	if !ok {
		return nil, false
	}
	return settings, true
}

func (h *OrganizationSettingsHandler) ensureRegistrationSettingDefinition() (*settingsEntities.SettingDefinition, error) {
	settingDef, err := h.repo.GetSettingDefinition(registrationSettingsDomain, registrationSettingsKey)
	if err != nil {
		return nil, err
	}

	if settingDef != nil {
		return settingDef, nil
	}

	schemaJSON, err := json.Marshal(map[string]interface{}{
		"type":                 "object",
		"additionalProperties": true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration settings schema: %w", err)
	}

	dataJSON, err := json.Marshal(defaultRegistrationSettings())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration settings defaults: %w", err)
	}

	settingDef = &settingsEntities.SettingDefinition{
		Domain:  registrationSettingsDomain,
		Key:     registrationSettingsKey,
		Version: 1,
		Schema:  schemaJSON,
		Data:    dataJSON,
	}

	if err := h.repo.CreateSettingDefinition(settingDef); err != nil {
		return nil, err
	}

	return settingDef, nil
}