package entities

import (
	baseCore "github.com/ae-base-server/pkg/core"
)

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
