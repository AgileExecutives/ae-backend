package entities

import (
	"time"

	"github.com/ae-base-server/pkg/core"
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

// Entity wrappers for core.Module interface
type InvoiceNumberEntity struct{}

func NewInvoiceNumberEntity() *InvoiceNumberEntity {
	return &InvoiceNumberEntity{}
}

func (e *InvoiceNumberEntity) TableName() string {
	return "invoice_numbers"
}

func (e *InvoiceNumberEntity) GetModel() interface{} {
	return &InvoiceNumber{}
}

func (e *InvoiceNumberEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}

type InvoiceNumberLogEntity struct{}

func NewInvoiceNumberLogEntity() *InvoiceNumberLogEntity {
	return &InvoiceNumberLogEntity{}
}

func (e *InvoiceNumberLogEntity) TableName() string {
	return "invoice_number_logs"
}

func (e *InvoiceNumberLogEntity) GetModel() interface{} {
	return &InvoiceNumberLog{}
}

func (e *InvoiceNumberLogEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}

// InvoiceNumberResponse for API responses
type InvoiceNumberResponse struct {
	ID             uint      `json:"id"`
	TenantID       uint      `json:"tenant_id"`
	OrganizationID uint      `json:"organization_id"`
	InvoiceNumber  string    `json:"invoice_number"`
	Year           int       `json:"year"`
	Month          int       `json:"month"`
	Sequence       int       `json:"sequence"`
	Status         string    `json:"status"`
	GeneratedAt    time.Time `json:"generated_at"`
	Format         string    `json:"format,omitempty"`
}

// ToResponse converts InvoiceNumberLog to response format
func (log *InvoiceNumberLog) ToResponse() InvoiceNumberResponse {
	return InvoiceNumberResponse{
		ID:             log.ID,
		TenantID:       log.TenantID,
		OrganizationID: log.OrganizationID,
		InvoiceNumber:  log.InvoiceNumber,
		Year:           log.Year,
		Month:          log.Month,
		Sequence:       log.Sequence,
		Status:         log.Status,
		GeneratedAt:    log.GeneratedAt,
	}
}
