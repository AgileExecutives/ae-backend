package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"gorm.io/gorm"
)

// InvoiceAdapterHandler adapts session invoicing to the invoice module
type InvoiceAdapterHandler struct {
	db               *gorm.DB
	invoiceModuleURL string // URL of the invoice module endpoint
}

// NewInvoiceAdapterHandler creates a new invoice adapter handler
func NewInvoiceAdapterHandler(db *gorm.DB, invoiceModuleURL string) *InvoiceAdapterHandler {
	return &InvoiceAdapterHandler{
		db:               db,
		invoiceModuleURL: invoiceModuleURL,
	}
}

// CreateInvoiceFromSessionsRequest represents the request to create an invoice from sessions
// This accepts the output from the unbilled-sessions endpoint
type CreateInvoiceFromSessionsRequest struct {
	Client         entities.ClientWithUnbilledSessionsResponse `json:"client" binding:"required"`
	OrganizationID uint                                        `json:"organization_id" binding:"required"`
	InvoiceNumber  string                                      `json:"invoice_number" binding:"required"`
	InvoiceDate    string                                      `json:"invoice_date" binding:"required"` // RFC3339 format
	DueDate        *string                                     `json:"due_date,omitempty"`              // RFC3339 format
	TemplateID     *uint                                       `json:"template_id,omitempty"`
}

// InvoiceModuleCreateRequest represents the request format for the invoice module
type InvoiceModuleCreateRequest struct {
	OrganizationID  uint                    `json:"organization_id"`
	InvoiceNumber   string                  `json:"invoice_number"`
	InvoiceDate     time.Time               `json:"invoice_date"`
	DueDate         *time.Time              `json:"due_date,omitempty"`
	CustomerName    string                  `json:"customer_name"`
	CustomerAddress string                  `json:"customer_address,omitempty"`
	CustomerEmail   string                  `json:"customer_email,omitempty"`
	CustomerTaxID   string                  `json:"customer_tax_id,omitempty"`
	TaxRate         float64                 `json:"tax_rate"`
	Currency        string                  `json:"currency"`
	PaymentTerms    string                  `json:"payment_terms,omitempty"`
	PaymentMethod   string                  `json:"payment_method,omitempty"`
	Notes           string                  `json:"notes,omitempty"`
	InternalNote    string                  `json:"internal_note,omitempty"`
	TemplateID      *uint                   `json:"template_id,omitempty"`
	Items           []InvoiceModuleItemData `json:"items"`
}

// InvoiceModuleItemData represents an invoice item for the invoice module
type InvoiceModuleItemData struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TaxRate     float64 `json:"tax_rate"`
}

// CreateInvoiceFromSessions creates an invoice in the invoice module from sessions
// @Summary Create invoice from sessions
// @Description Adapts session data and creates an invoice in the invoice module
// @Tags client-invoices
// @Accept json
// @Produce json
// @Param invoice body CreateInvoiceFromSessionsRequest true "Invoice from sessions request"
// @Success 201 {object} map[string]interface{} "Invoice created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Client or sessions not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /client-invoices/from-sessions [post]
func (h *InvoiceAdapterHandler) CreateInvoiceFromSessions(c *gin.Context) {
	var req CreateInvoiceFromSessionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant_id from context
	tenantIDValue, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}
	tenantID, ok := tenantIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid tenant_id"})
		return
	}

	// Validate that client has sessions
	if len(req.Client.Sessions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no sessions provided"})
		return
	}

	// Build invoice request for invoice module
	invoiceReq := h.buildInvoiceModuleRequest(req)

	// Parse dates
	invoiceDate, err := time.Parse(time.RFC3339, req.InvoiceDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice_date format, use RFC3339"})
		return
	}
	invoiceReq.InvoiceDate = invoiceDate

	if req.DueDate != nil {
		dueDate, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid due_date format, use RFC3339"})
			return
		}
		invoiceReq.DueDate = &dueDate
	}

	// Forward to invoice module
	if h.invoiceModuleURL != "" {
		response, err := h.forwardToInvoiceModule(invoiceReq, tenantID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create invoice: %v", err)})
			return
		}
		c.JSON(http.StatusCreated, response)
	} else {
		// Return what would be sent (for testing/development)
		c.JSON(http.StatusCreated, gin.H{
			"message": "Invoice module URL not configured, showing request that would be sent",
			"request": invoiceReq,
		})
	}
}

// buildInvoiceModuleRequest builds the invoice module request from sessions and client data
func (h *InvoiceAdapterHandler) buildInvoiceModuleRequest(
	req CreateInvoiceFromSessionsRequest,
) InvoiceModuleCreateRequest {
	// Build customer information from cost provider or client
	customerName := req.Client.FirstName + " " + req.Client.LastName
	customerAddress := ""

	if req.Client.CostProvider != nil {
		customerName = req.Client.CostProvider.Organization
		if req.Client.CostProvider.Department != "" {
			customerName += " - " + req.Client.CostProvider.Department
		}

		// Build address from cost provider
		addressParts := []string{}
		if req.Client.CostProvider.StreetAddress != "" {
			addressParts = append(addressParts, req.Client.CostProvider.StreetAddress)
		}
		if req.Client.CostProvider.Zip != "" || req.Client.CostProvider.City != "" {
			cityLine := req.Client.CostProvider.Zip
			if req.Client.CostProvider.City != "" {
				if cityLine != "" {
					cityLine += " "
				}
				cityLine += req.Client.CostProvider.City
			}
			if cityLine != "" {
				addressParts = append(addressParts, cityLine)
			}
		}
		if len(addressParts) > 0 {
			customerAddress = addressParts[0]
			for i := 1; i < len(addressParts); i++ {
				customerAddress += "\n" + addressParts[i]
			}
		}
	}

	// Build invoice items from sessions
	items := make([]InvoiceModuleItemData, len(req.Client.Sessions))
	for i, session := range req.Client.Sessions {
		description := fmt.Sprintf("%s - %s (%d min)",
			session.Type,
			session.OriginalDate.Format("2006-01-02"),
			session.DurationMin)

		// Determine unit price
		unitPrice := 0.0
		if req.Client.UnitPrice != nil {
			unitPrice = *req.Client.UnitPrice
		}

		items[i] = InvoiceModuleItemData{
			Description: description,
			Quantity:    float64(session.NumberUnits),
			UnitPrice:   unitPrice,
			TaxRate:     0.0, // This should come from configuration or cost provider settings
		}
	}

	return InvoiceModuleCreateRequest{
		OrganizationID:  req.OrganizationID,
		InvoiceNumber:   req.InvoiceNumber,
		CustomerName:    customerName,
		CustomerAddress: customerAddress,
		CustomerEmail:   req.Client.Email,
		TaxRate:         0.0, // Default or from configuration
		Currency:        "EUR",
		TemplateID:      req.TemplateID,
		Items:           items,
	}
}

// forwardToInvoiceModule sends the request to the invoice module
func (h *InvoiceAdapterHandler) forwardToInvoiceModule(req InvoiceModuleCreateRequest, tenantID uint) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", h.invoiceModuleURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	// Pass tenant context
	httpReq.Header.Set("X-Tenant-ID", fmt.Sprintf("%d", tenantID))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invoice module returned status %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}
