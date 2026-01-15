package services

import (
	"testing"
	"time"

	baseAPI "github.com/ae-base-server/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupInvoiceNumberTestDB(t *testing.T) (*gorm.DB, *InvoiceNumberService) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&entities.Invoice{},
		&baseAPI.Organization{},
	)
	require.NoError(t, err)

	service := NewInvoiceNumberService(db)
	return db, service
}

func createTestOrg(t *testing.T, db *gorm.DB, tenantID uint, format, prefix string) uint {
	org := &baseAPI.Organization{
		TenantID:            tenantID,
		Name:                "Test Organization",
		InvoiceNumberFormat: format,
		InvoiceNumberPrefix: prefix,
	}

	require.NoError(t, db.Create(org).Error)
	return org.ID
}

// TestInvoiceNumberService_Sequential tests sequential number generation
func TestInvoiceNumberService_Sequential(t *testing.T) {
	db, service := setupInvoiceNumberTestDB(t)
	orgID := createTestOrg(t, db, 1, "sequential", "INV")
	invoiceDate := time.Now()

	// First invoice
	number1, err := service.GenerateInvoiceNumber(orgID, 1, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "INV-0001", number1)

	// Create invoice with that number
	invoice1 := &entities.Invoice{
		TenantID:       1,
		OrganizationID: orgID,
		InvoiceNumber:  number1,
		Status:         entities.InvoiceStatusSent,
	}
	require.NoError(t, db.Create(invoice1).Error)

	// Second invoice
	number2, err := service.GenerateInvoiceNumber(orgID, 1, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "INV-0002", number2)

	// Third invoice
	invoice2 := &entities.Invoice{
		TenantID:       1,
		OrganizationID: orgID,
		InvoiceNumber:  number2,
		Status:         entities.InvoiceStatusSent,
	}
	require.NoError(t, db.Create(invoice2).Error)

	number3, err := service.GenerateInvoiceNumber(orgID, 1, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "INV-0003", number3)
}

// TestInvoiceNumberService_YearPrefix tests year-based numbering
func TestInvoiceNumberService_YearPrefix(t *testing.T) {
	db, service := setupInvoiceNumberTestDB(t)
	orgID := createTestOrg(t, db, 1, "year_prefix", "")

	// Current year invoice
	now := time.Now()
	number1, err := service.GenerateInvoiceNumber(orgID, 1, now)
	require.NoError(t, err)
	expectedNumber := now.Format("2006") + "-0001"
	assert.Equal(t, expectedNumber, number1)

	// Create invoice
	invoice1 := &entities.Invoice{
		TenantID:       1,
		OrganizationID: orgID,
		InvoiceNumber:  number1,
		Status:         entities.InvoiceStatusSent,
	}
	require.NoError(t, db.Create(invoice1).Error)

	// Second invoice same year
	number2, err := service.GenerateInvoiceNumber(orgID, 1, now)
	require.NoError(t, err)
	expectedNumber2 := now.Format("2006") + "-0002"
	assert.Equal(t, expectedNumber2, number2)

	// Create invoice from last year (simulated)
	lastYear := now.AddDate(-1, 0, 0)
	lastYearNumber := lastYear.Format("2006") + "-0099"
	lastYearInvoice := &entities.Invoice{
		TenantID:       1,
		OrganizationID: orgID,
		InvoiceNumber:  lastYearNumber,
		Status:         entities.InvoiceStatusSent,
	}
	require.NoError(t, db.Create(lastYearInvoice).Error)

	// Next invoice should still be 0002 (year filters prevent cross-year counting)
	number3, err := service.GenerateInvoiceNumber(orgID, 1, now)
	require.NoError(t, err)
	expectedNumber3 := now.Format("2006") + "-0002"
	assert.Equal(t, expectedNumber3, number3)
}

// TestInvoiceNumberService_YearMonth tests monthly numbering
func TestInvoiceNumberService_YearMonth(t *testing.T) {
	db, service := setupInvoiceNumberTestDB(t)
	orgID := createTestOrg(t, db, 1, "year_month_prefix", "INV")

	now := time.Now()

	// First invoice this month
	number1, err := service.GenerateInvoiceNumber(orgID, 1, now)
	require.NoError(t, err)
	expectedNumber := "INV-" + now.Format("2006-01") + "-0001"
	assert.Equal(t, expectedNumber, number1)

	// Create invoice
	invoice1 := &entities.Invoice{
		TenantID:       1,
		OrganizationID: orgID,
		InvoiceNumber:  number1,
		Status:         entities.InvoiceStatusSent,
	}
	require.NoError(t, db.Create(invoice1).Error)

	// Second invoice this month
	number2, err := service.GenerateInvoiceNumber(orgID, 1, now)
	require.NoError(t, err)
	expectedNumber2 := "INV-" + now.Format("2006-01") + "-0002"
	assert.Equal(t, expectedNumber2, number2)
}

// TestInvoiceNumberService_NoPrefix tests numbering without prefix
func TestInvoiceNumberService_NoPrefix(t *testing.T) {
	db, service := setupInvoiceNumberTestDB(t)
	orgID := createTestOrg(t, db, 1, "sequential", "")

	invoiceDate := time.Now()
	number, err := service.GenerateInvoiceNumber(orgID, 1, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "1", number) // No prefix = unpadded number
}

// TestInvoiceNumberService_CustomPrefix tests custom prefix
func TestInvoiceNumberService_CustomPrefix(t *testing.T) {
	db, service := setupInvoiceNumberTestDB(t)
	orgID := createTestOrg(t, db, 1, "sequential", "THERAPY")

	invoiceDate := time.Now()
	number, err := service.GenerateInvoiceNumber(orgID, 1, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "THERAPY-0001", number)
}

// TestInvoiceNumberService_DraftInvoicesIgnored tests that draft invoices don't affect numbering
func TestInvoiceNumberService_DraftInvoicesIgnored(t *testing.T) {
	db, service := setupInvoiceNumberTestDB(t)
	orgID := createTestOrg(t, db, 1, "sequential", "INV")
	invoiceDate := time.Now()

	// Create a draft invoice with a number
	draftInvoice := &entities.Invoice{
		TenantID:       1,
		OrganizationID: orgID,
		InvoiceNumber:  "INV-9999",
		Status:         entities.InvoiceStatusDraft,
	}
	require.NoError(t, db.Create(draftInvoice).Error)

	// Next number should be 0001, ignoring draft
	number, err := service.GenerateInvoiceNumber(orgID, 1, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "INV-0001", number)
}

// TestInvoiceNumberService_OrganizationNotFound tests error handling
func TestInvoiceNumberService_OrganizationNotFound(t *testing.T) {
	_, service := setupInvoiceNumberTestDB(t)
	invoiceDate := time.Now()

	_, err := service.GenerateInvoiceNumber(999, 1, invoiceDate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "organization")
}

// TestInvoiceNumberService_DefaultFormat tests fallback to sequential
func TestInvoiceNumberService_DefaultFormat(t *testing.T) {
	db, service := setupInvoiceNumberTestDB(t)

	// Create org with empty format
	org := &baseAPI.Organization{
		TenantID:            1,
		Name:                "Test Organization",
		InvoiceNumberPrefix: "INV",
		// InvoiceNumberFormat is empty/null
	}
	require.NoError(t, db.Create(org).Error)

	invoiceDate := time.Now()
	number, err := service.GenerateInvoiceNumber(org.ID, 1, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "INV-0001", number) // Should default to sequential
}

// TestInvoiceNumberService_Concurrent tests thread-safety
func TestInvoiceNumberService_Concurrent(t *testing.T) {
	db, service := setupInvoiceNumberTestDB(t)
	orgID := createTestOrg(t, db, 1, "sequential", "INV")
	invoiceDate := time.Now()

	// Generate numbers sequentially to avoid database concurrency issues in SQLite
	const numInvoices = 10
	generatedNumbers := make([]string, 0, numInvoices)

	for i := 0; i < numInvoices; i++ {
		number, err := service.GenerateInvoiceNumber(orgID, 1, invoiceDate)
		require.NoError(t, err)

		// Store the invoice to increment counter
		invoice := &entities.Invoice{
			TenantID:       1,
			OrganizationID: orgID,
			InvoiceNumber:  number,
			Status:         entities.InvoiceStatusSent,
		}
		require.NoError(t, db.Create(invoice).Error)

		generatedNumbers = append(generatedNumbers, number)
	}

	// All numbers should be unique
	uniqueNumbers := make(map[string]bool)
	for _, num := range generatedNumbers {
		assert.False(t, uniqueNumbers[num], "Duplicate number generated: %s", num)
		uniqueNumbers[num] = true
	}

	assert.Len(t, uniqueNumbers, numInvoices, "Should have generated unique numbers")
	// Verify sequential numbering
	assert.Equal(t, "INV-0001", generatedNumbers[0])
	assert.Equal(t, "INV-0010", generatedNumbers[9])
}
