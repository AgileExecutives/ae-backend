package entities

import (
	"github.com/ae-base-server/pkg/core"
	settingsEntities "github.com/ae-base-server/pkg/settings/entities"
)

// SettingEntity represents the settings value entity.
type SettingEntity struct{}

func NewSettingEntity() core.Entity {
	return &SettingEntity{}
}

func (e *SettingEntity) TableName() string {
	return "settings"
}

func (e *SettingEntity) GetModel() interface{} {
	return &settingsEntities.Setting{}
}

func (e *SettingEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}
