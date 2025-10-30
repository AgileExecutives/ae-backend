package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DemoHandler handles demo module endpoints
type DemoHandler struct{}

// NewDemoHandler creates a new demo handler
func NewDemoHandler() *DemoHandler {
	return &DemoHandler{}
}

// GetDemo returns demo information
// @Summary Get demo information
// @Description Get information about the demo module
// @Tags demo
// @Produce json
// @Success 200 {object} map[string]interface{} "Demo information"
// @Router /api/v1/modules/demo/demo [get]
func (h *DemoHandler) GetDemo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Demo module is working!",
		"module":  "demo",
		"version": "1.0.0",
	})
}

// GetDemoItems returns all demo items
// @Summary Get all demo items
// @Description Retrieve all demo items from the demo module
// @Tags demo
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "List of demo items"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/v1/modules/demo/items [get]
func (h *DemoHandler) GetDemoItems(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": []gin.H{
			{"id": 1, "name": "Demo Item 1", "description": "First demo item"},
			{"id": 2, "name": "Demo Item 2", "description": "Second demo item"},
		},
		"total": 2,
	})
}

// GetDemoItem returns a specific demo item
// @Summary Get demo item by ID
// @Description Retrieve a specific demo item by its ID
// @Tags demo
// @Produce json
// @Security BearerAuth
// @Param id path int true "Demo Item ID"
// @Success 200 {object} map[string]interface{} "Demo item details"
// @Failure 404 {object} map[string]string "Demo item not found"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/v1/modules/demo/items/{id} [get]
func (h *DemoHandler) GetDemoItem(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"id":          id,
		"name":        "Demo Item " + id,
		"description": "Demo item with ID " + id,
		"module":      "demo",
	})
}

// CreateDemoItem creates a new demo item
// @Summary Create a new demo item
// @Description Create a new demo item in the demo module
// @Tags demo
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param item body map[string]interface{} true "Demo item data"
// @Success 201 {object} map[string]interface{} "Created demo item"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/v1/modules/demo/items [post]
func (h *DemoHandler) CreateDemoItem(c *gin.Context) {
	var item map[string]interface{}
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Simulate creation
	item["id"] = 999
	item["created_at"] = "2025-10-28T17:00:00Z"
	item["module"] = "demo"

	c.JSON(http.StatusCreated, gin.H{
		"message": "Demo item created successfully",
		"data":    item,
	})
}

// UpdateDemoItem updates a demo item
// @Summary Update a demo item
// @Description Update an existing demo item by ID
// @Tags demo
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Demo Item ID"
// @Param item body map[string]interface{} true "Updated demo item data"
// @Success 200 {object} map[string]interface{} "Updated demo item"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Demo item not found"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/v1/modules/demo/items/{id} [put]
func (h *DemoHandler) UpdateDemoItem(c *gin.Context) {
	id := c.Param("id")
	var item map[string]interface{}
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Simulate update
	item["id"] = id
	item["updated_at"] = "2025-10-28T17:00:00Z"
	item["module"] = "demo"

	c.JSON(http.StatusOK, gin.H{
		"message": "Demo item updated successfully",
		"data":    item,
	})
}

// DeleteDemoItem deletes a demo item
// @Summary Delete a demo item
// @Description Delete a demo item by ID
// @Tags demo
// @Produce json
// @Security BearerAuth
// @Param id path int true "Demo Item ID"
// @Success 200 {object} map[string]string "Deletion confirmation"
// @Failure 404 {object} map[string]string "Demo item not found"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/v1/modules/demo/items/{id} [delete]
func (h *DemoHandler) DeleteDemoItem(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Demo item deleted successfully",
		"id":      id,
		"module":  "demo",
	})
}
