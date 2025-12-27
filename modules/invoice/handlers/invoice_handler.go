package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unburdy/invoice-module/entities"
	"github.com/unburdy/invoice-module/services"
	"github.com/unburdy/invoice-module/utils"
)

// InvoiceHandler handles invoice-related HTTP requests
type InvoiceHandler struct {
	service *services.InvoiceService
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(service *services.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{service: service}
}

// getTenantID extracts tenant_id from gin context
func getTenantID(c *gin.Context) (uint, error) {
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		return 0, gin.Error{Err: http.ErrAbortHandler, Type: gin.ErrorTypePublic, Meta: "tenant_id required"}
	}
	tenantID, ok := tenantIDValue.(uint)
	if !ok {
		return 0, gin.Error{Err: http.ErrAbortHandler, Type: gin.ErrorTypePublic, Meta: "invalid tenant_id"}
	}
	return tenantID, nil
}

// CreateInvoice creates a new invoice
// @Summary Create a new invoice
// @Description Create a new invoice with line items
// @Tags Invoices
// @Accept json
// @Produce json
// @Param invoice body entities.CreateInvoiceRequest true "Invoice data"
// @Success 201 {object} entities.InvoiceResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /invoices [post]
// @Security BearerAuth
func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id required"})
		return
	}
	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id"})
		return
	}

	var req entities.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.service.CreateInvoice(c.Request.Context(), uint(tenantID), uint(userID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invoice.ToResponse())
}

// GetInvoice retrieves an invoice by ID
// @Summary Get invoice by ID
// @Description Retrieve a single invoice by its ID
// @Tags Invoices
// @Produce json
// @Param id path int true "Invoice ID"
// @Success 200 {object} entities.InvoiceResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /invoices/{id} [get]
// @Security BearerAuth
func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	invoiceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice_id"})
		return
	}

	invoice, err := h.service.GetInvoice(c.Request.Context(), uint(tenantID), uint(invoiceID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		return
	}

	c.JSON(http.StatusOK, invoice.ToResponse())
}

// ListInvoices lists invoices with filters
// @Summary List invoices
// @Description Get a paginated list of invoices with optional filters
// @Tags Invoices
// @Produce json
// @Param organization_id query int false "Filter by organization ID"
// @Param status query string false "Filter by status" Enums(draft, sent, paid, overdue, cancelled)
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /invoices [get]
// @Security BearerAuth
func (h *InvoiceHandler) ListInvoices(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	// Parse filters
	var organizationID *uint
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		if orgID, err := strconv.ParseUint(orgIDStr, 10, 32); err == nil {
			orgIDVal := uint(orgID)
			organizationID = &orgIDVal
		}
	}

	var status *entities.InvoiceStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := entities.InvoiceStatus(statusStr)
		status = &s
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	invoices, total, err := h.service.ListInvoices(c.Request.Context(), uint(tenantID), organizationID, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]entities.InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		responses[i] = invoice.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
		"page":    page,
		"limit":   pageSize,
		"total":   total,
	})
}

// UpdateInvoice updates an invoice
// @Summary Update invoice
// @Description Update an existing invoice's details
// @Tags Invoices
// @Accept json
// @Produce json
// @Param id path int true "Invoice ID"
// @Param invoice body entities.UpdateInvoiceRequest true "Updated invoice data"
// @Success 200 {object} entities.InvoiceResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /invoices/{id} [put]
// @Security BearerAuth
func (h *InvoiceHandler) UpdateInvoice(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	invoiceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice_id"})
		return
	}

	var req entities.UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.service.UpdateInvoice(c.Request.Context(), uint(tenantID), uint(invoiceID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice.ToResponse())
}

// DeleteInvoice deletes an invoice
// @Summary Delete invoice
// @Description Soft delete an invoice by ID
// @Tags Invoices
// @Produce json
// @Param id path int true "Invoice ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /invoices/{id} [delete]
// @Security BearerAuth
func (h *InvoiceHandler) DeleteInvoice(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	invoiceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice_id"})
		return
	}

	if err := h.service.DeleteInvoice(c.Request.Context(), uint(tenantID), uint(invoiceID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Invoice deleted successfully"})
}

// MarkAsPaid marks an invoice as paid
// @Summary Mark invoice as paid
// @Description Mark an invoice as paid with payment date
// @Tags Invoices
// @Accept json
// @Produce json
// @Param id path int true "Invoice ID"
// @Param payment body map[string]string true "Payment date (RFC3339 format)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /invoices/{id}/mark-paid [post]
// @Security BearerAuth
func (h *InvoiceHandler) MarkAsPaid(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	invoiceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice_id"})
		return
	}

	var req struct {
		PaymentDate string `json:"payment_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	paymentDate, err := utils.ParseTime(req.PaymentDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment_date format"})
		return
	}

	if err := h.service.MarkAsPaid(c.Request.Context(), uint(tenantID), uint(invoiceID), paymentDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Invoice marked as paid"})
}
