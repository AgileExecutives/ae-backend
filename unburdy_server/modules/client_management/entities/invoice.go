package entities

import (
	"time"

	baseAPI "github.com/ae-base-server/api"
	"gorm.io/gorm"
)

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

// Invoice represents an invoice entity
type Invoice struct {
	ID             uint                  `gorm:"primarykey" json:"id"`
	TenantID       uint                  `gorm:"not null;index:idx_invoice_tenant" json:"tenant_id"`
	UserID         uint                  `gorm:"not null;index:idx_invoice_user" json:"user_id"`
	OrganizationID uint                  `gorm:"not null;index:idx_invoice_organization" json:"organization_id"`
	Organization   *baseAPI.Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	InvoiceDate    time.Time             `gorm:"not null" json:"invoice_date"`
	InvoiceNumber  string                `gorm:"size:100;not null;uniqueIndex:idx_invoice_number_tenant" json:"invoice_number"`
	NumberUnits    int                   `gorm:"not null;default:0" json:"number_units"`
	SumAmount      float64               `gorm:"column:subtotal_amount;type:decimal(10,2);not null;default:0" json:"sum_amount"`
	TaxAmount      float64               `gorm:"type:decimal(10,2);not null;default:0" json:"tax_amount"`
	TotalAmount    float64               `gorm:"type:decimal(10,2);not null;default:0" json:"total_amount"`
	PayedDate      *time.Time            `gorm:"column:payment_date" json:"payed_date,omitempty"`
	Status         InvoiceStatus         `gorm:"size:20;not null;default:'draft'" json:"status"`
	NumReminders   int                   `gorm:"column:num_reminders;not null;default:0" json:"num_reminders"`
	LatestReminder *time.Time            `gorm:"column:latest_reminder" json:"latest_reminder,omitempty"`
	DocumentID     *uint                 `gorm:"index:idx_invoice_document" json:"document_id,omitempty"`

	// Workflow timestamps
	EmailSentAt    *time.Time `json:"email_sent_at,omitempty"`
	ReminderSentAt *time.Time `json:"reminder_sent_at,omitempty"`
	FinalizedAt    *time.Time `json:"finalized_at,omitempty"`
	CancelledAt    *time.Time `json:"cancelled_at,omitempty"`

	// Credit note support
	IsCreditNote          bool  `gorm:"not null;default:false" json:"is_credit_note"`
	CreditNoteReferenceID *uint `gorm:"index" json:"credit_note_reference_id,omitempty"`

	InvoiceItems   []InvoiceItem   `gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE" json:"invoice_items,omitempty"`
	ClientInvoices []ClientInvoice `gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE" json:"client_invoices,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName specifies the table name for the Invoice model
func (Invoice) TableName() string {
	return "invoices"
}

// InvoiceItem represents an invoice item linked to a session or extra effort
type InvoiceItem struct {
	ID              uint    `gorm:"primarykey" json:"id"`
	InvoiceID       uint    `gorm:"not null;index:idx_invoice_item_invoice" json:"invoice_id"`
	ItemType        string  `gorm:"size:50;default:'session'" json:"item_type"` // session, extra_effort, custom
	SourceEffortID  *uint   `gorm:"index:idx_invoice_item_effort" json:"source_effort_id,omitempty"`
	SessionID       *uint   `gorm:"index:idx_invoice_item_session" json:"session_id,omitempty"`
	Description     string  `gorm:"type:text" json:"description"`
	NumberUnits     float64 `gorm:"type:decimal(10,2);not null;default:0" json:"number_units"`
	UnitPrice       float64 `gorm:"type:decimal(10,2);not null;default:0" json:"unit_price"`
	TotalAmount     float64 `gorm:"type:decimal(10,2);not null;default:0" json:"total_amount"`
	UnitDurationMin *int    `json:"unit_duration_min,omitempty"` // Duration for extra efforts
	IsEditable      bool    `gorm:"default:true" json:"is_editable"`

	// VAT handling
	VATRate          float64 `gorm:"type:decimal(5,2);not null;default:0" json:"vat_rate"`
	VATExempt        bool    `gorm:"not null;default:false" json:"vat_exempt"`
	VATExemptionText string  `gorm:"size:500" json:"vat_exemption_text,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for the InvoiceItem model
func (InvoiceItem) TableName() string {
	return "invoice_items"
}

// ClientInvoice represents the junction table linking invoices to clients and their sessions
type ClientInvoice struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	InvoiceID      uint           `gorm:"not null;index:idx_client_invoice_invoice" json:"invoice_id"`
	Invoice        *Invoice       `gorm:"foreignKey:InvoiceID" json:"invoice,omitempty"`
	ClientID       uint           `gorm:"not null;index:idx_client_invoice_client" json:"client_id"`
	Client         *Client        `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	CostProviderID uint           `gorm:"not null;index:idx_client_invoice_cost_provider" json:"cost_provider_id"`
	CostProvider   *CostProvider  `gorm:"foreignKey:CostProviderID" json:"cost_provider,omitempty"`
	SessionID      uint           `gorm:"not null;uniqueIndex:idx_client_invoice_session" json:"session_id"`
	Session        *Session       `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	InvoiceItemID  uint           `gorm:"not null;index:idx_client_invoice_item" json:"invoice_item_id"`
	InvoiceItem    *InvoiceItem   `gorm:"foreignKey:InvoiceItemID" json:"invoice_item,omitempty"`
	ExtraEffortID  *uint          `gorm:"index:idx_client_invoice_extra_effort" json:"extra_effort_id,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for the ClientInvoice model
func (ClientInvoice) TableName() string {
	return "client_invoices"
}

// CreateInvoiceRequest represents the request payload for creating an invoice from sessions
type CreateInvoiceRequest struct {
	ClientID   uint   `json:"client_id" binding:"required" example:"1"`
	SessionIDs []uint `json:"session_ids" binding:"required" example:"1,2,3"`
}

// InvoiceGenerationParameters contains the control parameters for invoice generation
type InvoiceGenerationParameters struct {
	InvoiceNumber   string  `json:"invoice_number,omitempty"`
	InvoiceDate     string  `json:"invoice_date,omitempty"`
	TaxRate         float64 `json:"tax_rate,omitempty"`
	GeneratePDF     *bool   `json:"generate_pdf,omitempty"` // Default true in handler
	TemplateID      *uint   `json:"template_id,omitempty"`
	SessionFromDate *string `json:"session_from_date,omitempty"` // Filter sessions by start date (YYYY-MM-DD)
	SessionToDate   *string `json:"session_to_date,omitempty"`   // Filter sessions by end date (YYYY-MM-DD)
}

// CreateInvoiceDirectRequest represents the request payload for creating a complete invoice directly
// The structure matches the output from the unbilled-sessions endpoint plus generation parameters
type CreateInvoiceDirectRequest struct {
	// Unbilled client data (matching ClientWithUnbilledSessionsResponse)
	UnbilledClient ClientWithUnbilledSessionsResponse `json:"unbilledClient" binding:"required"`

	// Invoice generation parameters
	Parameters InvoiceGenerationParameters `json:"parameters"`
}

// UpdateInvoiceRequest represents the request payload for updating an invoice
type UpdateInvoiceRequest struct {
	Status     *InvoiceStatus `json:"status,omitempty" example:"sent"`
	SessionIDs []uint         `json:"session_ids,omitempty" example:"1,2,3"`
}

// CreateDraftInvoiceRequest represents the request for creating a draft invoice
type CreateDraftInvoiceRequest struct {
	ClientID        uint                    `json:"client_id" binding:"required" example:"1"`
	SessionIDs      []uint                  `json:"session_ids,omitempty" example:"1,2,3"`
	ExtraEffortIDs  []uint                  `json:"extra_effort_ids,omitempty" example:"5,6"`
	CustomLineItems []CustomLineItemRequest `json:"custom_line_items,omitempty"`
}

// CustomLineItemRequest represents a custom line item for an invoice
type CustomLineItemRequest struct {
	Description      string  `json:"description" binding:"required" example:"Additional consultation"`
	NumberUnits      float64 `json:"number_units" example:"1"`
	UnitPrice        float64 `json:"unit_price" example:"150.00"`
	VATCategory      string  `json:"vat_category,omitempty" example:"exempt_heilberuf"` // exempt_heilberuf, taxable_standard, taxable_reduced
	VATRate          float64 `json:"vat_rate,omitempty" example:"19.00"`
	VATExempt        bool    `json:"vat_exempt,omitempty" example:"false"`
	VATExemptionText string  `json:"vat_exemption_text,omitempty"`
}

// UpdateDraftInvoiceRequest represents the request for updating a draft invoice
type UpdateDraftInvoiceRequest struct {
	AddSessionIDs        []uint                  `json:"add_session_ids,omitempty" example:"7,8"`
	RemoveSessionIDs     []uint                  `json:"remove_session_ids,omitempty" example:"1"`
	AddExtraEffortIDs    []uint                  `json:"add_extra_effort_ids,omitempty" example:"9"`
	RemoveExtraEffortIDs []uint                  `json:"remove_extra_effort_ids,omitempty" example:"5"`
	CustomLineItems      []CustomLineItemRequest `json:"custom_line_items,omitempty"`
}

// MarkInvoiceAsPaidRequest represents the request for marking an invoice as paid
type MarkInvoiceAsPaidRequest struct {
	PaymentDate      *time.Time `json:"payment_date,omitempty" example:"2026-01-08T00:00:00Z"`
	PaymentReference string     `json:"payment_reference,omitempty" example:"TRANSFER-123456"`
}

// CreateCreditNoteRequest represents the request for creating a credit note
type CreateCreditNoteRequest struct {
	LineItemIDs []uint     `json:"line_item_ids" binding:"required" example:"1,2,3"`
	Reason      string     `json:"reason" binding:"required" example:"Customer dissatisfaction - partial refund"`
	CreditDate  *time.Time `json:"credit_date,omitempty" example:"2026-01-08T00:00:00Z"`
}

// InvoiceResponse represents the response format for invoice data
type InvoiceResponse struct {
	ID             uint                          `json:"id"`
	TenantID       uint                          `json:"tenant_id"`
	UserID         uint                          `json:"user_id"`
	OrganizationID uint                          `json:"organization_id"`
	Organization   *baseAPI.OrganizationResponse `json:"organization,omitempty"`
	InvoiceDate    time.Time                     `json:"invoice_date"`
	InvoiceNumber  string                        `json:"invoice_number"`
	NumberUnits    int                           `json:"number_units"`
	SumAmount      float64                       `json:"sum_amount"`
	TaxAmount      float64                       `json:"tax_amount"`
	TotalAmount    float64                       `json:"total_amount"`
	PayedDate      *time.Time                    `json:"payed_date,omitempty"`
	Status         InvoiceStatus                 `json:"status"`
	NumReminders   int                           `json:"num_reminders"`
	LatestReminder *time.Time                    `json:"latest_reminder,omitempty"`
	DocumentID     *uint                         `json:"document_id,omitempty"`
	DocumentURL    string                        `json:"document_url,omitempty"`
	InvoiceItems   []InvoiceItemResponse         `json:"invoice_items,omitempty"`
	Clients        []ClientInvoiceResponse       `json:"clients,omitempty"`
	VATBreakdown   *VATBreakdownResponse         `json:"vat_breakdown,omitempty"`
	CreatedAt      time.Time                     `json:"created_at"`
	UpdatedAt      time.Time                     `json:"updated_at"`
}

// VATBreakdownResponse represents the VAT breakdown for an invoice
type VATBreakdownResponse struct {
	Subtotal   float64                    `json:"subtotal"`
	TotalTax   float64                    `json:"total_tax"`
	GrandTotal float64                    `json:"grand_total"`
	Items      []VATBreakdownItemResponse `json:"items"`
}

// VATBreakdownItemResponse represents a single VAT rate breakdown
type VATBreakdownItemResponse struct {
	VATRate       float64 `json:"vat_rate"`
	NetAmount     float64 `json:"net_amount"`
	TaxAmount     float64 `json:"tax_amount"`
	GrossAmount   float64 `json:"gross_amount"`
	ExemptionText string  `json:"exemption_text,omitempty"`
}

// ClientInvoiceResponse represents a client and their sessions within an invoice
type ClientInvoiceResponse struct {
	ClientID       uint                  `json:"client_id"`
	Client         *ClientResponse       `json:"client,omitempty"`
	CostProviderID uint                  `json:"cost_provider_id"`
	CostProvider   *CostProviderResponse `json:"cost_provider,omitempty"`
	Sessions       []SessionResponse     `json:"sessions"`
}

// InvoiceItemResponse represents the response format for invoice item data
type InvoiceItemResponse struct {
	ID              uint      `json:"id"`
	InvoiceID       uint      `json:"invoice_id"`
	ItemType        string    `json:"item_type"`
	SourceEffortID  *uint     `json:"source_effort_id,omitempty"`
	Description     string    `json:"description"`
	NumberUnits     float64   `json:"number_units"`
	UnitPrice       float64   `json:"unit_price"`
	TotalAmount     float64   `json:"total_amount"`
	UnitDurationMin *int      `json:"unit_duration_min,omitempty"`
	IsEditable      bool      `json:"is_editable"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ClientWithUnbilledSessionsResponse represents a client with sessions not yet invoiced
type ClientWithUnbilledSessionsResponse struct {
	ClientResponse
	Sessions     []SessionResponse     `json:"sessions"`
	ExtraEfforts []ExtraEffortResponse `json:"extra_efforts"`
}

// InvoiceAPIResponse represents the API response for a single invoice
type InvoiceAPIResponse struct {
	Success bool            `json:"success" example:"true"`
	Message string          `json:"message" example:"Invoice retrieved successfully"`
	Data    InvoiceResponse `json:"data"`
}

// InvoiceListAPIResponse represents the API response for invoice list
type InvoiceListAPIResponse struct {
	Success bool              `json:"success" example:"true"`
	Message string            `json:"message" example:"Invoices retrieved successfully"`
	Data    []InvoiceResponse `json:"data"`
	Page    int               `json:"page" example:"1"`
	Limit   int               `json:"limit" example:"10"`
	Total   int               `json:"total" example:"100"`
}

// ClientSessionsAPIResponse represents the API response for clients with unbilled sessions
type ClientSessionsAPIResponse struct {
	Success bool                                 `json:"success" example:"true"`
	Message string                               `json:"message" example:"Clients with unbilled sessions retrieved successfully"`
	Data    []ClientWithUnbilledSessionsResponse `json:"data"`
}

// InvoiceDeleteResponse represents the API response for invoice deletion
type InvoiceDeleteResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Invoice deleted successfully"`
}

// ToResponse converts an Invoice to InvoiceResponse
func (i *Invoice) ToResponse() InvoiceResponse {
	response := InvoiceResponse{
		ID:             i.ID,
		TenantID:       i.TenantID,
		UserID:         i.UserID,
		OrganizationID: i.OrganizationID,
		InvoiceDate:    i.InvoiceDate,
		InvoiceNumber:  i.InvoiceNumber,
		NumberUnits:    i.NumberUnits,
		SumAmount:      i.SumAmount,
		TaxAmount:      i.TaxAmount,
		TotalAmount:    i.TotalAmount,
		PayedDate:      i.PayedDate,
		Status:         i.Status,
		NumReminders:   i.NumReminders,
		LatestReminder: i.LatestReminder,
		DocumentID:     i.DocumentID,
		CreatedAt:      i.CreatedAt,
		UpdatedAt:      i.UpdatedAt,
	}

	// DocumentURL will be populated by handler using document service

	if i.Organization != nil {
		orgResp := i.Organization.ToResponse()
		response.Organization = &orgResp
	}

	// Populate invoice items
	if len(i.InvoiceItems) > 0 {
		response.InvoiceItems = make([]InvoiceItemResponse, len(i.InvoiceItems))
		for idx, item := range i.InvoiceItems {
			response.InvoiceItems[idx] = item.ToResponse()
		}
	}

	// Group client invoices by client
	if len(i.ClientInvoices) > 0 {
		clientMap := make(map[uint]*ClientInvoiceResponse)

		for _, ci := range i.ClientInvoices {
			if _, exists := clientMap[ci.ClientID]; !exists {
				clientInvResp := &ClientInvoiceResponse{
					ClientID:       ci.ClientID,
					CostProviderID: ci.CostProviderID,
					Sessions:       []SessionResponse{},
				}

				if ci.Client != nil {
					clientResp := ci.Client.ToResponse()
					clientInvResp.Client = &clientResp
				}

				if ci.CostProvider != nil {
					costProviderResp := ci.CostProvider.ToResponse()
					clientInvResp.CostProvider = &costProviderResp
				}

				clientMap[ci.ClientID] = clientInvResp
			}

			// Add session to client
			if ci.Session != nil {
				sessionResp := ci.Session.ToResponse()
				clientMap[ci.ClientID].Sessions = append(clientMap[ci.ClientID].Sessions, sessionResp)
			}
		}

		// Convert map to slice
		response.Clients = make([]ClientInvoiceResponse, 0, len(clientMap))
		for _, clientInv := range clientMap {
			response.Clients = append(response.Clients, *clientInv)
		}
	}

	return response
}

// ToResponseWithVATBreakdown converts an Invoice to InvoiceResponse with VAT breakdown
func (i *Invoice) ToResponseWithVATBreakdown(vatBreakdown *VATBreakdownResponse) InvoiceResponse {
	response := i.ToResponse()
	response.VATBreakdown = vatBreakdown
	return response
}

// ToResponse converts an InvoiceItem to InvoiceItemResponse
func (ii *InvoiceItem) ToResponse() InvoiceItemResponse {
	return InvoiceItemResponse{
		ID:              ii.ID,
		InvoiceID:       ii.InvoiceID,
		ItemType:        ii.ItemType,
		SourceEffortID:  ii.SourceEffortID,
		Description:     ii.Description,
		NumberUnits:     ii.NumberUnits,
		UnitPrice:       ii.UnitPrice,
		TotalAmount:     ii.TotalAmount,
		UnitDurationMin: ii.UnitDurationMin,
		IsEditable:      ii.IsEditable,
		CreatedAt:       ii.CreatedAt,
		UpdatedAt:       ii.UpdatedAt,
	}
}
