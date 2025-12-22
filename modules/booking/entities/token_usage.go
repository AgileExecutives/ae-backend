package entities

import (
	"time"

	"gorm.io/gorm"
)

// TokenUsage tracks usage statistics for booking link tokens
type TokenUsage struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	TokenID     string     `gorm:"not null;uniqueIndex;size:64" json:"token_id"` // SHA256 hash of token
	TenantID    uint       `gorm:"not null;index" json:"tenant_id"`
	TemplateID  uint       `gorm:"not null;index" json:"template_id"`
	ClientID    uint       `gorm:"not null;index" json:"client_id"`
	UseCount    int        `gorm:"not null;default:0" json:"use_count"`
	MaxUseCount int        `gorm:"not null;default:0" json:"max_use_count"` // 0 means unlimited
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// TableName specifies the table name for TokenUsage
func (TokenUsage) TableName() string {
	return "booking_token_usage"
}

// IsExpired checks if the token has expired
func (tu *TokenUsage) IsExpired() bool {
	if tu.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*tu.ExpiresAt)
}

// HasReachedLimit checks if the token has reached its usage limit
func (tu *TokenUsage) HasReachedLimit() bool {
	if tu.MaxUseCount == 0 {
		return false // Unlimited usage
	}
	return tu.UseCount >= tu.MaxUseCount
}

// CanBeUsed checks if the token can still be used
func (tu *TokenUsage) CanBeUsed() bool {
	return !tu.IsExpired() && !tu.HasReachedLimit()
}

// IncrementUsage increments the usage counter and updates last used timestamp
func (tu *TokenUsage) IncrementUsage() {
	tu.UseCount++
	now := time.Now()
	tu.LastUsedAt = &now
}
