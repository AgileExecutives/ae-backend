package services

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/invoice-module/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestInvoiceNumberConcurrency tests concurrent invoice finalization
// to ensure no duplicate invoice numbers are generated
func TestInvoiceNumberConcurrency(t *testing.T) {
	// Skip if using SQLite (no real advisory locks)
	t.Skip("SQLite doesn't support PostgreSQL advisory locks - run against real Postgres")

	db := setupTestDB(t)
	service := &InvoiceService{db: db}
	ctx := context.Background()

	// Create 10 draft invoices
	invoices := make([]*entities.Invoice, 10)
	for i := 0; i < 10; i++ {
		invoices[i] = createTestInvoice(t, db, entities.InvoiceStatusDraft, true)
	}

	// Finalize all invoices concurrently
	var wg sync.WaitGroup
	errors := make(chan error, 10)
	results := make(chan string, 10)

	for _, inv := range invoices {
		wg.Add(1)
		go func(invoice *entities.Invoice) {
			defer wg.Done()

			result, err := service.FinalizeInvoice(ctx, invoice.TenantID, invoice.ID, invoice.UserID)
			if err != nil {
				errors <- err
				return
			}
			results <- result.InvoiceNumber
		}(inv)
	}

	wg.Wait()
	close(errors)















































































































































































































}	wg.Wait()	}		}(i)			service.FinalizeInvoice(ctx, invoices[idx].TenantID, invoices[idx].ID, invoices[idx].UserID)			defer wg.Done()		go func(idx int) {		wg.Add(1)	for i := 0; i < b.N; i++ {	var wg sync.WaitGroup	b.ResetTimer()	}		invoices[i] = createTestInvoice(&testing.T{}, db, entities.InvoiceStatusDraft, true)	for i := 0; i < b.N; i++ {	invoices := make([]*entities.Invoice, b.N)	// Create draft invoices	ctx := context.Background()	service := &InvoiceService{db: db}	db := setupTestDB(&testing.T{})func BenchmarkConcurrentFinalize(b *testing.B) {// BenchmarkConcurrentFinalize benchmarks concurrent finalization}	}		service.FinalizeInvoice(ctx, invoices[i].TenantID, invoices[i].ID, invoices[i].UserID)	for i := 0; i < b.N; i++ {	b.ResetTimer()	}		invoices[i] = createTestInvoice(&testing.T{}, db, entities.InvoiceStatusDraft, true)	for i := 0; i < b.N; i++ {	invoices := make([]*entities.Invoice, b.N)	// Create draft invoices	ctx := context.Background()	service := &InvoiceService{db: db}	db := setupTestDB(&testing.T{})func BenchmarkFinalizeInvoice(b *testing.B) {// BenchmarkFinalizeInvoice benchmarks the finalization performance}	assert.NotEmpty(t, result2.InvoiceNumber)	assert.NotEmpty(t, result1.InvoiceNumber)	// The exact behavior depends on whether invoice numbers are scoped by organization	// Both should get number "1" (or similar) since they're in different organizations	require.NoError(t, err)	result2, err := service.FinalizeInvoice(ctx, inv2.TenantID, inv2.ID, inv2.UserID)	require.NoError(t, err)	result1, err := service.FinalizeInvoice(ctx, inv1.TenantID, inv1.ID, inv1.UserID)	// Finalize both	db.Save(inv2)	inv2.OrganizationID = 2	inv2 := createTestInvoice(t, db, entities.InvoiceStatusDraft, true)	db.Save(inv1)	inv1.OrganizationID = 1	inv1 := createTestInvoice(t, db, entities.InvoiceStatusDraft, true)	// Create invoices for different organizations	ctx := context.Background()	service := &InvoiceService{db: db}	db := setupTestDB(t)func TestInvoiceNumberByOrganization(t *testing.T) {// TestInvoiceNumberByOrganization tests that different organizations get separate number sequences}	// Additional assertions would depend on the number format	assert.NotEqual(t, num1, num3)	// (no gap from the failed second invoice)	// The third invoice should get the next number after the first	num3 := result3.InvoiceNumber	require.NoError(t, err)	result3, err := service.FinalizeInvoice(ctx, inv3.TenantID, inv3.ID, inv3.UserID)	inv3 := createTestInvoice(t, db, entities.InvoiceStatusDraft, true)	// Create third invoice and finalize successfully	require.Error(t, err) // Should fail validation	_, err = service.FinalizeInvoice(ctx, inv2.TenantID, inv2.ID, inv2.UserID)	inv2 := createTestInvoice(t, db, entities.InvoiceStatusDraft, false)	// Create second invoice without items (will fail validation)	num1 := result1.InvoiceNumber	require.NoError(t, err)	result1, err := service.FinalizeInvoice(ctx, inv1.TenantID, inv1.ID, inv1.UserID)	inv1 := createTestInvoice(t, db, entities.InvoiceStatusDraft, true)	// Create first invoice and finalize successfully	ctx := context.Background()	service := &InvoiceService{db: db}	db := setupTestDB(t)func TestInvoiceNumberRollback(t *testing.T) {// TestInvoiceNumberRollback tests that failed finalization doesn't consume invoice numbers}	}		seen[num] = true		assert.False(t, seen[num], "Duplicate invoice number: %s", num)	for _, num := range numbers {	seen := make(map[string]bool)	// Verify all numbers are unique	}		numbers[i] = result.InvoiceNumber		require.NoError(t, err)		result, err := service.FinalizeInvoice(ctx, invoice.TenantID, invoice.ID, invoice.UserID)		invoice := createTestInvoice(t, db, entities.InvoiceStatusDraft, true)	for i := 0; i < 5; i++ {	numbers := make([]string, 5)	// Create and finalize 5 invoices sequentially	ctx := context.Background()	service := &InvoiceService{db: db}	db := setupTestDB(t)func TestInvoiceNumberSequencing(t *testing.T) {// TestInvoiceNumberSequencing tests that invoice numbers are generated in sequence}	}		})			assert.NotEqual(t, invoice.InvoiceNumber, result.InvoiceNumber)			assert.NotEmpty(t, result.InvoiceNumber)			// For now, just verify we got a non-empty number			// The actual pattern matching would depend on the invoice_number service			require.NoError(t, err)			result, err := service.FinalizeInvoice(ctx, invoice.TenantID, invoice.ID, invoice.UserID)			invoice := createTestInvoice(t, db, entities.InvoiceStatusDraft, true)			// For this test, we'll mock it by creating invoices with the expected format			// This would normally come from the settings system			// Create settings for this format			ctx := context.Background()			service := &InvoiceService{db: db}			db := setupTestDB(t)		t.Run(tt.name, func(t *testing.T) {	for _, tt := range tests {	}		},			expectedPattern: `^RE-20\d{2}-\d{4}$`, // RE-2026-0001			prefix:          "RE",			format:          "year-prefix",			name:            "year-prefix with custom prefix",		{		},			expectedPattern: `^INV-\d+$`, // INV-1			prefix:          "INV",			format:          "sequential",			name:            "sequential with prefix",		{		},			expectedPattern: `^20\d{2}-\d{2}-\d{4}$`, // 2026-01-0001			prefix:          "",			format:          "year-month",			name:            "year-month",		{		},			expectedPattern: `^20\d{2}-\d{4}$`, // 2026-0001			prefix:          "",			format:          "year-prefix",			name:            "year-prefix",		{		},			expectedPattern: `^\d+$`, // Just numbers			prefix:          "",			format:          "sequential",			name:            "sequential",		{	}{		expectedPattern string		prefix         string		format         string		name           string	tests := []struct {func TestInvoiceNumberFormats(t *testing.T) {// TestInvoiceNumberFormats tests different invoice number formats}	assert.Equal(t, 10, len(numbers), "Should have 10 unique invoice numbers")	}		assert.Equal(t, 1, count, "Invoice number %s was generated %d times (should be 1)", num, count)	for num, count := range numbers {	// Verify all numbers are unique	}		numbers[num]++	for num := range results {	numbers := make(map[string]int)	// Collect all invoice numbers	}		t.Errorf("Finalization error: %v", err)	for err := range errors {	// Check for errors	close(results)