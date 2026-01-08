package tests

import (
	"testing"
	"time"

	baseAPI "github.com/ae-base-server/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"github.com/unburdy/unburdy-server-api/modules/client_management/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupInvoiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&entities.Client{},
		&entities.CostProvider{},
		&entities.Session{},
		&entities.Invoice{},
		&entities.InvoiceItem{},
		&entities.ClientInvoice{},
		&entities.ExtraEffort{},
		&baseAPI.Organization{},
	)
	require.NoError(t, err)

	return db
}

func createTestData(t *testing.T, db *gorm.DB, tenantID, userID uint) (uint, uint, []uint) {
	org := &baseAPI.Organization{
		TenantID: tenantID,
		Name:     "Test Organization",
	}
	require.NoError(t, db.Create(org).Error)

	costProvider := &entities.CostProvider{
		TenantID:     tenantID,
		Organization: "Test Insurance",
	}
	require.NoError(t, db.Create(costProvider).Error)

	client := &entities.Client{
		TenantID:       tenantID,
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john.doe@example.com",
		CostProviderID: &costProvider.ID,
	}
	require.NoError(t, db.Create(client).Error)

	sessionIDs := make([]uint, 3)
	for i := 0; i < 3; i++ {
		session := &entities.Session{
			TenantID:          tenantID,
			ClientID:          client.ID,
			OriginalDate:      time.Now().AddDate(0, 0, -i),
			OriginalStartTime: time.Now().AddDate(0, 0, -i),
			DurationMin:       60,
			Type:              "therapy",
			NumberUnits:       1,
			Status:            "conducted",
		}
		require.NoError(t, db.Create(session).Error)
		sessionIDs[i] = session.ID
	}

	return client.ID, costProvider.ID, sessionIDs
}

func TestCreateInvoice(t *testing.T) {
	db := setupInvoiceTestDB(t)
	service := services.NewInvoiceService(db)

	const tenantID uint = 1
	const userID uint = 1

	clientID, _, sessionIDs := createTestData(t, db, tenantID, userID)

	req := entities.CreateInvoiceRequest{
		ClientID:   clientID,
		SessionIDs: sessionIDs,
	}

	invoice, err := service.CreateInvoice(req, tenantID, userID)
	assert.NoError(t, err)
	assert.NotNil(t, invoice)
	assert.Len(t, invoice.ClientInvoices, 3)
	assert.Equal(t, clientID, invoice.ClientInvoices[0].ClientID)
	assert.Equal(t, 3, invoice.NumberUnits)
	assert.Equal(t, entities.InvoiceStatusDraft, invoice.Status)
	assert.Len(t, invoice.InvoiceItems, 3)
}

func TestCreateInvoice_ClientNotFound(t *testing.T) {
	db := setupInvoiceTestDB(t)
	service := services.NewInvoiceService(db)

	req := entities.CreateInvoiceRequest{
		ClientID:   999,
		SessionIDs: []uint{1, 2, 3},
	}

	_, err := service.CreateInvoice(req, 1, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client not found")
}

func TestCreateInvoice_SessionAlreadyInvoiced(t *testing.T) {
	db := setupInvoiceTestDB(t)
	service := services.NewInvoiceService(db)

	const tenantID uint = 1
	const userID uint = 1

	clientID, _, sessionIDs := createTestData(t, db, tenantID, userID)

	req1 := entities.CreateInvoiceRequest{
		ClientID:   clientID,
		SessionIDs: sessionIDs,
	}
	_, err := service.CreateInvoice(req1, tenantID, userID)
	require.NoError(t, err)

	req2 := entities.CreateInvoiceRequest{
		ClientID:   clientID,
		SessionIDs: sessionIDs,
	}
	_, err = service.CreateInvoice(req2, tenantID, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already invoiced")
}

func TestGetInvoiceByID(t *testing.T) {
	db := setupInvoiceTestDB(t)
	service := services.NewInvoiceService(db)

	const tenantID uint = 1
	const userID uint = 1

	clientID, _, sessionIDs := createTestData(t, db, tenantID, userID)

	req := entities.CreateInvoiceRequest{
		ClientID:   clientID,
		SessionIDs: sessionIDs,
	}
	created, err := service.CreateInvoice(req, tenantID, userID)
	require.NoError(t, err)

	invoice, err := service.GetInvoiceByID(created.ID, tenantID, userID)
	assert.NoError(t, err)
	assert.NotNil(t, invoice)
	assert.Equal(t, created.ID, invoice.ID)
	assert.NotNil(t, invoice.Organization)
	assert.NotEmpty(t, invoice.ClientInvoices)
	assert.Len(t, invoice.InvoiceItems, 3)
}

func TestGetInvoices(t *testing.T) {
	db := setupInvoiceTestDB(t)
	service := services.NewInvoiceService(db)

	const tenantID uint = 1
	const userID uint = 1

	clientID, _, sessionIDs := createTestData(t, db, tenantID, userID)

	// Create first invoice
	req1 := entities.CreateInvoiceRequest{
		ClientID:   clientID,
		SessionIDs: []uint{sessionIDs[0]},
	}
	_, err := service.CreateInvoice(req1, tenantID, userID)
	require.NoError(t, err)

	// Wait a bit to ensure different timestamp for invoice number
	time.Sleep(time.Millisecond * 10)

	// Create second invoice with different session
	req2 := entities.CreateInvoiceRequest{
		ClientID:   clientID,
		SessionIDs: []uint{sessionIDs[1]},
	}
	_, err = service.CreateInvoice(req2, tenantID, userID)
	require.NoError(t, err)

	// Get invoices
	invoices, total, err := service.GetInvoices(1, 10, tenantID, userID)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, invoices, 2)
}

func TestUpdateInvoiceStatus(t *testing.T) {
	db := setupInvoiceTestDB(t)
	service := services.NewInvoiceService(db)

	const tenantID uint = 1
	const userID uint = 1

	clientID, _, sessionIDs := createTestData(t, db, tenantID, userID)

	req := entities.CreateInvoiceRequest{
		ClientID:   clientID,
		SessionIDs: sessionIDs,
	}
	created, err := service.CreateInvoice(req, tenantID, userID)
	require.NoError(t, err)

	newStatus := entities.InvoiceStatusSent
	updateReq := entities.UpdateInvoiceRequest{
		Status: &newStatus,
	}
	updated, err := service.UpdateInvoice(created.ID, tenantID, userID, updateReq)
	assert.NoError(t, err)
	assert.Equal(t, entities.InvoiceStatusSent, updated.Status)
}

func TestUpdateInvoiceStatus_Payed(t *testing.T) {
	db := setupInvoiceTestDB(t)
	service := services.NewInvoiceService(db)

	const tenantID uint = 1
	const userID uint = 1

	clientID, _, sessionIDs := createTestData(t, db, tenantID, userID)

	req := entities.CreateInvoiceRequest{
		ClientID:   clientID,
		SessionIDs: sessionIDs,
	}
	created, err := service.CreateInvoice(req, tenantID, userID)
	require.NoError(t, err)

	newStatus := entities.InvoiceStatusPaid
	updateReq := entities.UpdateInvoiceRequest{
		Status: &newStatus,
	}
	updated, err := service.UpdateInvoice(created.ID, tenantID, userID, updateReq)
	assert.NoError(t, err)
	assert.Equal(t, entities.InvoiceStatusPaid, updated.Status)
	assert.NotNil(t, updated.PayedDate)
}

func TestUpdateInvoiceStatus_Reminder(t *testing.T) {
	db := setupInvoiceTestDB(t)
	service := services.NewInvoiceService(db)

	const tenantID uint = 1
	const userID uint = 1

	clientID, _, sessionIDs := createTestData(t, db, tenantID, userID)

	req := entities.CreateInvoiceRequest{
		ClientID:   clientID,
		SessionIDs: sessionIDs,
	}
	created, err := service.CreateInvoice(req, tenantID, userID)
	require.NoError(t, err)

	newStatus := entities.InvoiceStatusReminder
	updateReq := entities.UpdateInvoiceRequest{
		Status: &newStatus,
	}
	updated, err := service.UpdateInvoice(created.ID, tenantID, userID, updateReq)
	assert.NoError(t, err)
	assert.Equal(t, entities.InvoiceStatusReminder, updated.Status)
	assert.Equal(t, 1, updated.NumReminders)
	assert.NotNil(t, updated.LatestReminder)
}

func TestUpdateInvoiceItems(t *testing.T) {
	db := setupInvoiceTestDB(t)
	service := services.NewInvoiceService(db)

	const tenantID uint = 1
	const userID uint = 1

	clientID, _, sessionIDs := createTestData(t, db, tenantID, userID)

	req := entities.CreateInvoiceRequest{
		ClientID:   clientID,
		SessionIDs: sessionIDs[:2],
	}
	created, err := service.CreateInvoice(req, tenantID, userID)
	require.NoError(t, err)
	assert.Len(t, created.InvoiceItems, 2)

	updateReq := entities.UpdateInvoiceRequest{
		SessionIDs: sessionIDs,
	}
	updated, err := service.UpdateInvoice(created.ID, tenantID, userID, updateReq)
	assert.NoError(t, err)
	assert.Len(t, updated.InvoiceItems, 3)
	assert.Equal(t, 3, updated.NumberUnits)
}

func TestDeleteInvoice(t *testing.T) {
	db := setupInvoiceTestDB(t)
	service := services.NewInvoiceService(db)

	const tenantID uint = 1
	const userID uint = 1

	clientID, _, sessionIDs := createTestData(t, db, tenantID, userID)

	req := entities.CreateInvoiceRequest{
		ClientID:   clientID,
		SessionIDs: sessionIDs,
	}
	created, err := service.CreateInvoice(req, tenantID, userID)
	require.NoError(t, err)

	err = service.DeleteInvoice(created.ID, tenantID, userID)
	assert.NoError(t, err)

	_, err = service.GetInvoiceByID(created.ID, tenantID, userID)
	assert.Error(t, err)

	// Check that invoice items are also deleted (only count non-soft-deleted)
	var count int64
	db.Model(&entities.InvoiceItem{}).Where("invoice_id = ? AND deleted_at IS NULL", created.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestGetClientsWithUnbilledSessions(t *testing.T) {
	db := setupInvoiceTestDB(t)
	service := services.NewInvoiceService(db)

	const tenantID uint = 1
	const userID uint = 1

	clientID, _, sessionIDs := createTestData(t, db, tenantID, userID)

	clients, err := service.GetClientsWithUnbilledSessions(tenantID, userID)
	assert.NoError(t, err)
	assert.Len(t, clients, 1)
	assert.Equal(t, clientID, clients[0].ID)
	assert.Len(t, clients[0].Sessions, 3)

	req := entities.CreateInvoiceRequest{
		ClientID:   clientID,
		SessionIDs: sessionIDs[:2],
	}
	_, err = service.CreateInvoice(req, tenantID, userID)
	require.NoError(t, err)

	clients, err = service.GetClientsWithUnbilledSessions(tenantID, userID)
	assert.NoError(t, err)
	assert.Len(t, clients, 1)
	assert.Len(t, clients[0].Sessions, 1)

	req2 := entities.CreateInvoiceRequest{
		ClientID:   clientID,
		SessionIDs: []uint{sessionIDs[2]},
	}
	_, err = service.CreateInvoice(req2, tenantID, userID)
	require.NoError(t, err)

	clients, err = service.GetClientsWithUnbilledSessions(tenantID, userID)
	assert.NoError(t, err)
	assert.Len(t, clients, 0)
}

func TestMultiTenantIsolation(t *testing.T) {
	db := setupInvoiceTestDB(t)
	service := services.NewInvoiceService(db)

	clientID1, _, sessionIDs1 := createTestData(t, db, 1, 1)
	req1 := entities.CreateInvoiceRequest{
		ClientID:   clientID1,
		SessionIDs: sessionIDs1,
	}
	invoice1, err := service.CreateInvoice(req1, 1, 1)
	require.NoError(t, err)

	clientID2, _, sessionIDs2 := createTestData(t, db, 2, 2)
	req2 := entities.CreateInvoiceRequest{
		ClientID:   clientID2,
		SessionIDs: sessionIDs2,
	}
	invoice2, err := service.CreateInvoice(req2, 2, 2)
	require.NoError(t, err)

	invoices, total, err := service.GetInvoices(1, 10, 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, invoices, 1)
	assert.Equal(t, invoice1.ID, invoices[0].ID)

	invoices, total, err = service.GetInvoices(1, 10, 2, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, invoices, 1)
	assert.Equal(t, invoice2.ID, invoices[0].ID)

	_, err = service.GetInvoiceByID(invoice2.ID, 1, 1)
	assert.Error(t, err)
}
