package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type StaticHandler struct {
	staticDir string
}

// NewStaticHandler creates a new static file handler
func NewStaticHandler(staticDir string) *StaticHandler {
	return &StaticHandler{
		staticDir: staticDir,
	}
}

// ServeAsset serves static assets with proper headers
// @Summary Serve static assets
// @Description Serve static files like CSS, JS, images with appropriate headers
// @Tags static
// @Param path path string true "Asset path"
// @Success 200 "Asset content"
// @Failure 404 {object} models.ErrorResponse
// @Router /assets/{path} [get]
func (h *StaticHandler) ServeAsset(c *gin.Context) {
	path := c.Param("path")

	// Security check: prevent directory traversal
	if strings.Contains(path, "..") || strings.Contains(path, "~") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid path"})
		return
	}

	fullPath := filepath.Join(h.staticDir, "assets", path)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
		return
	}

	// Set appropriate content type and cache headers
	h.setContentHeaders(c, path)

	c.File(fullPath)
}

// ServeTemplate serves email/PDF templates for preview
// @Summary Preview templates
// @Description Serve template files for preview purposes
// @Tags static
// @Security BearerAuth
// @Param type path string true "Template type (email|pdf)"
// @Param template path string true "Template name"
// @Success 200 "Template content"
// @Failure 404 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /templates/{type}/{template} [get]
func (h *StaticHandler) ServeTemplate(c *gin.Context) {
	templateType := c.Param("type")
	templateName := c.Param("template")

	// Validate template type
	if templateType != "email" && templateType != "pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template type"})
		return
	}

	// Security checks
	if strings.Contains(templateName, "..") || strings.Contains(templateName, "~") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid template name"})
		return
	}

	var fullPath string
	if templateType == "email" {
		fullPath = filepath.Join(h.staticDir, "email_templates", templateName)
	} else {
		fullPath = filepath.Join(h.staticDir, "pdf_templates", templateName)
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	// Set HTML content type
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")

	c.File(fullPath)
}

// ServeLogo serves the company logo
// @Summary Get company logo
// @Description Serve the company logo in various formats
// @Tags static
// @Param format query string false "Logo format (svg|png|jpg)" default(svg)
// @Success 200 "Logo file"
// @Failure 404 {object} models.ErrorResponse
// @Router /logo [get]
func (h *StaticHandler) ServeLogo(c *gin.Context) {
	format := c.DefaultQuery("format", "svg")

	var filename string
	switch format {
	case "svg":
		filename = "logo.svg"
	case "png":
		filename = "logo.png"
	case "jpg", "jpeg":
		filename = "logo.jpg"
	case "full":
		filename = "logo-full.svg"
	default:
		filename = "logo.svg"
	}

	fullPath := filepath.Join(h.staticDir, "images", filename)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// Fallback to default SVG logo
		fullPath = filepath.Join(h.staticDir, "images", "logo.svg")
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Logo not found"})
			return
		}
	}

	// Set appropriate headers
	h.setContentHeaders(c, filename)
	c.Header("Cache-Control", "public, max-age=86400") // Cache for 1 day

	c.File(fullPath)
}

// ListAssets lists available static assets
// @Summary List static assets
// @Description Get a list of available static assets
// @Tags static
// @Security BearerAuth
// @Param type query string false "Asset type filter"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /assets [get]
func (h *StaticHandler) ListAssets(c *gin.Context) {
	assetType := c.Query("type")

	assets := make(map[string][]string)

	// List CSS files
	cssFiles, _ := h.listFiles(filepath.Join(h.staticDir, "assets", "css"), ".css")
	assets["css"] = cssFiles

	// List JS files
	jsFiles, _ := h.listFiles(filepath.Join(h.staticDir, "assets", "js"), ".js")
	assets["js"] = jsFiles

	// List image files
	imageFiles, _ := h.listFiles(filepath.Join(h.staticDir, "images"), ".svg", ".png", ".jpg", ".jpeg", ".gif")
	assets["images"] = imageFiles

	// List email templates
	emailTemplates, _ := h.listFiles(filepath.Join(h.staticDir, "email_templates"), ".html")
	assets["email_templates"] = emailTemplates

	// List PDF templates
	pdfTemplates, _ := h.listFiles(filepath.Join(h.staticDir, "pdf_templates"), ".html")
	assets["pdf_templates"] = pdfTemplates

	// Filter by type if specified
	if assetType != "" {
		if filtered, exists := assets[assetType]; exists {
			c.JSON(http.StatusOK, gin.H{"type": assetType, "assets": filtered})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Asset type not found"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"assets": assets})
}

// setContentHeaders sets appropriate content type and cache headers
func (h *StaticHandler) setContentHeaders(c *gin.Context, filename string) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".css":
		c.Header("Content-Type", "text/css; charset=utf-8")
		c.Header("Cache-Control", "public, max-age=3600")
	case ".js":
		c.Header("Content-Type", "application/javascript; charset=utf-8")
		c.Header("Cache-Control", "public, max-age=3600")
	case ".svg":
		c.Header("Content-Type", "image/svg+xml")
		c.Header("Cache-Control", "public, max-age=86400")
	case ".png":
		c.Header("Content-Type", "image/png")
		c.Header("Cache-Control", "public, max-age=86400")
	case ".jpg", ".jpeg":
		c.Header("Content-Type", "image/jpeg")
		c.Header("Cache-Control", "public, max-age=86400")
	case ".gif":
		c.Header("Content-Type", "image/gif")
		c.Header("Cache-Control", "public, max-age=86400")
	case ".html":
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Header("Cache-Control", "no-cache")
	case ".pdf":
		c.Header("Content-Type", "application/pdf")
		c.Header("Cache-Control", "public, max-age=3600")
	default:
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Cache-Control", "no-cache")
	}
}

// listFiles lists files in a directory with specified extensions
func (h *StaticHandler) listFiles(dir string, extensions ...string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return files, err
	}

	extMap := make(map[string]bool)
	for _, ext := range extensions {
		extMap[strings.ToLower(ext)] = true
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if len(extensions) == 0 || extMap[ext] {
				files = append(files, entry.Name())
			}
		}
	}

	return files, nil
}
