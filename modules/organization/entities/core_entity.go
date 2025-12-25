package entities

import (
"github.com/ae-base-server/pkg/core"
)

// OrganizationEntity implements core.Entity for Organization model
type OrganizationEntity struct{}

func NewOrganizationEntity() core.Entity {
	return &OrganizationEntity{}
}

func (e *OrganizationEntity) TableName() string {
	return "organizations"
}

func (e *OrganizationEntity) GetModel() interface{} {
	return &Organization{}
}

func (e *OrganizationEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}
