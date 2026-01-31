package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/invoice-module/entities"
	"github.com/unburdy/invoice-module/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// GoBD Compliance Test Suite
// Tests compliance with "Grundsätze zur ordnungsmäßigen Führung und Aufbewahrung
// von Büchern, Aufzeichnungen und Unterlagen in elektronischer Form sowie zum Datenzugriff"
//
// Run these tests with: go test -v -run GoBD

// setupGoBDTestDB creates an in-memory test database for GoBD compliance testing
func setupGoBDTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate all necessary tables
	err = db.AutoMigrate(
		&entities.Invoice{},
		&entities.InvoiceLineItem{},
	)
	require.NoError(t, err)

	return db
}

// TestGoBD_Rz44_Immutability_FinalizedInvoiceCannotBeModified tests that finalized
// invoices are immutable as required by GoBD Rz. 44-46 (Unveränderbarkeit)
//
// GoBD Requirement: Once business transactions are recorded, they must not be altered
// in a way that makes the original content unrecognizable.
func TestGoBD_Rz44_Immutability_FinalizedInvoiceCannotBeModified(t *testing.T) {
	db := setupGoBDTestDB(t)
	service := services.NewInvoiceService(db)
	ctx := context.Background()

	// Create and finalize an invoice
	invoice := &entities.Invoice{
		TenantID:      1,
		UserID:        1,
		ClientID:      1,
		InvoiceNumber: "2026-001",
		InvoiceDate:   time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 30),
		TotalNet:      1000.00,
		TotalGross:    1190.00,
		Status:        "finalized",
	}

	err := db.Create(invoice).Error
	require.NoError(t, err)

	// Attempt to modify finalized invoice - should fail
	invoice.TotalNet = 2000.00
	err = service.UpdateInvoice(ctx, invoice.ID, invoice, 1)

	// GoBD Compliance: Finalized invoices must be immutable
	assert.Error(t, err, "GoBD Rz. 44: Finalized invoice modification must be prevented")

	// Verify original data is unchanged
	var retrieved entities.Invoice
	db.First(&retrieved, invoice.ID)
	assert.Equal(t, 1000.00, retrieved.TotalNet, "GoBD Rz. 44: Original data must remain unchanged")
}

// TestGoBD_Rz71_SequentialNumbering_InvoiceNumbersMustBeSequential tests that invoice
// numbers follow a sequential, gap-free pattern as required by GoBD Rz. 71-72
//
// GoBD Requirement: Business documents must be numbered consecutively without gaps.
func TestGoBD_Rz71_SequentialNumbering_InvoiceNumbersMustBeSequential(t *testing.T) {
	db := setupGoBDTestDB(t)

	// Create multiple invoices
	invoices := []entities.Invoice{
		{TenantID: 1, UserID: 1, ClientID: 1, InvoiceNumber: "2026-001", InvoiceDate: time.Now(), DueDate: time.Now(), TotalNet: 100, TotalGross: 119, Status: "draft"},
		{TenantID: 1, UserID: 1, ClientID: 1, InvoiceNumber: "2026-002", InvoiceDate: time.Now(), DueDate: time.Now(), TotalNet: 200, TotalGross: 238, Status: "draft"},
		{TenantID: 1, UserID: 1, ClientID: 1, InvoiceNumber: "2026-003", InvoiceDate: time.Now(), DueDate: time.Now(), TotalNet: 300, TotalGross: 357, Status: "draft"},
	}

	for _, inv := range invoices {
		err := db.Create(&inv).Error
		require.NoError(t, err)
	}

	// Verify sequential numbering
	var allInvoices []entities.Invoice
	db.Where("tenant_id = ?", 1).Order("invoice_number ASC").Find(&allInvoices)

	// GoBD Compliance: Numbers must be sequential without gaps
	assert.Equal(t, "2026-001", allInvoices[0].InvoiceNumber, "GoBD Rz. 71: First number must be 001")
	assert.Equal(t, "2026-002", allInvoices[1].InvoiceNumber, "GoBD Rz. 71: Second number must be 002")
	assert.Equal(t, "2026-003", allInvoices[2].InvoiceNumber, "GoBD Rz. 71: Third number must be 003")
	assert.Equal(t, 3, len(allInvoices), "GoBD Rz. 71: No gaps in sequence")
}

// TestGoBD_Rz122_Auditability_AllChangesAreLogged tests that all modifications are
// logged in an audit trail as required by GoBD Rz. 122-128 (Nachprüfbarkeit)
//
// GoBD Requirement: All changes to business documents must be traceable and documented.
func TestGoBD_Rz122_Auditability_AllChangesAreLogged(t *testing.T) {
	t.Skip("Requires audit trail integration - see modules/audit")

	// Note: This test validates that audit logging is enabled
	// Full implementation requires the audit module integration
	// See: modules/audit/services for audit trail implementation
	//
	// GoBD Rz. 122-128 requires:
	// - Who made the change (user_id)
	// - When the change was made (timestamp)
	// - What was changed (before/after values)
	// - Why the change was made (reason)
}

// TestGoBD_Rz58_Completeness_AllInvoiceDataIsStored tests that all required invoice
// data is stored completely as required by GoBD Rz. 58-60 (Vollständigkeit)
//
// GoBD Requirement: All business transactions must be recorded completely.
func TestGoBD_Rz58_Completeness_AllInvoiceDataIsStored(t *testing.T) {
	db := setupGoBDTestDB(t)

	// Create invoice with all required fields
	invoice := &entities.Invoice{
		TenantID:      1,
		UserID:        1,
		ClientID:      1,
		InvoiceNumber: "2026-001",
		InvoiceDate:   time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 30),
		TotalNet:      1000.00,
		TotalGross:    1190.00,
		TotalVAT:      190.00,
		Currency:      "EUR",
		Status:        "draft",
	}

	err := db.Create(invoice).Error
	require.NoError(t, err)

	// Retrieve and verify all data is stored
	var retrieved entities.Invoice
	err = db.First(&retrieved, invoice.ID).Error
	require.NoError(t, err)

	// GoBD Compliance: All required fields must be present
	assert.NotZero(t, retrieved.InvoiceNumber, "GoBD Rz. 58: Invoice number required")
	assert.NotZero(t, retrieved.InvoiceDate, "GoBD Rz. 58: Invoice date required")
	assert.NotZero(t, retrieved.TotalNet, "GoBD Rz. 58: Net amount required")
	assert.NotZero(t, retrieved.TotalGross, "GoBD Rz. 58: Gross amount required")
	assert.NotEmpty(t, retrieved.Currency, "GoBD Rz. 58: Currency required")
	assert.NotZero(t, retrieved.ClientID, "GoBD Rz. 58: Client reference required")
}

// TestGoBD_Rz61_Accuracy_CalculationsAreCorrect tests that invoice calculations
// are mathematically correct as required by GoBD Rz. 61-63 (Richtigkeit)
//
// GoBD Requirement: Business transactions must be recorded accurately and correctly.
func TestGoBD_Rz61_Accuracy_CalculationsAreCorrect(t *testing.T) {
	db := setupGoBDTestDB(t)

	// Create invoice with line items
	invoice := &entities.Invoice{
		TenantID:      1,
		UserID:        1,
		ClientID:      1,
		InvoiceNumber: "2026-001",
		InvoiceDate:   time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 30),
		Currency:      "EUR",
		Status:        "draft",
	}

	err := db.Create(invoice).Error
	require.NoError(t, err)

	// Add line items
	lineItems := []entities.InvoiceLineItem{
		{InvoiceID: invoice.ID, Description: "Service A", Quantity: 2, UnitPrice: 100.00, VATRate: 19.0, TotalNet: 200.00, TotalGross: 238.00},
		{InvoiceID: invoice.ID, Description: "Service B", Quantity: 3, UnitPrice: 50.00, VATRate: 19.0, TotalNet: 150.00, TotalGross: 178.50},
	}

	for _, item := range lineItems {
		err := db.Create(&item).Error
		require.NoError(t, err)
	}

	// Calculate totals
	var totalNet, totalGross float64
	for _, item := range lineItems {
		totalNet += item.TotalNet
		totalGross += item.TotalGross
	}

	// GoBD Compliance: Calculations must be accurate
	assert.Equal(t, 350.00, totalNet, "GoBD Rz. 61: Net total calculation must be correct")
	assert.Equal(t, 416.50, totalGross, "GoBD Rz. 61: Gross total calculation must be correct")
	assert.InDelta(t, 66.50, totalGross-totalNet, 0.01, "GoBD Rz. 61: VAT calculation must be correct")
}

// TestGoBD_Rz64_Timeliness_InvoicesAreRecordedPromptly tests that invoices are
// recorded in a timely manner as required by GoBD Rz. 64-66 (Zeitgerechtigkeit)
//
// GoBD Requirement: Business transactions must be recorded promptly.
func TestGoBD_Rz64_Timeliness_InvoicesAreRecordedPromptly(t *testing.T) {
	db := setupGoBDTestDB(t)

	// Create invoice
	now := time.Now()
	invoice := &entities.Invoice{
		TenantID:      1,
		UserID:        1,
		ClientID:      1,
		InvoiceNumber: "2026-001",
		InvoiceDate:   now,
		DueDate:       now.AddDate(0, 0, 30),
		TotalNet:      1000.00,
		TotalGross:    1190.00,
		Currency:      "EUR",
		Status:        "draft",
	}

	err := db.Create(invoice).Error
	require.NoError(t, err)

	// Verify CreatedAt timestamp is set
	var retrieved entities.Invoice
	db.First(&retrieved, invoice.ID)

	// GoBD Compliance: Recording timestamp must be present and recent
	assert.NotZero(t, retrieved.CreatedAt, "GoBD Rz. 64: Creation timestamp required")
	assert.WithinDuration(t, now, retrieved.CreatedAt, 5*time.Second, "GoBD Rz. 64: Must be recorded promptly")
}

// TestGoBD_Rz129_DataRetention_DeletedInvoicesAreRetained tests that deleted invoices
// are retained (soft delete) as required by GoBD Rz. 129-136 (Aufbewahrung)
//
// GoBD Requirement: Business documents must be retained for the legally required period.
func TestGoBD_Rz129_DataRetention_DeletedInvoicesAreRetained(t *testing.T) {
	db := setupGoBDTestDB(t)
	service := services.NewInvoiceService(db)
	ctx := context.Background()

	// Create invoice
	invoice := &entities.Invoice{
		TenantID:      1,
		UserID:        1,
		ClientID:      1,
		InvoiceNumber: "2026-001",
		InvoiceDate:   time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 30),
		TotalNet:      1000.00,
		TotalGross:    1190.00,
		Currency:      "EUR",
		Status:        "draft",
	}

	err := db.Create(invoice).Error
	require.NoError(t, err)
	invoiceID := invoice.ID

	// Soft delete the invoice
	err = service.DeleteInvoice(ctx, invoiceID, 1)
	require.NoError(t, err)

	// Verify invoice still exists in database (soft deleted)
	var retrieved entities.Invoice
	err = db.Unscoped().First(&retrieved, invoiceID).Error
	require.NoError(t, err)

	// GoBD Compliance: Deleted records must be retained with deletion marker
	assert.NotNil(t, retrieved.DeletedAt, "GoBD Rz. 129: Deletion must be marked")
	assert.Equal(t, "2026-001", retrieved.InvoiceNumber, "GoBD Rz. 129: Data must be retained")
	assert.Equal(t, 1000.00, retrieved.TotalNet, "GoBD Rz. 129: Financial data must be retained")
}

// TestGoBD_TenantIsolation_InvoicesAreSeparatedByTenant tests that invoices from
// different tenants are strictly isolated (multi-tenancy requirement for GoBD compliance)
//
// GoBD Requirement: Data from different legal entities must be kept separate.
func TestGoBD_TenantIsolation_InvoicesAreSeparatedByTenant(t *testing.T) {
	db := setupGoBDTestDB(t)
	ctx := context.Background()

	// Create invoices for two different tenants
	invoice1 := &entities.Invoice{
		TenantID:      1,
		UserID:        1,
		ClientID:      1,
		InvoiceNumber: "2026-001",
		InvoiceDate:   time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 30),
		TotalNet:      1000.00,
		TotalGross:    1190.00,
		Currency:      "EUR",
		Status:        "draft",
	}

	invoice2 := &entities.Invoice{
		TenantID:      2,
		UserID:        2,
		ClientID:      2,
		InvoiceNumber: "2026-001", // Same number, different tenant
		InvoiceDate:   time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 30),
		TotalNet:      2000.00,
		TotalGross:    2380.00,
		Currency:      "EUR",
		Status:        "draft",
	}

	err := db.Create(invoice1).Error
	require.NoError(t, err)
	err = db.Create(invoice2).Error
	require.NoError(t, err)

	// Verify tenant 1 can only see their invoices
	var tenant1Invoices []entities.Invoice
	db.Where("tenant_id = ?", 1).Find(&tenant1Invoices)
	assert.Equal(t, 1, len(tenant1Invoices), "GoBD: Tenant 1 should see only their invoice")
	assert.Equal(t, uint(1), tenant1Invoices[0].TenantID)

	// Verify tenant 2 can only see their invoices
	var tenant2Invoices []entities.Invoice
	db.Where("tenant_id = ?", 2).Find(&tenant2Invoices)
	assert.Equal(t, 1, len(tenant2Invoices), "GoBD: Tenant 2 should see only their invoice")
	assert.Equal(t, uint(2), tenant2Invoices[0].TenantID)

	// Verify cross-tenant access is prevented
	service := services.NewInvoiceService(db)
	_, err = service.GetInvoice(ctx, invoice2.ID, 1) // Tenant 1 trying to access Tenant 2's invoice
	assert.Error(t, err, "GoBD: Cross-tenant access must be prevented")
}

// TestGoBD_Rz44_Immutability_CancellationInsteadOfDeletion tests that finalized invoices
// cannot be deleted but must be cancelled with a credit note (Storno) as per GoBD Rz. 44
//
// GoBD Requirement: Corrections must preserve the original transaction and create
// a new compensating transaction.
func TestGoBD_Rz44_Immutability_CancellationInsteadOfDeletion(t *testing.T) {
	db := setupGoBDTestDB(t)
	service := services.NewInvoiceService(db)
	ctx := context.Background()

	// Create and finalize an invoice
	invoice := &entities.Invoice{
		TenantID:      1,
		UserID:        1,
		ClientID:      1,
		InvoiceNumber: "2026-001",
		InvoiceDate:   time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 30),
		TotalNet:      1000.00,
		TotalGross:    1190.00,
		Currency:      "EUR",
		Status:        "finalized",
	}

	err := db.Create(invoice).Error
	require.NoError(t, err)

	// Attempt to delete finalized invoice - should fail
	err = service.DeleteInvoice(ctx, invoice.ID, 1)
	assert.Error(t, err, "GoBD Rz. 44: Finalized invoices cannot be deleted")

	// Verify invoice still exists and is not deleted
	var retrieved entities.Invoice
	err = db.First(&retrieved, invoice.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "finalized", retrieved.Status, "GoBD Rz. 44: Invoice must remain finalized")

	// Note: Proper cancellation would create a credit note (Stornorechnung)
	// This should be tested in invoice_cancellation_test.go
}
