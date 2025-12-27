package entities

import (
	baseCore "github.com/ae-base-server/pkg/core"
)

// InvoiceEntity implements core.Entity for Invoice model
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

// InvoiceItemEntity implements core.Entity for InvoiceItem model
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
