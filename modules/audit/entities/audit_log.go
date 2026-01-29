package entities

import (
"encoding/json"
"time"

"gorm.io/gorm"
)

// AuditAction represents the type of action performed
type AuditAction string

const (
AuditActionInvoiceDraftCreated   AuditAction = "invoice_draft_created"
AuditActionInvoiceDraftUpdated   AuditAction = "invoice_draft_updated"
AuditActionInvoiceDraftCancelled AuditAction = "invoice_draft_cancelled"
AuditActionInvoiceFinalized      AuditAction = "invoice_finalized"
AuditActionInvoiceSent           AuditAction = "invoice_sent"
AuditActionInvoiceMarkedPaid     AuditAction = "invoice_marked_paid"
AuditActionInvoiceMarkedOverdue  AuditAction = "invoice_marked_overdue"
AuditActionReminderSent          AuditAction = "reminder_sent"
AuditActionCreditNoteCreated     AuditAction = "credit_note_created"
AuditActionXRechnungExported     AuditAction = "xrechnung_exported"
)

// EntityType represents the type of entity being audited
type EntityType string

const (
EntityTypeInvoice     EntityType = "invoice"
EntityTypeInvoiceItem EntityType = "invoice_item"
EntityTypeSession     EntityType = "session"
EntityTypeExtraEffort EntityType = "extra_effort"
)

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID         uint            `gorm:"primarykey" json:"id"`
	TenantID   uint            `gorm:"not null;index:idx_audit_tenant" json:"tenant_id"`
	UserID     uint            `gorm:"not null;index:idx_audit_user" json:"user_id"`
	EntityType EntityType      `gorm:"type:varchar(50);not null;index:idx_audit_entity" json:"entity_type"`
	EntityID   uint            `gorm:"not null;index:idx_audit_entity" json:"entity_id"`
	Action     AuditAction     `gorm:"type:varchar(100);not null;index:idx_audit_action" json:"action"`
	Metadata   json.RawMessage `gorm:"type:jsonb" json:"metadata,omitempty"`
	IPAddress  string          `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent  string          `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt  time.Time       `gorm:"not null;index:idx_audit_created" json:"created_at"`
}

// TableName specifies the table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate hook to set created_at
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	return nil
}

// AuditLogMetadata represents the structured metadata for audit logs
type AuditLogMetadata struct {
	InvoiceNumber   string                 `json:"invoice_number,omitempty"`
	InvoiceStatus   string                 `json:"invoice_status,omitempty"`
	TotalAmount     float64                `json:"total_amount,omitempty"`
	PaymentDate     *time.Time             `json:"payment_date,omitempty"`
	NumReminders    int                    `json:"num_reminders,omitempty"`
	CreditReference string                 `json:"credit_reference,omitempty"`
	Changes         map[string]interface{} `json:"changes,omitempty"`
	Reason          string                 `json:"reason,omitempty"`
	AdditionalInfo  map[string]interface{} `json:"additional_info,omitempty"`
}

// ToJSON converts metadata to JSON
func (m *AuditLogMetadata) ToJSON() (json.RawMessage, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// AuditLogResponse represents the API response for audit log
type AuditLogResponse struct {
	ID         uint                   `json:"id"`
	TenantID   uint                   `json:"tenant_id"`
	UserID     uint                   `json:"user_id"`
	EntityType EntityType             `json:"entity_type"`
	EntityID   uint                   `json:"entity_id"`
	Action     AuditAction            `json:"action"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// ToResponse converts AuditLog to AuditLogResponse
func (a *AuditLog) ToResponse() AuditLogResponse {
	response := AuditLogResponse{
		ID:         a.ID,
		TenantID:   a.TenantID,
		UserID:     a.UserID,
		EntityType: a.EntityType,
		EntityID:   a.EntityID,
		Action:     a.Action,
		IPAddress:  a.IPAddress,
		UserAgent:  a.UserAgent,
		CreatedAt:  a.CreatedAt,
	}

	if len(a.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(a.Metadata, &metadata); err == nil {
			response.Metadata = metadata
		}
	}

	return response
}

// AuditLogFilter represents filter criteria for audit log queries
type AuditLogFilter struct {
	TenantID   uint         `form:"tenant_id"`
	UserID     *uint        `form:"user_id"`
	EntityType *EntityType  `form:"entity_type"`
	EntityID   *uint        `form:"entity_id"`
	Action     *AuditAction `form:"action"`
	StartDate  *time.Time   `form:"start_date"`
	EndDate    *time.Time   `form:"end_date"`
	Page       int          `form:"page"`
	Limit      int          `form:"limit"`
}

// AuditLogListResponse represents the API response for audit log list
type AuditLogListResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Data    []AuditLogResponse `json:"data"`
	Page    int                `json:"page"`
	Limit   int                `json:"limit"`
	Total   int64              `json:"total"`
}
