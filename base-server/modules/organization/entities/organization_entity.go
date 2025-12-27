package entities

import (
	"github.com/ae-base-server/internal/models"
	"github.com/ae-base-server/pkg/core"
)

// OrganizationEntity represents the organization entity
type OrganizationEntity struct{}

// NewOrganizationEntity creates a new organization entity
func NewOrganizationEntity() core.Entity {
	return &OrganizationEntity{}
}

func (e *OrganizationEntity) TableName() string {
	return "organizations"
}

func (e *OrganizationEntity) GetModel() interface{} {
	return &models.Organization{}
}

func (e *OrganizationEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}
