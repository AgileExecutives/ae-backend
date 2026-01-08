package models

import (
	"time"

	"gorm.io/gorm"
)

// UserSettings represents user-specific settings in the system
type UserSettings struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty" swaggerignore:"true"`
	UserID    uint           `gorm:"not null;uniqueIndex" json:"user_id" binding:"required"`
	// User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"` // Disabled for migration
	Language string `gorm:"default:en" json:"language"`
	Timezone string `gorm:"default:UTC" json:"timezone"`
	Theme    string `gorm:"default:light" json:"theme"`
	Settings string `gorm:"type:jsonb;default:'{}'" json:"settings"`
}

// TableName specifies the table name for UserSettings
func (UserSettings) TableName() string {
	return "user_settings"
}

// UserSettingsResponse represents the API response structure for UserSettings
type UserSettingsResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Language  string    `json:"language"`
	Timezone  string    `json:"timezone"`
	Theme     string    `json:"theme"`
	Settings  string    `json:"settings"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts UserSettings to UserSettingsResponse
func (us *UserSettings) ToResponse() UserSettingsResponse {
	return UserSettingsResponse{
		ID:        us.ID,
		UserID:    us.UserID,
		Language:  us.Language,
		Timezone:  us.Timezone,
		Theme:     us.Theme,
		Settings:  us.Settings,
		UpdatedAt: us.UpdatedAt,
	}
}

// UserSettingsUpdateRequest represents the request structure for updating user settings
type UserSettingsUpdateRequest struct {
	Language string `json:"language"`
	Timezone string `json:"timezone"`
	Theme    string `json:"theme"`
	Settings string `json:"settings"`
}
