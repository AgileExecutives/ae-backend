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

// InvoiceHandler handles invoice-related HTTP requests
type InvoiceHandler struct {
	service *services.InvoiceService
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(service *services.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{
		service: service,
	}
}

// CreateInvoice handles creating a new invoice
// @Summary Create a new invoice
// @Description Create a new invoice with invoice items for specified sessions
// @Tags invoices
// @ID createInvoice
// @Accept json
// @Produce json
// @Param invoice body entities.CreateInvoiceRequest true "Invoice information with client ID and session IDs"
// @Success 201 {object} entities.InvoiceAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /invoices [post]
func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	var req entities.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get user ID: "+err.Error()))
		return
	}

	invoice, err := h.service.CreateInvoice(req, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to create invoice", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse("Invoice created successfully", invoice.ToResponse()))
}

// GetInvoice handles retrieving an invoice by ID
// @Summary Get an invoice by ID
// @Description Retrieve a specific invoice by ID with all associations preloaded
// @Tags invoices
// @ID getInvoiceById
// @Produce json
// @Param id path int true "Invoice ID"
// @Success 200 {object} entities.InvoiceAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /invoices/{id} [get]
func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Invalid invoice ID"))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get user ID: "+err.Error()))
		return
	}

	invoice, err := h.service.GetInvoiceByID(uint(id), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Invoice retrieved successfully", invoice.ToResponse()))
}

// GetAllInvoices handles retrieving all invoices with pagination
// @Summary Get all invoices
// @Description Retrieve all invoices for the authenticated user with pagination and all associations preloaded
// @Tags invoices
// @ID getInvoices
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} entities.InvoiceListAPIResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /invoices [get]
func (h *InvoiceHandler) GetAllInvoices(c *gin.Context) {
	page, limit := utils.GetPaginationParams(c)

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get user ID: "+err.Error()))
		return
	}

	invoices, total, err := h.service.GetInvoices(page, limit, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to fetch invoices", err.Error()))
		return
	}

	responses := make([]entities.InvoiceResponse, len(invoices))
	for i, inv := range invoices {
		responses[i] = inv.ToResponse()
	}

	c.JSON(http.StatusOK, models.SuccessListResponse(responses, page, limit, int(total)))
}

// UpdateInvoice handles updating an invoice
// @Summary Update an invoice
// @Description Update an invoice's status or invoice items (not both at once)
// @Tags invoices
// @ID updateInvoice
// @Accept json
// @Produce json
// @Param id path int true "Invoice ID"
// @Param invoice body entities.UpdateInvoiceRequest true "Updated invoice information (status OR session_ids)"
// @Success 200 {object} entities.InvoiceAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /invoices/{id} [put]
func (h *InvoiceHandler) UpdateInvoice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Invalid invoice ID"))
		return
	}

	var req entities.UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get user ID: "+err.Error()))
		return
	}

	invoice, err := h.service.UpdateInvoice(uint(id), tenantID, userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Failed to update invoice", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Invoice updated successfully", invoice.ToResponse()))
}

// DeleteInvoice handles deleting an invoice
// @Summary Delete an invoice
// @Description Delete an invoice and all its invoice items by ID
// @Tags invoices
// @ID deleteInvoice
// @Produce json
// @Param id path int true "Invoice ID"
// @Success 200 {object} entities.InvoiceDeleteResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /invoices/{id} [delete]
func (h *InvoiceHandler) DeleteInvoice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Invalid invoice ID"))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get user ID: "+err.Error()))
		return
	}

	if err := h.service.DeleteInvoice(uint(id), tenantID, userID); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Failed to delete invoice", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessMessageResponse("Invoice deleted successfully"))
}

// GetClientsWithUnbilledSessions handles retrieving clients with unbilled sessions
// @Summary Get clients with unbilled sessions
// @Description Retrieve all clients that have sessions not yet associated with any invoice
// @Tags invoices
// @ID getClientsWithUnbilledSessions
// @Produce json
// @Success 200 {object} entities.ClientSessionsAPIResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /invoices/clientsessions [get]
func (h *InvoiceHandler) GetClientsWithUnbilledSessions(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get user ID: "+err.Error()))
		return
	}

	clients, err := h.service.GetClientsWithUnbilledSessions(tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to fetch clients with unbilled sessions", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Clients with unbilled sessions retrieved successfully", clients))
}
