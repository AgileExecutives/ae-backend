package services

import (
	"fmt"

	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
)

// VATService handles VAT calculations and categorization for invoices
type VATService struct{}

// NewVATService creates a new VAT service
func NewVATService() *VATService {
	return &VATService{}
}

// VATCategory represents a VAT category with rate and exemption info
type VATCategory string

const (
	VATCategoryExemptHealthcare VATCategory = "exempt_heilberuf" // §4 Nr.14 UStG
	VATCategoryStandard         VATCategory = "taxable_standard" // 19%
	VATCategoryReduced          VATCategory = "taxable_reduced"  // 7%
	VATCategoryCustom           VATCategory = "custom"           // Custom rate
)

// VATCategoryInfo contains details about a VAT category
type VATCategoryInfo struct {
	Category      VATCategory
	Rate          float64
	IsExempt      bool
	ExemptionText string
	Description   string
}

// GetVATCategories returns all available VAT categories
func (s *VATService) GetVATCategories() []VATCategoryInfo {
	return []VATCategoryInfo{
		{
			Category:      VATCategoryExemptHealthcare,
			Rate:          0.0,
			IsExempt:      true,
			ExemptionText: "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG",
			Description:   "Healthcare services exempt from VAT (§4 Nr.14 UStG)",
		},
		{
			Category:      VATCategoryStandard,
			Rate:          19.0,
			IsExempt:      false,
			ExemptionText: "",
			Description:   "Standard VAT rate (19%)",
		},
		{
			Category:      VATCategoryReduced,
			Rate:          7.0,
			IsExempt:      false,
			ExemptionText: "",
			Description:   "Reduced VAT rate (7%)",
		},
	}
}

// GetVATCategoryByName returns VAT category info by category name
func (s *VATService) GetVATCategoryByName(category VATCategory) (*VATCategoryInfo, error) {
	categories := s.GetVATCategories()
	for _, cat := range categories {
		if cat.Category == category {
			return &cat, nil
		}
	}
	return nil, fmt.Errorf("VAT category not found: %s", category)
}

// GetVATCategoryByRate returns the VAT category for a given rate
func (s *VATService) GetVATCategoryByRate(rate float64, isExempt bool) VATCategory {
	if isExempt {
		return VATCategoryExemptHealthcare
	}

	if rate == 19.0 {
		return VATCategoryStandard
	}

	if rate == 7.0 {
		return VATCategoryReduced
	}

	return VATCategoryCustom
}

// ApplyVATCategory applies a VAT category to an invoice item
func (s *VATService) ApplyVATCategory(item *entities.InvoiceItem, category VATCategory) error {
	catInfo, err := s.GetVATCategoryByName(category)
	if err != nil {
		return err
	}

	item.VATRate = catInfo.Rate
	item.VATExempt = catInfo.IsExempt
	item.VATExemptionText = catInfo.ExemptionText

	return nil
}

// GetDefaultVATCategory returns the default VAT category based on item type
func (s *VATService) GetDefaultVATCategory(itemType string) VATCategory {
	switch itemType {
	case "session":
		// Healthcare sessions are typically VAT-exempt in Germany
		return VATCategoryExemptHealthcare
	case "extra_effort":
		// Extra efforts related to healthcare are also typically exempt
		return VATCategoryExemptHealthcare
	case "custom":
		// Custom items default to standard rate
		return VATCategoryStandard
	default:
		return VATCategoryStandard
	}
}

// InvoiceVATSummary represents VAT breakdown for an invoice
type InvoiceVATSummary struct {
	SubtotalAmount float64            `json:"subtotal_amount"`
	TaxAmount      float64            `json:"tax_amount"`
	TotalAmount    float64            `json:"total_amount"`
	VATBreakdown   []VATBreakdownItem `json:"vat_breakdown"`
}

// VATBreakdownItem represents a single VAT rate breakdown
type VATBreakdownItem struct {
	Rate          float64 `json:"rate"`
	IsExempt      bool    `json:"is_exempt"`
	ExemptionText string  `json:"exemption_text,omitempty"`
	NetAmount     float64 `json:"net_amount"`
	TaxAmount     float64 `json:"tax_amount"`
	GrossAmount   float64 `json:"gross_amount"`
}

// CalculateInvoiceVAT calculates VAT totals and provides a breakdown by rate
func (s *VATService) CalculateInvoiceVAT(items []entities.InvoiceItem) InvoiceVATSummary {
	// Group amounts by VAT rate
	rateMap := make(map[float64]*VATBreakdownItem)
	exemptMap := make(map[string]*VATBreakdownItem)

	var totalSubtotal float64
	var totalTax float64

	for _, item := range items {
		if item.VATExempt {
			// Group exempt items by exemption text
			key := item.VATExemptionText
			if key == "" {
				key = "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG"
			}

			if _, exists := exemptMap[key]; !exists {
				exemptMap[key] = &VATBreakdownItem{
					Rate:          0.0,
					IsExempt:      true,
					ExemptionText: key,
					NetAmount:     0,
					TaxAmount:     0,
					GrossAmount:   0,
				}
			}

			exemptMap[key].NetAmount += item.TotalAmount
			exemptMap[key].GrossAmount += item.TotalAmount
			totalSubtotal += item.TotalAmount
		} else {
			rate := item.VATRate

			if _, exists := rateMap[rate]; !exists {
				rateMap[rate] = &VATBreakdownItem{
					Rate:        rate,
					IsExempt:    false,
					NetAmount:   0,
					TaxAmount:   0,
					GrossAmount: 0,
				}
			}

			taxAmount := item.TotalAmount * (rate / 100.0)

			rateMap[rate].NetAmount += item.TotalAmount
			rateMap[rate].TaxAmount += taxAmount
			rateMap[rate].GrossAmount += item.TotalAmount + taxAmount

			totalSubtotal += item.TotalAmount
			totalTax += taxAmount
		}
	}

	// Build breakdown array
	var breakdown []VATBreakdownItem

	// Add exempt items first
	for _, item := range exemptMap {
		breakdown = append(breakdown, *item)
	}

	// Add taxable items sorted by rate
	for _, item := range rateMap {
		breakdown = append(breakdown, *item)
	}

	return InvoiceVATSummary{
		SubtotalAmount: totalSubtotal,
		TaxAmount:      totalTax,
		TotalAmount:    totalSubtotal + totalTax,
		VATBreakdown:   breakdown,
	}
}

// ValidateVATConfiguration validates that invoice items have correct VAT configuration
func (s *VATService) ValidateVATConfiguration(items []entities.InvoiceItem) error {
	for i, item := range items {
		// Check exempt items have exemption text
		if item.VATExempt && item.VATExemptionText == "" {
			return fmt.Errorf("item %d ('%s') is VAT exempt but missing exemption text", i+1, item.Description)
		}

		// Check taxable items have valid rate
		if !item.VATExempt && item.VATRate == 0 {
			return fmt.Errorf("item %d ('%s') is taxable but has 0%% VAT rate", i+1, item.Description)
		}

		// Check rate is reasonable (0-100%)
		if item.VATRate < 0 || item.VATRate > 100 {
			return fmt.Errorf("item %d ('%s') has invalid VAT rate: %.2f%%", i+1, item.Description, item.VATRate)
		}
	}

	return nil
}

// SetDefaultVATExemptionText sets the default exemption text if not already set
func (s *VATService) SetDefaultVATExemptionText(item *entities.InvoiceItem) {
	if item.VATExempt && item.VATExemptionText == "" {
		item.VATExemptionText = "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG"
	}
}

// CalculateLineItemTax calculates tax for a single line item
func (s *VATService) CalculateLineItemTax(netAmount float64, vatRate float64, vatExempt bool) float64 {
	if vatExempt {
		return 0.0
	}

	return netAmount * (vatRate / 100.0)
}

// CalculateLineItemGross calculates gross amount for a single line item
func (s *VATService) CalculateLineItemGross(netAmount float64, vatRate float64, vatExempt bool) float64 {
	taxAmount := s.CalculateLineItemTax(netAmount, vatRate, vatExempt)
	return netAmount + taxAmount
}
