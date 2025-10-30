package handlers

import (
	"net/http"

	"github.com/ae-base-server/internal/models"
	"github.com/ae-base-server/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PlanHandler struct {
	db *gorm.DB
}

// NewPlanHandler creates a new plan handler
func NewPlanHandler(db *gorm.DB) *PlanHandler {
	return &PlanHandler{db: db}
}

// GetPlans retrieves all plans with pagination
// @Summary Get all plans
// @Description Get a paginated list of all plans
// @Tags plans
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param active query bool false "Filter by active status"
// @Success 200 {object} models.APIResponse{data=models.ListResponse}
// @Failure 500 {object} models.ErrorResponse
// @Router /plans [get]
func (h *PlanHandler) GetPlans(c *gin.Context) {
	page, limit := utils.GetPaginationParams(c)
	offset := utils.GetOffset(page, limit)

	var plans []models.Plan
	var total int64

	query := h.db.Model(&models.Plan{})

	// Filter by active status if provided
	if activeStr := c.Query("active"); activeStr != "" {
		if activeStr == "true" {
			query = query.Where("active = ?", true)
		} else if activeStr == "false" {
			query = query.Where("active = ?", false)
		}
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to count plans", err.Error()))
		return
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve plans", err.Error()))
		return
	}

	// Convert to response format
	var responses []models.PlanResponse
	for _, plan := range plans {
		responses = append(responses, plan.ToResponse())
	}

	response := models.ListResponse{
		Data: responses,
		Pagination: models.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      int(total),
			TotalPages: utils.CalculateTotalPages(int(total), limit),
		},
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Plans retrieved successfully", response))
}

// GetPlan retrieves a specific plan by ID
// @Summary Get plan by ID
// @Description Get a specific plan by its ID
// @Tags plans
// @Produce json
// @Param id path int true "Plan ID"
// @Success 200 {object} models.APIResponse{data=models.PlanResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /plans/{id} [get]
func (h *PlanHandler) GetPlan(c *gin.Context) {
	id, err := utils.ValidateID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid plan ID", err.Error()))
		return
	}

	var plan models.Plan
	if err := h.db.First(&plan, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Plan not found", "Plan with specified ID does not exist"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve plan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Plan retrieved successfully", plan.ToResponse()))
}

// CreatePlan creates a new plan
// @Summary Create a new plan
// @Description Create a new subscription plan
// @Tags plans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.PlanCreateRequest true "Plan creation data"
// @Success 201 {object} models.APIResponse{data=models.PlanResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Router /plans [post]
func (h *PlanHandler) CreatePlan(c *gin.Context) {
	var req models.PlanCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	// Check if plan with same slug exists
	var existingPlan models.Plan
	if err := h.db.Where("slug = ?", req.Slug).First(&existingPlan).Error; err == nil {
		c.JSON(http.StatusConflict, models.ErrorResponseFunc("Plan already exists", "Plan with this slug already exists"))
		return
	}

	// Set default values
	if req.Currency == "" {
		req.Currency = "EUR"
	}
	if req.InvoicePeriod == "" {
		req.InvoicePeriod = "monthly"
	}
	if req.MaxUsers == 0 {
		req.MaxUsers = 10
	}
	if req.MaxClients == 0 {
		req.MaxClients = 100
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}

	plan := models.Plan{
		Name:          req.Name,
		Slug:          req.Slug,
		Description:   req.Description,
		Price:         req.Price,
		Currency:      req.Currency,
		InvoicePeriod: req.InvoicePeriod,
		MaxUsers:      req.MaxUsers,
		MaxClients:    req.MaxClients,
		Features:      req.Features,
		Active:        active,
	}

	if err := h.db.Create(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to create plan", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse("Plan created successfully", plan.ToResponse()))
}

// UpdatePlan updates an existing plan
// @Summary Update a plan
// @Description Update an existing plan by ID
// @Tags plans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Plan ID"
// @Param request body models.PlanUpdateRequest true "Plan update data"
// @Success 200 {object} models.APIResponse{data=models.PlanResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /plans/{id} [put]
func (h *PlanHandler) UpdatePlan(c *gin.Context) {
	id, err := utils.ValidateID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid plan ID", err.Error()))
		return
	}

	var req models.PlanUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	var plan models.Plan
	if err := h.db.First(&plan, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Plan not found", "Plan with specified ID does not exist"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve plan", err.Error()))
		return
	}

	// Update fields if provided
	if req.Name != "" {
		plan.Name = req.Name
	}
	if req.Description != "" {
		plan.Description = req.Description
	}
	if req.Price != nil {
		plan.Price = *req.Price
	}
	if req.Currency != "" {
		plan.Currency = req.Currency
	}
	if req.InvoicePeriod != "" {
		plan.InvoicePeriod = req.InvoicePeriod
	}
	if req.MaxUsers != nil {
		plan.MaxUsers = *req.MaxUsers
	}
	if req.MaxClients != nil {
		plan.MaxClients = *req.MaxClients
	}
	if req.Features != "" {
		plan.Features = req.Features
	}
	if req.Active != nil {
		plan.Active = *req.Active
	}

	if err := h.db.Save(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to update plan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Plan updated successfully", plan.ToResponse()))
}

// DeletePlan deletes a plan (soft delete)
// @Summary Delete a plan
// @Description Soft delete a plan by ID
// @Tags plans
// @Produce json
// @Security BearerAuth
// @Param id path int true "Plan ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /plans/{id} [delete]
func (h *PlanHandler) DeletePlan(c *gin.Context) {
	id, err := utils.ValidateID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid plan ID", err.Error()))
		return
	}

	var plan models.Plan
	if err := h.db.First(&plan, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Plan not found", "Plan with specified ID does not exist"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve plan", err.Error()))
		return
	}

	if err := h.db.Delete(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to delete plan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Plan deleted successfully", nil))
}
