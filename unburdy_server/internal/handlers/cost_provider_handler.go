package handlers

import (
	"net/http"
	"strconv"

	baseAPI "github.com/ae-saas-basic/ae-saas-basic/api"
	"github.com/ae-saas-basic/ae-saas-basic/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/unburdy-server-api/internal/models"
	"github.com/unburdy/unburdy-server-api/internal/services"
)

type CostProviderHandler struct {
	costProviderService *services.CostProviderService
}

func NewCostProviderHandler(costProviderService *services.CostProviderService) *CostProviderHandler {
	return &CostProviderHandler{
		costProviderService: costProviderService,
	}
}

// CreateCostProvider godoc
// @Summary Create a new cost provider
// @Description Create a new cost provider for the authenticated tenant
// @Tags cost-providers
// @Accept json
// @Produce json
// @Param cost_provider body models.CreateCostProviderRequest true "Cost provider data"
// @Success 201 {object} models.CostProviderResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /cost-providers [post]
func (h *CostProviderHandler) CreateCostProvider(c *gin.Context) {
	// Get authenticated tenant info
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant information required"})
		return
	}

	var req models.CreateCostProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	costProvider, err := h.costProviderService.CreateCostProvider(req, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, costProvider.ToResponse())
}

// GetCostProvider godoc
// @Summary Get a cost provider by ID
// @Description Get details of a specific cost provider
// @Tags cost-providers
// @Accept json
// @Produce json
// @Param id path int true "Cost Provider ID"
// @Success 200 {object} models.CostProviderResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /cost-providers/{id} [get]
func (h *CostProviderHandler) GetCostProvider(c *gin.Context) {
	// Get authenticated tenant info
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant information required"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost provider ID"})
		return
	}

	costProvider, err := h.costProviderService.GetCostProviderByID(uint(id), tenantID)
	if err != nil {
		if err.Error() == "cost provider not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cost provider not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, costProvider.ToResponse())
}

// GetAllCostProviders godoc
// @Summary Get all cost providers
// @Description Get a paginated list of all cost providers for the authenticated tenant
// @Tags cost-providers
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "data: array of CostProviderResponse, total: int, page: int, limit: int"
// @Failure 401 {object} map[string]interface{} "error: string"
// @Failure 500 {object} map[string]interface{} "error: string"
// @Security BearerAuth
// @Router /cost-providers [get]
func (h *CostProviderHandler) GetAllCostProviders(c *gin.Context) {
	// Get authenticated tenant info
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant information required"})
		return
	}

	page, limit := utils.GetPaginationParams(c)

	costProviders, total, err := h.costProviderService.GetAllCostProviders(page, limit, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	costProviderResponses := make([]models.CostProviderResponse, len(costProviders))
	for i, costProvider := range costProviders {
		costProviderResponses[i] = costProvider.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"cost_providers": costProviderResponses,
		"total":          total,
		"page":           page,
		"limit":          limit,
	})
}

// UpdateCostProvider godoc
// @Summary Update a cost provider
// @Description Update an existing cost provider's information
// @Tags cost-providers
// @Accept json
// @Produce json
// @Param id path int true "Cost Provider ID"
// @Param cost_provider body models.UpdateCostProviderRequest true "Cost provider update data"
// @Success 200 {object} models.CostProviderResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /cost-providers/{id} [put]
func (h *CostProviderHandler) UpdateCostProvider(c *gin.Context) {
	// Get authenticated tenant info
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant information required"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost provider ID"})
		return
	}

	var req models.UpdateCostProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	costProvider, err := h.costProviderService.UpdateCostProvider(uint(id), tenantID, req)
	if err != nil {
		if err.Error() == "cost provider not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cost provider not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, costProvider.ToResponse())
}

// DeleteCostProvider godoc
// @Summary Delete a cost provider
// @Description Delete a cost provider by ID
// @Tags cost-providers
// @Accept json
// @Produce json
// @Param id path int true "Cost Provider ID"
// @Success 200 {object} map[string]interface{} "message: string"
// @Failure 401 {object} map[string]interface{} "error: string"
// @Failure 404 {object} map[string]interface{} "error: string"
// @Failure 500 {object} map[string]interface{} "error: string"
// @Security BearerAuth
// @Router /cost-providers/{id} [delete]
func (h *CostProviderHandler) DeleteCostProvider(c *gin.Context) {
	// Get authenticated tenant info
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant information required"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cost provider ID"})
		return
	}

	err = h.costProviderService.DeleteCostProvider(uint(id), tenantID)
	if err != nil {
		if err.Error() == "cost provider not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cost provider not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cost provider deleted successfully"})
}

// SearchCostProviders godoc
// @Summary Search cost providers
// @Description Search cost providers by organization name or contact name
// @Tags cost-providers
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "data: array of CostProviderResponse, total: int, page: int, limit: int, query: string"
// @Failure 400 {object} map[string]interface{} "error: string"
// @Failure 401 {object} map[string]interface{} "error: string"
// @Failure 500 {object} map[string]interface{} "error: string"
// @Security BearerAuth
// @Router /cost-providers/search [get]
func (h *CostProviderHandler) SearchCostProviders(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	// Get authenticated tenant info
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant information required"})
		return
	}

	page, limit := utils.GetPaginationParams(c)

	costProviders, total, err := h.costProviderService.SearchCostProviders(query, page, limit, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	costProviderResponses := make([]models.CostProviderResponse, len(costProviders))
	for i, costProvider := range costProviders {
		costProviderResponses[i] = costProvider.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"cost_providers": costProviderResponses,
		"total":          total,
		"page":           page,
		"limit":          limit,
		"query":          query,
	})
}
