package handlers

import (
	"net/http"
	"strconv"
	"strings"

	baseAPI "github.com/ae-base-server/api"
	"github.com/ae-base-server/modules/templates/entities"
	"github.com/ae-base-server/modules/templates/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TemplateHandler handles template-related HTTP requests
type TemplateHandler struct {
	service *services.TemplateService
	db      *gorm.DB
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(service *services.TemplateService, db *gorm.DB) *TemplateHandler {
	return &TemplateHandler{
		service: service,
		db:      db,
	}
}

// CreateTemplate creates a new template
// @Summary Create template
// @Description Create a new email/PDF template with content stored in MinIO
// @Tags Templates
// @ID createTemplate
// @Accept json
// @Produce json
//
//	@Param request body services.CreateTemplateRequest true "Template creation request" example({
//	  "name": "Welcome Email Template",
//	  "description": "Email template for welcoming new users",
//	  "template_type": "email",
//	  "channel": "EMAIL",
//	  "subject": "Welcome to {{.OrganizationName}}!",
//	  "content": "<h1>Hello {{.FirstName}}!</h1><p>Welcome to {{.OrganizationName}}.</p>",
//	  "variables": ["FirstName", "OrganizationName"]
//	})
//
// @Success 201 {object} entities.TemplateAPIResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates [post]
// @Security BearerAuth
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	// Get organization ID from auth middleware
	var organizationID *uint
	if user, err := baseAPI.GetUser(c); err == nil && user.OrganizationID > 0 {
		orgIDVal := user.OrganizationID
		organizationID = &orgIDVal
	}

	var req services.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.TenantID = uint(tenantID)
	req.OrganizationID = organizationID

	tmpl, err := h.service.CreateTemplate(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entities.TemplateAPIResponse{
		Success: true,
		Message: "Template created successfully",
		Data:    tmpl.ToResponse(),
	})
}

// GetTemplate retrieves a template by ID
// @Summary Get template
// @Description Get template metadata by ID with preview URL and variables
// @Tags Templates
// @ID getTemplate
// @Produce json
// @Param id path int true "Template ID" example(1)
// @Success 200 {object} entities.TemplateAPIResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/{id} [get]
// @Security BearerAuth
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template_id"})
		return
	}

	tmpl, err := h.service.GetTemplate(c.Request.Context(), uint(tenantID), uint(templateID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	c.JSON(http.StatusOK, entities.TemplateAPIResponse{
		Success: true,
		Message: "Template retrieved successfully",
		Data:    tmpl.ToResponse(),
	})
}

// ListTemplates lists templates with filters
// @Summary List templates
// @Description List templates with optional filters and pagination (organization_id from auth middleware)
// @Tags Templates
// @ID listTemplates
// @Produce json
// @Param channel query string false "Template channel (EMAIL, DOCUMENT)" example("EMAIL")
// @Param template_key query string false "Template key (password_reset, invoice, etc.)" example("welcome_email")
// @Param is_active query bool false "Active status filter" example(true)
// @Param page query int false "Page number (default 1)" example(1)
// @Param page_size query int false "Page size (default 20)" example(10)
// @Success 200 {object} entities.TemplateListResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates [get]
// @Security BearerAuth
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	// Get organization ID from auth middleware
	var organizationID *uint
	if user, err := baseAPI.GetUser(c); err == nil && user.OrganizationID > 0 {
		orgIDVal := user.OrganizationID
		organizationID = &orgIDVal
	}

	channel := c.Query("channel")
	templateKey := c.Query("template_key")

	// Normalize channel to uppercase (accept "email" or "EMAIL")
	if channel != "" {
		channel = strings.ToUpper(channel)
	}

	var isActive *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			isActive = &active
		}
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	templates, total, err := h.service.ListTemplates(
		c.Request.Context(),
		uint(tenantID),
		organizationID,
		channel,
		templateKey,
		isActive,
		page,
		pageSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	responses := make([]entities.TemplateResponse, len(templates))
	for i, tmpl := range templates {
		responses[i] = tmpl.ToResponse()
	}

	c.JSON(http.StatusOK, entities.TemplateListResponse{
		Success: true,
		Message: "Templates retrieved successfully",
		Data:    responses,
		Page:    page,
		Limit:   pageSize,
		Total:   total,
	})
}

// UpdateTemplate updates an existing template
// @Summary Update template
// @Description Update template metadata or content (creates new version)
// @Tags Templates
// @Accept json
// @Produce json
// @Param id path int true "Template ID" example(1)
//
//	@Param request body services.UpdateTemplateRequest true "Update data" example({
//	  "name": "Updated Welcome Email Template",
//	  "description": "Updated email template for welcoming new users with enhanced styling",
//	  "content": "<h1>Hello {{.FirstName}} {{.LastName}}!</h1><p>Welcome to {{.OrganizationName}}. We're excited to have you!</p>",
//	  "variables": ["FirstName", "LastName", "OrganizationName"],
//	  "is_active": true
//	})
//
// @Success 200 {object} entities.TemplateAPIResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/{id} [put]
// @ID updateTemplate
// @Security BearerAuth
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	// Get organization ID from auth middleware
	var organizationID *uint
	if user, err := baseAPI.GetUser(c); err == nil && user.OrganizationID > 0 {
		orgIDVal := user.OrganizationID
		organizationID = &orgIDVal
	}

	templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template_id"})
		return
	}

	var req services.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.OrganizationID = organizationID

	tmpl, err := h.service.UpdateTemplate(c.Request.Context(), uint(tenantID), uint(templateID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entities.TemplateAPIResponse{
		Success: true,
		Message: "Template updated successfully",
		Data:    tmpl.ToResponse(),
	})
}

// DeleteTemplate deletes a template
// @Summary Delete template
// @Description Soft delete a template (marks as deleted)
// @Tags Templates
// @ID deleteTemplate
// @Param id path int true "Template ID" example(1)
// @Success 204 "Template deleted successfully"
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/{id} [delete]
// @Security BearerAuth
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template_id"})
		return
	}

	if err := h.service.DeleteTemplate(c.Request.Context(), uint(tenantID), uint(templateID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// RenderTemplate renders template with custom data
// @Summary Render template
// @Description Render template with provided data variables
// @Tags Templates
// @Accept json
// @Produce html
// @Param id path int true "Template ID" example(1)
//
//	@Param data body entities.RenderTemplateRequest true "Template data"
//	  "FirstName": "Alice",
//	  "LastName": "Johnson",
//	  "OrganizationName": "Tech Innovators Inc",
//	  "Email": "alice.johnson@techinnovators.com"
//	})
//
// @Success 200 {string} string "Rendered HTML content" example("<h1>Hello Alice Johnson!</h1><p>Welcome to Tech Innovators Inc. We're excited to have you!</p>")
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/{id}/render [post]
// @ID renderTemplate
// @Security BearerAuth
func (h *TemplateHandler) RenderTemplate(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template_id"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	html, err := h.service.RenderTemplate(c.Request.Context(), uint(tenantID), uint(templateID), data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// GetTemplateContent retrieves the raw HTML content of a template
// @Summary Get template content
// @Description Get the raw HTML content of a template (without rendering variables)
// @Tags Templates
// @Produce text/html
// @Param id path int true "Template ID" example(1)
// @Success 200 {string} string "Raw HTML content"
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/{id}/content [get]
// @ID getTemplateContent
// @Security BearerAuth
func (h *TemplateHandler) GetTemplateContent(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template_id"})
		return
	}

	_, content, err := h.service.GetTemplateWithContent(c.Request.Context(), uint(tenantID), uint(templateID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

// GetDefaultTemplate gets the default template for a type
// @Summary Get default template
// @Description Get the default template by channel and template key (organization from auth middleware)
// @Tags Templates
// @Produce json
// @Param channel query string false "Template channel (EMAIL, DOCUMENT)" example("EMAIL")
// @Param template_type query string true "Template type (password_reset, invoice, etc.)" example("welcome_email")
// @Success 200 {object} entities.TemplateAPIResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/default [get]
// @ID getDefaultTemplate
// @Security BearerAuth
func (h *TemplateHandler) GetDefaultTemplate(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	templateType := c.Query("template_type")
	if templateType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "template_type required"})
		return
	}

	// Get organization ID from auth middleware
	var organizationID *uint
	if user, err := baseAPI.GetUser(c); err == nil && user.OrganizationID > 0 {
		orgIDVal := user.OrganizationID
		organizationID = &orgIDVal
	}

	tmpl, err := h.service.GetDefaultTemplate(c.Request.Context(), uint(tenantID), organizationID, templateType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no default template found"})
		return
	}

	c.JSON(http.StatusOK, entities.TemplateAPIResponse{
		Success: true,
		Message: "Default template retrieved successfully",
		Data:    tmpl.ToResponse(),
	})
}

// DuplicateTemplate creates a copy of an existing template
// @Summary Duplicate template
// @Description Create a copy of an existing template with a new name
// @Tags Templates
// @Accept json
// @Produce json
// @Param id path int true "Template ID" example(1)
// @Param request body entities.DuplicateTemplateRequest true "Duplicate request with name" example({"name": "Copy of Welcome Email Template"})
// @Success 201 {object} entities.TemplateAPIResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/{id}/duplicate [post]
// @ID duplicateTemplate
// @Security BearerAuth
func (h *TemplateHandler) DuplicateTemplate(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template_id"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := h.service.DuplicateTemplate(c.Request.Context(), uint(tenantID), uint(templateID), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entities.TemplateAPIResponse{
		Success: true,
		Message: "Template duplicated successfully",
		Data:    tmpl.ToResponse(),
	})
}
