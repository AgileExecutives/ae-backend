package handlers

import (
	"net/http"

	"github.com/ae-saas-basic/ae-saas-basic/internal/models"
	"github.com/ae-saas-basic/ae-saas-basic/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CustomerHandler struct {
	db *gorm.DB
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(db *gorm.DB) *CustomerHandler {
	return &CustomerHandler{db: db}
}

// GetCustomers retrieves all customers with pagination and tenant isolation
// @Summary Get all customers
// @Description Get a paginated list of customers for the authenticated tenant
// @Tags customers
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param active query bool false "Filter by active status"
// @Success 200 {object} models.APIResponse{data=models.ListResponse}
// @Failure 500 {object} models.ErrorResponse
// @Router /customers [get]
func (h *CustomerHandler) GetCustomers(c *gin.Context) {
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

// GetCustomer retrieves a specific customer by ID with tenant isolation
// @Summary Get customer by ID
// @Description Get a specific customer by its ID within the authenticated tenant
// @Tags customers
// @Produce json
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Success 200 {object} models.APIResponse{data=models.CustomerResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /customers/{id} [get]
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
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
// @Description Create a new customer within the authenticated tenant
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CustomerCreateRequest true "Customer creation data"
// @Success 201 {object} models.APIResponse{data=models.CustomerResponse}
// @Failure 400 {object} models.ErrorResponse
// @Router /customers [post]
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
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
// @Description Update an existing customer by ID within the authenticated tenant
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Param request body models.CustomerUpdateRequest true "Customer update data"
// @Success 200 {object} models.APIResponse{data=models.CustomerResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
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

// DeleteCustomer deletes a customer (soft delete)
// @Summary Delete a customer
// @Description Soft delete a customer by ID within the authenticated tenant
// @Tags customers
// @Produce json
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
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
