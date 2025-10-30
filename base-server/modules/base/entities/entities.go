package entities

import (
	"github.com/ae-base-server/internal/models"
	"github.com/ae-base-server/pkg/core"
)

// UserEntity implements core.Entity for User model
type UserEntity struct{}

func NewUserEntity() core.Entity {
	return &UserEntity{}
}

func (e *UserEntity) TableName() string {
	return "users"
}

func (e *UserEntity) GetModel() interface{} {
	return &models.User{}
}

func (e *UserEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}

// TenantEntity implements core.Entity for Tenant model
type TenantEntity struct{}

func NewTenantEntity() core.Entity {
	return &TenantEntity{}
}

func (e *TenantEntity) TableName() string {
	return "tenants"
}

func (e *TenantEntity) GetModel() interface{} {
	return &models.Tenant{}
}

func (e *TenantEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}

// NewsletterEntity implements core.Entity for Newsletter model
type NewsletterEntity struct{}

func NewNewsletterEntity() core.Entity {
	return &NewsletterEntity{}
}

func (e *NewsletterEntity) TableName() string {
	return "newsletters"
}

func (e *NewsletterEntity) GetModel() interface{} {
	return &models.Newsletter{}
}

func (e *NewsletterEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}

// ContactEntity implements core.Entity for Contact model
type ContactEntity struct{}

func NewContactEntity() core.Entity {
	return &ContactEntity{}
}

func (e *ContactEntity) TableName() string {
	return "contacts"
}

func (e *ContactEntity) GetModel() interface{} {
	return &models.Contact{}
}

func (e *ContactEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}

// TokenBlacklistEntity implements core.Entity for TokenBlacklist model
type TokenBlacklistEntity struct{}

func NewTokenBlacklistEntity() core.Entity {
	return &TokenBlacklistEntity{}
}

func (e *TokenBlacklistEntity) TableName() string {
	return "token_blacklists"
}

func (e *TokenBlacklistEntity) GetModel() interface{} {
	return &models.TokenBlacklist{}
}

func (e *TokenBlacklistEntity) GetMigrations() []core.Migration {
	return []core.Migration{}
}
