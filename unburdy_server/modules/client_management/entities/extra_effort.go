package entities

import (
	"time"

	"gorm.io/gorm"
)

// ExtraEffort represents unbilled therapeutic work outside scheduled sessions
type ExtraEffort struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	TenantID      uint           `gorm:"not null;index:idx_extra_efforts_tenant" json:"tenant_id"`
	ClientID      uint           `gorm:"not null;index:idx_extra_efforts_client" json:"client_id"`
	Client        *Client        `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	SessionID     *uint          `gorm:"index:idx_extra_efforts_session" json:"session_id"` // NULL if not session-related
	Session       *Session       `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	EffortType    string         `gorm:"size:50;not null" json:"effort_type"` // preparation, consultation, parent_meeting, documentation, other
	EffortDate    time.Time      `gorm:"not null;index:idx_extra_efforts_date" json:"effort_date"`
	DurationMin   int            `gorm:"not null" json:"duration_min"`
	Description   string         `gorm:"type:text" json:"description"`
	Billable      bool           `gorm:"default:true" json:"billable"`
	BillingStatus string         `gorm:"size:20;default:'delivered';index:idx_extra_efforts_billing_status" json:"billing_status"` // delivered, invoice-draft, billed, excluded
	InvoiceItemID *uint          `json:"invoice_item_id,omitempty"`                                                                // Link when billed
	InvoiceItem   *InvoiceItem   `gorm:"foreignKey:InvoiceItemID" json:"invoice_item,omitempty"`
	CreatedBy     uint           `json:"created_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-" swaggerignore:"true"`
}

// TableName specifies the table name for the ExtraEffort model
func (ExtraEffort) TableName() string {
	return "extra_efforts"
}

// CreateExtraEffortRequest represents the request payload for creating an extra effort
type CreateExtraEffortRequest struct {
	ClientID    uint   `json:"client_id" binding:"required" example:"1"`
	SessionID   *uint  `json:"session_id,omitempty" example:"5"`
	EffortType  string `json:"effort_type" binding:"required,oneof=preparation consultation parent_meeting documentation other" example:"preparation"`
	EffortDate  string `json:"effort_date" binding:"required" example:"2025-12-30"`
	DurationMin int    `json:"duration_min" binding:"required,min=1,max=480" example:"20"`
	Description string `json:"description,omitempty" example:"Copied therapy materials"`
	Billable    *bool  `json:"billable,omitempty" example:"true"`
}

// UpdateExtraEffortRequest represents the request payload for updating an extra effort
type UpdateExtraEffortRequest struct {
	EffortType  *string `json:"effort_type,omitempty" binding:"omitempty,oneof=preparation consultation parent_meeting documentation other" example:"consultation"`
	EffortDate  *string `json:"effort_date,omitempty" example:"2025-12-30"`
	DurationMin *int    `json:"duration_min,omitempty" binding:"omitempty,min=1,max=480" example:"30"`
	Description *string `json:"description,omitempty" example:"Updated description"`
	Billable    *bool   `json:"billable,omitempty" example:"false"`
}

// ExtraEffortResponse represents the response format for extra effort data
type ExtraEffortResponse struct {
	ID            uint      `json:"id"`
	ClientID      uint      `json:"client_id"`
	SessionID     *uint     `json:"session_id"`
	EffortType    string    `json:"effort_type"`
	EffortDate    time.Time `json:"effort_date"`
	DurationMin   int       `json:"duration_min"`
	Description   string    `json:"description"`
	Billable      bool      `json:"billable"`
	BillingStatus string    `json:"billing_status"`
	CreatedAt     time.Time `json:"created_at"`
}

// ExtraEffortAPIResponse represents the API response for a single extra effort
type ExtraEffortAPIResponse struct {
	Success bool                `json:"success" example:"true"`
	Message string              `json:"message" example:"Extra effort retrieved successfully"`
	Data    ExtraEffortResponse `json:"data"`
}

// ExtraEffortListAPIResponse represents the API response for extra effort list
type ExtraEffortListAPIResponse struct {
	Success bool                  `json:"success" example:"true"`
	Message string                `json:"message" example:"Extra efforts retrieved successfully"`
	Data    []ExtraEffortResponse `json:"data"`
	Total   int                   `json:"total" example:"10"`
}

// ExtraEffortDeleteResponse represents the API response for extra effort deletion
type ExtraEffortDeleteResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Extra effort deleted successfully"`
}
