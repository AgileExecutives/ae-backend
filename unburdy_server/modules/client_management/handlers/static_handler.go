package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/unburdy/unburdy-server-api/internal/models"
	"github.com/unburdy/unburdy-server-api/modules/client_management/services"
)

// StaticHandler handles static file serving with registration token authentication
type StaticHandler struct {
	clientService *services.ClientService
}

// NewStaticHandler creates a new static file handler
func NewStaticHandler(clientService *services.ClientService) *StaticHandler {
	return &StaticHandler{
		clientService: clientService,
	}
}

// ListStaticJSON lists all available JSON files in the statics/json directory
// @Summary List available static JSON files (registration token auth)
// @ID listStaticFilesWithToken
// @Description Get a list of all JSON files available in the statics/json directory, authenticated by registration token
// @Tags clients
// @Param token path string true "Registration token"
// @Success 200 {object} map[string]interface{} "List of available JSON files"
// @Failure 400 {object} map[string]string "Invalid token"
// @Failure 500 {object} map[string]string "Failed to read directory"
// @Router /clients/static/{token} [get]
func (h *StaticHandler) ListStaticJSON(c *gin.Context) {
	token := c.Param("token")

	// Validate registration token
	_, err := h.clientService.ValidateRegistrationToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Access denied", err.Error()))
		return
	}

	// Read the statics/json directory
	entries, err := os.ReadDir("./statics/json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read directory"})
		return
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			// Remove the .json extension for the API response
			filename := strings.TrimSuffix(entry.Name(), ".json")
			files = append(files, filename)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"available_files": files,
		"base_url":        "/api/v1/clients/static/" + token + "/",
		"example_usage":   "GET /api/v1/clients/static/" + token + "/{filename}",
		"security_note":   "Only JSON files from statics/json/ directory are accessible",
		"restrictions":    "Filenames must be alphanumeric with hyphens/underscores only",
		"note":            "Drop any .json file in ./statics/json/ directory to make it available",
	})
}

// ServeStaticJSON serves ONLY JSON files from the statics/json directory
// @Summary Serve static JSON files with registration token authentication
// @ID getStaticFileWithToken
// @Description Securely serve JSON data files from statics/json directory only. Authenticated by registration token. Prevents access to other directories or file types.
// @Tags clients
// @Param token path string true "Registration token"
// @Param filename path string true "JSON filename (without .json extension)" example("bundeslaender")
// @Success 200 {object} map[string]interface{} "JSON file content"
// @Failure 400 {object} map[string]string "Invalid file name or token"
// @Failure 404 {object} map[string]string "File not found"
// @Failure 500 {object} map[string]string "Failed to read file"
// @Router /clients/static/{token}/{filename} [get]
func (h *StaticHandler) ServeStaticJSON(c *gin.Context) {
	token := c.Param("token")
	fileName := c.Param("filename")

	// Validate registration token
	_, err := h.clientService.ValidateRegistrationToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Access denied", err.Error()))
		return
	}

	// Basic validation first
	if fileName == "" || len(fileName) > 100 {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Check for system files (should return 404 for security)
	if strings.HasPrefix(fileName, ".") {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Check for null byte injection (should return 404 for security)
	if strings.Contains(fileName, "\x00") {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Only allow alphanumeric characters, hyphens, and underscores (return 400 for invalid chars)
	for _, char := range fileName {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file name"})
			return
		}
	}

	// Path traversal checks (after character validation)
	if strings.Contains(fileName, "..") ||
		strings.Contains(fileName, "/") ||
		strings.Contains(fileName, "\\") ||
		strings.Contains(fileName, "~") {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Enforce case sensitivity by reading the directory and checking exact match first
	entries, err := os.ReadDir("./statics/json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read directory"})
		return
	}

	expectedFilename := fileName + ".json"
	found := false
	for _, entry := range entries {
		if entry.Name() == expectedFilename {
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Construct the full file path - always look in statics/json directory ONLY
	fullPath := filepath.Join("./statics/json", fileName+".json")

	// Read and return the JSON file content
	data, err := os.ReadFile(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Set content type to JSON and return raw content
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", data)
}
