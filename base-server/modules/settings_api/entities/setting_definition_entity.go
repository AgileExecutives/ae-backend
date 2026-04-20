package entities

import (
	"github.com/ae-base-server/pkg/core"
	settingsEntities "github.com/ae-base-server/pkg/settings/entities"
)

// SettingDefinitionEntity represents the setting definitions entity.
type SettingDefinitionEntity struct{}

func NewSettingDefinitionEntity() core.Entity {
	return &SettingDefinitionEntity{}
}

func (e *SettingDefinitionEntity) TableName() string {
	return "setting_definitions"
}

func (e *SettingDefinitionEntity) GetModel() interface{} {
	return &settingsEntities.SettingDefinition{}
}

func (e *SettingDefinitionEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}
