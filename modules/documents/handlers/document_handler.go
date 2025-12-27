package handlers

import (
	"io"
	"net/http"
	"strconv"
	"time"

	baseAPI "github.com/ae-base-server/api"
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/documents-module/entities"
	"github.com/unburdy/documents-module/services"
	"gorm.io/gorm"
)

// DocumentHandler handles document-related HTTP requests
type DocumentHandler struct {
	service *services.DocumentService
	db      *gorm.DB
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler(service *services.DocumentService, db *gorm.DB) *DocumentHandler {
	return &DocumentHandler{
		service: service,
		db:      db,
	}
}

// RegisterRoutes registers document-related routes
func (h *DocumentHandler) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	router.POST("", ctx.Auth.RequireAuth(), h.UploadDocument)
	router.GET("", ctx.Auth.RequireAuth(), h.ListDocuments)
	router.GET("/:id", ctx.Auth.RequireAuth(), h.GetDocument)
	router.GET("/:id/download", ctx.Auth.RequireAuth(), h.DownloadDocument)
	router.DELETE("/:id", ctx.Auth.RequireAuth(), h.DeleteDocument)
}

// GetPrefix returns the route prefix
func (h *DocumentHandler) GetPrefix() string {
	return "/documents"
}

// GetMiddleware returns route middleware
func (h *DocumentHandler) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

// GetSwaggerTags returns swagger tags for documentation
func (h *DocumentHandler) GetSwaggerTags() []string {
	return []string{"documents"}
}

// UploadDocument godoc
// @Summary Upload a document
// @Description Upload a document file with metadata to MinIO storage
// @Tags Documents
// @ID uploadDocument
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Document file to upload"
// @Param document_type formData string true "Document type (invoice, contract, report, etc.)"
// @Param bucket formData string false "Storage bucket (default: documents)"
// @Param path formData string false "Storage path (default: filename)"
// @Param reference_type formData string false "Reference type (invoice, client, session)"
// @Param reference_id formData int false "Reference ID"
// @Param organization_id formData int false "Organization ID"
// @Success 201 {object} entities.DocumentResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /documents [post]
// @Security BearerAuth
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}

	// Parse multipart form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "File is required"})
		return
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to read file"})
		return
	}

	// Parse form data
	documentType := c.PostForm("document_type")
	if documentType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "document_type is required"})
		return
	}

	bucket := c.PostForm("bucket")
	if bucket == "" {
		bucket = "documents"
	}

	path := c.PostForm("path")
	if path == "" {
		path = header.Filename
	}

	// Build request
	req := entities.StoreDocumentRequest{
		DocumentType:  documentType,
		ReferenceType: c.PostForm("reference_type"),
		FileName:      header.Filename,
		Content:       content,
		Bucket:        bucket,
		Path:          path,
		ContentType:   header.Header.Get("Content-Type"),
	}

	// Parse optional fields
	if orgIDStr := c.PostForm("organization_id"); orgIDStr != "" {
		if orgID, err := strconv.ParseUint(orgIDStr, 10, 32); err == nil {
			orgIDUint := uint(orgID)
			req.OrganizationID = &orgIDUint
		}
	}

	if refIDStr := c.PostForm("reference_id"); refIDStr != "" {
		if refID, err := strconv.ParseUint(refIDStr, 10, 32); err == nil {
			refIDUint := uint(refID)
			req.ReferenceID = &refIDUint
		}
	}

	// Store document
	doc, err := h.service.StoreDocument(c.Request.Context(), tenantID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entities.DocumentAPIResponse{
		Success: true,
		Message: "Document uploaded successfully",
		Data:    doc.ToResponse(),
	})
}

// GetDocument godoc
// @Summary Get document metadata
// @Description Retrieve metadata for a specific document by ID
// @Tags Documents
// @ID getDocument
// @Produce json
// @Param id path int true "Document ID"
// @Success 200 {object} entities.DocumentResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /documents/{id} [get]
// @Security BearerAuth
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid document ID"})
		return
	}

	doc, err := h.service.GetDocument(c.Request.Context(), uint(id), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entities.DocumentAPIResponse{
		Success: true,
		Message: "Document retrieved successfully",
		Data:    doc.ToResponse(),
	})
}

// DownloadDocument godoc
// @Summary Download document
// @Description Get a pre-signed URL for downloading a document (valid for 1 hour)
// @Tags Documents
// @ID downloadDocument
// @Produce json
// @Param id path int true "Document ID"
// @Success 200 {object} map[string]interface{}{success=bool,download_url=string,expires_in=int}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /documents/{id}/download [get]
// @Security BearerAuth
func (h *DocumentHandler) DownloadDocument(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid document ID"})
		return
	}

	// Get document metadata first
	doc, err := h.service.GetDocument(c.Request.Context(), uint(id), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Generate download URL (valid for 1 hour)
	expiresIn := time.Hour
	downloadURL, err := h.service.GetDownloadURL(c.Request.Context(), uint(id), tenantID, expiresIn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entities.DownloadURLResponse{
		Success:     true,
		Message:     "Download URL generated successfully",
		DocumentID:  doc.ID,
		FileName:    doc.FileName,
		DownloadURL: downloadURL,
		ExpiresAt:   time.Now().Add(expiresIn),
	})
}

// ListDocuments retrieves paginated list of documents
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}

	var req entities.ListDocumentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	documents, total, err := h.service.ListDocuments(c.Request.Context(), tenantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Convert to response format
	responses := make([]entities.DocumentResponse, len(documents))
	for i, doc := range documents {
		responses[i] = doc.ToResponse()
	}

	c.JSON(http.StatusOK, entities.DocumentListResponse{
		Success: true,
		Message: "Documents retrieved successfully",
		Data:    responses,
		Page:    req.Page,
		Limit:   req.Limit,
		Total:   total,
	})
}

// DeleteDocument godoc
// @Summary Delete document
// @Description Soft delete a document (marks as deleted, does not remove from storage)
// @Tags Documents
// @ID deleteDocument
// @Produce json
// @Param id path int true "Document ID"
// @Success 200 {object} map[string]interface{}{success=bool,message=string}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /documents/{id} [delete]
// @Security BearerAuth
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid document ID"})
		return
	}

	err = h.service.DeleteDocument(c.Request.Context(), uint(id), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Document deleted successfully",
	})
}
