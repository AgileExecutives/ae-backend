package handlers

import (
	"net/http"
	"strconv"

	baseAPI "github.com/ae-base-server/api"
	"github.com/ae-base-server/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/organization-module/entities"
	"github.com/unburdy/organization-module/services"
)

// OrganizationHandler handles organization-related HTTP requests
type OrganizationHandler struct {
	service *services.OrganizationService
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(service *services.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		service: service,
	}
}

// CreateOrganization handles creating a new organization
// @Summary Create a new organization
// @Description Create a new organization with the provided information
// @Tags organizations
// @ID createOrganization
// @Accept json
// @Produce json
// @Param organization body entities.CreateOrganizationRequest true "Organization information"
// @Success 201 {object} entities.OrganizationAPIResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /organizations [post]
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var req entities.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get user ID: "+err.Error()))
		return
	}

	organization, err := h.service.CreateOrganization(req, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, baseAPI.SuccessResponse("Organization created successfully", organization.ToResponse()))
}

// GetOrganization handles retrieving an organization by ID
// @Summary Get an organization by ID
// @Description Retrieve a specific organization by ID
// @Tags organizations
// @ID getOrganizationById
// @Produce json
// @Param id path int true "Organization ID"
// @Success 200 {object} entities.OrganizationAPIResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 404 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /organizations/{id} [get]
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", "Invalid organization ID"))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get user ID: "+err.Error()))
		return
	}

	organization, err := h.service.GetOrganizationByID(uint(id), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("Not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, baseAPI.SuccessResponse("", organization.ToResponse()))
}

// GetAllOrganizations handles retrieving all organizations with pagination
// @Summary Get all organizations
// @Description Retrieve all organizations for the authenticated user with pagination
// @Tags organizations
// @ID getOrganizations
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} entities.OrganizationListAPIResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /organizations [get]
func (h *OrganizationHandler) GetAllOrganizations(c *gin.Context) {
	page, limit := utils.GetPaginationParams(c)

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get user ID: "+err.Error()))
		return
	}

	organizations, total, err := h.service.GetOrganizations(page, limit, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	responses := make([]entities.OrganizationResponse, len(organizations))
	for i, org := range organizations {
		responses[i] = org.ToResponse()
	}

	c.JSON(http.StatusOK, baseAPI.SuccessListResponse(responses, page, limit, int(total)))
}

// UpdateOrganization handles updating an organization
// @Summary Update an organization
// @Description Update an organization's information
// @Tags organizations
// @ID updateOrganization
// @Accept json
// @Produce json
// @Param id path int true "Organization ID"
// @Param organization body entities.UpdateOrganizationRequest true "Updated organization information"
// @Success 200 {object} entities.OrganizationAPIResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 404 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", "Invalid organization ID"))
		return
	}

	var req entities.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get user ID: "+err.Error()))
		return
	}

	organization, err := h.service.UpdateOrganization(uint(id), tenantID, userID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("Not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, baseAPI.SuccessResponse("Organization updated successfully", organization.ToResponse()))
}

// DeleteOrganization handles deleting an organization
// @Summary Delete an organization
// @Description Delete an organization by ID
// @Tags organizations
// @ID deleteOrganization
// @Produce json
// @Param id path int true "Organization ID"
// @Success 200 {object} entities.OrganizationDeleteResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 404 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", "Invalid organization ID"))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get user ID: "+err.Error()))
		return
	}

	if err := h.service.DeleteOrganization(uint(id), tenantID, userID); err != nil {
		c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("Not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, baseAPI.SuccessMessageResponse("Organization deleted successfully"))
}
