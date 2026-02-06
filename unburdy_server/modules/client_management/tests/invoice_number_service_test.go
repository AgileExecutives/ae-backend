package tests

import (
	"encoding/json"
	"testing"
	"time"

	baseAPI "github.com/ae-base-server/api"
	settingsEntities "github.com/ae-base-server/pkg/settings/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"github.com/unburdy/unburdy-server-api/modules/client_management/services"
	"github.com/unburdy/unburdy-server-api/modules/client_management/settings"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupInvoiceNumberTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto migrate tables
	err = db.AutoMigrate(
		&entities.Invoice{},
		&baseAPI.Organization{},
		&settingsEntities.SettingDefinition{},
		&settingsEntities.Setting{},
	)
	require.NoError(t, err)

	return db
}

func createInvoiceNumberSettings(t *testing.T, db *gorm.DB, tenantID uint, format, prefix string) {
	// Create invoice number settings using the proper JSONB structure
	invoiceNumberData := settings.InvoiceNumberSettings{
		Format: format,
		Prefix: prefix,
	}

	dataJSON, err := json.Marshal(invoiceNumberData)
	require.NoError(t, err)

	setting := &settingsEntities.Setting{
		TenantID: tenantID,
		Domain:   settings.DomainBilling,
		Key:      settings.KeyInvoiceNumber,
		Version:  1,
		Data:     datatypes.JSON(dataJSON),
	}
	require.NoError(t, db.Create(setting).Error)
}

func TestInvoiceNumberService_Sequential_NoPrefix(t *testing.T) {
	db := setupInvoiceNumberTestDB(t)
	tenantID := uint(1)
	orgID := uint(1)

	// Create settings
	createInvoiceNumberSettings(t, db, tenantID, "sequential", "")

	service := services.NewInvoiceNumberService(db)

	// Test first invoice number
	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	num1, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "1", num1)

	// Create invoice with that number
	invoice1 := &entities.Invoice{
		TenantID:       tenantID,
		UserID:         1,
		OrganizationID: orgID,
		InvoiceNumber:  num1,
		InvoiceDate:    invoiceDate,
		Status:         entities.InvoiceStatusFinalized,
	}
	require.NoError(t, db.Create(invoice1).Error)

	// Test second invoice number
	num2, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "2", num2)

	// Create invoice with that number
	invoice2 := &entities.Invoice{
		TenantID:       tenantID,
		UserID:         1,
		OrganizationID: orgID,
		InvoiceNumber:  num2,
		InvoiceDate:    invoiceDate,
		Status:         entities.InvoiceStatusFinalized,
	}
	require.NoError(t, db.Create(invoice2).Error)

	// Test third invoice number
	num3, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "3", num3)
}

func TestInvoiceNumberService_Sequential_WithPrefix(t *testing.T) {
	db := setupInvoiceNumberTestDB(t)
	tenantID := uint(1)
	orgID := uint(1)

	// Create settings with prefix
	createInvoiceNumberSettings(t, db, tenantID, "sequential", "INV")

	service := services.NewInvoiceNumberService(db)

	// Test first invoice number
	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	num1, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "INV-0001", num1)

	// Create invoice
	invoice1 := &entities.Invoice{
		TenantID:       tenantID,
		UserID:         1,
		OrganizationID: orgID,
		InvoiceNumber:  num1,
		InvoiceDate:    invoiceDate,
		Status:         entities.InvoiceStatusFinalized,
	}
	require.NoError(t, db.Create(invoice1).Error)

	// Test second invoice number
	num2, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "INV-0002", num2)

	// Create invoice
	invoice2 := &entities.Invoice{
		TenantID:       tenantID,
		UserID:         1,
		OrganizationID: orgID,
		InvoiceNumber:  num2,
		InvoiceDate:    invoiceDate,
		Status:         entities.InvoiceStatusFinalized,
	}
	require.NoError(t, db.Create(invoice2).Error)

	// Test third invoice number
	num3, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "INV-0003", num3)
}

func TestInvoiceNumberService_YearPrefix_NoPrefix(t *testing.T) {
	db := setupInvoiceNumberTestDB(t)
	tenantID := uint(1)
	orgID := uint(1)

	// Create settings
	createInvoiceNumberSettings(t, db, tenantID, "year_prefix", "")

	service := services.NewInvoiceNumberService(db)

	// Test first invoice number in 2026
	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	num1, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "2026-0001", num1)

	// Create invoice
	invoice1 := &entities.Invoice{
		TenantID:       tenantID,
		UserID:         1,
		OrganizationID: orgID,
		InvoiceNumber:  num1,
		InvoiceDate:    invoiceDate,
		Status:         entities.InvoiceStatusFinalized,
	}
	require.NoError(t, db.Create(invoice1).Error)

	// Test second invoice number in 2026
	num2, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "2026-0002", num2)

	// Create invoice
	invoice2 := &entities.Invoice{
		TenantID:       tenantID,
		UserID:         1,
		OrganizationID: orgID,
		InvoiceNumber:  num2,
		InvoiceDate:    invoiceDate,
		Status:         entities.InvoiceStatusFinalized,
	}
	require.NoError(t, db.Create(invoice2).Error)

	// Test first invoice number in 2027 (should reset)
	invoiceDate2027 := time.Date(2027, 1, 15, 0, 0, 0, 0, time.UTC)
	num3, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate2027)
	require.NoError(t, err)
	assert.Equal(t, "2027-0001", num3)
}

func TestInvoiceNumberService_YearPrefix_WithPrefix(t *testing.T) {
	db := setupInvoiceNumberTestDB(t)
	tenantID := uint(1)
	orgID := uint(1)

	// Create settings with prefix
	createInvoiceNumberSettings(t, db, tenantID, "year_prefix", "INV")

	service := services.NewInvoiceNumberService(db)

	// Test first invoice number
	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	num1, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "INV-2026-0001", num1)

	// Create invoice
	invoice1 := &entities.Invoice{
		TenantID:       tenantID,
		UserID:         1,
		OrganizationID: orgID,
		InvoiceNumber:  num1,
		InvoiceDate:    invoiceDate,
		Status:         entities.InvoiceStatusFinalized,
	}
	require.NoError(t, db.Create(invoice1).Error)

	// Test second invoice number
	num2, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "INV-2026-0002", num2)
}

func TestInvoiceNumberService_YearMonthPrefix_NoPrefix(t *testing.T) {
	db := setupInvoiceNumberTestDB(t)
	tenantID := uint(1)
	orgID := uint(1)

	// Create settings
	createInvoiceNumberSettings(t, db, tenantID, "year_month_prefix", "")

	service := services.NewInvoiceNumberService(db)

	// Test first invoice number in January 2026
	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	num1, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "2026-01-0001", num1)

	// Create invoice
	invoice1 := &entities.Invoice{
		TenantID:       tenantID,
		UserID:         1,
		OrganizationID: orgID,
		InvoiceNumber:  num1,
		InvoiceDate:    invoiceDate,
		Status:         entities.InvoiceStatusFinalized,
	}
	require.NoError(t, db.Create(invoice1).Error)

	// Test second invoice number in January 2026
	num2, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "2026-01-0002", num2)

	// Create invoice
	invoice2 := &entities.Invoice{
		TenantID:       tenantID,
		UserID:         1,
		OrganizationID: orgID,
		InvoiceNumber:  num2,
		InvoiceDate:    invoiceDate,
		Status:         entities.InvoiceStatusFinalized,
	}
	require.NoError(t, db.Create(invoice2).Error)

	// Test first invoice number in February 2026 (should reset)
	invoiceDateFeb := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	num3, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDateFeb)
	require.NoError(t, err)
	assert.Equal(t, "2026-02-0001", num3)
}

func TestInvoiceNumberService_YearMonthPrefix_WithPrefix(t *testing.T) {
	db := setupInvoiceNumberTestDB(t)
	tenantID := uint(1)
	orgID := uint(1)

	// Create settings with prefix
	createInvoiceNumberSettings(t, db, tenantID, "year_month_prefix", "INV")

	service := services.NewInvoiceNumberService(db)

	// Test first invoice number
	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	num1, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "INV-2026-01-0001", num1)

	// Create invoice
	invoice1 := &entities.Invoice{
		TenantID:       tenantID,
		UserID:         1,
		OrganizationID: orgID,
		InvoiceNumber:  num1,
		InvoiceDate:    invoiceDate,
		Status:         entities.InvoiceStatusFinalized,
	}
	require.NoError(t, db.Create(invoice1).Error)

	// Test second invoice number
	num2, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "INV-2026-01-0002", num2)
}

func TestInvoiceNumberService_Sequential_ExcludesYearBasedFormats(t *testing.T) {
	db := setupInvoiceNumberTestDB(t)
	tenantID := uint(1)
	orgID := uint(1)

	// Create settings for sequential format (no prefix)
	createInvoiceNumberSettings(t, db, tenantID, "sequential", "")

	service := services.NewInvoiceNumberService(db)
	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create an invoice with year-based format (should be excluded from sequential counting)
	yearBasedInvoice := &entities.Invoice{
		TenantID:       tenantID,
		UserID:         1,
		OrganizationID: orgID,
		InvoiceNumber:  "2026-0001",
		InvoiceDate:    invoiceDate,
		Status:         entities.InvoiceStatusFinalized,
	}
	require.NoError(t, db.Create(yearBasedInvoice).Error)

	// Generate sequential number (should start at 1, not consider "2026-0001")
	num1, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "1", num1, "Sequential format should exclude year-based invoice numbers")

	// Create invoice with sequential number
	invoice1 := &entities.Invoice{
		TenantID:       tenantID,
		UserID:         1,
		OrganizationID: orgID,
		InvoiceNumber:  num1,
		InvoiceDate:    invoiceDate,
		Status:         entities.InvoiceStatusFinalized,
	}
	require.NoError(t, db.Create(invoice1).Error)

	// Next sequential should be 2
	num2, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "2", num2)
}

func TestInvoiceNumberService_Sequential_ExcludesMultipleYearFormats(t *testing.T) {
	db := setupInvoiceNumberTestDB(t)
	tenantID := uint(1)
	orgID := uint(1)

	// Create settings for sequential format (no prefix)
	createInvoiceNumberSettings(t, db, tenantID, "sequential", "")

	service := services.NewInvoiceNumberService(db)
	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create invoices with various year-based formats (all should be excluded)
	yearFormats := []string{
		"2025-0001",
		"2026-0001",
		"2027-0001",
		"2024-0100",
	}

	for _, format := range yearFormats {
		invoice := &entities.Invoice{
			TenantID:       tenantID,
			UserID:         1,
			OrganizationID: orgID,
			InvoiceNumber:  format,
			InvoiceDate:    invoiceDate,
			Status:         entities.InvoiceStatusFinalized,
		}
		require.NoError(t, db.Create(invoice).Error)
	}

	// Generate sequential number (should start at 1, ignoring all year-based numbers)
	num1, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "1", num1, "Sequential format should exclude all year-based invoice numbers")
}

func TestInvoiceNumberService_MultipleOrganizations(t *testing.T) {
	db := setupInvoiceNumberTestDB(t)
	tenant1ID := uint(1)
	tenant2ID := uint(2)
	org1ID := uint(1)
	org2ID := uint(2)

	// Create settings for both tenants with different prefixes to avoid conflicts
	createInvoiceNumberSettings(t, db, tenant1ID, "sequential", "ORG1")
	createInvoiceNumberSettings(t, db, tenant2ID, "sequential", "ORG2")

	service := services.NewInvoiceNumberService(db)
	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create invoice for org 1 tenant 1
	num1_org1, err := service.GenerateInvoiceNumber(org1ID, tenant1ID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "ORG1-0001", num1_org1)

	invoice1_org1 := &entities.Invoice{
		TenantID:       tenant1ID,
		UserID:         1,
		OrganizationID: org1ID,
		InvoiceNumber:  num1_org1,
		InvoiceDate:    invoiceDate,
		Status:         entities.InvoiceStatusFinalized,
	}
	require.NoError(t, db.Create(invoice1_org1).Error)

	// Create invoice for org 2 tenant 2 (should also start at 1)
	num1_org2, err := service.GenerateInvoiceNumber(org2ID, tenant2ID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "ORG2-0001", num1_org2)

	invoice1_org2 := &entities.Invoice{
		TenantID:       tenant2ID,
		UserID:         1,
		OrganizationID: org2ID,
		InvoiceNumber:  num1_org2,
		InvoiceDate:    invoiceDate,
		Status:         entities.InvoiceStatusFinalized,
	}
	require.NoError(t, db.Create(invoice1_org2).Error)

	// Next invoice for org 1 should be 2
	num2_org1, err := service.GenerateInvoiceNumber(org1ID, tenant1ID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "ORG1-0002", num2_org1)

	// Next invoice for org 2 should also be 2
	num2_org2, err := service.GenerateInvoiceNumber(org2ID, tenant2ID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "ORG2-0002", num2_org2)
}

func TestInvoiceNumberService_DefaultFormat(t *testing.T) {
	db := setupInvoiceNumberTestDB(t)
	tenantID := uint(1)
	orgID := uint(1)

	// Don't create any settings - should use defaults (sequential with no prefix)
	service := services.NewInvoiceNumberService(db)
	invoiceDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// Should default to sequential format
	num1, err := service.GenerateInvoiceNumber(orgID, tenantID, invoiceDate)
	require.NoError(t, err)
	assert.Equal(t, "1", num1, "No settings should default to sequential format")
}
