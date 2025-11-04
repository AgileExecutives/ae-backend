package handlers

import (
	"net/http"
	"strconv"

	baseAPI "github.com/ae-base-server/api"
	"github.com/ae-base-server/pkg/utils"
	"github.com/gin-gonic/gin"

	"github.com/unburdy/unburdy-server-api/internal/models"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"github.com/unburdy/unburdy-server-api/modules/client_management/services"
)

// CostProviderHandler handles cost provider-related HTTP requests
type CostProviderHandler struct {
	costProviderService *services.CostProviderService
}

// NewCostProviderHandler creates a new cost provider handler
func NewCostProviderHandler(costProviderService *services.CostProviderService) *CostProviderHandler {
	return &CostProviderHandler{
		costProviderService: costProviderService,
	}
}

// CreateCostProvider handles creating a new cost provider
// @Summary Create a new cost provider
// @ID createCostProvider
// @Description Create a new cost provider with the provided information
// @Tags cost-providers
// @Accept json
// @Produce json
// @Param cost_provider body entities.CreateCostProviderRequest true "Cost provider information"
// @Success 201 {object} models.APIResponse{data=entities.CostProviderResponse} "Created cost provider"
// @Failure 400 {object} models.APIResponse "Bad request"
// @Failure 401 {object} models.APIResponse "Unauthorized"
// @Failure 500 {object} models.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /cost-providers [post]
func (h *CostProviderHandler) CreateCostProvider(c *gin.Context) {
	var req entities.CreateCostProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	costProvider, err := h.costProviderService.CreateCostProvider(req, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse("Cost provider created successfully", costProvider.ToResponse()))
}

// GetCostProvider handles retrieving a cost provider by ID
// @Summary Get a cost provider by ID
// @ID getCostProviderById
// @Description Retrieve a specific cost provider by their ID
// @Tags cost-providers
// @Produce json
// @Param id path int true "Cost Provider ID"
// @Success 200 {object} models.APIResponse{data=entities.CostProviderResponse} "Cost provider found"
// @Failure 400 {object} models.APIResponse "Bad request"
// @Failure 401 {object} models.APIResponse "Unauthorized"
// @Failure 404 {object} models.APIResponse "Cost provider not found"
// @Failure 500 {object} models.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /cost-providers/{id} [get]
func (h *CostProviderHandler) GetCostProvider(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Invalid cost provider ID"))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	costProvider, err := h.costProviderService.GetCostProviderByID(uint(id), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Cost provider retrieved successfully", costProvider.ToResponse()))
}

// GetAllCostProviders handles retrieving all cost providers with pagination
// @Summary Get all cost providers
// @ID getCostProviders
// @Description Retrieve all cost providers with optional pagination
// @Tags cost-providers
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of cost providers per page (respects DEFAULT_PAGE_LIMIT and MAX_PAGE_LIMIT env vars)" default(200)
// @Success 200 {object} models.APIResponse{data=models.ListResponse} "Cost providers retrieved successfully"
// @Failure 401 {object} models.APIResponse "Unauthorized"
// @Failure 500 {object} models.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /cost-providers [get]
func (h *CostProviderHandler) GetAllCostProviders(c *gin.Context) {
	page, limit := utils.GetPaginationParams(c)

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	costProviders, total, err := h.costProviderService.GetAllCostProviders(page, limit, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	responses := make([]entities.CostProviderResponse, len(costProviders))
	for i, costProvider := range costProviders {
		responses[i] = costProvider.ToResponse()
	}

	c.JSON(http.StatusOK, models.SuccessListResponse(responses, page, limit, int(total)))
}

// UpdateCostProvider handles updating a cost provider
// @Summary Update a cost provider
// @ID updateCostProvider
// @Description Update a cost provider's information
// @Tags cost-providers
// @Accept json
// @Produce json
// @Param id path int true "Cost Provider ID"
// @Param cost_provider body entities.UpdateCostProviderRequest true "Updated cost provider information"
// @Success 200 {object} models.APIResponse{data=entities.CostProviderResponse} "Updated cost provider"
// @Failure 400 {object} models.APIResponse "Bad request"
// @Failure 401 {object} models.APIResponse "Unauthorized"
// @Failure 404 {object} models.APIResponse "Cost provider not found"
// @Failure 500 {object} models.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /cost-providers/{id} [put]
func (h *CostProviderHandler) UpdateCostProvider(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Invalid cost provider ID"))
		return
	}

	var req entities.UpdateCostProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	costProvider, err := h.costProviderService.UpdateCostProvider(uint(id), tenantID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Cost provider updated successfully", costProvider.ToResponse()))
}

// DeleteCostProvider handles deleting a cost provider
// @Summary Delete a cost provider
// @ID deleteCostProvider
// @Description Soft delete a cost provider by ID
// @Tags cost-providers
// @Produce json
// @Param id path int true "Cost Provider ID"
// @Success 200 {object} models.APIResponse "Cost provider deleted successfully"
// @Failure 400 {object} models.APIResponse "Bad request"
// @Failure 401 {object} models.APIResponse "Unauthorized"
// @Failure 404 {object} models.APIResponse "Cost provider not found"
// @Failure 500 {object} models.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /cost-providers/{id} [delete]
func (h *CostProviderHandler) DeleteCostProvider(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Invalid cost provider ID"))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	err = h.costProviderService.DeleteCostProvider(uint(id), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessMessageResponse("Cost provider deleted successfully"))
}

// SearchCostProviders handles searching cost providers
// @Summary Search cost providers
// @ID searchCostProviders
// @Description Search cost providers by organization name or contact name
// @Tags cost-providers
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of cost providers per page (respects DEFAULT_PAGE_LIMIT and MAX_PAGE_LIMIT env vars)" default(200)
// @Success 200 {object} models.APIResponse{data=models.ListResponse} "Search results"
// @Failure 400 {object} models.APIResponse "Bad request"
// @Failure 401 {object} models.APIResponse "Unauthorized"
// @Failure 500 {object} models.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /cost-providers/search [get]
func (h *CostProviderHandler) SearchCostProviders(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Search query is required"))
		return
	}

	page, limit := utils.GetPaginationParams(c)

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	costProviders, total, err := h.costProviderService.SearchCostProviders(query, page, limit, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	responses := make([]entities.CostProviderResponse, len(costProviders))
	for i, costProvider := range costProviders {
		responses[i] = costProvider.ToResponse()
	}

	c.JSON(http.StatusOK, models.SuccessListResponse(responses, page, limit, int(total)))
}
