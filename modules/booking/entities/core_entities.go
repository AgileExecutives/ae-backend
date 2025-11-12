package entities

import (
	"github.com/ae-base-server/pkg/core"
)

// BookingTemplateEntity implements core.Entity for BookingTemplate model
type BookingTemplateEntity struct{}

func NewBookingTemplateEntity() core.Entity {
	return &BookingTemplateEntity{}
}

func (e *BookingTemplateEntity) TableName() string {
	return "booking_templates"
}

func (e *BookingTemplateEntity) GetModel() interface{} {
	return &BookingTemplate{}
}

func (e *BookingTemplateEntity) GetMigrations() []core.Migration {
	return []core.Migration{} // No custom migrations needed, GORM handles basic schema
}
