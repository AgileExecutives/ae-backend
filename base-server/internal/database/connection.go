// Package database provides backward compatibility for internal database operations.
//
// This package re-exports functions from pkg/database and provides base-server
// specific seeding functionality. New code should import pkg/database directly
// for connection operations and use this package only for seeding.
package database

// Re-export connection utilities from pkg/database for backward compatibility
import (
	pkgDB "github.com/ae-base-server/pkg/database"
	"gorm.io/gorm"
)

// Config re-exports the database configuration type
type Config = pkgDB.Config

// Connect re-exports the database connection function
func Connect(config Config) (*gorm.DB, error) {
	return pkgDB.Connect(config)
}

// ConnectWithAutoCreate re-exports the auto-create database connection function
func ConnectWithAutoCreate(config Config) (*gorm.DB, error) {
	return pkgDB.ConnectWithAutoCreate(config)
}

// CreateDatabaseIfNotExists re-exports the database creation function
func CreateDatabaseIfNotExists(config Config) error {
	return pkgDB.CreateDatabaseIfNotExists(config)
}

// GetDefaultConfig re-exports the default configuration function
func GetDefaultConfig() Config {
	return pkgDB.GetDefaultConfig()
}
