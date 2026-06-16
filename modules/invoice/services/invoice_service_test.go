package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ae/shared-modules/invoice/entities"
)

func makeCreateReq(orgID uint, num string) *entities.CreateInvoiceRequest {
	return &entities.CreateInvoiceRequest{
		OrganizationID: orgID,
		InvoiceNumber:  num,
		InvoiceDate:    time.Now(),
		CustomerName:   "Test Customer",
		Items: []entities.InvoiceItemData{
			{Description: "Service A", Quantity: 2, UnitPrice: 100.00, TaxRate: 19},
		},
	}
}

func TestInvoiceService_CreateInvoice(t *testing.T) {
	ctx := context.Background()
	db := setupInvoiceDB(t)
	svc := NewInvoiceService(db)

	t.Run("creates invoice with calculated amounts", func(t *testing.T) {
		req := makeCreateReq(1, "INV-CRUD-001")
		invoice, err := svc.CreateInvoice(ctx, 1, 10, req)
		require.NoError(t, err)
		assert.Equal(t, "INV-CRUD-001", invoice.InvoiceNumber)
		assert.Equal(t, entities.InvoiceStatusDraft, invoice.Status)
		assert.InDelta(t, 200.0, invoice.SubtotalAmount, 0.01)
		assert.InDelta(t, 38.0, invoice.TaxAmount, 0.01)
		assert.InDelta(t, 238.0, invoice.TotalAmount, 0.01)
	})

	t.Run("creates invoice with zero tax rate", func(t *testing.T) {
		req := &entities.CreateInvoiceRequest{
			OrganizationID: 1,
			InvoiceNumber:  "INV-CRUD-002",
			InvoiceDate:    time.Now(),
			CustomerName:   "Zero Tax",
			Items: []entities.InvoiceItemData{
				{Description: "Item", Quantity: 1, UnitPrice: 50.00, TaxRate: 0},
			},
		}
		invoice, err := svc.CreateInvoice(ctx, 1, 10, req)
		require.NoError(t, err)
		assert.InDelta(t, 50.0, invoice.SubtotalAmount, 0.01)
		assert.InDelta(t, 0.0, invoice.TaxAmount, 0.01)
	})
}

func TestInvoiceService_GetInvoice(t *testing.T) {
	ctx := context.Background()
	db := setupInvoiceDB(t)
	svc := NewInvoiceService(db)

	inv := createTestInvoice(t, db, 1, entities.InvoiceStatusDraft, "INV-GET-001")

	t.Run("found", func(t *testing.T) {
		result, err := svc.GetInvoice(ctx, 1, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, inv.ID, result.ID)
	})

	t.Run("wrong tenant returns error", func(t *testing.T) {
		_, err := svc.GetInvoice(ctx, 2, inv.ID)
		require.Error(t, err)
	})

	t.Run("not found returns error", func(t *testing.T) {
		_, err := svc.GetInvoice(ctx, 1, 9999)
		require.Error(t, err)
	})
}

func TestInvoiceService_ListInvoices(t *testing.T) {
	ctx := context.Background()
	db := setupInvoiceDB(t)
	svc := NewInvoiceService(db)

	_ = createTestInvoice(t, db, 1, entities.InvoiceStatusDraft, "INV-LIST-001")
	_ = createTestInvoice(t, db, 1, entities.InvoiceStatusSent, "INV-LIST-002")
	_ = createTestInvoice(t, db, 1, entities.InvoiceStatusPaid, "INV-LIST-003")
	_ = createTestInvoice(t, db, 2, entities.InvoiceStatusDraft, "INV-LIST-004")

	t.Run("lists all for tenant", func(t *testing.T) {
		results, total, err := svc.ListInvoices(ctx, 1, nil, nil, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 3)
	})

	t.Run("filter by status", func(t *testing.T) {
		status := entities.InvoiceStatusDraft
		results, total, err := svc.ListInvoices(ctx, 1, nil, &status, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, results, 1)
	})

	t.Run("filter by organization", func(t *testing.T) {
		orgID := uint(1)
		_, _, err := svc.ListInvoices(ctx, 1, &orgID, nil, 1, 10)
		require.NoError(t, err)
	})

	t.Run("tenant isolation", func(t *testing.T) {
		results, total, err := svc.ListInvoices(ctx, 2, nil, nil, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, results, 1)
	})

	t.Run("pagination", func(t *testing.T) {
		results, total, err := svc.ListInvoices(ctx, 1, nil, nil, 1, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 2)
	})
}

func TestInvoiceService_UpdateInvoice(t *testing.T) {
	ctx := context.Background()
	db := setupInvoiceDB(t)
	svc := NewInvoiceService(db)

	inv := createTestInvoice(t, db, 1, entities.InvoiceStatusDraft, "INV-UPD-001")

	t.Run("update customer name", func(t *testing.T) {
		name := "New Customer"
		result, err := svc.UpdateInvoice(ctx, 1, inv.ID, &entities.UpdateInvoiceRequest{
			CustomerName: &name,
		})
		require.NoError(t, err)
		assert.Equal(t, "New Customer", result.CustomerName)
	})

	t.Run("update status to paid sets payment_date", func(t *testing.T) {
		status := entities.InvoiceStatusPaid
		result, err := svc.UpdateInvoice(ctx, 1, inv.ID, &entities.UpdateInvoiceRequest{
			Status: &status,
		})
		require.NoError(t, err)
		assert.Equal(t, entities.InvoiceStatusPaid, result.Status)
		assert.NotNil(t, result.PaymentDate)
	})

	t.Run("not found returns error", func(t *testing.T) {
		name := "X"
		_, err := svc.UpdateInvoice(ctx, 1, 9999, &entities.UpdateInvoiceRequest{CustomerName: &name})
		require.Error(t, err)
	})
}

func TestInvoiceService_DeleteInvoice(t *testing.T) {
	ctx := context.Background()
	db := setupInvoiceDB(t)
	svc := NewInvoiceService(db)

	inv := createTestInvoice(t, db, 1, entities.InvoiceStatusDraft, "INV-DEL-001")

	err := svc.DeleteInvoice(ctx, 1, inv.ID)
	require.NoError(t, err)
	_, err2 := svc.GetInvoice(ctx, 1, inv.ID)
	require.Error(t, err2)
}

func TestInvoiceService_MarkAsPaid(t *testing.T) {
	ctx := context.Background()
	db := setupInvoiceDB(t)
	svc := NewInvoiceService(db)

	inv := createTestInvoice(t, db, 1, entities.InvoiceStatusSent, "INV-PAY-001")

	err := svc.MarkAsPaid(ctx, 1, inv.ID, time.Now())
	require.NoError(t, err)

	var check entities.Invoice
	db.First(&check, inv.ID)
	assert.Equal(t, entities.InvoiceStatusPaid, check.Status)
	assert.NotNil(t, check.PaymentDate)
}

func TestInvoiceService_LinkDocument(t *testing.T) {
	ctx := context.Background()
	db := setupInvoiceDB(t)
	svc := NewInvoiceService(db)

	inv := createTestInvoice(t, db, 1, entities.InvoiceStatusDraft, "INV-LINK-001")

	err := svc.LinkDocument(ctx, 1, inv.ID, 42)
	require.NoError(t, err)

	var check entities.Invoice
	db.First(&check, inv.ID)
	require.NotNil(t, check.DocumentID)
	assert.Equal(t, uint(42), *check.DocumentID)
}

func TestInvoiceService_GetInvoiceSettings(t *testing.T) {
	ctx := context.Background()
	db := setupInvoiceDB(t)
	svc := NewInvoiceService(db)

	settings, err := svc.GetInvoiceSettings(ctx, 1, 1)
	require.NoError(t, err)
	assert.NotNil(t, settings)
	assert.Equal(t, "INV", settings["invoice_prefix"])
}
