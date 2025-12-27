package handlers

import (
	"net/http"
	"strconv"

	baseAPI "github.com/ae-base-server/api"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/templates-module/entities"
	"github.com/unburdy/templates-module/services"
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
// @Param request body handlers.CreateTemplateRequest true "Template data"
// @Success 201 {object} handlers.TemplateResponse
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

	var req services.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.TenantID = uint(tenantID)

	tmpl, err := h.service.CreateTemplate(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tmpl.ToResponse())
}

// GetTemplate retrieves a template by ID
// @Summary Get template
// @Description Get template metadata by ID
// @Tags Templates
// @Produce json
// @Param id path int true "Template ID"
// @Success 200 {object} entities.TemplateResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /templates/{id} [get]
// @ID getTemplate
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

	c.JSON(http.StatusOK, tmpl.ToResponse())
}

// GetTemplateContent retrieves template with full HTML content
// @Summary Get template content
// @Description Get template metadata and HTML content
// @Tags Templates
// @Produce json
// @Param id path int true "Template ID"
// @Success 200 {object} handlers.SuccessResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/{id}/content [get]
// @ID getTemplateContent
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

	tmpl, content, err := h.service.GetTemplateWithContent(c.Request.Context(), uint(tenantID), uint(templateID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	response := tmpl.ToResponse()
	c.JSON(http.StatusOK, gin.H{
		"template": response,
		"content":  content,
	})
}

// ListTemplates lists templates with filters
// @Summary List templates
// @Description List templates with optional filters and pagination
// @Tags Templates
// @Produce json
// @Param organization_id query int false "Organization ID filter"
// @Param template_type query string false "Template type (email, pdf, invoice, document)"
// @Param is_active query bool false "Active status filter"
// @Param page query int false "Page number (default 1)"
// @Param page_size query int false "Page size (default 20)"
// @Success 200 {object} handlers.SuccessResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates [get]
// @ID listTemplates
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	// Parse filters
	var organizationID *uint
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		if orgID, err := strconv.ParseUint(orgIDStr, 10, 32); err == nil {
			orgIDVal := uint(orgID)
			organizationID = &orgIDVal
		}
	}

	templateType := c.Query("template_type")

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
		templateType,
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

	c.JSON(http.StatusOK, gin.H{
		"data":        responses,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (int(total) + pageSize - 1) / pageSize,
	})
}

// UpdateTemplate updates an existing template
// @Summary Update template
// @Description Update template metadata or content (creates new version)
// @Tags Templates
// @Accept json
// @Produce json
// @Param id path int true "Template ID"
// @Param request body services.UpdateTemplateRequest true "Update data"
// @Success 200 {object} entities.TemplateResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /templates/{id} [put]
// @ID updateTemplate
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
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

	var req services.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := h.service.UpdateTemplate(c.Request.Context(), uint(tenantID), uint(templateID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tmpl.ToResponse())
}

// DeleteTemplate deletes a template
// @Summary Delete template
// @Description Soft delete a template
// @Tags Templates
// @Param id path int true "Template ID"
// @Success 204
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /templates/{id} [delete]
// @ID deleteTemplate
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

// PreviewTemplate renders template with sample data
// @Summary Preview template
// @Description Render template with sample data for preview
// @Tags Templates
// @Produce html
// @Param id path int true "Template ID"
// @Success 200 {string} string "HTML content"
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /templates/{id}/preview [get]
// @ID previewTemplate
func (h *TemplateHandler) PreviewTemplate(c *gin.Context) {
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

	html, err := h.service.PreviewTemplate(c.Request.Context(), uint(tenantID), uint(templateID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// RenderTemplate renders template with custom data
// @Summary Render template
// @Description Render template with provided data
// @Tags Templates
// @Accept json
// @Produce html
// @Param id path int true "Template ID"
// @Param data body map[string]interface{} true "Template data"
// @Success 200 {string} string "HTML content"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /templates/{id}/render [post]
// @ID renderTemplate
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

// DuplicateTemplate creates a copy of a template
// @Summary Duplicate template
// @Description Create a copy of an existing template
// @Tags Templates
// @Accept json
// @Produce json
// @Param id path int true "Template ID"
// @Param request body map[string]string true "New template name"
// @Success 201 {object} entities.TemplateResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /templates/{id}/duplicate [post]
// @ID duplicateTemplate
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

	c.JSON(http.StatusCreated, tmpl.ToResponse())
}

// GetDefaultTemplate gets the default template for a type
// @Summary Get default template
// @Description Get the default template for a specific type and organization
// @Tags Templates
// @Produce json
// @Param template_type query string true "Template type (email, pdf, invoice, document)"
// @Param organization_id query int false "Organization ID (falls back to system default if not found)"
// @Success 200 {object} entities.TemplateResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /templates/default [get]
// @ID getDefaultTemplate
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

	var organizationID *uint
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		if orgID, err := strconv.ParseUint(orgIDStr, 10, 32); err == nil {
			orgIDVal := uint(orgID)
			organizationID = &orgIDVal
		}
	}

	tmpl, err := h.service.GetDefaultTemplate(c.Request.Context(), uint(tenantID), organizationID, templateType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no default template found"})
		return
	}

	c.JSON(http.StatusOK, tmpl.ToResponse())
}
