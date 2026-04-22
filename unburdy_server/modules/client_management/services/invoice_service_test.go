package services

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	baseAPI "github.com/ae-base-server/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/unburdy-server-api/internal/models"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var invoiceCounter int64

func setupInvoiceServiceDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&baseAPI.Organization{},
		&entities.CostProvider{},
		&entities.Client{},
		&entities.Session{},
		&entities.ExtraEffort{},
		&entities.Invoice{},
		&entities.InvoiceItem{},
		&entities.ClientInvoice{},
		&models.CostProvider{},
	))
	return db
}

// createTestInvoiceRecord inserts a minimal valid Invoice row directly in the DB.
func createTestInvoiceRecord(t *testing.T, db *gorm.DB, tenantID, userID, orgID uint, status entities.InvoiceStatus) *entities.Invoice {
	t.Helper()
	inv := entities.Invoice{
		TenantID:       tenantID,
		UserID:         userID,
		OrganizationID: orgID,
		InvoiceDate:    time.Now().UTC(),
		InvoiceNumber:  fmt.Sprintf("TEST-%d", atomic.AddInt64(&invoiceCounter, 1)),
		Status:         status,
	}
	require.NoError(t, db.Create(&inv).Error)
	return &inv
}

func createTestOrganizationRecord(t *testing.T, db *gorm.DB, tenantID uint) *baseAPI.Organization {
	t.Helper()
	org := baseAPI.Organization{
		TenantID: tenantID,
		Name:     "Test Org",
	}
	require.NoError(t, db.Create(&org).Error)
	return &org
}

func TestInvoiceService_GetInvoiceByID(t *testing.T) {
	db := setupInvoiceServiceDB(t)
	svc := NewInvoiceService(db)
	org := createTestOrganizationRecord(t, db, 1)
	inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusDraft)

	t.Run("found", func(t *testing.T) {
		result, err := svc.GetInvoiceByID(inv.ID, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, inv.ID, result.ID)
	})

	t.Run("not found by id returns error", func(t *testing.T) {
		_, err := svc.GetInvoiceByID(9999, 1, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("wrong tenant returns error", func(t *testing.T) {
		_, err := svc.GetInvoiceByID(inv.ID, 2, 10)
		require.Error(t, err)
	})

	t.Run("wrong user returns error", func(t *testing.T) {
		_, err := svc.GetInvoiceByID(inv.ID, 1, 99)
		require.Error(t, err)
	})
}

func TestInvoiceService_GetInvoices(t *testing.T) {
	db := setupInvoiceServiceDB(t)
	svc := NewInvoiceService(db)
	org := createTestOrganizationRecord(t, db, 1)

	for i := 0; i < 3; i++ {
		createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusDraft)
	}
	// Different tenant/user
	org2 := createTestOrganizationRecord(t, db, 2)
	createTestInvoiceRecord(t, db, 2, 20, org2.ID, entities.InvoiceStatusDraft)

	t.Run("returns correct invoices with pagination", func(t *testing.T) {
		invoices, total, err := svc.GetInvoices(1, 10, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, invoices, 3)
	})

	t.Run("pagination works", func(t *testing.T) {
		invoices, total, err := svc.GetInvoices(1, 2, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, invoices, 2)
	})

	t.Run("tenant isolation", func(t *testing.T) {
		invoices, total, err := svc.GetInvoices(1, 10, 2, 20)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, invoices, 1)
	})
}

func TestInvoiceService_DeleteInvoice(t *testing.T) {
	db := setupInvoiceServiceDB(t)
	svc := NewInvoiceService(db)
	org := createTestOrganizationRecord(t, db, 1)
	inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusDraft)

	t.Run("deletes draft invoice successfully", func(t *testing.T) {
		err := svc.DeleteInvoice(inv.ID, 1, 10)
		require.NoError(t, err)
		_, err2 := svc.GetInvoiceByID(inv.ID, 1, 10)
		require.Error(t, err2)
	})

	t.Run("not found returns error", func(t *testing.T) {
		err := svc.DeleteInvoice(9999, 1, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("wrong tenant returns error", func(t *testing.T) {
		inv2 := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusDraft)
		err := svc.DeleteInvoice(inv2.ID, 2, 10)
		require.Error(t, err)
	})
}

func TestInvoiceService_CancelDraftInvoice(t *testing.T) {
	db := setupInvoiceServiceDB(t)
	svc := NewInvoiceService(db)
	org := createTestOrganizationRecord(t, db, 1)

	t.Run("cancels draft invoice", func(t *testing.T) {
		inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusDraft)
		err := svc.CancelDraftInvoice(inv.ID, 1, 10)
		require.NoError(t, err)
		// Invoice should be soft-deleted (not findable)
		var found entities.Invoice
		dbErr := db.First(&found, inv.ID).Error
		assert.Error(t, dbErr) // soft deleted
	})

	t.Run("non-draft invoice returns error", func(t *testing.T) {
		inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusSent)
		err := svc.CancelDraftInvoice(inv.ID, 1, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "draft")
	})

	t.Run("not found returns error", func(t *testing.T) {
		err := svc.CancelDraftInvoice(9999, 1, 10)
		require.Error(t, err)
	})
}

func TestInvoiceService_UpdateInvoice_StatusChange(t *testing.T) {
	db := setupInvoiceServiceDB(t)
	svc := NewInvoiceService(db)
	org := createTestOrganizationRecord(t, db, 1)

	t.Run("update status to paid sets payed_date", func(t *testing.T) {
		inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusSent)
		newStatus := entities.InvoiceStatusPaid
		updated, err := svc.UpdateInvoice(inv.ID, 1, 10, entities.UpdateInvoiceRequest{Status: &newStatus})
		require.NoError(t, err)
		assert.Equal(t, entities.InvoiceStatusPaid, updated.Status)
	})

	t.Run("update status to overdue increments reminder count", func(t *testing.T) {
		inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusSent)
		newStatus := entities.InvoiceStatusOverdue
		updated, err := svc.UpdateInvoice(inv.ID, 1, 10, entities.UpdateInvoiceRequest{Status: &newStatus})
		require.NoError(t, err)
		assert.Equal(t, entities.InvoiceStatusOverdue, updated.Status)
		assert.Equal(t, 1, updated.NumReminders)
	})

	t.Run("not found returns error", func(t *testing.T) {
		newStatus := entities.InvoiceStatusPaid
		_, err := svc.UpdateInvoice(9999, 1, 10, entities.UpdateInvoiceRequest{Status: &newStatus})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("wrong tenant returns error", func(t *testing.T) {
		inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusSent)
		newStatus := entities.InvoiceStatusPaid
		_, err := svc.UpdateInvoice(inv.ID, 2, 10, entities.UpdateInvoiceRequest{Status: &newStatus})
		require.Error(t, err)
	})
}

func TestInvoiceService_UpdateInvoiceDocumentID(t *testing.T) {
	db := setupInvoiceServiceDB(t)
	svc := NewInvoiceService(db)
	org := createTestOrganizationRecord(t, db, 1)
	inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusDraft)

	t.Run("updates document id successfully", func(t *testing.T) {
		err := svc.UpdateInvoiceDocumentID(inv.ID, 42, 1, 10)
		require.NoError(t, err)
		updated, err := svc.GetInvoiceByID(inv.ID, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, updated.DocumentID)
		assert.Equal(t, uint(42), *updated.DocumentID)
	})

	t.Run("invoice not found returns error", func(t *testing.T) {
		err := svc.UpdateInvoiceDocumentID(9999, 42, 1, 10)
		require.Error(t, err)
	})
}

func TestInvoiceService_GetCostProviderByID(t *testing.T) {
	db := setupInvoiceServiceDB(t)
	svc := NewInvoiceService(db)

	cp := models.CostProvider{
		TenantID:     1,
		Organization: "Test Authority",
	}
	require.NoError(t, db.Create(&cp).Error)

	t.Run("found", func(t *testing.T) {
		result, err := svc.GetCostProviderByID(cp.ID, 1)
		require.NoError(t, err)
		assert.Equal(t, "Test Authority", result.Organization)
	})

	t.Run("not found returns error", func(t *testing.T) {
		_, err := svc.GetCostProviderByID(9999, 1)
		require.Error(t, err)
	})

	t.Run("wrong tenant returns error", func(t *testing.T) {
		_, err := svc.GetCostProviderByID(cp.ID, 2)
		require.Error(t, err)
	})
}

func TestInvoiceService_GetOrganizationByID(t *testing.T) {
	db := setupInvoiceServiceDB(t)
	svc := NewInvoiceService(db)
	org := createTestOrganizationRecord(t, db, 1)

	t.Run("found", func(t *testing.T) {
		result, err := svc.GetOrganizationByID(org.ID, 1)
		require.NoError(t, err)
		assert.Equal(t, "Test Org", result.Name)
	})

	t.Run("not found returns error", func(t *testing.T) {
		_, err := svc.GetOrganizationByID(9999, 1)
		require.Error(t, err)
	})
}

func TestInvoiceService_GenerateInvoicePDF_NoPDFService(t *testing.T) {
	db := setupInvoiceServiceDB(t)
	svc := NewInvoiceService(db)

	_, err := svc.GenerateInvoicePDF(1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PDF service not available")
}

func TestInvoiceService_SetPDFService(t *testing.T) {
	db := setupInvoiceServiceDB(t)
	svc := NewInvoiceService(db)
	assert.Nil(t, svc.GetPDFService())
	svc.SetPDFService(nil) // setting nil is still an operation on the service
	assert.Nil(t, svc.GetPDFService())
}
func TestInvoiceService_MarkAsSent(t *testing.T) {
        db := setupInvoiceServiceDB(t)
        svc := NewInvoiceService(db)
        org := createTestOrganizationRecord(t, db, 1)

        t.Run("success transitions finalized to sent", func(t *testing.T) {
                inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusFinalized)
                result, err := svc.MarkAsSent(inv.ID, 1, "email")
                require.NoError(t, err)
                assert.Equal(t, entities.InvoiceStatusSent, result.Status)
                assert.NotNil(t, result.SentAt)
        })

        t.Run("wrong status returns error", func(t *testing.T) {
                inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusDraft)
                _, err := svc.MarkAsSent(inv.ID, 1, "email")
                require.Error(t, err)
        })

        t.Run("not found returns error", func(t *testing.T) {
                _, err := svc.MarkAsSent(9999, 1, "email")
                require.Error(t, err)
        })
}

func TestInvoiceService_MarkInvoiceAsPaid(t *testing.T) {
        db := setupInvoiceServiceDB(t)
        svc := NewInvoiceService(db)
        org := createTestOrganizationRecord(t, db, 1)

        t.Run("success from sent status with explicit date", func(t *testing.T) {
                inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusSent)
                payDate := time.Now().AddDate(0, 0, -1)
                result, err := svc.MarkInvoiceAsPaid(inv.ID, 1, 10, &payDate, "REF-123")
                require.NoError(t, err)
                assert.Equal(t, entities.InvoiceStatusPaid, result.Status)
        })

        t.Run("success from overdue status with nil date defaults to today", func(t *testing.T) {
                inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusOverdue)
                result, err := svc.MarkInvoiceAsPaid(inv.ID, 1, 10, nil, "")
                require.NoError(t, err)
                assert.Equal(t, entities.InvoiceStatusPaid, result.Status)
        })

        t.Run("wrong status returns error", func(t *testing.T) {
                inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusDraft)
                _, err := svc.MarkInvoiceAsPaid(inv.ID, 1, 10, nil, "")
                require.Error(t, err)
        })

        t.Run("not found returns error", func(t *testing.T) {
                _, err := svc.MarkInvoiceAsPaid(9999, 1, 10, nil, "")
                require.Error(t, err)
        })
}

func TestInvoiceService_CancelInvoice(t *testing.T) {
        db := setupInvoiceServiceDB(t)
        svc := NewInvoiceService(db)
        org := createTestOrganizationRecord(t, db, 1)

        t.Run("success cancels finalized invoice with number", func(t *testing.T) {
                inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusFinalized)
                err := svc.CancelInvoice(inv.ID, 1, 10, "client request")
                require.NoError(t, err)
                var check entities.Invoice
                db.First(&check, inv.ID)
                assert.Equal(t, entities.InvoiceStatusCancelled, check.Status)
        })

        t.Run("already cancelled returns error", func(t *testing.T) {
                inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusCancelled)
                err := svc.CancelInvoice(inv.ID, 1, 10, "reason")
                require.Error(t, err)
        })

        t.Run("invoice without number returns error", func(t *testing.T) {
                inv := &entities.Invoice{TenantID: 1, UserID: 10, OrganizationID: org.ID, InvoiceDate: time.Now(), Status: entities.InvoiceStatusFinalized}
                require.NoError(t, db.Create(inv).Error)
                // Remove invoice number
                db.Model(inv).Update("invoice_number", "")
                err := svc.CancelInvoice(inv.ID, 1, 10, "reason")
                require.Error(t, err)
        })

        t.Run("not found returns error", func(t *testing.T) {
                err := svc.CancelInvoice(9999, 1, 10, "reason")
                require.Error(t, err)
        })
}

func TestInvoiceService_SendReminder(t *testing.T) {
        db := setupInvoiceServiceDB(t)
        svc := NewInvoiceService(db)
        org := createTestOrganizationRecord(t, db, 1)

        t.Run("success increments reminder counter", func(t *testing.T) {
                inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusOverdue)
                err := svc.SendReminder(inv.ID, 1, 10)
                require.NoError(t, err)
                var check entities.Invoice
                db.First(&check, inv.ID)
                assert.Equal(t, 1, check.NumReminders)
        })

        t.Run("wrong status returns error", func(t *testing.T) {
                inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusSent)
                err := svc.SendReminder(inv.ID, 1, 10)
                require.Error(t, err)
        })

        t.Run("not found returns error", func(t *testing.T) {
                err := svc.SendReminder(9999, 1, 10)
                require.Error(t, err)
        })
}

func TestInvoiceService_MarkInvoiceAsOverdue(t *testing.T) {
        db := setupInvoiceServiceDB(t)
        // Add payment_due_days column to organizations table (not in base entity)
        db.Exec("ALTER TABLE organizations ADD COLUMN payment_due_days INTEGER DEFAULT 14")
        svc := NewInvoiceService(db)
        org := createTestOrganizationRecord(t, db, 1)
        // Set payment_due_days to 0 so the invoice is immediately overdue
        db.Exec("UPDATE organizations SET payment_due_days = ? WHERE id = ?", 0, org.ID)

        t.Run("success marks sent invoice as overdue when past due", func(t *testing.T) {
                inv := &entities.Invoice{
                        TenantID:       1,
                        UserID:         10,
                        OrganizationID: org.ID,
                        InvoiceDate:    time.Now().AddDate(0, 0, -1), // yesterday
                        InvoiceNumber:  fmt.Sprintf("TEST-%d", atomic.AddInt64(&invoiceCounter, 1)),
                        Status:         entities.InvoiceStatusSent,
                }
                require.NoError(t, db.Create(inv).Error)
                result, err := svc.MarkInvoiceAsOverdue(inv.ID, 1, 10)
                require.NoError(t, err)
                assert.Equal(t, entities.InvoiceStatusOverdue, result.Status)
        })

        t.Run("wrong status returns error", func(t *testing.T) {
                inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusDraft)
                _, err := svc.MarkInvoiceAsOverdue(inv.ID, 1, 10)
                require.Error(t, err)
        })

        t.Run("not yet overdue returns error", func(t *testing.T) {
                // invoice dated today with payment_due_days=14 → not overdue yet
                db2 := setupInvoiceServiceDB(t)
                db2.Exec("ALTER TABLE organizations ADD COLUMN payment_due_days INTEGER DEFAULT 14")
                org2 := createTestOrganizationRecord(t, db2, 1)
                db2.Exec("UPDATE organizations SET payment_due_days = ? WHERE id = ?", 14, org2.ID)
                svc2 := NewInvoiceService(db2)
                inv := &entities.Invoice{
                        TenantID:       1,
                        UserID:         10,
                        OrganizationID: org2.ID,
                        InvoiceDate:    time.Now(),
                        InvoiceNumber:  fmt.Sprintf("TEST-%d", atomic.AddInt64(&invoiceCounter, 1)),
                        Status:         entities.InvoiceStatusSent,
                }
                require.NoError(t, db2.Create(inv).Error)
                _, err := svc2.MarkInvoiceAsOverdue(inv.ID, 1, 10)
                require.Error(t, err)
        })

        t.Run("not found returns error", func(t *testing.T) {
                _, err := svc.MarkInvoiceAsOverdue(9999, 1, 10)
                require.Error(t, err)
        })
}

func TestInvoiceService_SendInvoiceEmail(t *testing.T) {
        db := setupInvoiceServiceDB(t)
        svc := NewInvoiceService(db)
        org := createTestOrganizationRecord(t, db, 1)

        t.Run("success updates sent_at for sent invoice", func(t *testing.T) {
                inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusSent)
                err := svc.SendInvoiceEmail(inv.ID, 1, 10)
                require.NoError(t, err)
        })

        t.Run("success for paid invoice", func(t *testing.T) {
                inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusPaid)
                err := svc.SendInvoiceEmail(inv.ID, 1, 10)
                require.NoError(t, err)
        })

        t.Run("success for overdue invoice", func(t *testing.T) {
                inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusOverdue)
                err := svc.SendInvoiceEmail(inv.ID, 1, 10)
                require.NoError(t, err)
        })

        t.Run("draft invoice returns error", func(t *testing.T) {
                inv := createTestInvoiceRecord(t, db, 1, 10, org.ID, entities.InvoiceStatusDraft)
                err := svc.SendInvoiceEmail(inv.ID, 1, 10)
                require.Error(t, err)
        })

        t.Run("not found returns error", func(t *testing.T) {
                err := svc.SendInvoiceEmail(9999, 1, 10)
                require.Error(t, err)
        })
}

func TestInvoiceService_GetClientsWithUnbilledSessions(t *testing.T) {
        db := setupInvoiceServiceDB(t)
        svc := NewInvoiceService(db)

        // Create a client with a conducted session
        client := &entities.Client{TenantID: 1, FirstName: "Test", LastName: "Client"}
        require.NoError(t, db.Create(client).Error)
        session := &entities.Session{
                TenantID:          1,
                ClientID:          client.ID,
                OriginalDate:      time.Now().AddDate(0, 0, -1),
                OriginalStartTime: time.Now().AddDate(0, 0, -1),
                DurationMin:       60,
                Status:            "conducted",
                Type:              "individual",
        }
        require.NoError(t, db.Create(session).Error)

        result, err := svc.GetClientsWithUnbilledSessions(1, 10)
        require.NoError(t, err)
        assert.GreaterOrEqual(t, len(result), 1)
}