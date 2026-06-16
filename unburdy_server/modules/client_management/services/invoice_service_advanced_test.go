package services

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	settingsEntities "github.com/ae/base-server/pkg/settings/entities"
	baseAPI "github.com/ae/base-server/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var advInvoiceCounter int64

func advInvoiceNum() string {
	return fmt.Sprintf("ADV-%d", atomic.AddInt64(&advInvoiceCounter, 1))
}

func setupInvoiceAdvancedDB(t *testing.T) *gorm.DB {
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
		&settingsEntities.Setting{},
		&settingsEntities.SettingDefinition{},
	))
	return db
}

func createInvTestCostProvider(t *testing.T, db *gorm.DB, tenantID uint) *entities.CostProvider {
	t.Helper()
	cp := entities.CostProvider{
		TenantID:     tenantID,
		Organization: "Test Insurance",
	}
	require.NoError(t, db.Create(&cp).Error)
	return &cp
}

func createInvTestClient(t *testing.T, db *gorm.DB, tenantID uint, cpID *uint) *entities.Client {
	t.Helper()
	client := entities.Client{
		TenantID:       tenantID,
		FirstName:      "Test",
		LastName:       "Client",
		CostProviderID: cpID,
	}
	require.NoError(t, db.Create(&client).Error)
	return &client
}

func createInvTestOrg(t *testing.T, db *gorm.DB, tenantID uint) *baseAPI.Organization {
	t.Helper()
	org := baseAPI.Organization{TenantID: tenantID, Name: "Test Org"}
	require.NoError(t, db.Create(&org).Error)
	return &org
}

func createInvConductedSession(t *testing.T, db *gorm.DB, tenantID, clientID uint) *entities.Session {
	t.Helper()
	session := entities.Session{
		TenantID:          tenantID,
		ClientID:          clientID,
		Status:            "conducted",
		NumberUnits:       1,
		OriginalDate:      time.Now(),
		OriginalStartTime: time.Now(),
	}
	require.NoError(t, db.Create(&session).Error)
	return &session
}

// TestInvoiceService_CreateInvoice_WithSessions tests the legacy CreateInvoice function
func TestInvoiceService_CreateInvoice_WithSessions(t *testing.T) {
	db := setupInvoiceAdvancedDB(t)
	svc := NewInvoiceService(db)

	cp := createInvTestCostProvider(t, db, 1)
	cpID := cp.ID
	client := createInvTestClient(t, db, 1, &cpID)
	org := createInvTestOrg(t, db, 1)
	_ = org

	session := createInvConductedSession(t, db, 1, client.ID)

	t.Run("success creates invoice with sessions", func(t *testing.T) {
		req := entities.CreateInvoiceRequest{
			ClientID:   client.ID,
			SessionIDs: []uint{session.ID},
		}
		inv, err := svc.CreateInvoice(req, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, entities.InvoiceStatusDraft, inv.Status)
		assert.Equal(t, 1, inv.NumberUnits)
	})

	t.Run("client not found returns error", func(t *testing.T) {
		req := entities.CreateInvoiceRequest{
			ClientID:   9999,
			SessionIDs: []uint{session.ID},
		}
		_, err := svc.CreateInvoice(req, 1, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client not found")
	})

	t.Run("client without cost provider returns error", func(t *testing.T) {
		clientNoCp := createInvTestClient(t, db, 1, nil)
		req := entities.CreateInvoiceRequest{
			ClientID:   clientNoCp.ID,
			SessionIDs: []uint{session.ID},
		}
		_, err := svc.CreateInvoice(req, 1, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cost provider")
	})

	t.Run("no org for tenant returns error", func(t *testing.T) {
		// Tenant 99 has no org
		req := entities.CreateInvoiceRequest{
			ClientID:   client.ID,
			SessionIDs: []uint{session.ID},
		}
		_, err := svc.CreateInvoice(req, 99, 10)
		require.Error(t, err)
	})

	t.Run("sessions not found returns error", func(t *testing.T) {
		req := entities.CreateInvoiceRequest{
			ClientID:   client.ID,
			SessionIDs: []uint{99999},
		}
		_, err := svc.CreateInvoice(req, 1, 10)
		require.Error(t, err)
	})
}

// TestInvoiceService_CreateDraftInvoice tests CreateDraftInvoice with custom line items (no sessions)
func TestInvoiceService_CreateDraftInvoice(t *testing.T) {
	db := setupInvoiceAdvancedDB(t)
	svc := NewInvoiceService(db)

	cp := createInvTestCostProvider(t, db, 1)
	cpID := cp.ID
	client := createInvTestClient(t, db, 1, &cpID)
	_ = createInvTestOrg(t, db, 1)

	t.Run("creates draft with custom line items", func(t *testing.T) {
		req := entities.CreateDraftInvoiceRequest{
			ClientID: client.ID,
			CustomLineItems: []entities.CustomLineItemRequest{
				{
					Description:      "Consultation",
					NumberUnits:      2,
					UnitPrice:        100.0,
					VATCategory:      "taxable_standard",
					VATRate:          19.0,
					VATExempt:        false,
				},
			},
		}
		inv, err := svc.CreateDraftInvoice(req, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, entities.InvoiceStatusDraft, inv.Status)
		assert.NotZero(t, inv.ID)
	})

	t.Run("no items returns error", func(t *testing.T) {
		req := entities.CreateDraftInvoiceRequest{
			ClientID: client.ID,
		}
		_, err := svc.CreateDraftInvoice(req, 1, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one")
	})

	t.Run("client without cost provider returns error", func(t *testing.T) {
		clientNoCp := createInvTestClient(t, db, 1, nil)
		req := entities.CreateDraftInvoiceRequest{
			ClientID: clientNoCp.ID,
			CustomLineItems: []entities.CustomLineItemRequest{
				{Description: "Item", NumberUnits: 1, UnitPrice: 50.0, VATRate: 19},
			},
		}
		_, err := svc.CreateDraftInvoice(req, 1, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cost provider")
	})

	t.Run("client not found returns error", func(t *testing.T) {
		req := entities.CreateDraftInvoiceRequest{
			ClientID: 9999,
			CustomLineItems: []entities.CustomLineItemRequest{
				{Description: "Item", NumberUnits: 1, UnitPrice: 50.0, VATRate: 19},
			},
		}
		_, err := svc.CreateDraftInvoice(req, 1, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client not found")
	})
}

// TestInvoiceService_CreateDraftInvoice_WithSessions tests CreateDraftInvoice with conducted sessions
func TestInvoiceService_CreateDraftInvoice_WithSessions(t *testing.T) {
	db := setupInvoiceAdvancedDB(t)
	svc := NewInvoiceService(db)

	cp := createInvTestCostProvider(t, db, 1)
	cpID := cp.ID
	client := createInvTestClient(t, db, 1, &cpID)
	_ = createInvTestOrg(t, db, 1)

	session := createInvConductedSession(t, db, 1, client.ID)

	t.Run("creates draft from sessions", func(t *testing.T) {
		req := entities.CreateDraftInvoiceRequest{
			ClientID:   client.ID,
			SessionIDs: []uint{session.ID},
		}
		inv, err := svc.CreateDraftInvoice(req, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, entities.InvoiceStatusDraft, inv.Status)
	})

	t.Run("sessions not in conducted status returns error", func(t *testing.T) {
		// Create a session that's not in conducted status
		pendingSession := entities.Session{
			TenantID:          1,
			ClientID:          client.ID,
			Status:            "scheduled",
			NumberUnits:       1,
			OriginalDate:      time.Now(),
			OriginalStartTime: time.Now(),
		}
		require.NoError(t, db.Create(&pendingSession).Error)

		req := entities.CreateDraftInvoiceRequest{
			ClientID:   client.ID,
			SessionIDs: []uint{pendingSession.ID},
		}
		_, err := svc.CreateDraftInvoice(req, 1, 10)
		require.Error(t, err)
	})
}

// TestInvoiceService_FinalizeInvoice tests FinalizeInvoice
func TestInvoiceService_FinalizeInvoice(t *testing.T) {
	db := setupInvoiceAdvancedDB(t)
	svc := NewInvoiceService(db)

	cp := createInvTestCostProvider(t, db, 1)
	cpID := cp.ID
	client := createInvTestClient(t, db, 1, &cpID)
	org := createInvTestOrg(t, db, 1)

	t.Run("finalizes draft with valid VAT items", func(t *testing.T) {
		// Create draft invoice directly
		inv := entities.Invoice{
			TenantID:       1,
			UserID:         10,
			OrganizationID: org.ID,
			InvoiceDate:    time.Now(),
			InvoiceNumber:  advInvoiceNum(),
			Status:         entities.InvoiceStatusDraft,
		}
		require.NoError(t, db.Create(&inv).Error)

		// Add invoice item with valid VAT (taxable with rate)
		item := entities.InvoiceItem{
			InvoiceID:   inv.ID,
			Description: "Service",
			NumberUnits: 1,
			UnitPrice:   100.0,
			VATRate:     19.0,
			VATExempt:   false,
		}
		require.NoError(t, db.Create(&item).Error)

		// Link client invoice (for preload)
		ci := entities.ClientInvoice{
			InvoiceID:      inv.ID,
			ClientID:       client.ID,
			CostProviderID: cp.ID,
		}
		require.NoError(t, db.Create(&ci).Error)

		result, err := svc.FinalizeInvoice(inv.ID, 1, 10, nil)
		require.NoError(t, err)
		assert.Equal(t, entities.InvoiceStatusFinalized, result.Status)
		assert.NotEmpty(t, result.InvoiceNumber)
	})

	t.Run("cannot finalize non-draft invoice", func(t *testing.T) {
		inv := entities.Invoice{
			TenantID:       1,
			UserID:         10,
			OrganizationID: org.ID,
			InvoiceDate:    time.Now(),
			InvoiceNumber:  "TEST-SENT-001",
			Status:         entities.InvoiceStatusSent,
		}
		require.NoError(t, db.Create(&inv).Error)
		item := entities.InvoiceItem{
			InvoiceID:   inv.ID,
			Description: "Service",
			VATRate:     19.0,
		}
		require.NoError(t, db.Create(&item).Error)

		_, err := svc.FinalizeInvoice(inv.ID, 1, 10, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "draft")
	})

	t.Run("cannot finalize invoice without items", func(t *testing.T) {
		inv := entities.Invoice{
			TenantID:       1,
			UserID:         10,
			OrganizationID: org.ID,
			InvoiceDate:    time.Now(),
			InvoiceNumber:  advInvoiceNum(),
			Status:         entities.InvoiceStatusDraft,
		}
		require.NoError(t, db.Create(&inv).Error)

		_, err := svc.FinalizeInvoice(inv.ID, 1, 10, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "line item")
	})

	t.Run("VAT exempt without exemption text fails", func(t *testing.T) {
		inv := entities.Invoice{
			TenantID:       1,
			UserID:         10,
			OrganizationID: org.ID,
			InvoiceDate:    time.Now(),
			InvoiceNumber:  advInvoiceNum(),
			Status:         entities.InvoiceStatusDraft,
		}
		require.NoError(t, db.Create(&inv).Error)

		item := entities.InvoiceItem{
			InvoiceID:   inv.ID,
			Description: "Exempt Service",
			VATExempt:   true,
			// Missing VATExemptionText - should fail validation
		}
		require.NoError(t, db.Create(&item).Error)

		_, err := svc.FinalizeInvoice(inv.ID, 1, 10, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exemption text")
	})

	t.Run("not found returns error", func(t *testing.T) {
		_, err := svc.FinalizeInvoice(9999, 1, 10, nil)
		require.Error(t, err)
	})
}

// TestInvoiceService_CreateCreditNote tests CreateCreditNote
func TestInvoiceService_CreateCreditNote(t *testing.T) {
	db := setupInvoiceAdvancedDB(t)
	svc := NewInvoiceService(db)

	cp := createInvTestCostProvider(t, db, 1)
	cpID := cp.ID
	client := createInvTestClient(t, db, 1, &cpID)
	org := createInvTestOrg(t, db, 1)

	t.Run("creates credit note for finalized invoice", func(t *testing.T) {
		// Create sent invoice (credit notes require sent/paid/overdue)
		inv := entities.Invoice{
			TenantID:       1,
			UserID:         10,
			OrganizationID: org.ID,
			InvoiceDate:    time.Now(),
			InvoiceNumber:  advInvoiceNum(),
			Status:         entities.InvoiceStatusSent,
		}
		require.NoError(t, db.Create(&inv).Error)

		item := entities.InvoiceItem{
			InvoiceID:   inv.ID,
			Description: "Service",
			NumberUnits: 1,
			UnitPrice:   100.0,
			VATRate:     19.0,
		}
		require.NoError(t, db.Create(&item).Error)

		ci := entities.ClientInvoice{
			InvoiceID:      inv.ID,
			ClientID:       client.ID,
			CostProviderID: cp.ID,
		}
		require.NoError(t, db.Create(&ci).Error)

		req := entities.CreateCreditNoteRequest{
			LineItemIDs: []uint{item.ID},
			Reason:      "Customer dispute",
		}
		creditNote, err := svc.CreateCreditNote(inv.ID, 1, 10, req)
		require.NoError(t, err)
		assert.NotNil(t, creditNote)
		assert.Equal(t, entities.InvoiceStatusSent, creditNote.Status) // credit notes are immediately sent
	})

	t.Run("invoice not found returns error", func(t *testing.T) {
		req := entities.CreateCreditNoteRequest{
			LineItemIDs: []uint{1},
			Reason:      "Error",
		}
		_, err := svc.CreateCreditNote(9999, 1, 10, req)
		require.Error(t, err)
	})

	t.Run("non-finalized invoice returns error", func(t *testing.T) {
		inv := entities.Invoice{
			TenantID:       1,
			UserID:         10,
			OrganizationID: org.ID,
			InvoiceDate:    time.Now(),
			InvoiceNumber:  advInvoiceNum(),
			Status:         entities.InvoiceStatusDraft,
		}
		require.NoError(t, db.Create(&inv).Error)

		req := entities.CreateCreditNoteRequest{
			LineItemIDs: []uint{1},
			Reason:      "Error",
		}
		_, err := svc.CreateCreditNote(inv.ID, 1, 10, req)
		require.Error(t, err)
	})
}

// TestInvoiceService_UpdateDraftInvoice tests UpdateDraftInvoice
func TestInvoiceService_UpdateDraftInvoice(t *testing.T) {
	db := setupInvoiceAdvancedDB(t)
	svc := NewInvoiceService(db)

	cp := createInvTestCostProvider(t, db, 1)
	cpID := cp.ID
	_ = createInvTestClient(t, db, 1, &cpID)
	org := createInvTestOrg(t, db, 1)

	t.Run("cannot update non-draft invoice", func(t *testing.T) {
		inv := entities.Invoice{
			TenantID:       1,
			UserID:         10,
			OrganizationID: org.ID,
			InvoiceDate:    time.Now(),
			InvoiceNumber:  "INV-SENT-UPDATE",
			Status:         entities.InvoiceStatusSent,
		}
		require.NoError(t, db.Create(&inv).Error)

		req := entities.UpdateDraftInvoiceRequest{
			CustomLineItems: []entities.CustomLineItemRequest{
				{Description: "New", NumberUnits: 1, UnitPrice: 50},
			},
		}
		_, err := svc.UpdateDraftInvoice(inv.ID, 1, 10, req)
		require.Error(t, err)
	})

	t.Run("not found returns error", func(t *testing.T) {
		req := entities.UpdateDraftInvoiceRequest{}
		_, err := svc.UpdateDraftInvoice(9999, 1, 10, req)
		require.Error(t, err)
	})
}

// TestInvoiceService_CancelClientInvoice tests CancelClientInvoice
func TestInvoiceService_CancelClientInvoice(t *testing.T) {
	db := setupInvoiceAdvancedDB(t)
	svc := NewInvoiceService(db)

	org := createInvTestOrg(t, db, 1)

	t.Run("cancels finalized invoice", func(t *testing.T) {
		inv := entities.Invoice{
			TenantID:       1,
			UserID:         10,
			OrganizationID: org.ID,
			InvoiceDate:    time.Now(),
			InvoiceNumber:  "INV-CLI-CANCEL-001",
			Status:         entities.InvoiceStatusFinalized,
		}
		require.NoError(t, db.Create(&inv).Error)

		err := svc.CancelClientInvoice(inv.ID, 1, 10, "Test reason")
		require.NoError(t, err)

		var check entities.Invoice
		db.First(&check, inv.ID)
		assert.Equal(t, entities.InvoiceStatusCancelled, check.Status)
	})

	t.Run("not found returns error", func(t *testing.T) {
		err := svc.CancelClientInvoice(9999, 1, 10, "reason")
		require.Error(t, err)
	})
}
func TestInvoiceService_CreateDraftInvoice_WithCustomerName(t *testing.T) {
        db := setupInvoiceAdvancedDB(t)
        svc := NewInvoiceService(db)

        cp := createInvTestCostProvider(t, db, 1)
        _ = createInvTestOrg(t, db, 1) // required by CreateDraftInvoice
        client := createInvTestClient(t, db, 1, &cp.ID)

        // Passing CustomerName directly in request covers the "Priority 1" branch (lines 300-310)
        req := entities.CreateDraftInvoiceRequest{
                ClientID:              client.ID,
                CustomerName:          "Direct Corp",
                CustomerAddress:       "123 Direct St",
                CustomerAddressExt:    "Suite 1",
                CustomerZip:           "10001",
                CustomerCity:          "Berlin",
                CustomerCountry:       "DE",
                CustomerContactPerson: "John Doe",
                CustomerDepartment:    "Finance",
                CustomerEmail:         "direct@corp.de",
                CustomLineItems: []entities.CustomLineItemRequest{
                        {Description: "Custom Item", NumberUnits: 1, UnitPrice: 100, VATExempt: true, VATExemptionText: "Exempt"},
                },
        }

        inv, err := svc.CreateDraftInvoice(req, 1, 10)
        require.NoError(t, err)
        require.NotNil(t, inv)
        assert.Equal(t, "Direct Corp", inv.CustomerName)
        assert.Equal(t, "123 Direct St", inv.CustomerAddress)
        assert.Equal(t, "10001", inv.CustomerZip)
}

func TestInvoiceService_CreateDraftInvoice_ClientContactFallback(t *testing.T) {
        db := setupInvoiceAdvancedDB(t)
        svc := NewInvoiceService(db)

        // Cost provider with NO organization name — triggers Priority 3 (client contact fallback)
        emptyCP := entities.CostProvider{TenantID: 1, Organization: ""}
        require.NoError(t, db.Create(&emptyCP).Error)
        _ = createInvTestOrg(t, db, 1) // required by CreateDraftInvoice

        // Create a client with contact name fields and contact email
        client := entities.Client{
                TenantID:         1,
                FirstName:        "Alice",
                LastName:         "Fallback",
                Email:            "alice@example.com",
                ContactFirstName: "Contact",
                ContactLastName:  "Person",
                ContactEmail:     "contact@example.com",
                StreetAddress:    "Main St 1",
                Zip:              "12345",
                City:             "Berlin",
                CostProviderID:   &emptyCP.ID,
        }
        require.NoError(t, db.Create(&client).Error)

        // No CustomerName in request — triggers Priority 3 (client contact fallback)
        req := entities.CreateDraftInvoiceRequest{
                ClientID: client.ID,
                CustomLineItems: []entities.CustomLineItemRequest{
                        {Description: "Item", NumberUnits: 1, UnitPrice: 50, VATExempt: true, VATExemptionText: "Exempt"},
                },
        }

        inv, err := svc.CreateDraftInvoice(req, 1, 10)
        require.NoError(t, err)
        require.NotNil(t, inv)
        // Should fall back to client name
        assert.Equal(t, "Alice Fallback", inv.CustomerName)
}

func TestInvoiceService_CreateDraftInvoice_WithExtraEfforts(t *testing.T) {
	db := setupInvoiceAdvancedDB(t)
	svc := NewInvoiceService(db)

	cp := createInvTestCostProvider(t, db, 1)
	cpID := cp.ID
	_ = createInvTestOrg(t, db, 1)

	client := entities.Client{
		TenantID:       1,
		FirstName:      "Bob",
		LastName:       "Effort",
		CostProviderID: &cpID,
	}
	require.NoError(t, db.Create(&client).Error)

	// Create a billable delivered extra effort
	effort := entities.ExtraEffort{
		TenantID:      1,
		ClientID:      client.ID,
		EffortType:    "documentation",
		EffortDate:    time.Now(),
		DurationMin:   30,
		Description:   "Patient notes",
		Billable:      true,
		BillingStatus: "delivered",
	}
	require.NoError(t, db.Create(&effort).Error)

	req := entities.CreateDraftInvoiceRequest{
		ClientID:       client.ID,
		ExtraEffortIDs: []uint{effort.ID},
	}

	inv, err := svc.CreateDraftInvoice(req, 1, 10)
	require.NoError(t, err)
	require.NotNil(t, inv)
	assert.Equal(t, entities.InvoiceStatusDraft, inv.Status)
	// Invoice should have one extra_effort line item
	assert.GreaterOrEqual(t, len(inv.InvoiceItems), 1)
}

func TestInvoiceService_CreateDraftInvoice_InvalidVATCategory(t *testing.T) {
	db := setupInvoiceAdvancedDB(t)
	svc := NewInvoiceService(db)

	cp := createInvTestCostProvider(t, db, 1)
	cpID := cp.ID
	_ = createInvTestOrg(t, db, 1)

	client := entities.Client{
		TenantID:       1,
		FirstName:      "Carl",
		LastName:       "VATTest",
		CostProviderID: &cpID,
	}
	require.NoError(t, db.Create(&client).Error)

	// Invalid VAT category triggers error-fallback path in CreateDraftInvoice
	req := entities.CreateDraftInvoiceRequest{
		ClientID: client.ID,
		CustomLineItems: []entities.CustomLineItemRequest{
			{
				Description: "Test item",
				NumberUnits: 2,
				UnitPrice:   80,
				VATCategory: "INVALID_CATEGORY", // triggers ApplyVATCategory error fallback
				VATRate:     7.0,
			},
		},
	}

	inv, err := svc.CreateDraftInvoice(req, 1, 10)
	require.NoError(t, err)
	require.NotNil(t, inv)
	// Should succeed with fallback to manual VAT settings (VATRate 7.0)
	require.GreaterOrEqual(t, len(inv.InvoiceItems), 1)
	assert.Equal(t, 7.0, inv.InvoiceItems[0].VATRate)
}

func TestInvoiceService_UpdateDraftInvoice_AddSessions(t *testing.T) {
	db := setupInvoiceAdvancedDB(t)
	svc := NewInvoiceService(db)

	cp := createInvTestCostProvider(t, db, 1)
	cpID := cp.ID
	_ = createInvTestOrg(t, db, 1)

	client := entities.Client{
		TenantID:       1,
		FirstName:      "Dave",
		LastName:       "AddSession",
		CostProviderID: &cpID,
	}
	require.NoError(t, db.Create(&client).Error)

	// Create a draft invoice with a custom line item (to have an existing item)
	initialReq := entities.CreateDraftInvoiceRequest{
		ClientID: client.ID,
		CustomLineItems: []entities.CustomLineItemRequest{
			{Description: "Initial item", NumberUnits: 1, UnitPrice: 50, VATExempt: true},
		},
	}
	inv, err := svc.CreateDraftInvoice(initialReq, 1, 10)
	require.NoError(t, err)

	// Create a conducted session to add
	addSession := entities.Session{
		TenantID:          1,
		ClientID:          client.ID,
		Status:            "conducted",
		NumberUnits:       1,
		DurationMin:       60,
		OriginalDate:      time.Now(),
		OriginalStartTime: time.Now(),
	}
	require.NoError(t, db.Create(&addSession).Error)

	updateReq := entities.UpdateDraftInvoiceRequest{
		AddSessionIDs: []uint{addSession.ID},
	}

	updated, err := svc.UpdateDraftInvoice(inv.ID, 1, 10, updateReq)
	require.NoError(t, err)
	require.NotNil(t, updated)
	// Should now have at least 2 items (original + added session)
	assert.GreaterOrEqual(t, len(updated.InvoiceItems), 2)
}

func TestInvoiceService_UpdateDraftInvoice_AddExtraEfforts(t *testing.T) {
	db := setupInvoiceAdvancedDB(t)
	svc := NewInvoiceService(db)

	cp := createInvTestCostProvider(t, db, 1)
	cpID := cp.ID
	_ = createInvTestOrg(t, db, 1)

	client := entities.Client{
		TenantID:       1,
		FirstName:      "Eve",
		LastName:       "AddEffort",
		CostProviderID: &cpID,
	}
	require.NoError(t, db.Create(&client).Error)

	// Create a draft invoice with a custom line item (to have an existing item)
	initialReq := entities.CreateDraftInvoiceRequest{
		ClientID: client.ID,
		CustomLineItems: []entities.CustomLineItemRequest{
			{Description: "Initial item", NumberUnits: 1, UnitPrice: 50, VATExempt: true},
		},
	}
	inv, err := svc.CreateDraftInvoice(initialReq, 1, 10)
	require.NoError(t, err)

	// Create a billable delivered effort to add
	addEffort := entities.ExtraEffort{
		TenantID:      1,
		ClientID:      client.ID,
		EffortType:    "preparation",
		EffortDate:    time.Now(),
		DurationMin:   20,
		Description:   "Prep work",
		Billable:      true,
		BillingStatus: "delivered",
	}
	require.NoError(t, db.Create(&addEffort).Error)

	updateReq := entities.UpdateDraftInvoiceRequest{
		AddExtraEffortIDs: []uint{addEffort.ID},
	}

	updated, err := svc.UpdateDraftInvoice(inv.ID, 1, 10, updateReq)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.GreaterOrEqual(t, len(updated.InvoiceItems), 2)
}

func TestInvoiceService_UpdateDraftInvoice_CustomLineItems(t *testing.T) {
	db := setupInvoiceAdvancedDB(t)
	svc := NewInvoiceService(db)

	cp := createInvTestCostProvider(t, db, 1)
	cpID := cp.ID
	_ = createInvTestOrg(t, db, 1)

	client := entities.Client{
		TenantID:       1,
		FirstName:      "Frank",
		LastName:       "CustomLine",
		CostProviderID: &cpID,
	}
	require.NoError(t, db.Create(&client).Error)

	// Create draft invoice
	initialReq := entities.CreateDraftInvoiceRequest{
		ClientID: client.ID,
		CustomLineItems: []entities.CustomLineItemRequest{
			{Description: "Old item", NumberUnits: 1, UnitPrice: 100, VATExempt: true},
		},
	}
	inv, err := svc.CreateDraftInvoice(initialReq, 1, 10)
	require.NoError(t, err)

	// Replace with new custom line items (including invalid VAT category fallback path)
	updateReq := entities.UpdateDraftInvoiceRequest{
		CustomLineItems: []entities.CustomLineItemRequest{
			{Description: "New item A", NumberUnits: 2, UnitPrice: 60, VATRate: 19.0},
			{Description: "New item B", NumberUnits: 1, UnitPrice: 30, VATCategory: "INVALID_CATEGORY", VATRate: 7.0},
		},
	}

	updated, err := svc.UpdateDraftInvoice(inv.ID, 1, 10, updateReq)
	require.NoError(t, err)
	require.NotNil(t, updated)
	// Should have exactly 2 new custom items
	assert.Equal(t, 2, len(updated.InvoiceItems))
}

func TestInvoiceService_CreateInvoiceDirect(t *testing.T) {
	db := setupInvoiceAdvancedDB(t)
	// These columns are added via PostgreSQL migrations; add them manually for SQLite
	db.Exec("ALTER TABLE organizations ADD COLUMN extra_efforts_billing_mode TEXT DEFAULT ''")
	db.Exec("ALTER TABLE organizations ADD COLUMN extra_efforts_config BLOB")
	db.Exec("ALTER TABLE organizations ADD COLUMN line_item_single_unit_text TEXT DEFAULT ''")
	db.Exec("ALTER TABLE organizations ADD COLUMN line_item_double_unit_text TEXT DEFAULT ''")
	svc := NewInvoiceService(db)

	cp := createInvTestCostProvider(t, db, 1)
	cpID := cp.ID
	_ = createInvTestOrg(t, db, 1)

	client := entities.Client{
		TenantID:       1,
		FirstName:      "Grace",
		LastName:       "Direct",
		CostProviderID: &cpID,
	}
	require.NoError(t, db.Create(&client).Error)

	session := createInvConductedSession(t, db, 1, client.ID)

	t.Run("success creates invoice from unbilled client data", func(t *testing.T) {
		req := entities.CreateInvoiceDirectRequest{
			UnbilledClient: entities.ClientWithUnbilledSessionsResponse{
				ClientResponse: entities.ClientResponse{
					ID:       client.ID,
					TenantID: 1,
				},
				Sessions: []entities.SessionResponse{
					{ID: session.ID, TenantID: 1, ClientID: client.ID},
				},
			},
			Parameters: entities.InvoiceGenerationParameters{
				InvoiceNumber: advInvoiceNum(),
				InvoiceDate:   "2024-06-01",
			},
		}
		inv, err := svc.CreateInvoiceDirect(req, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, inv)
		assert.Equal(t, entities.InvoiceStatusDraft, inv.Status)
	})

	t.Run("client not found returns error", func(t *testing.T) {
		req := entities.CreateInvoiceDirectRequest{
			UnbilledClient: entities.ClientWithUnbilledSessionsResponse{
				ClientResponse: entities.ClientResponse{
					ID:       9999,
					TenantID: 1,
				},
			},
		}
		_, err := svc.CreateInvoiceDirect(req, 1, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client not found")
	})

	t.Run("invalid invoice date returns error", func(t *testing.T) {
		req := entities.CreateInvoiceDirectRequest{
			UnbilledClient: entities.ClientWithUnbilledSessionsResponse{
				ClientResponse: entities.ClientResponse{ID: client.ID, TenantID: 1},
			},
			Parameters: entities.InvoiceGenerationParameters{
				InvoiceDate: "not-a-date",
			},
		}
		_, err := svc.CreateInvoiceDirect(req, 1, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid invoice date format")
	})

	t.Run("session not found returns error", func(t *testing.T) {
		req := entities.CreateInvoiceDirectRequest{
			UnbilledClient: entities.ClientWithUnbilledSessionsResponse{
				ClientResponse: entities.ClientResponse{ID: client.ID, TenantID: 1},
				Sessions: []entities.SessionResponse{
					{ID: 9999, TenantID: 1, ClientID: client.ID},
				},
			},
		}
		_, err := svc.CreateInvoiceDirect(req, 1, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestInvoiceService_UpdateInvoice_WithSessions(t *testing.T) {
	db := setupInvoiceAdvancedDB(t)
	svc := NewInvoiceService(db)

	cp := createInvTestCostProvider(t, db, 1)
	cpID := cp.ID
	_ = createInvTestOrg(t, db, 1)

	client := entities.Client{
		TenantID:       1,
		FirstName:      "Henry",
		LastName:       "Update",
		CostProviderID: &cpID,
	}
	require.NoError(t, db.Create(&client).Error)

	session1 := createInvConductedSession(t, db, 1, client.ID)

	// Create an invoice with session1 via CreateInvoice
	createReq := entities.CreateInvoiceRequest{
		ClientID:   client.ID,
		SessionIDs: []uint{session1.ID},
	}
	inv, err := svc.CreateInvoice(createReq, 1, 10)
	require.NoError(t, err)

	// Create another session to replace with
	session2 := createInvConductedSession(t, db, 1, client.ID)

	// UpdateInvoice with new sessions (replaces existing items)
	status := inv.Status
	updateReq := entities.UpdateInvoiceRequest{
		Status:     &status,
		SessionIDs: nil,
	}
	// First test: update status only (no session IDs)
	updated, err := svc.UpdateInvoice(inv.ID, 1, 10, updateReq)
	require.NoError(t, err)
	assert.Equal(t, inv.Status, updated.Status)

	// Second test: update with new session IDs
	updateReq2 := entities.UpdateInvoiceRequest{
		SessionIDs: []uint{session2.ID},
	}
	updated2, err := svc.UpdateInvoice(inv.ID, 1, 10, updateReq2)
	require.NoError(t, err)
	require.NotNil(t, updated2)
}
