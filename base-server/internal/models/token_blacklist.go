package models

import (
	"time"

	"gorm.io/gorm"
)

// TokenBlacklist represents blacklisted JWT tokens
type TokenBlacklist struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TokenID   string         `gorm:"not null;uniqueIndex" json:"token_id" binding:"required"`
	UserID    uint           `gorm:"not null" json:"user_id" binding:"required"`
	// User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"` // Temporarily disabled
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Reason    string    `json:"reason"`
}

// TableName specifies the table name for TokenBlacklist
func (TokenBlacklist) TableName() string {
	return "token_blacklist"
}

// IsBlacklisted checks if a token ID is blacklisted
func (tb *TokenBlacklist) IsBlacklisted() bool {
	return time.Now().Before(tb.ExpiresAt)
}
