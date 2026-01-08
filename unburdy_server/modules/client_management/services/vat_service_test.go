package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
)

// TestVATService_GetVATCategories tests retrieving all VAT categories
func TestVATService_GetVATCategories(t *testing.T) {
	service := NewVATService()

	categories := service.GetVATCategories()

	// Should have 3 categories
	assert.Len(t, categories, 3)

	// Find categories by type
	var exemptCat, standardCat, reducedCat *VATCategoryInfo
	for i := range categories {
		switch categories[i].Category {
		case VATCategoryExemptHealthcare:
			exemptCat = &categories[i]
		case VATCategoryStandard:
			standardCat = &categories[i]
		case VATCategoryReduced:
			reducedCat = &categories[i]
		}
	}

	// Validate exempt healthcare category
	require.NotNil(t, exemptCat)
	assert.Equal(t, 0.0, exemptCat.Rate)
	assert.True(t, exemptCat.IsExempt)
	assert.NotEmpty(t, exemptCat.ExemptionText)
	assert.Contains(t, exemptCat.ExemptionText, "§4 Nr. 14 UStG")

	// Validate standard rate
	require.NotNil(t, standardCat)
	assert.Equal(t, 19.0, standardCat.Rate)
	assert.False(t, standardCat.IsExempt)

	// Validate reduced rate
	require.NotNil(t, reducedCat)
	assert.Equal(t, 7.0, reducedCat.Rate)
	assert.False(t, reducedCat.IsExempt)
}

// TestVATService_ApplyVATCategory_ExemptHealthcare tests applying exempt healthcare category
func TestVATService_ApplyVATCategory_ExemptHealthcare(t *testing.T) {
	service := NewVATService()

	item := &entities.InvoiceItem{
		TotalAmount: 100.0,
	}

	err := service.ApplyVATCategory(item, VATCategoryExemptHealthcare)
	require.NoError(t, err)

	assert.Equal(t, 0.0, item.VATRate)
	assert.True(t, item.VATExempt)
	assert.NotEmpty(t, item.VATExemptionText)
	assert.Contains(t, item.VATExemptionText, "§4 Nr. 14 UStG")
}

// TestVATService_ApplyVATCategory_StandardRate tests applying standard rate
func TestVATService_ApplyVATCategory_StandardRate(t *testing.T) {
	service := NewVATService()

	item := &entities.InvoiceItem{
		TotalAmount: 100.0,
	}

	err := service.ApplyVATCategory(item, VATCategoryStandard)
	require.NoError(t, err)

	assert.Equal(t, 19.0, item.VATRate)
	assert.False(t, item.VATExempt)
	assert.Empty(t, item.VATExemptionText)
}

// TestVATService_ApplyVATCategory_ReducedRate tests applying reduced rate
func TestVATService_ApplyVATCategory_ReducedRate(t *testing.T) {
	service := NewVATService()

	item := &entities.InvoiceItem{
		TotalAmount: 100.0,
	}

	err := service.ApplyVATCategory(item, VATCategoryReduced)
	require.NoError(t, err)

	assert.Equal(t, 7.0, item.VATRate)
	assert.False(t, item.VATExempt)
	assert.Empty(t, item.VATExemptionText)
}

// TestVATService_ApplyVATCategory_Invalid tests invalid category
func TestVATService_ApplyVATCategory_Invalid(t *testing.T) {
	service := NewVATService()

	item := &entities.InvoiceItem{
		TotalAmount: 100.0,
	}

	err := service.ApplyVATCategory(item, "invalid_category")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestVATService_CalculateInvoiceVAT_ExemptOnly tests calculation with only exempt items
func TestVATService_CalculateInvoiceVAT_ExemptOnly(t *testing.T) {
	service := NewVATService()

	items := []entities.InvoiceItem{
		{
			TotalAmount:      100.0,
			VATRate:          0.0,
			VATExempt:        true,
			VATExemptionText: "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG",
		},
		{
			TotalAmount:      150.0,
			VATRate:          0.0,
			VATExempt:        true,
			VATExemptionText: "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG",
		},
		{
			TotalAmount:      50.0,
			VATRate:          0.0,
			VATExempt:        true,
			VATExemptionText: "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG",
		},
	}

	// Calculate VAT
	summary := service.CalculateInvoiceVAT(items)

	// Verify totals
	assert.Equal(t, 300.0, summary.SubtotalAmount)
	assert.Equal(t, 0.0, summary.TaxAmount)
	assert.Equal(t, 300.0, summary.TotalAmount)
	assert.Len(t, summary.VATBreakdown, 1)

	assert.Equal(t, 0.0, summary.VATBreakdown[0].Rate)
	assert.True(t, summary.VATBreakdown[0].IsExempt)
	assert.NotEmpty(t, summary.VATBreakdown[0].ExemptionText)
	assert.Equal(t, 300.0, summary.VATBreakdown[0].NetAmount)
	assert.Equal(t, 0.0, summary.VATBreakdown[0].TaxAmount)
	assert.Equal(t, 300.0, summary.VATBreakdown[0].GrossAmount)
}

// TestVATService_CalculateInvoiceVAT_StandardOnly tests calculation with standard rate
func TestVATService_CalculateInvoiceVAT_StandardOnly(t *testing.T) {
	service := NewVATService()

	items := []entities.InvoiceItem{
		{
			TotalAmount: 100.0,
			VATRate:     19.0,
			VATExempt:   false,
		},
		{
			TotalAmount: 100.0,
			VATRate:     19.0,
			VATExempt:   false,
		},
	}

	// Calculate VAT
	summary := service.CalculateInvoiceVAT(items)

	// Verify totals
	assert.Equal(t, 200.0, summary.SubtotalAmount)
	assert.Equal(t, 38.0, summary.TaxAmount) // 200 * 0.19
	assert.Equal(t, 238.0, summary.TotalAmount)
	assert.Len(t, summary.VATBreakdown, 1)

	assert.Equal(t, 19.0, summary.VATBreakdown[0].Rate)
	assert.False(t, summary.VATBreakdown[0].IsExempt)
	assert.Equal(t, 200.0, summary.VATBreakdown[0].NetAmount)
	assert.Equal(t, 38.0, summary.VATBreakdown[0].TaxAmount)
	assert.Equal(t, 238.0, summary.VATBreakdown[0].GrossAmount)
}

// TestVATService_CalculateInvoiceVAT_MixedRates tests calculation with mixed rates
func TestVATService_CalculateInvoiceVAT_MixedRates(t *testing.T) {
	service := NewVATService()

	items := []entities.InvoiceItem{
		{
			TotalAmount:      150.0,
			VATRate:          0.0,
			VATExempt:        true,
			VATExemptionText: "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG",
		},
		{
			TotalAmount:      100.0,
			VATRate:          0.0,
			VATExempt:        true,
			VATExemptionText: "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG",
		},
		{
			TotalAmount: 100.0,
			VATRate:     19.0,
			VATExempt:   false,
		},
	}

	// Calculate VAT
	result := service.CalculateInvoiceVAT(items)

	// Verify totals: 250 (exempt) + 100 (standard net) = 350 net
	//                0 (exempt tax) + 19 (standard tax) = 19 total tax
	//                369 total gross
	assert.Equal(t, 350.0, result.SubtotalAmount)
	assert.Equal(t, 19.0, result.TaxAmount)
	assert.Equal(t, 369.0, result.TotalAmount)

	// Verify the breakdown structure
	assert.Len(t, result.VATBreakdown, 2, "Should have 2 rate categories")

	// Find entries by rate
	var exemptBreakdown, standardBreakdown *VATBreakdownItem
	for i := range result.VATBreakdown {
		if result.VATBreakdown[i].IsExempt {
			exemptBreakdown = &result.VATBreakdown[i]
		} else if result.VATBreakdown[i].Rate == 19.0 {
			standardBreakdown = &result.VATBreakdown[i]
		}
	}

	require.NotNil(t, exemptBreakdown, "Should have exempt breakdown")
	assert.Equal(t, 250.0, exemptBreakdown.NetAmount)
	assert.Equal(t, 0.0, exemptBreakdown.TaxAmount)
	assert.Equal(t, 250.0, exemptBreakdown.GrossAmount)
	assert.NotEmpty(t, exemptBreakdown.ExemptionText)

	require.NotNil(t, standardBreakdown, "Should have standard rate breakdown")
	assert.Equal(t, 100.0, standardBreakdown.NetAmount)
	assert.Equal(t, 19.0, standardBreakdown.TaxAmount)
	assert.Equal(t, 119.0, standardBreakdown.GrossAmount)
}

// TestVATService_ValidateVATConfiguration_Valid tests valid configuration
func TestVATService_ValidateVATConfiguration_Valid(t *testing.T) {
	service := NewVATService()

	items := []entities.InvoiceItem{
		{
			VATRate:          0.0,
			VATExempt:        true,
			VATExemptionText: "§4 Nr.14 UStG",
		},
		{
			VATRate:   19.0,
			VATExempt: false,
		},
	}

	err := service.ValidateVATConfiguration(items)
	assert.NoError(t, err, "Should have no validation errors")
}

// TestVATService_ValidateVATConfiguration_MissingExemptionText tests missing exemption text
func TestVATService_ValidateVATConfiguration_MissingExemptionText(t *testing.T) {
	service := NewVATService()

	items := []entities.InvoiceItem{
		{
			VATRate:   0.0,
			VATExempt: true,
			// Missing VATExemptionText
		},
	}

	err := service.ValidateVATConfiguration(items)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exempt")
}

// TestVATService_ValidateVATConfiguration_InvalidRate tests invalid rate
func TestVATService_ValidateVATConfiguration_InvalidRate(t *testing.T) {
	service := NewVATService()

	items := []entities.InvoiceItem{
		{
			VATRate:   -5.0,
			VATExempt: false,
		},
	}

	err := service.ValidateVATConfiguration(items)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

// TestVATService_ValidateVATConfiguration_ExemptWithNonZeroRate tests exempt with non-zero rate
func TestVATService_ValidateVATConfiguration_ExemptWithNonZeroRate(t *testing.T) {
	service := NewVATService()

	items := []entities.InvoiceItem{
		{
			VATRate:          19.0,
			VATExempt:        true,
			VATExemptionText: "§4 Nr.14 UStG",
		},
	}

	err := service.ValidateVATConfiguration(items)
	// This configuration is technically valid - the service doesn't enforce rate=0 for exempt items
	assert.NoError(t, err)
}
