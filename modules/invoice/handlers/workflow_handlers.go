package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// FinalizeInvoice finalizes a draft invoice
// @Summary Finalize invoice
// @Description Finalize a draft invoice by generating invoice number and changing status
// @Tags Invoices
// @Produce json
// @Param id path int true "Invoice ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /invoices/{id}/finalize [post]
// @Security BearerAuth
func (h *InvoiceHandler) FinalizeInvoice(c *gin.Context) {
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

	invoiceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice_id"})
		return
	}

	invoice, err := h.service.FinalizeInvoice(c.Request.Context(), tenantID, uint(invoiceID), userID)
	if err != nil {
		// Check for specific error types
		if err.Error() == "can only finalize invoices in draft status" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "invoice must have at least one line item" {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Invoice finalized successfully",
		"data":    invoice.ToResponse(),
	})
}

// MarkInvoiceAsSent marks a finalized invoice as sent
// @Summary Mark invoice as sent
// @Description Mark a finalized invoice as sent (e.g., after emailing to customer)
// @Tags Invoices
// @Produce json
// @Param id path int true "Invoice ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /invoices/{id}/send [post]
// @Security BearerAuth
func (h *InvoiceHandler) MarkInvoiceAsSent(c *gin.Context) {
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

	invoice, err := h.service.MarkAsSent(c.Request.Context(), tenantID, uint(invoiceID))
	if err != nil {
		if err.Error() == "can only mark finalized invoices as sent" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Invoice marked as sent",
		"data":    invoice.ToResponse(),
	})
}

// MarkInvoiceAsPaid marks an invoice as paid
// @Summary Mark invoice as paid
// @Description Record payment for an invoice
// @Tags Invoices
// @Accept json
// @Produce json
// @Param id path int true "Invoice ID"
// @Param payment body object{payment_date=string,payment_method=string} true "Payment details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /invoices/{id}/pay [post]
// @Security BearerAuth
func (h *InvoiceHandler) MarkInvoiceAsPaid(c *gin.Context) {
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
		PaymentDate   string `json:"payment_date"`
		PaymentMethod string `json:"payment_method"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse payment date or use current time
	var paymentDate time.Time
	if req.PaymentDate != "" {
		paymentDate, err = time.Parse(time.RFC3339, req.PaymentDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment_date format, use RFC3339"})
			return
		}
	} else {
		paymentDate = time.Now()
	}

	invoice, err := h.service.MarkAsPaidWithAmount(c.Request.Context(), tenantID, uint(invoiceID), paymentDate, req.PaymentMethod)
	if err != nil {
		if err.Error() == "can only mark sent or overdue invoices as paid" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Invoice marked as paid",
		"data":    invoice.ToResponse(),
	})
}

// SendInvoiceReminder sends a payment reminder
// @Summary Send payment reminder
// @Description Send a payment reminder for an overdue or sent invoice
// @Tags Invoices
// @Produce json
// @Param id path int true "Invoice ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /invoices/{id}/remind [post]
// @Security BearerAuth
func (h *InvoiceHandler) SendInvoiceReminder(c *gin.Context) {
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

	invoice, err := h.service.SendReminder(c.Request.Context(), tenantID, uint(invoiceID))
	if err != nil {
		if err.Error() == "can only send reminders for sent or overdue invoices" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reminder sent successfully",
		"data":    invoice.ToResponse(),
	})
}

// CancelInvoice cancels an invoice
// @Summary Cancel invoice
// @Description Cancel an invoice with optional reason
// @Tags Invoices
// @Accept json
// @Produce json
// @Param id path int true "Invoice ID"
// @Param cancellation body object{reason=string} false "Cancellation reason"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /invoices/{id}/cancel [post]
// @Security BearerAuth
func (h *InvoiceHandler) CancelInvoice(c *gin.Context) {
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
		Reason string `json:"reason"`
	}
	// Reason is optional
	_ = c.ShouldBindJSON(&req)

	invoice, err := h.service.CancelInvoice(c.Request.Context(), tenantID, uint(invoiceID), req.Reason)
	if err != nil {
		if err.Error() == "cannot cancel paid invoices" || err.Error() == "invoice is already cancelled" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Invoice cancelled successfully",
		"data":    invoice.ToResponse(),
	})
}

// GenerateInvoicePDF generates a PDF for an invoice
// @Summary Generate invoice PDF
// @Description Generate and store a PDF document for an invoice
// @Tags Invoices
// @Accept json
// @Produce json
// @Param id path int true "Invoice ID"
// @Param template body object{template_id=int} false "Optional template ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /invoices/{id}/generate-pdf [post]
// @Security BearerAuth
func (h *InvoiceHandler) GenerateInvoicePDF(c *gin.Context) {
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
		TemplateID *uint `json:"template_id"`
	}
	_ = c.ShouldBindJSON(&req)

	invoice, err := h.service.GenerateInvoicePDF(c.Request.Context(), tenantID, uint(invoiceID), req.TemplateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "PDF generated successfully",
		"data":    invoice.ToResponse(),
	})
}
