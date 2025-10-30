package entities

import (
	"github.com/ae-base-server/internal/models"
	"github.com/ae-base-server/pkg/core"
	"gorm.io/gorm"
)

// EmailEntity implements the core.Entity interface for email entities
type EmailEntity struct{}

// NewEmailEntity creates a new email entity instance
func NewEmailEntity() *EmailEntity {
	return &EmailEntity{}
}

// Name returns the entity name
func (e *EmailEntity) Name() string {
	return "email"
}

// TableName returns the database table name
func (e *EmailEntity) TableName() string {
	return "emails"
}

// GetModel returns the GORM model for the entity
func (e *EmailEntity) GetModel() interface{} {
	return &models.Email{}
}

// GetMigrations returns custom migrations for the entity
func (e *EmailEntity) GetMigrations() []core.Migration {
	return []core.Migration{
		&EmailMigration001{},
	}
}

// EmailMigration001 implements the Migration interface for creating emails table
type EmailMigration001 struct{}

// Version returns the migration version
func (m *EmailMigration001) Version() string {
	return "001_create_emails_table"
}

// Up applies the migration
func (m *EmailMigration001) Up(db *gorm.DB) error {
	return db.AutoMigrate(&models.Email{})
}

// Down rolls back the migration
func (m *EmailMigration001) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&models.Email{})
}
