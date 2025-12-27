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
	InvoiceStatusGenerated InvoiceStatus = "generated"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusReminder  InvoiceStatus = "reminder"
	InvoiceStatusPayed     InvoiceStatus = "payed"
)

// Invoice represents an invoice entity
type Invoice struct {
	ID             uint                  `gorm:"primarykey" json:"id"`
	TenantID       uint                  `gorm:"not null;index:idx_invoice_tenant" json:"tenant_id"`
	UserID         uint                  `gorm:"not null;index:idx_invoice_user" json:"user_id"`
	ClientID       uint                  `gorm:"not null;index:idx_invoice_client" json:"client_id"`
	Client         *Client               `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	CostProviderID uint                  `gorm:"not null;index:idx_invoice_cost_provider" json:"cost_provider_id"`
	CostProvider   *CostProvider         `gorm:"foreignKey:CostProviderID" json:"cost_provider,omitempty"`
	OrganizationID uint                  `gorm:"not null;index:idx_invoice_organization" json:"organization_id"`
	Organization   *baseAPI.Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	InvoiceDate    time.Time             `gorm:"not null" json:"invoice_date"`
	InvoiceNumber  string                `gorm:"size:100;not null;uniqueIndex:idx_invoice_number_tenant" json:"invoice_number"`
	NumberUnits    int                   `gorm:"not null;default:0" json:"number_units"`
	SumAmount      float64               `gorm:"type:decimal(10,2);not null;default:0" json:"sum_amount"`
	TaxAmount      float64               `gorm:"type:decimal(10,2);not null;default:0" json:"tax_amount"`
	TotalAmount    float64               `gorm:"type:decimal(10,2);not null;default:0" json:"total_amount"`
	PayedDate      *time.Time            `json:"payed_date,omitempty"`
	Status         InvoiceStatus         `gorm:"size:20;not null;default:'draft'" json:"status"`
	NumReminders   int                   `gorm:"not null;default:0" json:"num_reminders"`
	LatestReminder *time.Time            `json:"latest_reminder,omitempty"`
	InvoiceItems   []InvoiceItem         `gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE" json:"invoice_items,omitempty"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
	DeletedAt      gorm.DeletedAt        `gorm:"index" json:"-"`
}

// TableName specifies the table name for the Invoice model
func (Invoice) TableName() string {
	return "invoices"
}

// InvoiceItem represents an invoice item linked to a session
type InvoiceItem struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	InvoiceID uint           `gorm:"not null;index:idx_invoice_item_invoice" json:"invoice_id"`
	SessionID uint           `gorm:"not null;uniqueIndex:idx_invoice_item_session" json:"session_id"`
	Session   *Session       `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for the InvoiceItem model
func (InvoiceItem) TableName() string {
	return "invoice_items"
}

// CreateInvoiceRequest represents the request payload for creating an invoice
type CreateInvoiceRequest struct {
	ClientID   uint   `json:"client_id" binding:"required" example:"1"`
	SessionIDs []uint `json:"session_ids" binding:"required" example:"1,2,3"`
}

// UpdateInvoiceRequest represents the request payload for updating an invoice
type UpdateInvoiceRequest struct {
	Status     *InvoiceStatus `json:"status,omitempty" example:"sent"`
	SessionIDs []uint         `json:"session_ids,omitempty" example:"1,2,3"`
}

// InvoiceResponse represents the response format for invoice data
type InvoiceResponse struct {
	ID             uint                          `json:"id"`
	TenantID       uint                          `json:"tenant_id"`
	UserID         uint                          `json:"user_id"`
	ClientID       uint                          `json:"client_id"`
	Client         *ClientResponse               `json:"client,omitempty"`
	CostProviderID uint                          `json:"cost_provider_id"`
	CostProvider   *CostProviderResponse         `json:"cost_provider,omitempty"`
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
	InvoiceItems   []InvoiceItemResponse         `json:"invoice_items,omitempty"`
	CreatedAt      time.Time                     `json:"created_at"`
	UpdatedAt      time.Time                     `json:"updated_at"`
}

// InvoiceItemResponse represents the response format for invoice item data
type InvoiceItemResponse struct {
	ID        uint             `json:"id"`
	InvoiceID uint             `json:"invoice_id"`
	SessionID uint             `json:"session_id"`
	Session   *SessionResponse `json:"session,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// ClientWithUnbilledSessionsResponse represents a client with sessions not yet invoiced
type ClientWithUnbilledSessionsResponse struct {
	ClientResponse
	Sessions []SessionResponse `json:"sessions"`
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
		ClientID:       i.ClientID,
		CostProviderID: i.CostProviderID,
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
		CreatedAt:      i.CreatedAt,
		UpdatedAt:      i.UpdatedAt,
	}

	if i.Client != nil {
		clientResp := i.Client.ToResponse()
		response.Client = &clientResp
	}

	if i.CostProvider != nil {
		costProviderResp := i.CostProvider.ToResponse()
		response.CostProvider = &costProviderResp
	}

	if i.Organization != nil {
		orgResp := i.Organization.ToResponse()
		response.Organization = &orgResp
	}

	if len(i.InvoiceItems) > 0 {
		response.InvoiceItems = make([]InvoiceItemResponse, len(i.InvoiceItems))
		for idx, item := range i.InvoiceItems {
			response.InvoiceItems[idx] = item.ToResponse()
		}
	}

	return response
}

// ToResponse converts an InvoiceItem to InvoiceItemResponse
func (ii *InvoiceItem) ToResponse() InvoiceItemResponse {
	response := InvoiceItemResponse{
		ID:        ii.ID,
		InvoiceID: ii.InvoiceID,
		SessionID: ii.SessionID,
		CreatedAt: ii.CreatedAt,
		UpdatedAt: ii.UpdatedAt,
	}

	if ii.Session != nil {
		sessionResp := ii.Session.ToResponse()
		response.Session = &sessionResp
	}

	return response
}
