package handlers

import (
	"net/http"
	"strconv"
	"strings"

	baseAPI "github.com/ae-base-server/api"
	"github.com/ae-base-server/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/unburdy-server-api/internal/models"
	"github.com/unburdy/unburdy-server-api/internal/services"
	auditEntities "github.com/unburdy/unburdy-server-api/modules/audit/entities"
	auditServices "github.com/unburdy/unburdy-server-api/modules/audit/services"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	clientServices "github.com/unburdy/unburdy-server-api/modules/client_management/services"
)

// InvoiceHandler handles invoice-related HTTP requests
type InvoiceHandler struct {
	service          *clientServices.InvoiceService
	xrechnungService *services.XRechnungService
	auditService     *auditServices.AuditService
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(service *clientServices.InvoiceService, xrechnungService *services.XRechnungService, auditServiceRaw interface{}) *InvoiceHandler {
	var auditService *auditServices.AuditService
	if auditServiceRaw != nil {
		if svc, ok := auditServiceRaw.(*auditServices.AuditService); ok {
			auditService = svc
		}
	}

	return &InvoiceHandler{
		service:          service,
		xrechnungService: xrechnungService,
		auditService:     auditService,
	}
}

// logAudit logs an audit event if audit service is available
func (h *InvoiceHandler) logAudit(c *gin.Context, action auditEntities.AuditAction, entityID uint, metadata *auditEntities.AuditLogMetadata) {
	if h.auditService == nil {
		return
	}

	tenantID, _ := baseAPI.GetTenantID(c)
	userID, _ := baseAPI.GetUserID(c)

	_ = h.auditService.LogEvent(auditServices.LogEventRequest{
		TenantID:   tenantID,
		UserID:     userID,
		EntityType: auditEntities.EntityTypeInvoice,
		EntityID:   entityID,
		Action:     action,
		Metadata:   metadata,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
}

// CreateInvoice handles creating a new invoice
// @Summary Create a new invoice
// @Description Create a new invoice with invoice items for specified sessions
// @Tags client-invoices
// @ID createInvoice
// @Accept json
// @Produce json
// @Param invoice body entities.CreateInvoiceRequest true "Invoice information with client ID and session IDs"
// @Success 201 {object} entities.InvoiceAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices [post]
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

// CreateDraftInvoice handles creating a new draft invoice
// @Summary Create a draft invoice
// @Description Create a new draft invoice with sessions, extra efforts, and/or custom line items. Items are reserved with status 'invoice-draft'.
// @Tags client-invoices
// @ID createDraftInvoice
// @Accept json
// @Produce json
// @Param invoice body entities.CreateDraftInvoiceRequest true "Draft invoice information"
// @Success 201 {object} entities.InvoiceAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/draft [post]
func (h *InvoiceHandler) CreateDraftInvoice(c *gin.Context) {
	var req entities.CreateDraftInvoiceRequest
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

	invoice, err := h.service.CreateDraftInvoice(req, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to create draft invoice", err.Error()))
		return
	}

	// Log audit event
	h.logAudit(c, auditEntities.AuditActionInvoiceDraftCreated, invoice.ID, &auditEntities.AuditLogMetadata{
		TotalAmount: invoice.TotalAmount,
		AdditionalInfo: map[string]interface{}{
			"num_sessions":     len(req.SessionIDs),
			"num_efforts":      len(req.ExtraEffortIDs),
			"num_custom_items": len(req.CustomLineItems),
		},
	})

	c.JSON(http.StatusCreated, models.SuccessResponse("Draft invoice created successfully", invoice.ToResponse()))
}

// UpdateDraftInvoice handles editing an existing draft invoice
// @Summary Update a draft invoice
// @Description Edit a draft invoice by adding/removing items and recalculating totals
// @Tags invoices
// @ID updateDraftInvoice
// @Accept json
// @Produce json
// @Param id path int true "Invoice ID"
// @Param request body entities.UpdateDraftInvoiceRequest true "Update request"
// @Success 200 {object} entities.InvoiceAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id} [put]
func (h *InvoiceHandler) UpdateDraftInvoice(c *gin.Context) {
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

	var req entities.UpdateDraftInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	invoice, err := h.service.UpdateDraftInvoice(uint(id), tenantID, userID, req)
	if err != nil {
		if err.Error() == "can only edit invoices in draft status" {
			c.JSON(http.StatusForbidden, models.ErrorResponseFunc("Forbidden", err.Error()))
			return
		}
		if err.Error() == "invoice not found: record not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", "Invoice not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	// Log audit event
	h.logAudit(c, auditEntities.AuditActionInvoiceDraftUpdated, invoice.ID, &auditEntities.AuditLogMetadata{
		TotalAmount: invoice.TotalAmount,
		AdditionalInfo: map[string]interface{}{
			"add_sessions":     len(req.AddSessionIDs),
			"add_efforts":      len(req.AddExtraEffortIDs),
			"num_custom_items": len(req.CustomLineItems),
		},
	})

	c.JSON(http.StatusOK, models.SuccessResponse("Draft invoice updated successfully", invoice.ToResponse()))
}

// CancelDraftInvoice handles canceling a draft invoice
// @Summary Cancel a draft invoice
// @Description Cancel a draft invoice and revert all item statuses
// @Tags invoices
// @ID cancelDraftInvoice
// @Produce json
// @Param id path int true "Invoice ID"
// @Success 200 {object} map[string]interface{} "Success message"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id} [delete]
func (h *InvoiceHandler) CancelDraftInvoice(c *gin.Context) {
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

	err = h.service.CancelDraftInvoice(uint(id), tenantID, userID)
	if err != nil {
		if err.Error() == "can only cancel invoices in draft status" {
			c.JSON(http.StatusForbidden, models.ErrorResponseFunc("Forbidden", err.Error()))
			return
		}
		if err.Error() == "invoice not found: record not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", "Invoice not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	// Log audit event
	h.logAudit(c, auditEntities.AuditActionInvoiceDraftCancelled, uint(id), &auditEntities.AuditLogMetadata{
		AdditionalInfo: map[string]interface{}{
			"invoice_id": id,
		},
	})

	c.JSON(http.StatusOK, models.SuccessResponse("Draft invoice cancelled successfully", nil))
}

// FinalizeInvoice handles finalizing a draft invoice
// @Summary Finalize a draft invoice
// @Description Finalize a draft invoice by generating invoice number and changing status to 'finalized'
// @Tags invoices
// @ID finalizeInvoice
// @Accept json
// @Produce json
// @Param id path int true "Invoice ID"
// @Param request body entities.FinalizeInvoiceRequest false "Optional customer data to update"
// @Success 200 {object} entities.InvoiceAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 422 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id}/finalize [post]
func (h *InvoiceHandler) FinalizeInvoice(c *gin.Context) {
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

	// Parse optional request body for customer data
	var req entities.FinalizeInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body or invalid JSON, that's okay - use nil
		req = entities.FinalizeInvoiceRequest{}
	}

	invoice, err := h.service.FinalizeInvoice(uint(id), tenantID, userID, &req)
	if err != nil {
		if err.Error() == "can only finalize invoices in draft status" {
			c.JSON(http.StatusForbidden, models.ErrorResponseFunc("Forbidden", err.Error()))
			return
		}
		if err.Error() == "invoice not found: record not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", "Invoice not found"))
			return
		}
		// Validation errors return 422
		if err.Error() == "invoice must have at least one line item" ||
			strings.Contains(err.Error(), "missing VAT") ||
			strings.Contains(err.Error(), "government customer") {
			c.JSON(http.StatusUnprocessableEntity, models.ErrorResponseFunc("Validation error", err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	// Log audit event
	h.logAudit(c, auditEntities.AuditActionInvoiceFinalized, uint(id), &auditEntities.AuditLogMetadata{
		TotalAmount: invoice.TotalAmount,
		AdditionalInfo: map[string]interface{}{
			"invoice_number": invoice.InvoiceNumber,
			"status":         invoice.Status,
		},
	})

	c.JSON(http.StatusOK, models.SuccessResponse("Invoice finalized successfully", invoice.ToResponse()))
}

// MarkInvoiceAsSent marks a finalized invoice as sent
// @Summary Mark invoice as sent
// @Description Mark a finalized invoice as sent (changes status from finalized to sent). Requires send_method to be specified.
// @Tags client-invoices
// @ID markInvoiceAsSent
// @Accept json
// @Produce json
// @Param id path int true "Invoice ID"
// @Param request body entities.MarkInvoiceAsSentRequest true "Send method (email, manual, xrechnung)"
// @Success 200 {object} entities.InvoiceAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id}/mark-sent [post]
func (h *InvoiceHandler) MarkInvoiceAsSent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Invalid invoice ID"))
		return
	}

	var req entities.MarkInvoiceAsSentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	invoice, err := h.service.MarkAsSent(uint(id), tenantID, req.SendMethod)
	if err != nil {
		if err.Error() == "can only mark finalized invoices as sent" {
			c.JSON(http.StatusForbidden, models.ErrorResponseFunc("Forbidden", err.Error()))
			return
		}
		if strings.Contains(err.Error(), "invoice not found") {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", "Invoice not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	// Log audit event
	h.logAudit(c, auditEntities.AuditActionInvoiceSent, uint(id), &auditEntities.AuditLogMetadata{
		TotalAmount: invoice.TotalAmount,
		AdditionalInfo: map[string]interface{}{
			"invoice_number": invoice.InvoiceNumber,
			"status":         invoice.Status,
		},
	})

	c.JSON(http.StatusOK, models.SuccessResponse("Invoice marked as sent successfully", invoice.ToResponse()))
}

// SendInvoiceEmail handles sending an invoice via email
// @Summary Send invoice via email
// @Description Send an invoice to the client via email
// @Tags invoices
// @ID sendInvoiceEmail
// @Produce json
// @Param id path int true "Invoice ID"
// @Success 200 {object} map[string]interface{} "Success message"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id}/send-email [post]
func (h *InvoiceHandler) SendInvoiceEmail(c *gin.Context) {
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

	err = h.service.SendInvoiceEmail(uint(id), tenantID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "can only send email") {
			c.JSON(http.StatusForbidden, models.ErrorResponseFunc("Forbidden", err.Error()))
			return
		}
		if err.Error() == "invoice not found: record not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", "Invoice not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	// Log audit event
	h.logAudit(c, auditEntities.AuditActionInvoiceSent, uint(id), &auditEntities.AuditLogMetadata{
		AdditionalInfo: map[string]interface{}{
			"invoice_id": id,
		},
	})

	c.JSON(http.StatusOK, models.SuccessResponse("Invoice email sent successfully", nil))
}

// MarkInvoiceAsPaid handles marking an invoice as paid
// @Summary Mark an invoice as paid
// @Description Mark an invoice as paid with optional payment date and reference
// @Tags invoices
// @ID markInvoiceAsPaid
// @Accept json
// @Produce json
// @Param id path int true "Invoice ID"
// @Param request body entities.MarkInvoiceAsPaidRequest false "Payment details"
// @Success 200 {object} entities.InvoiceAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id}/mark-paid [post]
func (h *InvoiceHandler) MarkInvoiceAsPaid(c *gin.Context) {
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

	var req entities.MarkInvoiceAsPaidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body is provided, that's okay - we'll use defaults
		req = entities.MarkInvoiceAsPaidRequest{}
	}

	invoice, err := h.service.MarkInvoiceAsPaid(uint(id), tenantID, userID, req.PaymentDate, req.PaymentReference)
	if err != nil {
		if strings.Contains(err.Error(), "can only mark invoices as paid") {
			c.JSON(http.StatusForbidden, models.ErrorResponseFunc("Forbidden", err.Error()))
			return
		}
		if err.Error() == "invoice not found: record not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", "Invoice not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	// Log audit event
	metadata := &auditEntities.AuditLogMetadata{
		TotalAmount: invoice.TotalAmount,
		AdditionalInfo: map[string]interface{}{
			"invoice_number": invoice.InvoiceNumber,
		},
	}
	if req.PaymentDate != nil {
		metadata.AdditionalInfo["payment_date"] = req.PaymentDate.Format("2006-01-02")
	}
	if req.PaymentReference != "" {
		metadata.AdditionalInfo["payment_reference"] = req.PaymentReference
	}
	h.logAudit(c, auditEntities.AuditActionInvoiceMarkedPaid, uint(id), metadata)

	c.JSON(http.StatusOK, models.SuccessResponse("Invoice marked as paid successfully", invoice.ToResponse()))
}

// MarkInvoiceAsOverdue handles marking an invoice as overdue
// @Summary Mark an invoice as overdue
// @Description Mark an invoice as overdue if payment is past due date
// @Tags invoices
// @ID markInvoiceAsOverdue
// @Produce json
// @Param id path int true "Invoice ID"
// @Success 200 {object} entities.InvoiceAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 422 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id}/mark-overdue [post]
func (h *InvoiceHandler) MarkInvoiceAsOverdue(c *gin.Context) {
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

	invoice, err := h.service.MarkInvoiceAsOverdue(uint(id), tenantID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "can only mark invoices as overdue") {
			c.JSON(http.StatusForbidden, models.ErrorResponseFunc("Forbidden", err.Error()))
			return
		}
		if strings.Contains(err.Error(), "not yet overdue") {
			c.JSON(http.StatusUnprocessableEntity, models.ErrorResponseFunc("Validation error", err.Error()))
			return
		}
		if err.Error() == "invoice not found: record not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", "Invoice not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	// Log audit event
	h.logAudit(c, auditEntities.AuditActionInvoiceMarkedOverdue, uint(id), &auditEntities.AuditLogMetadata{
		TotalAmount: invoice.TotalAmount,
		AdditionalInfo: map[string]interface{}{
			"invoice_number": invoice.InvoiceNumber,
		},
	})

	c.JSON(http.StatusOK, models.SuccessResponse("Invoice marked as overdue successfully", invoice.ToResponse()))
}

// SendReminder handles sending a payment reminder for an overdue invoice
// @Summary Send payment reminder
// @Description Send a payment reminder email for an overdue invoice
// @Tags invoices
// @ID sendReminder
// @Produce json
// @Param id path int true "Invoice ID"
// @Success 200 {object} map[string]interface{} "Success message"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id}/reminder [post]
func (h *InvoiceHandler) SendReminder(c *gin.Context) {
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

	err = h.service.SendReminder(uint(id), tenantID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "can only send reminders") {
			c.JSON(http.StatusForbidden, models.ErrorResponseFunc("Forbidden", err.Error()))
			return
		}
		if err.Error() == "invoice not found: record not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", "Invoice not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	// Log audit event
	h.logAudit(c, auditEntities.AuditActionReminderSent, uint(id), &auditEntities.AuditLogMetadata{
		AdditionalInfo: map[string]interface{}{
			"invoice_id": id,
		},
	})

	c.JSON(http.StatusOK, models.SuccessResponse("Reminder sent successfully", nil))
}

// CreateCreditNote handles creating a credit note for an existing invoice
// @Summary Create a credit note
// @Description Create a credit note for an existing invoice with selected line items
// @Tags invoices
// @ID createCreditNote
// @Accept json
// @Produce json
// @Param id path int true "Original Invoice ID"
// @Param request body entities.CreateCreditNoteRequest true "Credit note details"
// @Success 201 {object} entities.InvoiceAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 422 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id}/credit-note [post]
func (h *InvoiceHandler) CreateCreditNote(c *gin.Context) {
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

	var req entities.CreateCreditNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	creditNote, err := h.service.CreateCreditNote(uint(id), tenantID, userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "can only create credit notes") {
			c.JSON(http.StatusForbidden, models.ErrorResponseFunc("Forbidden", err.Error()))
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", err.Error()))
			return
		}
		if strings.Contains(err.Error(), "do not belong to this invoice") {
			c.JSON(http.StatusUnprocessableEntity, models.ErrorResponseFunc("Validation error", err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	// Log audit event
	h.logAudit(c, auditEntities.AuditActionCreditNoteCreated, creditNote.ID, &auditEntities.AuditLogMetadata{
		TotalAmount: creditNote.TotalAmount,
		AdditionalInfo: map[string]interface{}{
			"original_invoice_id": id,
			"credit_note_number":  creditNote.InvoiceNumber,
			"reason":              req.Reason,
			"num_items":           len(req.LineItemIDs),
		},
	})

	c.JSON(http.StatusCreated, models.SuccessResponse("Credit note created successfully", creditNote.ToResponse()))
}

// GetInvoice handles retrieving an invoice by ID
// @Summary Get an invoice by ID
// @Description Retrieve a specific invoice by ID with all associations preloaded
// @Tags client-invoices
// @ID getInvoiceById
// @Produce json
// @Param id path int true "Invoice ID"
// @Success 200 {object} entities.InvoiceAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id} [get]
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

	// Calculate VAT breakdown
	vatService := clientServices.NewVATService()
	vatSummary := vatService.CalculateInvoiceVAT(invoice.InvoiceItems)

	// Convert VAT summary to response
	vatBreakdown := &entities.VATBreakdownResponse{
		Subtotal:   vatSummary.SubtotalAmount,
		TotalTax:   vatSummary.TaxAmount,
		GrandTotal: vatSummary.TotalAmount,
		Items:      make([]entities.VATBreakdownItemResponse, len(vatSummary.VATBreakdown)),
	}

	for i, item := range vatSummary.VATBreakdown {
		vatBreakdown.Items[i] = entities.VATBreakdownItemResponse{
			VATRate:       item.Rate,
			NetAmount:     item.NetAmount,
			TaxAmount:     item.TaxAmount,
			GrossAmount:   item.GrossAmount,
			ExemptionText: item.ExemptionText,
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Invoice retrieved successfully", invoice.ToResponseWithVATBreakdown(vatBreakdown)))
}

// GetAllInvoices handles retrieving all invoices with pagination
// @Summary Get all invoices
// @Description Retrieve all invoices for the authenticated user with pagination and all associations preloaded
// @Tags client-invoices
// @ID getInvoices
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {object} entities.InvoiceListAPIResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices [get]
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
// @Tags client-invoices
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
// @Router /client-invoices/{id} [put]
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
// @Tags client-invoices
// @ID deleteInvoice
// @Produce json
// @Param id path int true "Invoice ID"
// @Success 200 {object} entities.InvoiceDeleteResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id} [delete]
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

// CancelInvoice handles basic invoice cancellation according to GoBD requirements
// @Summary Cancel an invoice (basic)
// @Description Cancel an invoice that has not been sent (sent_at IS NULL). Does not revert sessions/extra efforts.
// @Tags invoices
// @ID cancelInvoice
// @Accept json
// @Produce json
// @Param id path int true "Invoice ID"
// @Param request body entities.CancelInvoiceRequest true "Cancellation reason"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse "Invoice already sent or invalid request"
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /invoices/{id}/cancel [post]
func (h *InvoiceHandler) CancelInvoice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Invalid invoice ID"))
		return
	}

	var req entities.CancelInvoiceRequest
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

	if err := h.service.CancelInvoice(uint(id), tenantID, userID, req.Reason); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Invoice not found", err.Error()))
		} else if strings.Contains(errMsg, "been sent") || strings.Contains(errMsg, "no number") || strings.Contains(errMsg, "already cancelled") {
			c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Cannot cancel invoice", err.Error()))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to cancel invoice", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessMessageResponse("Invoice cancelled successfully."))
}

// CancelClientInvoice handles canceling a client invoice with session/extra effort reversion
// @Summary Cancel a client invoice (extended)
// @Description Cancel a client invoice that has not been sent and revert all sessions to 'conducted' and extra efforts to 'unbilled' status
// @Tags client-invoices
// @ID cancelClientInvoice
// @Accept json
// @Produce json
// @Param id path int true "Invoice ID"
// @Param request body entities.CancelInvoiceRequest true "Cancellation reason"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse "Invoice already sent or invalid request"
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id}/cancel [post]
func (h *InvoiceHandler) CancelClientInvoice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Invalid invoice ID"))
		return
	}

	var req entities.CancelInvoiceRequest
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

	if err := h.service.CancelClientInvoice(uint(id), tenantID, userID, req.Reason); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Invoice not found", err.Error()))
		} else if strings.Contains(errMsg, "been sent") || strings.Contains(errMsg, "no number") || strings.Contains(errMsg, "already cancelled") {
			c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Cannot cancel invoice", err.Error()))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to cancel invoice", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessMessageResponse("Client invoice cancelled successfully. Sessions and extra efforts reverted to unbilled status."))
}

// GetClientsWithUnbilledSessions handles retrieving clients with unbilled sessions
// @Summary Get clients with unbilled sessions
// @Description Retrieve all clients that have sessions not yet associated with any invoice
// @Tags client-invoices
// @ID getClientsWithUnbilledSessions
// @Produce json
// @Success 200 {object} entities.ClientSessionsAPIResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/unbilled-sessions [get]
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

// ExportXRechnung handles exporting an invoice as XRechnung XML for German government invoicing
// @Summary Export invoice as XRechnung XML
// @Description Generate and download XRechnung-compliant UBL XML for a finalized invoice to a government customer
// @Tags client-invoices
// @ID exportXRechnung
// @Produce application/xml
// @Param id path int true "Invoice ID"
// @Success 200 {file} string "XRechnung XML file"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id}/xrechnung [get]
func (h *InvoiceHandler) ExportXRechnung(c *gin.Context) {
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

	// Fetch the invoice with all required relations
	invoice, err := h.service.GetInvoiceByID(uint(id), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Invoice not found", err.Error()))
		return
	}

	// Validate invoice status (must be sent, paid, or overdue)
	if invoice.Status != "sent" &&
		invoice.Status != "paid" &&
		invoice.Status != "overdue" {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid invoice status",
			"XRechnung can only be exported for finalized invoices (status: sent, paid, or overdue)"))
		return
	}

	// Fetch cost provider ID from client invoices
	if len(invoice.ClientInvoices) == 0 {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Missing client relationship", "Invoice has no client association"))
		return
	}
	costProviderID := invoice.ClientInvoices[0].CostProviderID

	// Fetch the cost provider (customer)
	costProvider, err := h.service.GetCostProviderByID(costProviderID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to fetch customer", err.Error()))
		return
	}

	// Validate government customer
	if !costProvider.IsGovernmentCustomer {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Not a government customer",
			"XRechnung is only available for government customers"))
		return
	}

	if costProvider.LeitwegID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Missing Leitweg-ID",
			"Government customer must have a Leitweg-ID for XRechnung export"))
		return
	}

	// Fetch the organization
	organization, err := h.service.GetOrganizationByID(invoice.OrganizationID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to fetch organization", err.Error()))
		return
	}

	// Generate XRechnung XML
	xmlData, err := h.xrechnungService.GenerateXRechnungXML(invoice, organization, costProvider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to generate XRechnung XML", err.Error()))
		return
	}

	// Log audit event
	h.logAudit(c, auditEntities.AuditActionXRechnungExported, uint(id), &auditEntities.AuditLogMetadata{
		AdditionalInfo: map[string]interface{}{
			"invoice_number": invoice.InvoiceNumber,
			"leitweg_id":     costProvider.LeitwegID,
			"customer_name":  costProvider.Organization,
		},
	})

	// Set headers for XML file download
	filename := "xrechnung_" + invoice.InvoiceNumber + ".xml"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", xmlData)
}

// GetVATCategories handles retrieving available VAT categories
// @Summary Get available VAT categories
// @Description Get list of available VAT categories with rates and exemption information
// @Tags client-invoices
// @ID getVATCategories
// @Produce json
// @Success 200 {array} map[string]interface{} "List of VAT categories with code, description, rate, is_exempt fields"
// @Failure 401 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/vat-categories [get]
func (h *InvoiceHandler) GetVATCategories(c *gin.Context) {
	vatService := clientServices.NewVATService()
	categories := vatService.GetVATCategories()

	c.JSON(http.StatusOK, models.SuccessResponse("VAT categories retrieved successfully", categories))
}

// DownloadInvoicePDF downloads the PDF for a finalized invoice
// @Summary Download invoice PDF
// @Description Download the PDF document for a finalized invoice
// @Tags client-invoices
// @ID downloadInvoicePDF
// @Param id path int true "Invoice ID"
// @Produce application/pdf
// @Success 200 {file} binary "PDF file"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 422 {object} models.ErrorResponse "Invoice not finalized"
// @Security BearerAuth
// @Router /client-invoices/{id}/pdf [get]
func (h *InvoiceHandler) DownloadInvoicePDF(c *gin.Context) {
	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid invoice ID"})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Tenant ID not found"})
		return
	}

	// Get invoice to check if it exists and is finalized
	userID, _ := baseAPI.GetUserID(c)
	invoice, err := h.service.GetInvoiceByID(uint(invoiceID), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Invoice not found"})
		return
	}

	if invoice.Status == "draft" {
		c.JSON(http.StatusUnprocessableEntity, models.ErrorResponse{Error: "Cannot download PDF for draft invoice. Please finalize first."})
		return
	}

	// Generate and return PDF
	pdfBytes, err := h.service.GenerateInvoicePDF(uint(invoiceID), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate PDF: " + err.Error()})
		return
	}

	// Set headers for PDF download
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=invoice-"+invoice.InvoiceNumber+".pdf")
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

	c.Data(http.StatusOK, "application/pdf", pdfBytes)

	// Log audit event
	h.logAudit(c, auditEntities.AuditActionInvoiceSent, uint(invoiceID), &auditEntities.AuditLogMetadata{
		InvoiceNumber: invoice.InvoiceNumber,
		Reason:        "Downloaded PDF for invoice " + invoice.InvoiceNumber,
	})
}

// PreviewInvoicePDF generates and returns PDF for preview (inline display)
// @Summary Preview invoice PDF
// @Description Generate and display PDF for preview without download
// @Tags client-invoices
// @ID previewInvoicePDF
// @Param id path int true "Invoice ID"
// @Produce application/pdf
// @Success 200 {file} binary "PDF file for preview"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id}/preview-pdf [get]
func (h *InvoiceHandler) PreviewInvoicePDF(c *gin.Context) {
	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid invoice ID"})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Tenant ID not found"})
		return
	}

	// Get invoice to check if it exists
	userID, _ := baseAPI.GetUserID(c)
	invoice, err := h.service.GetInvoiceByID(uint(invoiceID), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Invoice not found"})
		return
	}

	// Generate PDF (works for both draft and finalized invoices)
	pdfBytes, err := h.service.GenerateInvoicePDF(uint(invoiceID), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate PDF: " + err.Error()})
		return
	}

	// Set headers for inline PDF display
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline; filename=invoice-"+invoice.InvoiceNumber+"-preview.pdf")
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

	c.Data(http.StatusOK, "application/pdf", pdfBytes)

	// Log audit event
	h.logAudit(c, auditEntities.AuditActionInvoiceSent, uint(invoiceID), &auditEntities.AuditLogMetadata{
		InvoiceNumber: invoice.InvoiceNumber,
		Reason:        "Previewed PDF for invoice " + invoice.InvoiceNumber,
	})
}
