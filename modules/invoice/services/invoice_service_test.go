package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/invoice-module/entities"
	"gorm.io/gorm"
)

func TestCreateInvoice(t *testing.T) {
	db := setupTestDB(t)
	service := &InvoiceService{db: db}
	ctx := context.Background()

	tests := []struct {
		name    string
		request *entities.CreateInvoiceRequest
		wantErr bool
	}{
		{
			name: "success - create draft invoice",
			request: &entities.CreateInvoiceRequest{
				CustomerName:    "Test Customer Inc.",
				CustomerAddress: "123 Main St, Test City",
				CustomerEmail:   ptr("customer@test.com"),
				InvoiceDate:     time.Now(),
				DueDate:         ptr(time.Now().Add(30 * 24 * time.Hour)),
				Currency:        "EUR",
				Items: []entities.InvoiceItemRequest{
					{
						Description: "Consulting Services",
						Quantity:    10,
						UnitPrice:   150.0,
						TaxRate:     19.0,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success - invoice with all fields",
			request: &entities.CreateInvoiceRequest{
				CustomerName:           "Complete Customer GmbH",
				CustomerAddress:        "456 Business Ave, Commerce City",
				CustomerEmail:          ptr("accounting@complete.com"),
				CustomerTaxID:          ptr("DE123456789"),
				CustomerContactPerson:  ptr("John Doe"),
				CustomerDepartment:     ptr("Finance"),
				Subject:                ptr("Q4 2025 Consulting Services"),
				OurReference:           ptr("PROJ-2025-042"),
				YourReference:          ptr("PO-98765"),
				PONumber:               ptr("PO-98765"),
				InvoiceDate:            time.Now(),
				DueDate:                ptr(time.Now().Add(30 * 24 * time.Hour)),
				DeliveryDate:           ptr(time.Now().Add(-7 * 24 * time.Hour)),
				PerformancePeriodStart: ptr(time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)),
				PerformancePeriodEnd:   ptr(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)),
				NetTerms:               ptr(30),
				DiscountRate:           5.0,
				DiscountTerms:          ptr("2% if paid within 10 days"),
				Currency:               "EUR",
				Items: []entities.InvoiceItemRequest{
					{
						Description: "Senior Consultant - 100 hours",
						Quantity:    100,
						UnitPrice:   150.0,
						TaxRate:     19.0,
					},
					{
						Description: "Junior Consultant - 50 hours",
						Quantity:    50,
						UnitPrice:   80.0,
						TaxRate:     19.0,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success - VAT exempt invoice",
			request: &entities.CreateInvoiceRequest{
				CustomerName:    "EU Customer Ltd",
				CustomerAddress: "789 Euro Street, Brussels",
				CustomerTaxID:   ptr("BE987654321"),
				InvoiceDate:     time.Now(),
				Currency:        "EUR",
				Items: []entities.InvoiceItemRequest{
					{
						Description:      "International Consulting",
						Quantity:         1,
						UnitPrice:        5000.0,
						TaxRate:          0.0,
						VATExempt:        ptr(true),
						VATExemptionText: ptr("Reverse charge - Article 196 EU VAT Directive"),
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CreateInvoice(ctx, 1, 1, 1, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, entities.InvoiceStatusDraft, result.Status)
			assert.Equal(t, tt.request.CustomerName, result.CustomerName)
			assert.Equal(t, tt.request.Currency, result.Currency)
			assert.NotEmpty(t, result.Items)
			assert.Len(t, result.Items, len(tt.request.Items))

			// Verify calculated amounts
			assert.Greater(t, result.TotalAmount, 0.0)
			assert.GreaterOrEqual(t, result.TotalAmount, result.SubtotalAmount)

			// Verify optional fields if provided
			if tt.request.Subject != nil {
				assert.Equal(t, *tt.request.Subject, result.Subject)
			}
			if tt.request.OurReference != nil {
				assert.Equal(t, *tt.request.OurReference, result.OurReference)
			}
		})
	}
}

func TestUpdateInvoice(t *testing.T) {
	db := setupTestDB(t)
	service := &InvoiceService{db: db}
	ctx := context.Background()

	// Create initial invoice
	invoice := createTestInvoice(t, db, entities.InvoiceStatusDraft, true)

	tests := []struct {
		name        string
		request     *entities.UpdateInvoiceRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "success - update customer info",
			request: &entities.UpdateInvoiceRequest{
				CustomerName:    ptr("Updated Customer Name"),
				CustomerAddress: ptr("New Address 123"),
			},
			wantErr: false,
		},
		{
			name: "success - update references",
			request: &entities.UpdateInvoiceRequest{
				Subject:       ptr("Updated Subject"),
				OurReference:  ptr("REF-NEW-001"),
				YourReference: ptr("YOUR-REF-NEW"),
			},
			wantErr: false,
		},
		{
			name: "success - update dates",
			request: &entities.UpdateInvoiceRequest{
				DueDate:      ptr(time.Now().Add(45 * 24 * time.Hour)),
				DeliveryDate: ptr(time.Now().Add(-1 * 24 * time.Hour)),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.UpdateInvoice(ctx, invoice.TenantID, invoice.ID, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)

			// Verify updates
			if tt.request.CustomerName != nil {
				assert.Equal(t, *tt.request.CustomerName, result.CustomerName)
			}
			if tt.request.Subject != nil {
				assert.Equal(t, *tt.request.Subject, result.Subject)
			}
		})
	}
}

func TestDeleteInvoice(t *testing.T) {
	tests := []struct {
		name        string
		status      entities.InvoiceStatus
		wantErr     bool
		errContains string
	}{
		{
			name:    "success - delete draft",
			status:  entities.InvoiceStatusDraft,
			wantErr: false,
		},
		{
			name:        "error - cannot delete finalized",
			status:      entities.InvoiceStatusFinalized,
			wantErr:     true,
			errContains: "cannot delete finalized invoices",
		},
		{
			name:        "error - cannot delete sent",
			status:      entities.InvoiceStatusSent,
			wantErr:     true,
			errContains: "cannot delete finalized invoices",
		},
		{
			name:        "error - cannot delete paid",
			status:      entities.InvoiceStatusPaid,
			wantErr:     true,
			errContains: "cannot delete finalized invoices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			service := &InvoiceService{db: db}
			ctx := context.Background()

			invoice := createTestInvoice(t, db, tt.status, true)

			err := service.DeleteInvoice(ctx, invoice.TenantID, invoice.ID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)

				// Verify invoice still exists
				var found entities.Invoice
				err := db.First(&found, invoice.ID).Error
				assert.NoError(t, err)
				return
			}

			require.NoError(t, err)

			// Verify invoice is deleted
			var found entities.Invoice
			err = db.First(&found, invoice.ID).Error
			assert.Error(t, err)
			assert.Equal(t, gorm.ErrRecordNotFound, err)
		})
	}
}

func TestGetInvoice(t *testing.T) {
	db := setupTestDB(t)
	service := &InvoiceService{db: db}
	ctx := context.Background()

	invoice := createTestInvoice(t, db, entities.InvoiceStatusDraft, true)

	tests := []struct {
		name        string
		tenantID    uint
		invoiceID   uint
		wantErr     bool
		errContains string
	}{
		{
			name:      "success - get existing invoice",
			tenantID:  invoice.TenantID,
			invoiceID: invoice.ID,
			wantErr:   false,
		},
		{
			name:        "error - wrong tenant",
			tenantID:    999,
			invoiceID:   invoice.ID,
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:        "error - non-existent invoice",
			tenantID:    invoice.TenantID,
			invoiceID:   99999,
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetInvoice(ctx, tt.tenantID, tt.invoiceID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, invoice.ID, result.ID)
			assert.Equal(t, invoice.CustomerName, result.CustomerName)
		})
	}
}

func TestListInvoices(t *testing.T) {
	db := setupTestDB(t)
	service := &InvoiceService{db: db}
	ctx := context.Background()

	// Create test invoices
	createTestInvoice(t, db, entities.InvoiceStatusDraft, true)
	createTestInvoice(t, db, entities.InvoiceStatusFinalized, true)
	sent := createTestInvoice(t, db, entities.InvoiceStatusSent, true)
	createTestInvoice(t, db, entities.InvoiceStatusPaid, true)

	// Create invoice for different tenant
	other := createTestInvoice(t, db, entities.InvoiceStatusDraft, true)
	other.TenantID = 999
	db.Save(other)

	tests := []struct {
		name          string
		tenantID      uint
		status        *entities.InvoiceStatus
		expectedCount int
	}{
		{
			name:          "all invoices for tenant",
			tenantID:      1,
			status:        nil,
			expectedCount: 4,
		},
		{
			name:          "filter by status - sent",
			tenantID:      1,
			status:        &sent.Status,
			expectedCount: 1,
		},
		{
			name:          "different tenant",
			tenantID:      999,
			status:        nil,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ListInvoices(ctx, tt.tenantID, tt.status, 1, 100)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Len(t, result, tt.expectedCount)

			// Verify all invoices belong to the correct tenant
			for _, inv := range result {
				assert.Equal(t, tt.tenantID, inv.TenantID)
				if tt.status != nil {
					assert.Equal(t, *tt.status, inv.Status)
				}
			}
		})
	}
}

func TestInvoiceStatusTransitions(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus entities.InvoiceStatus
		action        string
		shouldSucceed bool
	}{
		// Finalize transitions
		{name: "draft → finalize → finalized", initialStatus: entities.InvoiceStatusDraft, action: "finalize", shouldSucceed: true},
		{name: "finalized → finalize → error", initialStatus: entities.InvoiceStatusFinalized, action: "finalize", shouldSucceed: false},
		{name: "sent → finalize → error", initialStatus: entities.InvoiceStatusSent, action: "finalize", shouldSucceed: false},

		// Send transitions
		{name: "finalized → send → sent", initialStatus: entities.InvoiceStatusFinalized, action: "send", shouldSucceed: true},
		{name: "draft → send → error", initialStatus: entities.InvoiceStatusDraft, action: "send", shouldSucceed: false},
		{name: "sent → send → error", initialStatus: entities.InvoiceStatusSent, action: "send", shouldSucceed: false},

		// Pay transitions
		{name: "sent → pay → paid", initialStatus: entities.InvoiceStatusSent, action: "pay", shouldSucceed: true},
		{name: "overdue → pay → paid", initialStatus: entities.InvoiceStatusOverdue, action: "pay", shouldSucceed: true},
		{name: "draft → pay → error", initialStatus: entities.InvoiceStatusDraft, action: "pay", shouldSucceed: false},
		{name: "finalized → pay → error", initialStatus: entities.InvoiceStatusFinalized, action: "pay", shouldSucceed: false},

		// Cancel transitions
		{name: "draft → cancel → cancelled", initialStatus: entities.InvoiceStatusDraft, action: "cancel", shouldSucceed: true},
		{name: "finalized → cancel → cancelled", initialStatus: entities.InvoiceStatusFinalized, action: "cancel", shouldSucceed: true},
		{name: "sent → cancel → cancelled", initialStatus: entities.InvoiceStatusSent, action: "cancel", shouldSucceed: true},
		{name: "paid → cancel → error", initialStatus: entities.InvoiceStatusPaid, action: "cancel", shouldSucceed: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			service := &InvoiceService{db: db}
			ctx := context.Background()

			invoice := createTestInvoice(t, db, tt.initialStatus, true)

			var err error
			switch tt.action {
			case "finalize":
				_, err = service.FinalizeInvoice(ctx, invoice.TenantID, invoice.ID, invoice.UserID)
			case "send":
				_, err = service.MarkAsSent(ctx, invoice.TenantID, invoice.ID)
			case "pay":
				_, err = service.MarkAsPaidWithAmount(ctx, invoice.TenantID, invoice.ID, time.Now(), "")
			case "cancel":
				_, err = service.CancelInvoice(ctx, invoice.TenantID, invoice.ID, "Test cancellation")
			}

			if tt.shouldSucceed {
				assert.NoError(t, err, "Expected %s to succeed", tt.name)
			} else {
				assert.Error(t, err, "Expected %s to fail", tt.name)
			}
		})
	}
}
