package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// ServeStaticJSON serves any JSON file from the statics/json directory
// DISABLED-SWAGGER: @Summary Serve static JSON files
// DISABLED-SWAGGER: @Description Dynamically serve any JSON data file from the statics/json directory. Just drop a .json file in the directory and it becomes available at /static/{filename}
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

	// Security check: prevent path traversal
	if strings.Contains(fileName, "..") || strings.Contains(fileName, "/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file name"})
		return
	}

	// Construct the full file path - always look in statics/json directory
	fullPath := filepath.Join("./statics/json", fileName+".json")

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

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
		"base_url":        "/static/",
		"example_usage":   "GET /static/{filename}",
		"note":            "Drop any .json file in ./statics/json/ directory to make it available",
	})
}
