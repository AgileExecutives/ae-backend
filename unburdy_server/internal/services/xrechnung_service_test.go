package services

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	baseAPI "github.com/ae/base-server/api"
	settingsEntities "github.com/ae/base-server/pkg/settings/entities"
	"github.com/unburdy/unburdy-server-api/internal/models"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupXRechnungDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	// Migrate settings table so GetPaymentTerms doesn't fail on missing table
	require.NoError(t, db.AutoMigrate(&settingsEntities.SettingDefinition{}, &settingsEntities.Setting{}))
	return db
}

func makeTestInvoice(status entities.InvoiceStatus) *entities.Invoice {
	return &entities.Invoice{
		InvoiceNumber: "RE-2024-001",
		InvoiceDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Status:        status,
		TotalAmount:   119.00,
		InvoiceItems: []entities.InvoiceItem{
			{
				Description: "Therapy session",
				NumberUnits: 1,
				UnitPrice:   100.00,
				TotalAmount: 100.00,
				VATRate:     19.0,
			},
		},
	}
}

func makeTestOrganization() *baseAPI.Organization {
	return &baseAPI.Organization{
		TenantID:         1,
		Name:             "Therapy GmbH",
		StreetAddress:    "Hauptstr. 1",
		Zip:              "10115",
		City:             "Berlin",
		TaxUstID:         "DE123456789",
		BankAccountIBAN:  "DE89370400440532013000",
		BankAccountBIC:   "COBADEFFXXX",
		BankAccountOwner: "Therapy GmbH",
	}
}

func makeTestCostProvider(isGovt bool, leitwegID string) *models.CostProvider {
	return &models.CostProvider{
		Organization:         "Health Authority Berlin",
		StreetAddress:        "Behördenstr. 1",
		Zip:                  "10001",
		City:                 "Berlin",
		IsGovernmentCustomer: isGovt,
		LeitwegID:            leitwegID,
	}
}

func TestXRechnungService_GenerateXRechnungXML_Validation(t *testing.T) {
	db := setupXRechnungDB(t)
	svc := NewXRechnungService(db)

	org := makeTestOrganization()
	cp := makeTestCostProvider(true, "04011000-12345-06")

	t.Run("draft invoice returns error", func(t *testing.T) {
		_, err := svc.GenerateXRechnungXML(makeTestInvoice(entities.InvoiceStatusDraft), org, cp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "finalized")
	})

	t.Run("finalized status returns error", func(t *testing.T) {
		_, err := svc.GenerateXRechnungXML(makeTestInvoice(entities.InvoiceStatusFinalized), org, cp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "finalized")
	})

	t.Run("cancelled status returns error", func(t *testing.T) {
		_, err := svc.GenerateXRechnungXML(makeTestInvoice(entities.InvoiceStatusCancelled), org, cp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "finalized")
	})

	t.Run("non-government customer returns error", func(t *testing.T) {
		_, err := svc.GenerateXRechnungXML(makeTestInvoice(entities.InvoiceStatusSent), org, makeTestCostProvider(false, "04011000-12345-06"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "government")
	})

	t.Run("missing leitweg-id returns error", func(t *testing.T) {
		_, err := svc.GenerateXRechnungXML(makeTestInvoice(entities.InvoiceStatusSent), org, makeTestCostProvider(true, ""))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Leitweg-ID")
	})
}

func TestXRechnungService_GenerateXRechnungXML_Success(t *testing.T) {
	db := setupXRechnungDB(t)
	svc := NewXRechnungService(db)
	org := makeTestOrganization()
	cp := makeTestCostProvider(true, "04011000-12345-06")

	for _, status := range []entities.InvoiceStatus{
		entities.InvoiceStatusSent,
		entities.InvoiceStatusPaid,
		entities.InvoiceStatusOverdue,
	} {
		status := status
		t.Run(string(status)+" invoice produces valid XML", func(t *testing.T) {
			invoice := makeTestInvoice(status)
			if status == entities.InvoiceStatusPaid {
				now := time.Now()
				invoice.PayedDate = &now
			}
			xmlData, err := svc.GenerateXRechnungXML(invoice, org, cp)
			require.NoError(t, err)
			assert.NotEmpty(t, xmlData)
			xmlStr := string(xmlData)
			assert.True(t, strings.HasPrefix(xmlStr, "<?xml"), "should start with XML declaration")
			assert.Contains(t, xmlStr, "ubl:Invoice")
			assert.Contains(t, xmlStr, "RE-2024-001")
			assert.Contains(t, xmlStr, "04011000-12345-06")
			assert.Contains(t, xmlStr, "EUR")
		})
	}
}

func TestXRechnungService_GenerateXRechnungXML_WithBankDetails(t *testing.T) {
	db := setupXRechnungDB(t)
	svc := NewXRechnungService(db)
	cp := makeTestCostProvider(true, "04011000-12345-06")

	org := makeTestOrganization()
	org.BankAccountIBAN = "DE89370400440532013000"

	invoice := makeTestInvoice(entities.InvoiceStatusSent)
	xmlData, err := svc.GenerateXRechnungXML(invoice, org, cp)
	require.NoError(t, err)
	assert.Contains(t, string(xmlData), "DE89370400440532013000")
}

func TestXRechnungService_GenerateXRechnungXML_NoBankDetails(t *testing.T) {
	db := setupXRechnungDB(t)
	svc := NewXRechnungService(db)
	cp := makeTestCostProvider(true, "04011000-12345-06")

	org := makeTestOrganization()
	org.BankAccountIBAN = "" // no bank account

	invoice := makeTestInvoice(entities.InvoiceStatusSent)
	xmlData, err := svc.GenerateXRechnungXML(invoice, org, cp)
	require.NoError(t, err)
	assert.NotEmpty(t, xmlData)
}

func TestXRechnungService_GenerateXRechnungXML_OrgWithEmailAndPhone(t *testing.T) {
	db := setupXRechnungDB(t)
	svc := NewXRechnungService(db)
	cp := makeTestCostProvider(true, "04011000-12345-06")

	// Org with email and phone covers the contact block in buildSupplierParty (lines 293-295)
	org := makeTestOrganization()
	org.Email = "info@therapy.de"
	org.Phone = "+4930123456"

	invoice := makeTestInvoice(entities.InvoiceStatusSent)
	xmlData, err := svc.GenerateXRechnungXML(invoice, org, cp)
	require.NoError(t, err)
	xmlStr := string(xmlData)
	assert.Contains(t, xmlStr, "info@therapy.de")
	assert.Contains(t, xmlStr, "+4930123456")
}

func TestXRechnungService_GenerateXRechnungXML_VATExemptNoText(t *testing.T) {
	db := setupXRechnungDB(t)
	svc := NewXRechnungService(db)
	org := makeTestOrganization()
	cp := makeTestCostProvider(true, "04011000-12345-06")

	// VATExempt item with NO VATExemptionText — covers default fallback in buildTaxTotal and getItemTaxCategory
	invoice := &entities.Invoice{
		InvoiceNumber: "RE-2024-002",
		InvoiceDate:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		Status:        entities.InvoiceStatusSent,
		TotalAmount:   100.00,
		InvoiceItems: []entities.InvoiceItem{
			{
				Description:  "Exempt session",
				NumberUnits:  1,
				UnitPrice:    100.00,
				TotalAmount:  100.00,
				VATExempt:    true,
				VATRate:      0,
				// VATExemptionText intentionally empty — triggers default text branch
			},
		},
	}

	xmlData, err := svc.GenerateXRechnungXML(invoice, org, cp)
	require.NoError(t, err)
	xmlStr := string(xmlData)
	// Should contain the default exemption text
	assert.Contains(t, xmlStr, "Umsatzsteuerfrei")
}
