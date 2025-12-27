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
