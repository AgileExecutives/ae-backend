package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/ae-base-server/modules/templates/services/storage"
	"github.com/gin-gonic/gin"
)

// PublicAssetHandler handles public template asset requests
type PublicAssetHandler struct {
	minioStorage *storage.MinIOStorage
}

// NewPublicAssetHandler creates a new public asset handler
func NewPublicAssetHandler(minioStorage *storage.MinIOStorage) *PublicAssetHandler {
	return &PublicAssetHandler{
		minioStorage: minioStorage,
	}
}

// GetAsset serves a template asset publicly (no authentication required)
// @Summary Get public template asset
// @Description Retrieve a template asset (image, CSS, etc.) without authentication
// @Tags Templates
// @ID getPublicTemplateAsset
// @Produce application/octet-stream
// @Param tenant path string true "Tenant ID"
// @Param template path string true "Template ID"
// @Param file path string true "Asset filename"
// @Success 200 {file} binary
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /public/templates/assets/{tenant}/{template}/{file} [get]
func (h *PublicAssetHandler) GetAsset(c *gin.Context) {
	tenantID := c.Param("tenant")
	templateID := c.Param("template")
	filename := c.Param("file")

	if tenantID == "" || templateID == "" || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant, template, and file are required"})
		return
	}

	// Construct object path: templates/{tenantID}/{templateID}/assets/{filename}
	objectPath := fmt.Sprintf("templates/%s/%s/assets/%s", tenantID, templateID, filename)

	// Get object from MinIO (using Retrieve method which returns []byte)
	data, err := h.minioStorage.Retrieve(c.Request.Context(), "templates", objectPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	// Determine content type from file extension
	contentType := getContentType(filename)

	// Set headers
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", fmt.Sprintf("%d", len(data)))
	c.Header("Cache-Control", "public, max-age=86400") // Cache for 24 hours

	// Write the file data
	if _, err := c.Writer.Write(data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stream asset"})
		return
	}
}

// getContentType determines content type from file extension
func getContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
