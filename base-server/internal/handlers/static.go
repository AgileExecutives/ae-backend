package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// ServeStaticJSON serves ONLY JSON files from the statics/json directory
// Security: This endpoint is restricted to JSON files in statics/json/ directory only
// DISABLED-SWAGGER: @Summary Serve static JSON files (JSON only, security restricted)
// DISABLED-SWAGGER: @Description Securely serve JSON data files from statics/json directory only. Prevents access to other directories or file types.
// DISABLED-SWAGGER: @Tags static
// DISABLED-SWAGGER: @Param filename path string true "JSON filename (without .json extension)" example("bundeslaender")
// DISABLED-SWAGGER: @Success 200 {object} map[string]interface{} "JSON file content"
// DISABLED-SWAGGER: @Failure 400 {object} map[string]string "Invalid file name"
// DISABLED-SWAGGER: @Failure 404 {object} map[string]string "File not found"
// DISABLED-SWAGGER: @Failure 500 {object} map[string]string "Failed to read file"
// DISABLED-SWAGGER: @Router /static/{filename} [get]
func ServeStaticJSON(c *gin.Context) {
	// Get the requested file name
	fileName := c.Param("filename")

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

// ListStaticJSON lists all available JSON files in the statics/json directory
// DISABLED-SWAGGER: @Summary List available static JSON files
// DISABLED-SWAGGER: @Description Get a list of all JSON files available in the statics/json directory
// DISABLED-SWAGGER: @Tags static
// DISABLED-SWAGGER: @Success 200 {object} map[string]interface{} "List of available JSON files"
// DISABLED-SWAGGER: @Failure 500 {object} map[string]string "Failed to read directory"
// DISABLED-SWAGGER: @Router /static [get]
func ListStaticJSON(c *gin.Context) {
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
		"base_url":        "/api/v1/static/",
		"example_usage":   "GET /api/v1/static/{filename}",
		"security_note":   "Only JSON files from statics/json/ directory are accessible",
		"restrictions":    "Filenames must be alphanumeric with hyphens/underscores only",
		"note":            "Drop any .json file in ./statics/json/ directory to make it available",
	})
}
