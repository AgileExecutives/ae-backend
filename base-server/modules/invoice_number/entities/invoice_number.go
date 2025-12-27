package entities

import (
	"time"

	"gorm.io/gorm"
)

// InvoiceNumber tracks invoice number sequences with PostgreSQL persistence
type InvoiceNumber struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	TenantID       uint           `gorm:"not null;index:idx_tenant_org_year_month" json:"tenant_id"`
	OrganizationID uint           `gorm:"index:idx_tenant_org_year_month" json:"organization_id"`
	Year           int            `gorm:"not null;index:idx_tenant_org_year_month" json:"year"`
	Month          int            `gorm:"not null;index:idx_tenant_org_year_month" json:"month"`
	Sequence       int            `gorm:"not null;default:0" json:"sequence"`
	LastNumber     string         `gorm:"size:50" json:"last_number"`
	Format         string         `gorm:"size:100" json:"format"` // e.g., "INV-{YYYY}-{SEQ:4}"
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name
func (InvoiceNumber) TableName() string {
	return "invoice_numbers"
}

// InvoiceNumberLog tracks all generated invoice numbers for audit trail
type InvoiceNumberLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TenantID       uint      `gorm:"not null;index:idx_tenant_invoice" json:"tenant_id"`
	OrganizationID uint      `gorm:"index:idx_org_invoice" json:"organization_id"`
	InvoiceNumber  string    `gorm:"not null;size:50;uniqueIndex:idx_tenant_invoice_number" json:"invoice_number"`
	Year           int       `gorm:"not null;index" json:"year"`
	Month          int       `gorm:"not null;index" json:"month"`
	Sequence       int       `gorm:"not null" json:"sequence"`
	ReferenceID    uint      `gorm:"index" json:"reference_id,omitempty"`    // Optional reference to invoice
	ReferenceType  string    `gorm:"size:50" json:"reference_type,omitempty"` // e.g., "invoice", "credit_note"
	Status         string    `gorm:"size:20;default:'active'" json:"status"` // active, voided, cancelled
	GeneratedAt    time.Time `gorm:"not null" json:"generated_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName specifies the table name
func (InvoiceNumberLog) TableName() string {
	return "invoice_number_logs"
}

// InvoiceNumberResponse is the JSON response format
type InvoiceNumberResponse struct {
	InvoiceNumber  string    `json:"invoice_number"`
	Year           int       `json:"year"`
	Month          int       `json:"month"`
	Sequence       int       `json:"sequence"`
	Format         string    `json:"format"`
	OrganizationID uint      `json:"organization_id"`
	GeneratedAt    time.Time `json:"generated_at"`
}

// ToResponse converts InvoiceNumberLog to response format
func (log *InvoiceNumberLog) ToResponse() InvoiceNumberResponse {
	return InvoiceNumberResponse{
		InvoiceNumber:  log.InvoiceNumber,
		Year:           log.Year,
		Month:          log.Month,
		Sequence:       log.Sequence,
		Format:         "", // Will be filled from config
		OrganizationID: log.OrganizationID,
		GeneratedAt:    log.GeneratedAt,
	}
}
