package entities

import (
	baseCore "github.com/ae-base-server/pkg/core"
)

// ClientEntity implements core.Entity for Client model
type ClientEntity struct{}

func NewClientEntity() baseCore.Entity {
	return &ClientEntity{}
}

func (e *ClientEntity) TableName() string {
	return "clients"
}

func (e *ClientEntity) GetModel() interface{} {
	return &Client{}
}

func (e *ClientEntity) GetMigrations() []baseCore.Migration {
	return []baseCore.Migration{}
}

// CostProviderEntity implements baseCore.Entity for CostProvider model
type CostProviderEntity struct{}

func NewCostProviderEntity() baseCore.Entity {
	return &CostProviderEntity{}
}

func (e *CostProviderEntity) TableName() string {
	return "cost_providers"
}

func (e *CostProviderEntity) GetModel() interface{} {
	return &CostProvider{}
}

func (e *CostProviderEntity) GetMigrations() []baseCore.Migration {
	return []baseCore.Migration{}
}

// SessionEntity implements baseCore.Entity for Session model
type SessionEntity struct{}

func NewSessionEntity() baseCore.Entity {
	return &SessionEntity{}
}

func (e *SessionEntity) TableName() string {
	return "sessions"
}

func (e *SessionEntity) GetModel() interface{} {
	return &Session{}
}

func (e *SessionEntity) GetMigrations() []baseCore.Migration {
	return []baseCore.Migration{}
}

// InvoiceEntity implements baseCore.Entity for Invoice model
type InvoiceEntity struct{}

func NewInvoiceEntity() baseCore.Entity {
	return &InvoiceEntity{}
}

func (e *InvoiceEntity) TableName() string {
	return "invoices"
}

func (e *InvoiceEntity) GetModel() interface{} {
	return &Invoice{}
}

func (e *InvoiceEntity) GetMigrations() []baseCore.Migration {
	return []baseCore.Migration{}
}

// InvoiceItemEntity implements baseCore.Entity for InvoiceItem model
type InvoiceItemEntity struct{}

func NewInvoiceItemEntity() baseCore.Entity {
	return &InvoiceItemEntity{}
}

func (e *InvoiceItemEntity) TableName() string {
	return "invoice_items"
}

func (e *InvoiceItemEntity) GetModel() interface{} {
	return &InvoiceItem{}
}

func (e *InvoiceItemEntity) GetMigrations() []baseCore.Migration {
	return []baseCore.Migration{}
}

// ClientInvoiceEntity implements baseCore.Entity for ClientInvoice model
type ClientInvoiceEntity struct{}

func NewClientInvoiceEntity() baseCore.Entity {
	return &ClientInvoiceEntity{}
}

func (e *ClientInvoiceEntity) TableName() string {
	return "client_invoices"
}

func (e *ClientInvoiceEntity) GetModel() interface{} {
	return &ClientInvoice{}
}

func (e *ClientInvoiceEntity) GetMigrations() []baseCore.Migration {
	return []baseCore.Migration{}
}

// ExtraEffortEntity implements baseCore.Entity for ExtraEffort model
type ExtraEffortEntity struct{}

func NewExtraEffortEntity() baseCore.Entity {
	return &ExtraEffortEntity{}
}

func (e *ExtraEffortEntity) TableName() string {
	return "extra_efforts"
}

func (e *ExtraEffortEntity) GetModel() interface{} {
	return &ExtraEffort{}
}

func (e *ExtraEffortEntity) GetMigrations() []baseCore.Migration {
	return []baseCore.Migration{}
}
