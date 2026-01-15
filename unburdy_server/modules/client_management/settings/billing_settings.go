package settings

import (
	"github.com/ae-base-server/pkg/settings/entities"
)

// Domain constants
const (
	DomainBilling = "billing"

	KeyBillingMode   = "mode"
	KeyInvoiceItems  = "invoice_items"
	KeyInvoiceNumber = "invoice_number"
	KeyPaymentTerms  = "payment_terms"
	KeyTax           = "tax"
)

// BillingTaxSettings represents tax configuration
type BillingTaxSettings struct {
	VATRate        float64 `json:"vat_rate"`
	ReducedVATRate float64 `json:"reduced_vat_rate"`
	VATExempt      bool    `json:"vat_exempt"`
}

// InvoiceNumberSettings represents invoice numbering configuration
type InvoiceNumberSettings struct {
	Format string `json:"format"` // sequential, year_prefix, year_month_prefix
	Prefix string `json:"prefix"`
}

// PaymentTermsSettings represents payment terms configuration
type PaymentTermsSettings struct {
	PaymentDueDays     int `json:"payment_due_days"`
	FirstReminderDays  int `json:"first_reminder_days"`
	SecondReminderDays int `json:"second_reminder_days"`
}

// InvoiceItemsSettings represents invoice line item text configuration
type InvoiceItemsSettings struct {
	SingleUnitText string `json:"single_unit_text"`
	DoubleUnitText string `json:"double_unit_text"`
}

// BillingModeSettings represents extra efforts billing configuration
type BillingModeSettings struct {
	ExtraEffortsBillingMode string             `json:"extra_efforts_billing_mode"`
	ExtraEffortsConfig      ExtraEffortsConfig `json:"extra_efforts_config"`
}

// ExtraEffortsConfig represents detailed extra efforts configuration
type ExtraEffortsConfig struct {
	BillableEffortTypes   []string `json:"billable_effort_types"`
	ModeBThresholdMinutes int      `json:"mode_b_threshold_minutes"`
	ModeDPreparationRatio float64  `json:"mode_d_preparation_ratio"`
}

// GetBillingSettingsDefinitions returns all billing settings schema definitions
func GetBillingSettingsDefinitions() []entities.SettingRegistration {
	return []entities.SettingRegistration{
		getBillingModeDefinition(),
		getInvoiceItemsDefinition(),
		getInvoiceNumberDefinition(),
		getPaymentTermsDefinition(),
		getTaxDefinition(),
	}
}

// getBillingModeDefinition returns the billing mode setting definition
func getBillingModeDefinition() entities.SettingRegistration {
	return entities.SettingRegistration{
		Domain:  DomainBilling,
		Key:     KeyBillingMode,
		Version: 1,
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"extra_efforts_billing_mode": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"automatic_double", "manual", "disabled"},
					"default": "automatic_double",
				},
				"extra_efforts_config": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"billable_effort_types": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "string",
							},
						},
						"mode_b_threshold_minutes": map[string]interface{}{
							"type":    "number",
							"default": 45,
						},
						"mode_d_preparation_ratio": map[string]interface{}{
							"type":    "number",
							"default": 0.3333333333333333,
						},
					},
				},
			},
		},
		Data: map[string]interface{}{
			"extra_efforts_billing_mode": "automatic_double",
			"extra_efforts_config": map[string]interface{}{
				"billable_effort_types": []string{
					"teacher_meeting",
					"parent_meeting",
					"therapy_planning",
					"case_meeting",
					"final_meeting",
				},
				"mode_b_threshold_minutes": 45,
				"mode_d_preparation_ratio": 0.3333333333333333,
			},
		},
	}
}

// getInvoiceItemsDefinition returns the invoice items setting definition
func getInvoiceItemsDefinition() entities.SettingRegistration {
	return entities.SettingRegistration{
		Domain:  DomainBilling,
		Key:     KeyInvoiceItems,
		Version: 1,
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"single_unit_text": map[string]interface{}{
					"type":    "string",
					"default": "Einzelstunde",
				},
				"double_unit_text": map[string]interface{}{
					"type":    "string",
					"default": "Doppelstunde",
				},
			},
		},
		Data: map[string]interface{}{
			"single_unit_text": "Einzelstunde",
			"double_unit_text": "Doppelstunde",
		},
	}
}

// getInvoiceNumberDefinition returns the invoice number setting definition
func getInvoiceNumberDefinition() entities.SettingRegistration {
	return entities.SettingRegistration{
		Domain:  DomainBilling,
		Key:     KeyInvoiceNumber,
		Version: 1,
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"format": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"sequential", "year_prefix", "year_month_prefix"},
					"default": "sequential",
				},
				"prefix": map[string]interface{}{
					"type":    "string",
					"default": "",
				},
			},
		},
		Data: map[string]interface{}{
			"format": "sequential",
			"prefix": "",
		},
	}
}

// getPaymentTermsDefinition returns the payment terms setting definition
func getPaymentTermsDefinition() entities.SettingRegistration {
	return entities.SettingRegistration{
		Domain:  DomainBilling,
		Key:     KeyPaymentTerms,
		Version: 1,
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"payment_due_days": map[string]interface{}{
					"type":    "integer",
					"default": 14,
				},
				"first_reminder_days": map[string]interface{}{
					"type":    "integer",
					"default": 7,
				},
				"second_reminder_days": map[string]interface{}{
					"type":    "integer",
					"default": 14,
				},
			},
		},
		Data: map[string]interface{}{
			"payment_due_days":     14,
			"first_reminder_days":  7,
			"second_reminder_days": 14,
		},
	}
}

// getTaxDefinition returns the tax setting definition
func getTaxDefinition() entities.SettingRegistration {
	return entities.SettingRegistration{
		Domain:  DomainBilling,
		Key:     KeyTax,
		Version: 1,
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"vat_rate": map[string]interface{}{
					"type":    "number",
					"default": 19.00,
				},
				"reduced_vat_rate": map[string]interface{}{
					"type":    "number",
					"default": 7.00,
				},
				"vat_exempt": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
			},
		},
		Data: map[string]interface{}{
			"vat_rate":         19.00,
			"reduced_vat_rate": 7.00,
			"vat_exempt":       true,
		},
	}
}

// GetDefaultBillingTax returns default tax settings
func GetDefaultBillingTax() *BillingTaxSettings {
	return &BillingTaxSettings{
		VATRate:        19.00,
		ReducedVATRate: 7.00,
		VATExempt:      true,
	}
}

// GetDefaultInvoiceNumber returns default invoice number settings
func GetDefaultInvoiceNumber() *InvoiceNumberSettings {
	return &InvoiceNumberSettings{
		Format: "sequential",
		Prefix: "",
	}
}

// GetDefaultPaymentTerms returns default payment terms settings
func GetDefaultPaymentTerms() *PaymentTermsSettings {
	return &PaymentTermsSettings{
		PaymentDueDays:     14,
		FirstReminderDays:  7,
		SecondReminderDays: 14,
	}
}

// GetDefaultInvoiceItems returns default invoice items settings
func GetDefaultInvoiceItems() *InvoiceItemsSettings {
	return &InvoiceItemsSettings{
		SingleUnitText: "Einzelstunde",
		DoubleUnitText: "Doppelstunde",
	}
}

// GetDefaultBillingMode returns default billing mode settings
func GetDefaultBillingMode() *BillingModeSettings {
	return &BillingModeSettings{
		ExtraEffortsBillingMode: "automatic_double",
		ExtraEffortsConfig: ExtraEffortsConfig{
			BillableEffortTypes: []string{
				"teacher_meeting",
				"parent_meeting",
				"therapy_planning",
				"case_meeting",
				"final_meeting",
			},
			ModeBThresholdMinutes: 45,
			ModeDPreparationRatio: 0.3333333333333333,
		},
	}
}
