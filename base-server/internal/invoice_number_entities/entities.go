package invoice_number_entities

import (
	"github.com/ae-base-server/pkg/core"
	"github.com/unburdy/invoice-number-module/entities"
)

// InvoiceNumberEntity represents the invoice number entity
type InvoiceNumberEntity struct{}

func NewInvoiceNumberEntity() core.Entity {
	return &InvoiceNumberEntity{}
}

func (e *InvoiceNumberEntity) TableName() string {
	return "invoice_numbers"
}

func (e *InvoiceNumberEntity) GetModel() interface{} {
	return &entities.InvoiceNumber{}
}

func (e *InvoiceNumberEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}

// InvoiceNumberLogEntity represents the invoice number log entity
type InvoiceNumberLogEntity struct{}

func NewInvoiceNumberLogEntity() core.Entity {
	return &InvoiceNumberLogEntity{}
}

func (e *InvoiceNumberLogEntity) TableName() string {
	return "invoice_number_logs"
}

func (e *InvoiceNumberLogEntity) GetModel() interface{} {
	return &entities.InvoiceNumberLog{}
}

func (e *InvoiceNumberLogEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}
