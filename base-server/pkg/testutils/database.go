package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Reduce test noise
	})
	require.NoError(t, err, "Failed to create test database")

	return db
}

// SetupTestDBWithLogging creates a test DB with query logging enabled
func SetupTestDBWithLogging(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	require.NoError(t, err, "Failed to create test database")

	return db
}

// CleanupTestDB closes the database connection
func CleanupTestDB(db *gorm.DB) {
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// MigrateTestDB runs auto-migration on test database
func MigrateTestDB(t *testing.T, db *gorm.DB, entities ...interface{}) {
	err := db.AutoMigrate(entities...)
	require.NoError(t, err, "Failed to migrate test database")
}

// TruncateTable removes all records from a table
func TruncateTable(db *gorm.DB, tableName string) error {
	return db.Exec("DELETE FROM " + tableName).Error
}

// GetRowCount returns the number of rows in a table
func GetRowCount(db *gorm.DB, tableName string) (int64, error) {
	var count int64
	err := db.Table(tableName).Count(&count).Error
	return count, err
}

// BeginTestTransaction starts a transaction for testing
// Caller should defer tx.Rollback() to ensure cleanup
func BeginTestTransaction(t *testing.T, db *gorm.DB) *gorm.DB {
	tx := db.Begin()
	require.NoError(t, tx.Error, "Failed to begin transaction")
	return tx
}
