package handlers

import (
	"net/http"
	"strconv"
	"time"

	baseAPI "github.com/ae-base-server/api"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/unburdy-server-api/modules/documents/services"
	"gorm.io/gorm"
)

// InvoiceNumberHandler handles invoice number generation requests
type InvoiceNumberHandler struct {
	service *services.InvoiceNumberService
	db      *gorm.DB
}

// NewInvoiceNumberHandler creates a new invoice number handler
func NewInvoiceNumberHandler(service *services.InvoiceNumberService, db *gorm.DB) *InvoiceNumberHandler {
	return &InvoiceNumberHandler{
		service: service,
		db:      db,
	}
}

// GenerateInvoiceNumber generates the next invoice number
// @Summary Generate next invoice number
// @Description Generate the next sequential invoice number for an organization
// @Tags Invoice Numbers
// @ID generateInvoiceNumber
// @Accept json
// @Produce json
// @Param request body handlers.GenerateInvoiceNumberRequest true "Invoice number configuration"
// @Success 200 {object} handlers.InvoiceNumberResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /invoice-numbers/generate [post]
// @Security BearerAuth
func (h *InvoiceNumberHandler) GenerateInvoiceNumber(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	var req GenerateInvoiceNumberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build configuration from request or use defaults
	config := services.DefaultInvoiceConfig()

	if req.Prefix != "" {
		config.Prefix = req.Prefix
	}
	if req.YearFormat != "" {
		config.YearFormat = req.YearFormat
	}
	if req.MonthFormat != "" {
		config.MonthFormat = req.MonthFormat
	}
	if req.Padding > 0 {
		config.Padding = req.Padding
	}
	if req.Separator != "" {
		config.Separator = req.Separator
	}
	if req.ResetMonthly != nil {
		config.ResetMonthly = *req.ResetMonthly
	}

	// Generate invoice number
	result, err := h.service.GenerateInvoiceNumber(
		c.Request.Context(),
		uint(tenantID),
		req.OrganizationID,
		config,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetCurrentSequence retrieves the current sequence without incrementing
// @Summary Get current sequence
// @Description Get the current invoice number sequence for an organization
// @Tags Invoice Numbers
// @ID getCurrentInvoiceSequence
// @Produce json
// @Param organization_id query int true "Organization ID"
// @Param year query int false "Year (defaults to current year)"
// @Param month query int false "Month (defaults to current month)"
// @Success 200 {object} handlers.SuccessResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /invoice-numbers/current [get]
// @Security BearerAuth
func (h *InvoiceNumberHandler) GetCurrentSequence(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	organizationIDStr := c.Query("organization_id")
	if organizationIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id required"})
		return
	}

	organizationID, err := strconv.ParseUint(organizationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization_id"})
		return
	}

	// Get year and month (default to current)
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if yearStr := c.Query("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}
	if monthStr := c.Query("month"); monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil {
			month = m
		}
	}

	sequence, err := h.service.GetCurrentSequence(
		c.Request.Context(),
		uint(tenantID),
		uint(organizationID),
		year,
		month,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"organization_id": organizationID,
		"year":            year,
		"month":           month,
		"sequence":        sequence,
	})
}

// GetInvoiceNumberHistory retrieves invoice number history
// @Summary Get invoice number history
// @Description Retrieve the history of generated invoice numbers
// @Tags Invoice Numbers
// @ID getInvoiceNumberHistory
// @Produce json
// @Param organization_id query int true "Organization ID"
// @Param year query int false "Filter by year"
// @Param month query int false "Filter by month"
// @Param page query int false "Page number (default 1)"
// @Param page_size query int false "Page size (default 20)"
// @Success 200 {object} handlers.SuccessResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /invoice-numbers/history [get]
// @Security BearerAuth
func (h *InvoiceNumberHandler) GetInvoiceNumberHistory(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	organizationIDStr := c.Query("organization_id")
	if organizationIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id required"})
		return
	}

	organizationID, err := strconv.ParseUint(organizationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization_id"})
		return
	}

	// Optional filters
	year := 0
	month := 0
	if yearStr := c.Query("year"); yearStr != "" {
		year, _ = strconv.Atoi(yearStr)
	}
	if monthStr := c.Query("month"); monthStr != "" {
		month, _ = strconv.Atoi(monthStr)
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

	logs, total, err := h.service.GetInvoiceNumberHistory(
		c.Request.Context(),
		uint(tenantID),
		uint(organizationID),
		year,
		month,
		page,
		pageSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        logs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (int(total) + pageSize - 1) / pageSize,
	})
}

// VoidInvoiceNumber marks an invoice number as voided
// @Summary Void invoice number
// @Description Mark an invoice number as voided in the audit log
// @Tags Invoice Numbers
// @ID voidInvoiceNumber
// @Accept json
// @Produce json
// @Param request body map[string]string true "Invoice number to void"
// @Success 200 {object} handlers.SuccessResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /invoice-numbers/void [post]
// @Security BearerAuth
func (h *InvoiceNumberHandler) VoidInvoiceNumber(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	var req struct {
		InvoiceNumber string `json:"invoice_number" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.VoidInvoiceNumber(c.Request.Context(), uint(tenantID), req.InvoiceNumber); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invoice number voided successfully"})
}
