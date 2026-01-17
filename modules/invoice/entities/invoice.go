package entities

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Invoice represents a generic invoice document
type Invoice struct {
	ID             uint          `gorm:"primaryKey" json:"id"`
	TenantID       uint          `gorm:"not null;index:idx_invoice_tenant;index:idx_invoice_tenant_status,priority:1" json:"tenant_id"`
	OrganizationID uint          `gorm:"not null;index:idx_invoice_organization;index:idx_invoice_org_status,priority:1" json:"organization_id"`
	UserID         uint          `gorm:"not null;index:idx_invoice_user" json:"user_id"`
	InvoiceNumber  string        `gorm:"size:100;not null;uniqueIndex:idx_invoice_number_tenant" json:"invoice_number"`
	InvoiceDate    time.Time     `gorm:"not null;index:idx_invoice_date" json:"invoice_date"`
	DueDate        *time.Time    `gorm:"index:idx_invoice_due_date" json:"due_date,omitempty"`
	Status         InvoiceStatus `gorm:"size:20;not null;default:'draft';index:idx_invoice_status;index:idx_invoice_tenant_status,priority:2;index:idx_invoice_org_status,priority:2" json:"status"`

	// Customer information
	CustomerName          string `gorm:"size:255;not null;default:'';index:idx_invoice_customer_name,type:gin" json:"customer_name"` // GIN index for text search
	CustomerAddress       string `gorm:"size:500" json:"customer_address,omitempty"`
	CustomerAddressExt    string `gorm:"size:255" json:"customer_address_ext,omitempty"`
	CustomerZip           string `gorm:"size:20" json:"customer_zip,omitempty"`
	CustomerCity          string `gorm:"size:100" json:"customer_city,omitempty"`
	CustomerCountry       string `gorm:"size:100" json:"customer_country,omitempty"`
	CustomerEmail         string `gorm:"size:255" json:"customer_email,omitempty"`
	CustomerTaxID         string `gorm:"size:100" json:"customer_tax_id,omitempty"`
	CustomerContactPerson string `gorm:"size:255" json:"customer_contact_person,omitempty"`
	CustomerDepartment    string `gorm:"size:255" json:"customer_department,omitempty"`

	// Business references
	Subject                string     `gorm:"size:500" json:"subject,omitempty"`
	OurReference           string     `gorm:"size:100" json:"our_reference,omitempty"`
	YourReference          string     `gorm:"size:100" json:"your_reference,omitempty"`
	PONumber               string     `gorm:"size:100" json:"po_number,omitempty"`
	DeliveryDate           *time.Time `json:"delivery_date,omitempty"`
	PerformancePeriodStart *time.Time `json:"performance_period_start,omitempty"`
	PerformancePeriodEnd   *time.Time `json:"performance_period_end,omitempty"`

	// Financial data
	SubtotalAmount float64 `gorm:"type:decimal(10,2);not null;default:0" json:"subtotal_amount"`
	TaxRate        float64 `gorm:"type:decimal(5,2);not null;default:0" json:"tax_rate"`
	TaxAmount      float64 `gorm:"type:decimal(10,2);not null;default:0" json:"tax_amount"`
	TotalAmount    float64 `gorm:"type:decimal(10,2);not null;default:0" json:"total_amount"`
	Currency       string  `gorm:"size:3;not null;default:'EUR'" json:"currency"`

	// Payment information
	PaymentTerms  string     `gorm:"size:500" json:"payment_terms,omitempty"`
	NetTerms      int        `gorm:"default:30" json:"net_terms"` // Payment due in days (e.g., 30 for NET 30)
	PaymentMethod string     `gorm:"size:50" json:"payment_method,omitempty"`
	PaymentDate   *time.Time `json:"payment_date,omitempty"`
	DiscountRate  float64    `gorm:"type:decimal(5,2);default:0" json:"discount_rate,omitempty"` // Early payment discount percentage
	DiscountTerms string     `gorm:"size:100" json:"discount_terms,omitempty"`                   // e.g., "2% within 10 days"

	// Document references
	DocumentID *uint `gorm:"index" json:"document_id,omitempty"` // Link to generated PDF document

	// Workflow tracking
	NumReminders       int        `gorm:"default:0" json:"num_reminders"`
	FinalizedAt        *time.Time `json:"finalized_at,omitempty"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty"`
	CancellationReason string     `gorm:"type:text" json:"cancellation_reason,omitempty"`
	ReminderSentAt     *time.Time `json:"reminder_sent_at,omitempty"`

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
	OrganizationID         uint              `json:"organization_id" binding:"required"`
	InvoiceNumber          string            `json:"invoice_number" binding:"required"`
	InvoiceDate            time.Time         `json:"invoice_date" binding:"required"`
	DueDate                *time.Time        `json:"due_date,omitempty"`
	CustomerName           string            `json:"customer_name" binding:"required"`
	CustomerAddress        string            `json:"customer_address,omitempty"`
	CustomerAddressExt     string            `json:"customer_address_ext,omitempty"`
	CustomerZip            string            `json:"customer_zip,omitempty"`
	CustomerCity           string            `json:"customer_city,omitempty"`
	CustomerCountry        string            `json:"customer_country,omitempty"`
	CustomerEmail          string            `json:"customer_email,omitempty"`
	CustomerTaxID          string            `json:"customer_tax_id,omitempty"`
	CustomerContactPerson  string            `json:"customer_contact_person,omitempty"`
	CustomerDepartment     string            `json:"customer_department,omitempty"`
	Subject                string            `json:"subject,omitempty"`
	OurReference           string            `json:"our_reference,omitempty"`
	YourReference          string            `json:"your_reference,omitempty"`
	PONumber               string            `json:"po_number,omitempty"`
	DeliveryDate           *time.Time        `json:"delivery_date,omitempty"`
	PerformancePeriodStart *time.Time        `json:"performance_period_start,omitempty"`
	PerformancePeriodEnd   *time.Time        `json:"performance_period_end,omitempty"`
	TaxRate                float64           `json:"tax_rate"`
	Currency               string            `json:"currency"`
	PaymentTerms           string            `json:"payment_terms,omitempty"`
	NetTerms               int               `json:"net_terms,omitempty"`
	PaymentMethod          string            `json:"payment_method,omitempty"`
	DiscountRate           float64           `json:"discount_rate,omitempty"`
	DiscountTerms          string            `json:"discount_terms,omitempty"`
	Notes                  string            `json:"notes,omitempty"`
	InternalNote           string            `json:"internal_note,omitempty"`
	TemplateHTML           string            `json:"template_html,omitempty"` // HTML template for PDF generation
	Items                  []InvoiceItemData `json:"items" binding:"required,min=1"`
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
	Status                 *InvoiceStatus     `json:"status,omitempty"`
	DueDate                *time.Time         `json:"due_date,omitempty"`
	CustomerName           *string            `json:"customer_name,omitempty"`
	CustomerAddress        *string            `json:"customer_address,omitempty"`
	CustomerAddressExt     *string            `json:"customer_address_ext,omitempty"`
	CustomerZip            *string            `json:"customer_zip,omitempty"`
	CustomerCity           *string            `json:"customer_city,omitempty"`
	CustomerCountry        *string            `json:"customer_country,omitempty"`
	CustomerEmail          *string            `json:"customer_email,omitempty"`
	CustomerContactPerson  *string            `json:"customer_contact_person,omitempty"`
	CustomerDepartment     *string            `json:"customer_department,omitempty"`
	Subject                *string            `json:"subject,omitempty"`
	OurReference           *string            `json:"our_reference,omitempty"`
	YourReference          *string            `json:"your_reference,omitempty"`
	PONumber               *string            `json:"po_number,omitempty"`
	DeliveryDate           *time.Time         `json:"delivery_date,omitempty"`
	PerformancePeriodStart *time.Time         `json:"performance_period_start,omitempty"`
	PerformancePeriodEnd   *time.Time         `json:"performance_period_end,omitempty"`
	PaymentTerms           *string            `json:"payment_terms,omitempty"`
	NetTerms               *int               `json:"net_terms,omitempty"`
	PaymentMethod          *string            `json:"payment_method,omitempty"`
	PaymentDate            *time.Time         `json:"payment_date,omitempty"`
	DiscountRate           *float64           `json:"discount_rate,omitempty"`
	DiscountTerms          *string            `json:"discount_terms,omitempty"`
	Notes                  *string            `json:"notes,omitempty"`
	InternalNote           *string            `json:"internal_note,omitempty"`
	Items                  *[]InvoiceItemData `json:"items,omitempty"`
}

// InvoiceResponse represents the response format for invoice data
type InvoiceResponse struct {
	ID                     uint                  `json:"id"`
	TenantID               uint                  `json:"tenant_id"`
	OrganizationID         uint                  `json:"organization_id"`
	UserID                 uint                  `json:"user_id"`
	InvoiceNumber          string                `json:"invoice_number"`
	InvoiceDate            time.Time             `json:"invoice_date"`
	DueDate                *time.Time            `json:"due_date,omitempty"`
	Status                 InvoiceStatus         `json:"status"`
	CustomerName           string                `json:"customer_name"`
	CustomerAddress        string                `json:"customer_address,omitempty"`
	CustomerAddressExt     string                `json:"customer_address_ext,omitempty"`
	CustomerZip            string                `json:"customer_zip,omitempty"`
	CustomerCity           string                `json:"customer_city,omitempty"`
	CustomerCountry        string                `json:"customer_country,omitempty"`
	CustomerEmail          string                `json:"customer_email,omitempty"`
	CustomerTaxID          string                `json:"customer_tax_id,omitempty"`
	CustomerContactPerson  string                `json:"customer_contact_person,omitempty"`
	CustomerDepartment     string                `json:"customer_department,omitempty"`
	Subject                string                `json:"subject,omitempty"`
	OurReference           string                `json:"our_reference,omitempty"`
	YourReference          string                `json:"your_reference,omitempty"`
	PONumber               string                `json:"po_number,omitempty"`
	DeliveryDate           *time.Time            `json:"delivery_date,omitempty"`
	PerformancePeriodStart *time.Time            `json:"performance_period_start,omitempty"`
	PerformancePeriodEnd   *time.Time            `json:"performance_period_end,omitempty"`
	SubtotalAmount         float64               `json:"subtotal_amount"`
	TaxRate                float64               `json:"tax_rate"`
	TaxAmount              float64               `json:"tax_amount"`
	TotalAmount            float64               `json:"total_amount"`
	Currency               string                `json:"currency"`
	PaymentTerms           string                `json:"payment_terms,omitempty"`
	NetTerms               int                   `json:"net_terms"`
	PaymentMethod          string                `json:"payment_method,omitempty"`
	PaymentDate            *time.Time            `json:"payment_date,omitempty"`
	DiscountRate           float64               `json:"discount_rate,omitempty"`
	DiscountTerms          string                `json:"discount_terms,omitempty"`
	DocumentID             *uint                 `json:"document_id,omitempty"`
	Notes                  string                `json:"notes,omitempty"`
	Items                  []InvoiceItemResponse `json:"items,omitempty"`
	NumReminders           int                   `json:"num_reminders"`
	CreatedAt              time.Time             `json:"created_at"`
	UpdatedAt              time.Time             `json:"updated_at"`
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
		ID:                     i.ID,
		TenantID:               i.TenantID,
		OrganizationID:         i.OrganizationID,
		UserID:                 i.UserID,
		InvoiceNumber:          i.InvoiceNumber,
		InvoiceDate:            i.InvoiceDate,
		DueDate:                i.DueDate,
		Status:                 i.Status,
		CustomerName:           i.CustomerName,
		CustomerAddress:        i.CustomerAddress,
		CustomerAddressExt:     i.CustomerAddressExt,
		CustomerZip:            i.CustomerZip,
		CustomerCity:           i.CustomerCity,
		CustomerCountry:        i.CustomerCountry,
		CustomerEmail:          i.CustomerEmail,
		CustomerTaxID:          i.CustomerTaxID,
		CustomerContactPerson:  i.CustomerContactPerson,
		CustomerDepartment:     i.CustomerDepartment,
		Subject:                i.Subject,
		OurReference:           i.OurReference,
		YourReference:          i.YourReference,
		PONumber:               i.PONumber,
		DeliveryDate:           i.DeliveryDate,
		PerformancePeriodStart: i.PerformancePeriodStart,
		PerformancePeriodEnd:   i.PerformancePeriodEnd,
		SubtotalAmount:         i.SubtotalAmount,
		TaxRate:                i.TaxRate,
		TaxAmount:              i.TaxAmount,
		TotalAmount:            i.TotalAmount,
		Currency:               i.Currency,
		PaymentTerms:           i.PaymentTerms,
		NetTerms:               i.NetTerms,
		PaymentMethod:          i.PaymentMethod,
		PaymentDate:            i.PaymentDate,
		DiscountRate:           i.DiscountRate,
		DiscountTerms:          i.DiscountTerms,
		DocumentID:             i.DocumentID,
		Notes:                  i.Notes,
		NumReminders:           i.NumReminders,
		CreatedAt:              i.CreatedAt,
		UpdatedAt:              i.UpdatedAt,
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
