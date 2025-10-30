package entities

import (
	"github.com/ae-base-server/internal/models"
	"github.com/ae-base-server/pkg/core"
)

// CustomerEntity implements core.Entity for Customer model
type CustomerEntity struct{}

func NewCustomerEntity() core.Entity {
	return &CustomerEntity{}
}

func (e *CustomerEntity) TableName() string {
	return "customers"
}

func (e *CustomerEntity) GetModel() interface{} {
	return &models.Customer{}
}

func (e *CustomerEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}

// PlanEntity implements core.Entity for Plan model
type PlanEntity struct{}

func NewPlanEntity() core.Entity {
	return &PlanEntity{}
}

func (e *PlanEntity) TableName() string {
	return "plans"
}

func (e *PlanEntity) GetModel() interface{} {
	return &models.Plan{}
}

func (e *PlanEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}
