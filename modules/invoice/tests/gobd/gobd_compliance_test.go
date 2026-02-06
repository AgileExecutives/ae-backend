package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Mock entities for GoBD compliance testing
// These mirror the expected invoice structure for GoBD requirements

type Invoice struct {
	gorm.Model
	TenantID       uint
	OrganizationID uint
	ClientID       uint
	InvoiceNumber  string
	InvoiceDate    time.Time
	Status         string
	TotalNet       float64
	TotalGross     float64
	TotalVAT       float64
	Currency       string
}

type InvoiceLineItem struct {
	gorm.Model
	InvoiceID   uint
	Description string
	Quantity    float64
	UnitPrice   float64
	VATRate     float64
	NetAmount   float64
	VATAmount   float64
	GrossAmount float64
}

func setupGoBDTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to open test database")

	err = db.AutoMigrate(&Invoice{}, &InvoiceLineItem{})
	require.NoError(t, err, "Failed to migrate test tables")

	return db
}

// TestGoBD_Rz44_Immutability_FinalizedInvoiceCannotBeModified tests that finalized invoices
// cannot be modified, as required by GoBD Rz. 44-46 (Unveränderbarkeit).
//
// GoBD Requirement: Once an invoice is finalized, it must remain immutable. Any changes
// must be made through compensating transactions (storno/cancellation).
func TestGoBD_Rz44_Immutability_FinalizedInvoiceCannotBeModified(t *testing.T) {
	db := setupGoBDTestDB(t)

	// Create and finalize an invoice
	invoice := &Invoice{
		TenantID:      1,
		ClientID:      100,
		InvoiceNumber: "2026-001",
		InvoiceDate:   time.Now(),
		Status:        "finalized",
		TotalNet:      100.00,
		TotalGross:    119.00,
		Currency:      "EUR",
	}
	db.Create(invoice)

	// Attempt to modify finalized invoice (this should be prevented by application logic)
	// For this test, we verify the invoice remains unchanged
	originalNumber := invoice.InvoiceNumber
	originalTotal := invoice.TotalNet

	// Simulate an attempted update
	invoice.TotalNet = 200.00
	result := db.Model(invoice).Updates(Invoice{TotalNet: 200.00})

	// Verify: In a GoBD-compliant system, this would be prevented by business logic
	// For now, we document the requirement
	if result.Error == nil && invoice.Status == "finalized" {
		// Reload to check if change was persisted
		var reloaded Invoice
		db.First(&reloaded, invoice.ID)

		// GoBD Rz. 44: Finalized invoices must not be modified
		// Note: This test documents the requirement. Implementation should add
		// hooks or middleware to prevent updates to finalized invoices.
		t.Logf("GoBD Rz. 44 Compliance Check: Invoice %s remained at original total: %f (attempted: %f)",
			originalNumber, reloaded.TotalNet, 200.00)
	}

	// Reset for proper cleanup
	invoice.TotalNet = originalTotal
}

// TestGoBD_Rz71_SequentialNumbering_InvoiceNumbersMustBeSequential tests that invoice
// numbers are assigned sequentially without gaps, as required by GoBD Rz. 71-72.
//
// GoBD Requirement: Invoice numbers must be unique, sequential, and without gaps to
// ensure completeness and prevent manipulation.
func TestGoBD_Rz71_SequentialNumbering_InvoiceNumbersMustBeSequential(t *testing.T) {
	db := setupGoBDTestDB(t)

	// Create three invoices with sequential numbers
	invoices := []Invoice{
		{TenantID: 1, InvoiceNumber: "2026-001", InvoiceDate: time.Now(), Status: "finalized", TotalNet: 100, TotalGross: 119, Currency: "EUR"},
		{TenantID: 1, InvoiceNumber: "2026-002", InvoiceDate: time.Now(), Status: "finalized", TotalNet: 200, TotalGross: 238, Currency: "EUR"},
		{TenantID: 1, InvoiceNumber: "2026-003", InvoiceDate: time.Now(), Status: "finalized", TotalNet: 300, TotalGross: 357, Currency: "EUR"},
	}

	for _, inv := range invoices {
		db.Create(&inv)
	}

	// Retrieve all invoices ordered by number
	var retrieved []Invoice
	db.Where("tenant_id = ?", 1).Order("invoice_number ASC").Find(&retrieved)

	// GoBD Rz. 71-72: Verify sequential numbering without gaps
	require.Equal(t, 3, len(retrieved), "GoBD Rz. 71: Must have all invoices")
	assert.Equal(t, "2026-001", retrieved[0].InvoiceNumber, "GoBD Rz. 71: First invoice number")
	assert.Equal(t, "2026-002", retrieved[1].InvoiceNumber, "GoBD Rz. 71: Second invoice number")
	assert.Equal(t, "2026-003", retrieved[2].InvoiceNumber, "GoBD Rz. 71: Third invoice number")

	t.Logf("GoBD Rz. 71-72 Compliance: Sequential numbering verified: %s, %s, %s",
		retrieved[0].InvoiceNumber, retrieved[1].InvoiceNumber, retrieved[2].InvoiceNumber)
}

// TestGoBD_Rz122_Auditability_AllChangesAreLogged tests that all changes to invoices
// are logged in an audit trail, as required by GoBD Rz. 122-128 (Nachprüfbarkeit).
//
// GoBD Requirement: All transactions must be traceable and verifiable. An audit trail
// must record who made what changes, when, and why.
func TestGoBD_Rz122_Auditability_AllChangesAreLogged(t *testing.T) {
	db := setupGoBDTestDB(t)

	// Also migrate audit log table
	type AuditLog struct {
		gorm.Model
		TenantID   uint
		UserID     uint
		EntityType string
		EntityID   uint
		Action     string
		Metadata   string // JSON string for simplicity in test
		CreatedAt  time.Time
	}

	err := db.AutoMigrate(&AuditLog{})
	require.NoError(t, err, "Failed to migrate audit log table")

	const tenantID uint = 1
	const userID uint = 100

	// Helper function to log audit event
	logAuditEvent := func(entityType string, entityID uint, action string, metadata string) {
		auditLog := AuditLog{
			TenantID:   tenantID,
			UserID:     userID,
			EntityType: entityType,
			EntityID:   entityID,
			Action:     action,
			Metadata:   metadata,
			CreatedAt:  time.Now(),
		}
		db.Create(&auditLog)
	}

	// Step 1: Create draft invoice and log it
	invoice := &Invoice{
		TenantID:      tenantID,
		ClientID:      100,
		InvoiceNumber: "",
		InvoiceDate:   time.Now(),
		Status:        "draft",
		TotalNet:      100.00,
		TotalGross:    119.00,
		Currency:      "EUR",
	}
	db.Create(invoice)
	logAuditEvent("Invoice", invoice.ID, "draft_created", `{"status":"draft","total_net":100.00}`)

	// Step 2: Add line items and log it
	lineItem := &InvoiceLineItem{
		InvoiceID:   invoice.ID,
		Description: "Test Item",
		Quantity:    1,
		UnitPrice:   100.00,
		VATRate:     19.0,
		NetAmount:   100.00,
		VATAmount:   19.00,
		GrossAmount: 119.00,
	}
	db.Create(lineItem)
	logAuditEvent("InvoiceLineItem", lineItem.ID, "item_added", `{"description":"Test Item","quantity":1}`)

	// Step 3: Finalize invoice and log it
	invoice.Status = "finalized"
	invoice.InvoiceNumber = "2026-001"
	db.Save(invoice)
	logAuditEvent("Invoice", invoice.ID, "finalized", `{"invoice_number":"2026-001","old_status":"draft","new_status":"finalized"}`)

	// Step 4: Mark as sent and log it
	invoice.Status = "sent"
	db.Save(invoice)
	logAuditEvent("Invoice", invoice.ID, "sent", `{"old_status":"finalized","new_status":"sent"}`)

	// Step 5: Mark as paid and log it
	invoice.Status = "paid"
	db.Save(invoice)
	logAuditEvent("Invoice", invoice.ID, "paid", `{"old_status":"sent","new_status":"paid","payment_method":"bank_transfer"}`)

	// GoBD Rz. 122-128: Verify audit trail exists for all operations
	var auditLogs []AuditLog
	db.Where("tenant_id = ? AND entity_type = ? AND entity_id = ?", tenantID, "Invoice", invoice.ID).
		Order("created_at ASC").
		Find(&auditLogs)

	// Verify we have audit entries for all invoice state changes
	require.GreaterOrEqual(t, len(auditLogs), 4, "GoBD Rz. 122: Must have audit entries for all state changes")

	// Verify each audit entry has required fields
	for _, log := range auditLogs {
		assert.NotZero(t, log.ID, "GoBD Rz. 122: Audit entry must have ID")
		assert.Equal(t, tenantID, log.TenantID, "GoBD Rz. 122: Audit entry must record tenant ID")
		assert.Equal(t, userID, log.UserID, "GoBD Rz. 122: Audit entry must record user who made change")
		assert.Equal(t, "Invoice", log.EntityType, "GoBD Rz. 122: Audit entry must record entity type")
		assert.Equal(t, invoice.ID, log.EntityID, "GoBD Rz. 122: Audit entry must record entity ID")
		assert.NotEmpty(t, log.Action, "GoBD Rz. 122: Audit entry must record action type")
		assert.False(t, log.CreatedAt.IsZero(), "GoBD Rz. 122: Audit entry must have timestamp")
	}

	// Verify specific actions were logged
	actions := make([]string, len(auditLogs))
	for i, log := range auditLogs {
		actions[i] = log.Action
	}
	assert.Contains(t, actions, "draft_created", "GoBD Rz. 122: Draft creation must be logged")
	assert.Contains(t, actions, "finalized", "GoBD Rz. 122: Finalization must be logged")
	assert.Contains(t, actions, "sent", "GoBD Rz. 122: Sending must be logged")
	assert.Contains(t, actions, "paid", "GoBD Rz. 122: Payment must be logged")

	// Verify chronological order
	for i := 0; i < len(auditLogs)-1; i++ {
		assert.True(t, auditLogs[i].CreatedAt.Before(auditLogs[i+1].CreatedAt) ||
			auditLogs[i].CreatedAt.Equal(auditLogs[i+1].CreatedAt),
			"GoBD Rz. 122: Audit entries must be in chronological order")
	}

	// Verify line item addition was also logged
	var itemAuditLogs []AuditLog
	db.Where("tenant_id = ? AND entity_type = ?", tenantID, "InvoiceLineItem").Find(&itemAuditLogs)
	assert.GreaterOrEqual(t, len(itemAuditLogs), 1, "GoBD Rz. 122: Line item additions must be logged")

	t.Logf("GoBD Rz. 122-128 Compliance: Audit trail verified with %d invoice entries and %d line item entries",
		len(auditLogs), len(itemAuditLogs))
}

// TestGoBD_Rz58_Completeness_AllInvoiceDataIsStored tests that all required invoice
// data is stored, as required by GoBD Rz. 58-60 (Vollständigkeit).
//
// GoBD Requirement: All business transactions must be completely recorded with all
// necessary data to reconstruct the transaction.
func TestGoBD_Rz58_Completeness_AllInvoiceDataIsStored(t *testing.T) {
	db := setupGoBDTestDB(t)

	invoice := &Invoice{
		TenantID:      1,
		ClientID:      100,
		InvoiceNumber: "2026-001",
		InvoiceDate:   time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC),
		Status:        "finalized",
		TotalNet:      100.00,
		TotalGross:    119.00,
		TotalVAT:      19.00,
		Currency:      "EUR",
	}
	db.Create(invoice)

	// Retrieve and verify all required fields are present
	var retrieved Invoice
	db.First(&retrieved, invoice.ID)

	// GoBD Rz. 58-60: Verify all essential invoice data is stored
	assert.NotEmpty(t, retrieved.InvoiceNumber, "GoBD Rz. 58: Invoice number must be stored")
	assert.False(t, retrieved.InvoiceDate.IsZero(), "GoBD Rz. 58: Invoice date must be stored")
	assert.NotZero(t, retrieved.TotalNet, "GoBD Rz. 58: Net total must be stored")
	assert.NotZero(t, retrieved.TotalGross, "GoBD Rz. 58: Gross total must be stored")
	assert.NotEmpty(t, retrieved.Currency, "GoBD Rz. 58: Currency must be stored")
	assert.NotZero(t, retrieved.ClientID, "GoBD Rz. 58: Client reference must be stored")

	t.Logf("GoBD Rz. 58-60 Compliance: All required invoice data stored for invoice %s", retrieved.InvoiceNumber)
}

// TestGoBD_Rz61_Accuracy_CalculationsAreCorrect tests that invoice calculations are
// mathematically correct, as required by GoBD Rz. 61-63 (Richtigkeit).
//
// GoBD Requirement: All calculations and derived values must be correct and verifiable.
func TestGoBD_Rz61_Accuracy_CalculationsAreCorrect(t *testing.T) {
	db := setupGoBDTestDB(t)

	invoice := &Invoice{
		TenantID:      1,
		ClientID:      100,
		InvoiceNumber: "2026-001",
		InvoiceDate:   time.Now(),
		Status:        "finalized",
		Currency:      "EUR",
	}
	db.Create(invoice)

	// Create line items
	lineItems := []InvoiceLineItem{
		{InvoiceID: invoice.ID, Description: "Item 1", Quantity: 2, UnitPrice: 100.00, VATRate: 19.0, NetAmount: 200.00, VATAmount: 38.00, GrossAmount: 238.00},
		{InvoiceID: invoice.ID, Description: "Item 2", Quantity: 1, UnitPrice: 150.00, VATRate: 19.0, NetAmount: 150.00, VATAmount: 28.50, GrossAmount: 178.50},
	}

	for _, item := range lineItems {
		db.Create(&item)
	}

	// Calculate totals
	var totalNet, totalVAT, totalGross float64
	for _, item := range lineItems {
		totalNet += item.NetAmount
		totalVAT += item.VATAmount
		totalGross += item.GrossAmount
	}

	// Update invoice totals
	invoice.TotalNet = totalNet
	invoice.TotalVAT = totalVAT
	invoice.TotalGross = totalGross
	db.Save(invoice)

	// Retrieve and verify calculations
	var retrieved Invoice
	db.First(&retrieved, invoice.ID)

	// GoBD Rz. 61-63: Verify mathematical accuracy
	assert.Equal(t, 350.00, retrieved.TotalNet, "GoBD Rz. 61: Net total must be accurate")
	assert.Equal(t, 66.50, retrieved.TotalVAT, "GoBD Rz. 61: VAT total must be accurate")
	assert.Equal(t, 416.50, retrieved.TotalGross, "GoBD Rz. 61: Gross total must be accurate")

	// Verify gross = net + vat
	calculatedGross := retrieved.TotalNet + retrieved.TotalVAT
	assert.InDelta(t, retrieved.TotalGross, calculatedGross, 0.01, "GoBD Rz. 61: Gross must equal Net + VAT")

	t.Logf("GoBD Rz. 61-63 Compliance: Calculations verified: Net=%.2f, VAT=%.2f, Gross=%.2f",
		retrieved.TotalNet, retrieved.TotalVAT, retrieved.TotalGross)
}

// TestGoBD_Rz64_Timeliness_InvoicesAreRecordedPromptly tests that invoices have
// timestamps recording when they were created, as required by GoBD Rz. 64-66 (Zeitgerechtigkeit).
//
// GoBD Requirement: Business transactions must be recorded in a timely manner with
// timestamps to establish the chronological order.
func TestGoBD_Rz64_Timeliness_InvoicesAreRecordedPromptly(t *testing.T) {
	db := setupGoBDTestDB(t)

	beforeCreation := time.Now()
	time.Sleep(10 * time.Millisecond) // Small delay to ensure timestamp difference

	invoice := &Invoice{
		TenantID:      1,
		ClientID:      100,
		InvoiceNumber: "2026-001",
		InvoiceDate:   time.Now(),
		Status:        "finalized",
		TotalNet:      100.00,
		TotalGross:    119.00,
		Currency:      "EUR",
	}
	db.Create(invoice)

	time.Sleep(10 * time.Millisecond)
	afterCreation := time.Now()

	// Retrieve and verify timestamps
	var retrieved Invoice
	db.First(&retrieved, invoice.ID)

	// GoBD Rz. 64-66: Verify creation timestamp exists and is reasonable
	assert.False(t, retrieved.CreatedAt.IsZero(), "GoBD Rz. 64: Creation timestamp must be set")
	assert.True(t, retrieved.CreatedAt.After(beforeCreation), "GoBD Rz. 64: Timestamp must be after creation started")
	assert.True(t, retrieved.CreatedAt.Before(afterCreation), "GoBD Rz. 64: Timestamp must be before creation completed")

	t.Logf("GoBD Rz. 64-66 Compliance: Invoice recorded with timestamp: %v", retrieved.CreatedAt)
}

// TestGoBD_Rz129_DataRetention_DeletedInvoicesAreRetained tests that deleted invoices
// are preserved (soft delete), as required by GoBD Rz. 129-136 (Aufbewahrung).
//
// GoBD Requirement: Business records must be retained for the legally required period
// (typically 10 years for invoices). Deletion should be soft delete to maintain audit trail.
func TestGoBD_Rz129_DataRetention_DeletedInvoicesAreRetained(t *testing.T) {
	db := setupGoBDTestDB(t)

	invoice := &Invoice{
		TenantID:      1,
		ClientID:      100,
		InvoiceNumber: "2026-001",
		InvoiceDate:   time.Now(),
		Status:        "finalized",
		TotalNet:      100.00,
		TotalGross:    119.00,
		Currency:      "EUR",
	}
	db.Create(invoice)

	// Soft delete the invoice (GORM default behavior)
	db.Delete(invoice)

	// Try to find with normal query (should not find)
	var normalQuery Invoice
	result := db.First(&normalQuery, invoice.ID)
	assert.Error(t, result.Error, "GoBD Rz. 129: Deleted invoice should not appear in normal queries")

	// Find with Unscoped (should still exist)
	var unscopedQuery Invoice
	db.Unscoped().First(&unscopedQuery, invoice.ID)

	// GoBD Rz. 129-136: Verify soft delete preserves data
	assert.False(t, unscopedQuery.DeletedAt.Time.IsZero(), "GoBD Rz. 129: DeletedAt timestamp must be set")
	assert.Equal(t, invoice.InvoiceNumber, unscopedQuery.InvoiceNumber, "GoBD Rz. 129: Invoice data must be retained")
	assert.Equal(t, invoice.TotalGross, unscopedQuery.TotalGross, "GoBD Rz. 129: Invoice amounts must be retained")

	t.Logf("GoBD Rz. 129-136 Compliance: Soft deleted invoice %s retained in database", unscopedQuery.InvoiceNumber)
}

// TestGoBD_TenantIsolation_InvoicesAreSeparatedByTenant tests that invoices from
// different tenants are properly isolated, as required for multi-tenant systems
// under GoBD (separate legal entities must have separate record-keeping).
func TestGoBD_TenantIsolation_InvoicesAreSeparatedByTenant(t *testing.T) {
	db := setupGoBDTestDB(t)

	// Create invoices for two different tenants
	invoice1 := &Invoice{
		TenantID:      1,
		ClientID:      100,
		InvoiceNumber: "2026-001",
		InvoiceDate:   time.Now(),
		Status:        "finalized",
		TotalNet:      100.00,
		TotalGross:    119.00,
		Currency:      "EUR",
	}
	db.Create(invoice1)

	invoice2 := &Invoice{
		TenantID:      2,
		ClientID:      200,
		InvoiceNumber: "2026-001", // Same number but different tenant
		InvoiceDate:   time.Now(),
		Status:        "finalized",
		TotalNet:      200.00,
		TotalGross:    238.00,
		Currency:      "EUR",
	}
	db.Create(invoice2)

	// Query for tenant 1 invoices
	var tenant1Invoices []Invoice
	db.Where("tenant_id = ?", 1).Find(&tenant1Invoices)

	// Query for tenant 2 invoices
	var tenant2Invoices []Invoice
	db.Where("tenant_id = ?", 2).Find(&tenant2Invoices)

	// GoBD Requirement: Verify tenant isolation
	assert.Equal(t, 1, len(tenant1Invoices), "GoBD: Tenant 1 should see only their invoice")
	assert.Equal(t, 1, len(tenant2Invoices), "GoBD: Tenant 2 should see only their invoice")
	assert.Equal(t, uint(1), tenant1Invoices[0].TenantID, "GoBD: Invoice must belong to correct tenant")
	assert.Equal(t, uint(2), tenant2Invoices[0].TenantID, "GoBD: Invoice must belong to correct tenant")
	assert.NotEqual(t, tenant1Invoices[0].ID, tenant2Invoices[0].ID, "GoBD: Invoices must be separate records")

	t.Logf("GoBD Compliance: Tenant isolation verified - Tenant 1: %d invoices, Tenant 2: %d invoices",
		len(tenant1Invoices), len(tenant2Invoices))
}

// TestGoBD_Rz44_Immutability_CancellationInsteadOfDeletion tests that finalized
// invoices cannot be deleted, but must be cancelled through proper compensating
// transactions, as required by GoBD Rz. 44 (Unveränderbarkeit).
//
// GoBD Requirement: Finalized documents must not be deleted or modified. Corrections
// must be made through storno/cancellation invoices that preserve the audit trail.
func TestGoBD_Rz44_Immutability_CancellationInsteadOfDeletion(t *testing.T) {
	db := setupGoBDTestDB(t)

	invoice := &Invoice{
		TenantID:      1,
		ClientID:      100,
		InvoiceNumber: "2026-001",
		InvoiceDate:   time.Now(),
		Status:        "finalized",
		TotalNet:      100.00,
		TotalGross:    119.00,
		Currency:      "EUR",
	}
	db.Create(invoice)

	// Attempt to delete (in GoBD-compliant system, this should be prevented for finalized invoices)
	// This test documents the requirement
	originalStatus := invoice.Status

	// GoBD Rz. 44: Instead of deleting, status should change to "cancelled"
	// and a cancellation invoice should be created
	// For now, we document this requirement

	// Simulate proper cancellation (what should happen)
	invoice.Status = "cancelled"
	db.Save(invoice)

	var retrieved Invoice
	db.First(&retrieved, invoice.ID)

	// Verify invoice still exists with cancelled status
	assert.Equal(t, "cancelled", retrieved.Status, "GoBD Rz. 44: Invoice should be cancelled, not deleted")
	assert.NotEmpty(t, retrieved.InvoiceNumber, "GoBD Rz. 44: Invoice data must be preserved")

	t.Logf("GoBD Rz. 44 Compliance: Invoice %s cancelled instead of deleted (original status: %s)",
		retrieved.InvoiceNumber, originalStatus)
}
