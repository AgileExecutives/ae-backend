package entities

import (
	baseCore "github.com/ae-base-server/pkg/core"
)

// DocumentEntity implements core.Entity for Document model
type DocumentEntity struct{}

func NewDocumentEntity() baseCore.Entity {
	return &DocumentEntity{}
}

func (e *DocumentEntity) TableName() string {
	return "documents"
}

func (e *DocumentEntity) GetModel() interface{} {
	return &Document{}
}

func (e *DocumentEntity) GetMigrations() []baseCore.Migration {
	return []baseCore.Migration{}
}

// TemplateEntity implements core.Entity for Template model
type TemplateEntity struct{}

func NewTemplateEntity() baseCore.Entity {
	return &TemplateEntity{}
}

func (e *TemplateEntity) TableName() string {
	return "templates"
}

func (e *TemplateEntity) GetModel() interface{} {
	return &Template{}
}

func (e *TemplateEntity) GetMigrations() []baseCore.Migration {
	return []baseCore.Migration{}
}

// InvoiceNumberEntity implements core.Entity for InvoiceNumber model
type InvoiceNumberEntity struct{}

func NewInvoiceNumberEntity() baseCore.Entity {
	return &InvoiceNumberEntity{}
}

func (e *InvoiceNumberEntity) TableName() string {
	return "invoice_numbers"
}

func (e *InvoiceNumberEntity) GetModel() interface{} {
	return &InvoiceNumber{}
}

func (e *InvoiceNumberEntity) GetMigrations() []baseCore.Migration {
	return []baseCore.Migration{}
}

// InvoiceNumberLogEntity implements core.Entity for InvoiceNumberLog model
type InvoiceNumberLogEntity struct{}

func NewInvoiceNumberLogEntity() baseCore.Entity {
	return &InvoiceNumberLogEntity{}
}

func (e *InvoiceNumberLogEntity) TableName() string {
	return "invoice_number_logs"
}

func (e *InvoiceNumberLogEntity) GetModel() interface{} {
	return &InvoiceNumberLog{}
}

func (e *InvoiceNumberLogEntity) GetMigrations() []baseCore.Migration {
	return []baseCore.Migration{}
}
