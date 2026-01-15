package entities

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Invoice represents a generic invoice document
type Invoice struct {
	ID             uint          `gorm:"primaryKey" json:"id"`
	TenantID       uint          `gorm:"not null;index:idx_invoice_tenant" json:"tenant_id"`
	OrganizationID uint          `gorm:"not null;index:idx_invoice_organization" json:"organization_id"`
	UserID         uint          `gorm:"not null;index:idx_invoice_user" json:"user_id"`
	InvoiceNumber  string        `gorm:"size:100;not null;uniqueIndex:idx_invoice_number_tenant" json:"invoice_number"`
	InvoiceDate    time.Time     `gorm:"not null" json:"invoice_date"`
	DueDate        *time.Time    `json:"due_date,omitempty"`
	Status         InvoiceStatus `gorm:"size:20;not null;default:'draft'" json:"status"`

	// Customer information
	CustomerName    string `gorm:"size:255;not null;default:''" json:"customer_name"`
	CustomerAddress string `gorm:"size:500" json:"customer_address,omitempty"`
	CustomerEmail   string `gorm:"size:255" json:"customer_email,omitempty"`
	CustomerTaxID   string `gorm:"size:100" json:"customer_tax_id,omitempty"`

	// Financial data
	SubtotalAmount float64 `gorm:"type:decimal(10,2);not null;default:0" json:"subtotal_amount"`
	TaxRate        float64 `gorm:"type:decimal(5,2);not null;default:0" json:"tax_rate"`
	TaxAmount      float64 `gorm:"type:decimal(10,2);not null;default:0" json:"tax_amount"`
	TotalAmount    float64 `gorm:"type:decimal(10,2);not null;default:0" json:"total_amount"`
	Currency       string  `gorm:"size:3;not null;default:'EUR'" json:"currency"`

	// Payment information
	PaymentTerms  string     `gorm:"size:500" json:"payment_terms,omitempty"`
	PaymentMethod string     `gorm:"size:50" json:"payment_method,omitempty"`
	PaymentDate   *time.Time `json:"payment_date,omitempty"`

	// Document references
	DocumentID *uint `gorm:"index" json:"document_id,omitempty"` // Link to generated PDF document

	// Workflow timestamps
	EmailSentAt    *time.Time `json:"email_sent_at,omitempty"`
	ReminderSentAt *time.Time `json:"reminder_sent_at,omitempty"`
	FinalizedAt    *time.Time `json:"finalized_at,omitempty"`
	CancelledAt    *time.Time `json:"cancelled_at,omitempty"`

	// Credit note support
	IsCreditNote          bool  `gorm:"not null;default:false" json:"is_credit_note"`
	CreditNoteReferenceID *uint `gorm:"index" json:"credit_note_reference_id,omitempty"` // References original invoice if this is a credit note

	// Additional data
	Notes        string         `gorm:"type:text" json:"notes,omitempty"`
	InternalNote string         `gorm:"type:text" json:"internal_note,omitempty"`
	Metadata     datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`

	// Line items
	Items []InvoiceItem `gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE" json:"items,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Invoice) TableName() string {
	return "invoices"
}

// InvoiceItem represents a line item on an invoice
type InvoiceItem struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	InvoiceID   uint    `gorm:"not null;index:idx_invoice_item_invoice" json:"invoice_id"`
	Position    int     `gorm:"not null;default:0" json:"position"`
	Description string  `gorm:"size:500;not null;default:''" json:"description"`
	Quantity    float64 `gorm:"type:decimal(10,3);not null;default:1" json:"quantity"`
	UnitPrice   float64 `gorm:"type:decimal(10,2);not null;default:0" json:"unit_price"`
	TaxRate     float64 `gorm:"type:decimal(5,2);not null;default:0" json:"tax_rate"`
	Amount      float64 `gorm:"type:decimal(10,2);not null;default:0" json:"amount"`

	// VAT handling
	VATRate          float64 `gorm:"type:decimal(5,2);not null;default:0" json:"vat_rate"`
	VATExempt        bool    `gorm:"not null;default:false" json:"vat_exempt"`
	VATExemptionText string  `gorm:"size:500" json:"vat_exemption_text,omitempty"`

	// Source references
	SessionID *uint `gorm:"index" json:"session_id,omitempty"` // Link to session if applicable

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (InvoiceItem) TableName() string {
	return "invoice_items"
}

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusFinalized InvoiceStatus = "finalized"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

// CreateInvoiceRequest represents a request to create an invoice
type CreateInvoiceRequest struct {
	OrganizationID  uint              `json:"organization_id" binding:"required"`
	InvoiceNumber   string            `json:"invoice_number" binding:"required"`
	InvoiceDate     time.Time         `json:"invoice_date" binding:"required"`
	DueDate         *time.Time        `json:"due_date,omitempty"`
	CustomerName    string            `json:"customer_name" binding:"required"`
	CustomerAddress string            `json:"customer_address,omitempty"`
	CustomerEmail   string            `json:"customer_email,omitempty"`
	CustomerTaxID   string            `json:"customer_tax_id,omitempty"`
	TaxRate         float64           `json:"tax_rate"`
	Currency        string            `json:"currency"`
	PaymentTerms    string            `json:"payment_terms,omitempty"`
	PaymentMethod   string            `json:"payment_method,omitempty"`
	Notes           string            `json:"notes,omitempty"`
	InternalNote    string            `json:"internal_note,omitempty"`
	TemplateHTML    string            `json:"template_html,omitempty"` // HTML template for PDF generation
	Items           []InvoiceItemData `json:"items" binding:"required,min=1"`
}

// InvoiceItemData represents invoice item data in requests
type InvoiceItemData struct {
	Description string  `json:"description" binding:"required"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" binding:"required"`
	TaxRate     float64 `json:"tax_rate"`
}

// UpdateInvoiceRequest represents a request to update an invoice
type UpdateInvoiceRequest struct {
	Status          *InvoiceStatus     `json:"status,omitempty"`
	DueDate         *time.Time         `json:"due_date,omitempty"`
	CustomerName    *string            `json:"customer_name,omitempty"`
	CustomerAddress *string            `json:"customer_address,omitempty"`
	CustomerEmail   *string            `json:"customer_email,omitempty"`
	PaymentTerms    *string            `json:"payment_terms,omitempty"`
	PaymentMethod   *string            `json:"payment_method,omitempty"`
	PaymentDate     *time.Time         `json:"payment_date,omitempty"`
	Notes           *string            `json:"notes,omitempty"`
	InternalNote    *string            `json:"internal_note,omitempty"`
	Items           *[]InvoiceItemData `json:"items,omitempty"`
}

// InvoiceResponse represents the response format for invoice data
type InvoiceResponse struct {
	ID              uint                  `json:"id"`
	TenantID        uint                  `json:"tenant_id"`
	OrganizationID  uint                  `json:"organization_id"`
	UserID          uint                  `json:"user_id"`
	InvoiceNumber   string                `json:"invoice_number"`
	InvoiceDate     time.Time             `json:"invoice_date"`
	DueDate         *time.Time            `json:"due_date,omitempty"`
	Status          InvoiceStatus         `json:"status"`
	CustomerName    string                `json:"customer_name"`
	CustomerAddress string                `json:"customer_address,omitempty"`
	CustomerEmail   string                `json:"customer_email,omitempty"`
	CustomerTaxID   string                `json:"customer_tax_id,omitempty"`
	SubtotalAmount  float64               `json:"subtotal_amount"`
	TaxRate         float64               `json:"tax_rate"`
	TaxAmount       float64               `json:"tax_amount"`
	TotalAmount     float64               `json:"total_amount"`
	Currency        string                `json:"currency"`
	PaymentTerms    string                `json:"payment_terms,omitempty"`
	PaymentMethod   string                `json:"payment_method,omitempty"`
	PaymentDate     *time.Time            `json:"payment_date,omitempty"`
	DocumentID      *uint                 `json:"document_id,omitempty"`
	Notes           string                `json:"notes,omitempty"`
	Items           []InvoiceItemResponse `json:"items,omitempty"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

// InvoiceItemResponse represents the response format for invoice item data
type InvoiceItemResponse struct {
	ID          uint    `json:"id"`
	Position    int     `json:"position"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TaxRate     float64 `json:"tax_rate"`
	Amount      float64 `json:"amount"`
}

// ToResponse converts Invoice to InvoiceResponse
func (i *Invoice) ToResponse() InvoiceResponse {
	resp := InvoiceResponse{
		ID:              i.ID,
		TenantID:        i.TenantID,
		OrganizationID:  i.OrganizationID,
		UserID:          i.UserID,
		InvoiceNumber:   i.InvoiceNumber,
		InvoiceDate:     i.InvoiceDate,
		DueDate:         i.DueDate,
		Status:          i.Status,
		CustomerName:    i.CustomerName,
		CustomerAddress: i.CustomerAddress,
		CustomerEmail:   i.CustomerEmail,
		CustomerTaxID:   i.CustomerTaxID,
		SubtotalAmount:  i.SubtotalAmount,
		TaxRate:         i.TaxRate,
		TaxAmount:       i.TaxAmount,
		TotalAmount:     i.TotalAmount,
		Currency:        i.Currency,
		PaymentTerms:    i.PaymentTerms,
		PaymentMethod:   i.PaymentMethod,
		PaymentDate:     i.PaymentDate,
		DocumentID:      i.DocumentID,
		Notes:           i.Notes,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
	}

	if len(i.Items) > 0 {
		resp.Items = make([]InvoiceItemResponse, len(i.Items))
		for idx, item := range i.Items {
			resp.Items[idx] = InvoiceItemResponse{
				ID:          item.ID,
				Position:    item.Position,
				Description: item.Description,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				TaxRate:     item.TaxRate,
				Amount:      item.Amount,
			}
		}
	}

	return resp
}
