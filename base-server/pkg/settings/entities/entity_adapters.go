package entities

import (
	"github.com/ae-base-server/pkg/core"
)

// SettingDefinitionEntity implements core.Entity for SettingDefinition model
type SettingDefinitionEntity struct{}

func NewSettingDefinitionEntity() core.Entity {
	return &SettingDefinitionEntity{}
}

func (e *SettingDefinitionEntity) TableName() string {
	return "setting_definitions"
}

func (e *SettingDefinitionEntity) GetModel() interface{} {
	return &SettingDefinition{}
}

func (e *SettingDefinitionEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}

// SettingEntity implements core.Entity for Setting model
type SettingEntity struct{}

func NewSettingEntity() core.Entity {
	return &SettingEntity{}
}

func (e *SettingEntity) TableName() string {
	return "settings"
}

func (e *SettingEntity) GetModel() interface{} {
	return &Setting{}
}

func (e *SettingEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}
