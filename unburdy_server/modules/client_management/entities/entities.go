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
