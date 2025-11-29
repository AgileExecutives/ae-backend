package handlers

import (
	"net/http"

	"github.com/ae-base-server/internal/models"
	_ "github.com/ae-base-server/modules/base/models" // Import models for swagger
	"github.com/ae-base-server/pkg/core"
	"github.com/ae-base-server/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CustomerHandlers provides customer management handlers
type CustomerHandlers struct {
	db     *gorm.DB
	logger core.Logger
}

// NewCustomerHandlers creates new customer handlers
func NewCustomerHandlers(db *gorm.DB, logger core.Logger) *CustomerHandlers {
	return &CustomerHandlers{
		db:     db,
		logger: logger,
	}
}

// GetCustomers retrieves all customers with pagination
// @Summary Get all customers
// @ID getCustomers
// @Description Get a paginated list of all customers
// @Tags customers
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} models.APIResponse{data=models.ListResponse}
// @Failure 400 {object} models.ErrorResponse
// @Router /customers [get]
func (h *CustomerHandlers) GetCustomers(c *gin.Context) {
	// Get user from context for tenant isolation
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("User not found", "User not authenticated"))
		return
	}
	user := userInterface.(*models.User)

	page, limit := utils.GetPaginationParams(c)
	offset := utils.GetOffset(page, limit)

	var customers []models.Customer
	var total int64

	query := h.db.Model(&models.Customer{}).Where("tenant_id = ?", user.TenantID)

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
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to count customers", err.Error()))
		return
	}

	// Get paginated results with preloaded relationships
	// Note: Tenant and Plan relations temporarily disabled due to GORM relation issues
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&customers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve customers", err.Error()))
		return
	}

	// Convert to response format
	var responses []models.CustomerResponse
	for _, customer := range customers {
		responses = append(responses, customer.ToResponse())
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

	c.JSON(http.StatusOK, models.SuccessResponse("Customers retrieved successfully", response))
}

// GetCustomer retrieves a specific customer by ID
// @Summary Get customer by ID
// @ID getCustomerById
// @Description Get a specific customer by its ID
// @Tags customers
// @Produce json
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Success 200 {object} models.APIResponse{data=models.CustomerResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /customers/{id} [get]
func (h *CustomerHandlers) GetCustomer(c *gin.Context) {
	// Get user from context for tenant isolation
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("User not found", "User not authenticated"))
		return
	}
	user := userInterface.(*models.User)

	id, err := utils.ValidateID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid customer ID", err.Error()))
		return
	}

	var customer models.Customer
	// Note: Plan and Tenant relations temporarily disabled due to GORM relation issues
	if err := h.db.Where("id = ? AND tenant_id = ?", id, user.TenantID).First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Customer not found", "Customer with specified ID does not exist"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve customer", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Customer retrieved successfully", customer.ToResponse()))
}

// CreateCustomer creates a new customer
// @Summary Create a new customer
// @ID createCustomer
// @Description Create a new customer
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param customer body models.CustomerRequest true "Customer data"
// @Success 201 {object} models.APIResponse{data=models.CustomerResponse}
// @Failure 400 {object} models.ErrorResponse
// @Router /customers [post]
func (h *CustomerHandlers) CreateCustomer(c *gin.Context) {
	// Get user from context for tenant isolation
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("User not found", "User not authenticated"))
		return
	}
	user := userInterface.(*models.User)

	var req models.CustomerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	// Ensure customer is created within user's organization
	req.TenantID = user.TenantID

	// Verify the plan exists
	var plan models.Plan
	if err := h.db.First(&plan, req.PlanID).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Plan not found", "Invalid plan ID"))
		return
	}

	customer := models.Customer{
		Name:          req.Name,
		Email:         req.Email,
		Phone:         req.Phone,
		Street:        req.Street,
		Zip:           req.Zip,
		City:          req.City,
		Country:       req.Country,
		TaxID:         req.TaxID,
		VAT:           req.VAT,
		PlanID:        req.PlanID,
		TenantID:      req.TenantID,
		Status:        "active",
		PaymentMethod: req.PaymentMethod,
		Active:        true,
	}

	if err := h.db.Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to create customer", err.Error()))
		return
	}

	// Note: Plan and Tenant relations temporarily disabled due to GORM relation issues
	// h.db.Preload("Plan").Preload("Tenant").First(&customer, customer.ID)

	c.JSON(http.StatusCreated, models.SuccessResponse("Customer created successfully", customer.ToResponse()))
}

// UpdateCustomer updates an existing customer
// @Summary Update a customer
// @ID updateCustomer
// @Description Update an existing customer by ID
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Param customer body models.CustomerRequest true "Updated customer data"
// @Success 200 {object} models.APIResponse{data=models.CustomerResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /customers/{id} [put]
func (h *CustomerHandlers) UpdateCustomer(c *gin.Context) {
	// Get user from context for tenant isolation
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("User not found", "User not authenticated"))
		return
	}
	user := userInterface.(*models.User)

	id, err := utils.ValidateID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid customer ID", err.Error()))
		return
	}

	var req models.CustomerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	var customer models.Customer
	if err := h.db.Where("id = ? AND tenant_id = ?", id, user.TenantID).First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Customer not found", "Customer with specified ID does not exist"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve customer", err.Error()))
		return
	}

	// Update fields if provided
	if req.Name != "" {
		customer.Name = req.Name
	}
	if req.Email != "" {
		customer.Email = req.Email
	}
	if req.Phone != "" {
		customer.Phone = req.Phone
	}
	if req.Street != "" {
		customer.Street = req.Street
	}
	if req.Zip != "" {
		customer.Zip = req.Zip
	}
	if req.City != "" {
		customer.City = req.City
	}
	if req.Country != "" {
		customer.Country = req.Country
	}
	if req.TaxID != "" {
		customer.TaxID = req.TaxID
	}
	if req.VAT != "" {
		customer.VAT = req.VAT
	}
	if req.PlanID != nil {
		// Verify the plan exists
		var plan models.Plan
		if err := h.db.First(&plan, *req.PlanID).Error; err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Plan not found", "Invalid plan ID"))
			return
		}
		customer.PlanID = *req.PlanID
	}
	if req.Status != "" {
		customer.Status = req.Status
	}
	if req.PaymentMethod != "" {
		customer.PaymentMethod = req.PaymentMethod
	}
	if req.Active != nil {
		customer.Active = *req.Active
	}

	if err := h.db.Save(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to update customer", err.Error()))
		return
	}

	// Note: Plan and Tenant relations temporarily disabled due to GORM relation issues
	// h.db.Preload("Plan").Preload("Tenant").First(&customer, customer.ID)

	c.JSON(http.StatusOK, models.SuccessResponse("Customer updated successfully", customer.ToResponse()))
}

// DeleteCustomer deletes a customer
// @Summary Delete a customer
// @ID deleteCustomer
// @Description Soft delete a customer by ID
// @Tags customers
// @Produce json
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /customers/{id} [delete]
func (h *CustomerHandlers) DeleteCustomer(c *gin.Context) {
	// Get user from context for tenant isolation
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("User not found", "User not authenticated"))
		return
	}
	user := userInterface.(*models.User)

	id, err := utils.ValidateID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid customer ID", err.Error()))
		return
	}

	var customer models.Customer
	if err := h.db.Where("id = ? AND tenant_id = ?", id, user.TenantID).First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Customer not found", "Customer with specified ID does not exist"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve customer", err.Error()))
		return
	}

	if err := h.db.Delete(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to delete customer", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Customer deleted successfully", nil))
}

// PlanHandlers provides subscription plan management handlers
type PlanHandlers struct {
	db     *gorm.DB
	logger core.Logger
}

// NewPlanHandlers creates new plan handlers
func NewPlanHandlers(db *gorm.DB, logger core.Logger) *PlanHandlers {
	return &PlanHandlers{
		db:     db,
		logger: logger,
	}
}

// GetPlans retrieves all available plans
// @Summary Get all plans
// @ID getPlans
// @Description Get a list of all available subscription plans
// @Tags plans
// @Produce json
// @Success 200 {array} models.Plan
// @Failure 500 {object} models.ErrorResponse
// @Router /plans [get]
func (h *PlanHandlers) GetPlans(c *gin.Context) {
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
// @ID getPlanById
// @Description Get a specific subscription plan by its ID
// @Tags plans
// @Produce json
// @Param id path int true "Plan ID"
// @Success 200 {object} models.APIResponse{data=models.Plan}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /plans/{id} [get]
func (h *PlanHandlers) GetPlan(c *gin.Context) {
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

// CreatePlan creates a new subscription plan
// @Summary Create a new plan
// @ID createPlan
// @Description Create a new subscription plan (admin only)
// @Tags plans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param plan body models.PlanRequest true "Plan data"
// @Success 201 {object} models.APIResponse{data=models.Plan}
// @Failure 400 {object} models.ErrorResponse
// @Router /plans [post]
func (h *PlanHandlers) CreatePlan(c *gin.Context) {
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
// @ID updatePlan
// @Description Update an existing subscription plan (admin only)
// @Tags plans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Plan ID"
// @Param plan body models.PlanRequest true "Updated plan data"
// @Success 200 {object} models.APIResponse{data=models.Plan}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /plans/{id} [put]
func (h *PlanHandlers) UpdatePlan(c *gin.Context) {
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

// DeletePlan deletes a plan
// @Summary Delete a plan
// @ID deletePlan
// @Description Delete a subscription plan (admin only)
// @Tags plans
// @Produce json
// @Security BearerAuth
// @Param id path int true "Plan ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /plans/{id} [delete]
func (h *PlanHandlers) DeletePlan(c *gin.Context) {
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
