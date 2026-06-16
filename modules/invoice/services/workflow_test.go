package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ae/shared-modules/invoice/entities"
)

func setupInvoiceDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger:                                   logger.Discard,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)

	// Use raw SQL to avoid GIN index (PostgreSQL-only) causing SQLite "USING" syntax error
	sqls := []string{
		`CREATE TABLE IF NOT EXISTS invoices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL,
			organization_id INTEGER NOT NULL DEFAULT 0,
			user_id INTEGER NOT NULL DEFAULT 0,
			invoice_number TEXT NOT NULL DEFAULT '' UNIQUE,
			invoice_date DATETIME NOT NULL,
			due_date DATETIME,
			status TEXT NOT NULL DEFAULT 'draft',
			customer_name TEXT NOT NULL DEFAULT '',
			customer_address TEXT, customer_address_ext TEXT, customer_zip TEXT,
			customer_city TEXT, customer_country TEXT, customer_email TEXT,
			customer_tax_id TEXT, customer_contact_person TEXT, customer_department TEXT,
			subject TEXT, our_reference TEXT, your_reference TEXT,
			po_number TEXT, delivery_date DATETIME,
			performance_period_start DATETIME, performance_period_end DATETIME,
			subtotal_amount REAL NOT NULL DEFAULT 0,
			tax_rate REAL NOT NULL DEFAULT 0,
			tax_amount REAL NOT NULL DEFAULT 0,
			total_amount REAL NOT NULL DEFAULT 0,
			currency TEXT NOT NULL DEFAULT 'EUR',
			payment_terms TEXT, net_terms INTEGER DEFAULT 30,
			payment_method TEXT, payment_date DATETIME,
			discount_rate REAL DEFAULT 0, discount_terms TEXT,
			document_id INTEGER,
			num_reminders INTEGER DEFAULT 0,
			finalized_at DATETIME, cancelled_at DATETIME,
			cancellation_reason TEXT, reminder_sent_at DATETIME, email_sent_at DATETIME,
			is_credit_note BOOLEAN NOT NULL DEFAULT 0,
			credit_note_reference_id INTEGER,
			notes TEXT, internal_note TEXT, metadata TEXT,
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS invoice_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			invoice_id INTEGER NOT NULL,
			position INTEGER NOT NULL DEFAULT 0,
			description TEXT NOT NULL DEFAULT '',
			quantity REAL NOT NULL DEFAULT 1,
			unit_price REAL NOT NULL DEFAULT 0,
			tax_rate REAL NOT NULL DEFAULT 0,
			amount REAL NOT NULL DEFAULT 0,
			vat_rate REAL NOT NULL DEFAULT 0,
			vat_exempt BOOLEAN NOT NULL DEFAULT 0,
			vat_exemption_text TEXT,
			session_id INTEGER,
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
		)`,
	}
	for _, sql := range sqls {
		require.NoError(t, db.Exec(sql).Error)
	}

	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	})
	return db
}

// createTestInvoice inserts an invoice with one item and returns the loaded record.
func createTestInvoice(t *testing.T, db *gorm.DB, tenantID uint, status entities.InvoiceStatus, invoiceNum string) entities.Invoice {
	t.Helper()
	invoiceDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	inv := entities.Invoice{
		TenantID:       tenantID,
		OrganizationID: 1,
		UserID:         1,
		InvoiceNumber:  invoiceNum,
		InvoiceDate:    invoiceDate,
		DueDate:        &dueDate,
		Status:         status,
		CustomerName:   "Test Customer GmbH",
		TotalAmount:    119.0,
		SubtotalAmount: 100.0,
		TaxAmount:      19.0,
		Currency:       "EUR",
	}
	require.NoError(t, db.Create(&inv).Error)

	item := entities.InvoiceItem{
		InvoiceID:   inv.ID,
		Description: "Consulting",
		Quantity:    1,
		UnitPrice:   100.0,
		VATRate:     19.0,
		Amount:      100.0,
	}
	require.NoError(t, db.Create(&item).Error)
	inv.Items = []entities.InvoiceItem{item}
	return inv
}

// ──────────────────────────────────────────────────────────────
// FinalizeInvoice
// ──────────────────────────────────────────────────────────────

func TestInvoiceWorkflow_FinalizeInvoice(t *testing.T) {
	ctx := context.Background()

	t.Run("finalizes draft invoice with pre-set number", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusDraft, "INV-001")

		result, err := svc.FinalizeInvoice(ctx, 1, inv.ID, 1)
		require.NoError(t, err)
		assert.Equal(t, entities.InvoiceStatusFinalized, result.Status)
		assert.NotNil(t, result.FinalizedAt)
	})

	t.Run("fails for non-draft invoice", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusSent, "INV-002")

		_, err := svc.FinalizeInvoice(ctx, 1, inv.ID, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "draft")
	})

	t.Run("fails when invoice not found", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		_, err := svc.FinalizeInvoice(ctx, 1, 9999, 1)
		require.Error(t, err)
	})

	t.Run("fails without items", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		invoiceDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		inv := entities.Invoice{
			TenantID: 1, OrganizationID: 1, UserID: 1,
			InvoiceNumber: "INV-003", InvoiceDate: invoiceDate,
			Status: entities.InvoiceStatusDraft, CustomerName: "Test", TotalAmount: 100,
		}
		require.NoError(t, db.Create(&inv).Error)
		_, err := svc.FinalizeInvoice(ctx, 1, inv.ID, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item")
	})

	t.Run("fails VAT validation: exempt without exemption text", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		invoiceDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		inv := entities.Invoice{
			TenantID: 1, OrganizationID: 1, UserID: 1,
			InvoiceNumber: "INV-004", InvoiceDate: invoiceDate,
			Status: entities.InvoiceStatusDraft, CustomerName: "Test", TotalAmount: 100,
		}
		require.NoError(t, db.Create(&inv).Error)
		item := entities.InvoiceItem{InvoiceID: inv.ID, Description: "Service", VATExempt: true, VATRate: 0}
		require.NoError(t, db.Create(&item).Error)

		_, err := svc.FinalizeInvoice(ctx, 1, inv.ID, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "VAT")
	})

	t.Run("fails VAT validation: non-exempt without VAT rate", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		invoiceDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		inv := entities.Invoice{
			TenantID: 1, OrganizationID: 1, UserID: 1,
			InvoiceNumber: "INV-005", InvoiceDate: invoiceDate,
			Status: entities.InvoiceStatusDraft, CustomerName: "Test", TotalAmount: 100,
		}
		require.NoError(t, db.Create(&inv).Error)
		item := entities.InvoiceItem{InvoiceID: inv.ID, Description: "Service", VATExempt: false, VATRate: 0}
		require.NoError(t, db.Create(&item).Error)

		_, err := svc.FinalizeInvoice(ctx, 1, inv.ID, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "VAT")
	})

	t.Run("fails invoice validation: customer name missing", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		invoiceDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		inv := entities.Invoice{
			TenantID: 1, OrganizationID: 1, UserID: 1,
			InvoiceNumber: "INV-006", InvoiceDate: invoiceDate,
			Status: entities.InvoiceStatusDraft, CustomerName: "", TotalAmount: 100,
		}
		require.NoError(t, db.Create(&inv).Error)
		item := entities.InvoiceItem{InvoiceID: inv.ID, Description: "Service", VATRate: 19}
		require.NoError(t, db.Create(&item).Error)

		_, err := svc.FinalizeInvoice(ctx, 1, inv.ID, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "customer")
	})

	t.Run("fails invoice validation: due date before invoice date", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		invoiceDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		dueDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) // before invoice date
		inv := entities.Invoice{
			TenantID: 1, OrganizationID: 1, UserID: 1,
			InvoiceNumber: "INV-007", InvoiceDate: invoiceDate, DueDate: &dueDate,
			Status: entities.InvoiceStatusDraft, CustomerName: "Test", TotalAmount: 100,
		}
		require.NoError(t, db.Create(&inv).Error)
		item := entities.InvoiceItem{InvoiceID: inv.ID, Description: "Service", VATRate: 19}
		require.NoError(t, db.Create(&item).Error)

		_, err := svc.FinalizeInvoice(ctx, 1, inv.ID, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "due date")
	})

	t.Run("fails invoice validation: total not positive", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		invoiceDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		inv := entities.Invoice{
			TenantID: 1, OrganizationID: 1, UserID: 1,
			InvoiceNumber: "INV-008", InvoiceDate: invoiceDate,
			Status: entities.InvoiceStatusDraft, CustomerName: "Test", TotalAmount: 0,
		}
		require.NoError(t, db.Create(&inv).Error)
		item := entities.InvoiceItem{InvoiceID: inv.ID, Description: "Service", VATRate: 19}
		require.NoError(t, db.Create(&item).Error)

		_, err := svc.FinalizeInvoice(ctx, 1, inv.ID, 1)
		require.Error(t, err)
	})

	t.Run("fails invoice validation: performance period end before start", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		invoiceDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		dueDate := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
		perfStart := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
		perfEnd := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC) // before start
		inv := entities.Invoice{
			TenantID: 1, OrganizationID: 1, UserID: 1,
			InvoiceNumber: "INV-009", InvoiceDate: invoiceDate, DueDate: &dueDate,
			Status:                 entities.InvoiceStatusDraft,
			CustomerName:           "Test",
			TotalAmount:            100,
			PerformancePeriodStart: &perfStart,
			PerformancePeriodEnd:   &perfEnd,
		}
		require.NoError(t, db.Create(&inv).Error)
		item := entities.InvoiceItem{InvoiceID: inv.ID, Description: "Service", VATRate: 19}
		require.NoError(t, db.Create(&item).Error)

		_, err := svc.FinalizeInvoice(ctx, 1, inv.ID, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "performance period")
	})
}

// ──────────────────────────────────────────────────────────────
// MarkAsSent
// ──────────────────────────────────────────────────────────────

func TestInvoiceWorkflow_MarkAsSent(t *testing.T) {
	ctx := context.Background()

	t.Run("marks finalized invoice as sent", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusFinalized, "INV-101")

		result, err := svc.MarkAsSent(ctx, 1, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, entities.InvoiceStatusSent, result.Status)
	})

	t.Run("fails for draft invoice", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusDraft, "INV-102")

		_, err := svc.MarkAsSent(ctx, 1, inv.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "finalized")
	})

	t.Run("fails for non-existent invoice", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		_, err := svc.MarkAsSent(ctx, 1, 9999)
		require.Error(t, err)
	})
}

// ──────────────────────────────────────────────────────────────
// MarkAsPaidWithAmount
// ──────────────────────────────────────────────────────────────

func TestInvoiceWorkflow_MarkAsPaid(t *testing.T) {
	ctx := context.Background()
	paymentDate := time.Now()

	t.Run("marks sent invoice as paid", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusSent, "INV-201")

		result, err := svc.MarkAsPaidWithAmount(ctx, 1, inv.ID, paymentDate, "bank_transfer")
		require.NoError(t, err)
		assert.Equal(t, entities.InvoiceStatusPaid, result.Status)
		assert.NotNil(t, result.PaymentDate)
		assert.Equal(t, "bank_transfer", result.PaymentMethod)
	})

	t.Run("marks overdue invoice as paid", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusOverdue, "INV-202")

		result, err := svc.MarkAsPaidWithAmount(ctx, 1, inv.ID, paymentDate, "cash")
		require.NoError(t, err)
		assert.Equal(t, entities.InvoiceStatusPaid, result.Status)
	})

	t.Run("fails for draft invoice", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusDraft, "INV-203")

		_, err := svc.MarkAsPaidWithAmount(ctx, 1, inv.ID, paymentDate, "cash")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sent")
	})

	t.Run("fails for non-existent invoice", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		_, err := svc.MarkAsPaidWithAmount(ctx, 1, 9999, paymentDate, "cash")
		require.Error(t, err)
	})
}

// ──────────────────────────────────────────────────────────────
// SendReminder
// ──────────────────────────────────────────────────────────────

func TestInvoiceWorkflow_SendReminder(t *testing.T) {
	ctx := context.Background()

	t.Run("sends reminder to sent invoice", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusSent, "INV-301")

		result, err := svc.SendReminder(ctx, 1, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, result.NumReminders)
		assert.NotNil(t, result.ReminderSentAt)
	})

	t.Run("sends reminder to overdue invoice", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusOverdue, "INV-302")

		result, err := svc.SendReminder(ctx, 1, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, result.NumReminders)
	})

	t.Run("overdue invoice stays overdue when past due date", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		pastDue := time.Now().AddDate(-1, 0, 0) // 1 year ago
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusSent, "INV-303")
		require.NoError(t, db.Model(&inv).Update("due_date", &pastDue).Error)

		result, err := svc.SendReminder(ctx, 1, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, entities.InvoiceStatusOverdue, result.Status)
	})

	t.Run("fails for draft invoice", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusDraft, "INV-304")

		_, err := svc.SendReminder(ctx, 1, inv.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reminder")
	})

	t.Run("fails for non-existent invoice", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		_, err := svc.SendReminder(ctx, 1, 9999)
		require.Error(t, err)
	})
}

// ──────────────────────────────────────────────────────────────
// CancelInvoice
// ──────────────────────────────────────────────────────────────

func TestInvoiceWorkflow_CancelInvoice(t *testing.T) {
	ctx := context.Background()

	t.Run("cancels draft invoice", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusDraft, "INV-401")

		result, err := svc.CancelInvoice(ctx, 1, inv.ID, "")
		require.NoError(t, err)
		assert.Equal(t, entities.InvoiceStatusCancelled, result.Status)
		assert.NotNil(t, result.CancelledAt)
	})

	t.Run("cancels sent invoice with reason", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusSent, "INV-402")

		result, err := svc.CancelInvoice(ctx, 1, inv.ID, "Customer requested cancellation")
		require.NoError(t, err)
		assert.Equal(t, entities.InvoiceStatusCancelled, result.Status)
		assert.Equal(t, "Customer requested cancellation", result.CancellationReason)
	})

	t.Run("fails for paid invoice", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusPaid, "INV-403")

		_, err := svc.CancelInvoice(ctx, 1, inv.ID, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "paid")
	})

	t.Run("fails for already cancelled invoice", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusCancelled, "INV-404")

		_, err := svc.CancelInvoice(ctx, 1, inv.ID, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cancelled")
	})

	t.Run("fails for non-existent invoice", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		_, err := svc.CancelInvoice(ctx, 1, 9999, "")
		require.Error(t, err)
	})
}

// ──────────────────────────────────────────────────────────────
// CheckOverdueInvoices
// ──────────────────────────────────────────────────────────────

func TestInvoiceWorkflow_CheckOverdueInvoices(t *testing.T) {
	ctx := context.Background()

	t.Run("marks sent invoice with past due date as overdue", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		pastDue := time.Now().AddDate(-1, 0, 0)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusSent, "INV-501")
		require.NoError(t, db.Model(&inv).Update("due_date", &pastDue).Error)

		err := svc.CheckOverdueInvoices(ctx)
		require.NoError(t, err)

		var updated entities.Invoice
		db.First(&updated, inv.ID)
		assert.Equal(t, entities.InvoiceStatusOverdue, updated.Status)
	})

	t.Run("does not affect paid invoices", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		pastDue := time.Now().AddDate(-1, 0, 0)
		inv := createTestInvoice(t, db, 1, entities.InvoiceStatusPaid, "INV-502")
		require.NoError(t, db.Model(&inv).Update("due_date", &pastDue).Error)

		err := svc.CheckOverdueInvoices(ctx)
		require.NoError(t, err)

		var updated entities.Invoice
		db.First(&updated, inv.ID)
		assert.Equal(t, entities.InvoiceStatusPaid, updated.Status)
	})

	t.Run("runs successfully with empty DB", func(t *testing.T) {
		db := setupInvoiceDB(t)
		svc := NewInvoiceService(db)
		err := svc.CheckOverdueInvoices(ctx)
		assert.NoError(t, err)
	})
}
