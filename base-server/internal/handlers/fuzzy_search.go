package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ae-base-server/internal/models"
	"github.com/ae-base-server/internal/services"
	"github.com/gin-gonic/gin"
)

// FuzzySearchHandler handles fuzzy search requests
type FuzzySearchHandler struct {
	fuzzySearchService *services.FuzzySearchService
}

// NewFuzzySearchHandler creates a new fuzzy search handler
func NewFuzzySearchHandler(fuzzySearchService *services.FuzzySearchService) *FuzzySearchHandler {
	return &FuzzySearchHandler{
		fuzzySearchService: fuzzySearchService,
	}
}

// SearchRequest represents a fuzzy search request
type SearchRequest struct {
	Query               string                 `json:"query" binding:"required,min=1"`
	EntityTypes         []string               `json:"entity_types,omitempty"`
	Filters             map[string]interface{} `json:"filters,omitempty"`
	SortBy              string                 `json:"sort_by,omitempty"`
	SortOrder           string                 `json:"sort_order,omitempty"`
	Offset              int                    `json:"offset"`
	Limit               int                    `json:"limit"`
	IncludeCount        bool                   `json:"include_count"`
	IncludeAggregations bool                   `json:"include_aggregations"`
	HighlightFields     []string               `json:"highlight_fields,omitempty"`
}

// QuickSearchRequest represents a simplified search request
type QuickSearchRequest struct {
	Query string   `json:"query" binding:"required,min=1"`
	Types []string `json:"types,omitempty"`
	Limit int      `json:"limit,omitempty"`
}

// EntityTypesResponse represents available entity types
type EntityTypesResponse struct {
	EntityTypes map[string]EntityTypeInfo `json:"entity_types"`
}

// EntityTypeInfo represents information about an entity type
type EntityTypeInfo struct {
	Name         string   `json:"name"`
	DisplayName  string   `json:"display_name"`
	SearchFields []string `json:"search_fields"`
	Description  string   `json:"description"`
	RequireAuth  bool     `json:"require_auth"`
}

// SearchConfigResponse represents search configuration
type SearchConfigResponse struct {
	Config      services.FuzzySearchConfig `json:"config"`
	EntityTypes map[string]EntityTypeInfo  `json:"entity_types"`
}

// Search performs fuzzy search across entities
func (h *FuzzySearchHandler) Search(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid search request",
			"details": err.Error(),
		})
		return
	}

	// Get user context for permission filtering
	var userID *uint
	var organizationID *uint

	if userInterface, exists := c.Get("user"); exists {
		if user, ok := userInterface.(*models.User); ok {
			userID = &user.ID
			organizationID = &user.TenantID
		}
	}

	// Build search options
	searchOptions := services.SearchOptions{
		Query:               req.Query,
		EntityTypes:         req.EntityTypes,
		Filters:             req.Filters,
		SortBy:              req.SortBy,
		SortOrder:           req.SortOrder,
		Offset:              req.Offset,
		Limit:               req.Limit,
		UserID:              userID,
		TenantID:      organizationID,
		IncludeCount:        req.IncludeCount,
		IncludeAggregations: req.IncludeAggregations,
		HighlightFields:     req.HighlightFields,
	}

	// Perform search
	response, err := h.fuzzySearchService.Search(searchOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Search failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// QuickSearch performs a simplified fuzzy search (GET endpoint)
func (h *FuzzySearchHandler) QuickSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Query parameter 'q' is required",
		})
		return
	}

	// Parse entity types from query parameter
	var entityTypes []string
	if typesParam := c.Query("types"); typesParam != "" {
		entityTypes = strings.Split(typesParam, ",")
	}

	// Parse limit
	limit := 10 // default
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Parse offset
	offset := 0 // default
	if offsetParam := c.Query("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get user context
	var userID *uint
	var organizationID *uint

	if userInterface, exists := c.Get("user"); exists {
		if user, ok := userInterface.(*models.User); ok {
			userID = &user.ID
			organizationID = &user.TenantID
		}
	}

	// Build search options
	searchOptions := services.SearchOptions{
		Query:          query,
		EntityTypes:    entityTypes,
		Offset:         offset,
		Limit:          limit,
		UserID:         userID,
		TenantID: organizationID,
	}

	// Perform search
	response, err := h.fuzzySearchService.Search(searchOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Search failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchInEntity searches within a specific entity type
func (h *FuzzySearchHandler) SearchInEntity(c *gin.Context) {
	entityType := c.Param("entity_type")
	if entityType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Entity type is required",
		})
		return
	}

	// Verify entity type exists
	entityTypes := h.fuzzySearchService.GetEntityTypes()
	if _, exists := entityTypes[entityType]; !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "Invalid entity type",
			"valid_types": getEntityTypeNames(entityTypes),
		})
		return
	}

	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid search request",
			"details": err.Error(),
		})
		return
	}

	// Override entity types with the specified one
	req.EntityTypes = []string{entityType}

	// Get user context
	var userID *uint
	var organizationID *uint

	if userInterface, exists := c.Get("user"); exists {
		if user, ok := userInterface.(*models.User); ok {
			userID = &user.ID
			organizationID = &user.TenantID
		}
	}

	// Build search options
	searchOptions := services.SearchOptions{
		Query:               req.Query,
		EntityTypes:         req.EntityTypes,
		Filters:             req.Filters,
		SortBy:              req.SortBy,
		SortOrder:           req.SortOrder,
		Offset:              req.Offset,
		Limit:               req.Limit,
		UserID:              userID,
		TenantID:      organizationID,
		IncludeCount:        req.IncludeCount,
		IncludeAggregations: req.IncludeAggregations,
		HighlightFields:     req.HighlightFields,
	}

	// Perform search
	response, err := h.fuzzySearchService.Search(searchOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Search failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetEntityTypes returns available entity types for search
func (h *FuzzySearchHandler) GetEntityTypes(c *gin.Context) {
	entityTypes := h.fuzzySearchService.GetEntityTypes()

	// Convert to response format
	responseTypes := make(map[string]EntityTypeInfo)
	for name, config := range entityTypes {
		searchFields := make([]string, len(config.SearchFields))
		for i, field := range config.SearchFields {
			searchFields[i] = field.Name
		}

		responseTypes[name] = EntityTypeInfo{
			Name:         name,
			DisplayName:  config.DisplayName,
			SearchFields: searchFields,
			RequireAuth:  config.Permissions.RequireAuth,
		}
	}

	c.JSON(http.StatusOK, EntityTypesResponse{
		EntityTypes: responseTypes,
	})
}

// GetSearchConfig returns current search configuration
func (h *FuzzySearchHandler) GetSearchConfig(c *gin.Context) {
	config := h.fuzzySearchService.GetConfig()
	entityTypes := h.fuzzySearchService.GetEntityTypes()

	// Convert entity types to response format
	responseTypes := make(map[string]EntityTypeInfo)
	for name, entityConfig := range entityTypes {
		searchFields := make([]string, len(entityConfig.SearchFields))
		for i, field := range entityConfig.SearchFields {
			searchFields[i] = field.Name
		}

		responseTypes[name] = EntityTypeInfo{
			Name:         name,
			DisplayName:  entityConfig.DisplayName,
			SearchFields: searchFields,
			Description:  fmt.Sprintf("Search within %s records", entityConfig.DisplayName),
			RequireAuth:  entityConfig.Permissions.RequireAuth,
		}
	}

	c.JSON(http.StatusOK, SearchConfigResponse{
		Config:      config,
		EntityTypes: responseTypes,
	})
}

// UpdateSearchConfig updates search configuration (admin only)
func (h *FuzzySearchHandler) UpdateSearchConfig(c *gin.Context) {
	var configUpdate services.FuzzySearchConfig
	if err := c.ShouldBindJSON(&configUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid configuration",
			"details": err.Error(),
		})
		return
	}

	// Validate configuration values
	if configUpdate.MinSearchLength < 1 {
		configUpdate.MinSearchLength = 1
	}
	if configUpdate.MaxResults <= 0 {
		configUpdate.MaxResults = 50
	}
	if configUpdate.ScoreThreshold < 0 {
		configUpdate.ScoreThreshold = 0
	}

	// Update configuration
	h.fuzzySearchService.UpdateConfig(configUpdate)

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated successfully",
		"config":  configUpdate,
	})
}

// SearchSuggestions provides search suggestions
func (h *FuzzySearchHandler) SearchSuggestions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Query parameter 'q' is required",
		})
		return
	}

	// For now, return entity type suggestions and common search tips
	entityTypes := h.fuzzySearchService.GetEntityTypes()
	suggestions := []string{
		"Try using different keywords",
		"Check your spelling",
		"Use more general terms",
	}

	// Add entity-specific suggestions
	for _, config := range entityTypes {
		suggestions = append(suggestions,
			fmt.Sprintf("Search %s by name", config.DisplayName),
			fmt.Sprintf("Find %s by email", config.DisplayName),
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"query":       query,
		"suggestions": suggestions,
	})
}

// RegisterCustomEntity allows registration of custom entity types (admin only)
func (h *FuzzySearchHandler) RegisterCustomEntity(c *gin.Context) {
	entityType := c.Param("entity_type")
	if entityType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Entity type is required",
		})
		return
	}

	var config services.EntityConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid entity configuration",
			"details": err.Error(),
		})
		return
	}

	// Basic validation
	if config.TableName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Table name is required",
		})
		return
	}

	if len(config.SearchFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least one search field is required",
		})
		return
	}

	// Register the entity
	h.fuzzySearchService.RegisterEntity(entityType, config)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Entity type registered successfully",
		"entity_type": entityType,
		"config":      config,
	})
}

// Helper functions

func getEntityTypeNames(entityTypes map[string]services.EntityConfig) []string {
	names := make([]string, 0, len(entityTypes))
	for name := range entityTypes {
		names = append(names, name)
	}
	return names
}

// SearchStats returns search statistics and analytics
func (h *FuzzySearchHandler) SearchStats(c *gin.Context) {
	// This could be enhanced with actual search analytics
	entityTypes := h.fuzzySearchService.GetEntityTypes()

	stats := gin.H{
		"total_entity_types": len(entityTypes),
		"entity_types":       getEntityTypeNames(entityTypes),
		"search_config":      h.fuzzySearchService.GetConfig(),
		"status":             "active",
	}

	c.JSON(http.StatusOK, stats)
}

// HealthCheck returns the health status of the search service
func (h *FuzzySearchHandler) HealthCheck(c *gin.Context) {
	entityTypes := h.fuzzySearchService.GetEntityTypes()

	c.JSON(http.StatusOK, gin.H{
		"status":       "healthy",
		"service":      "fuzzy_search",
		"entity_types": len(entityTypes),
		"last_check":   time.Now(),
	})
}
