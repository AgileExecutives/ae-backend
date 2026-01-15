package services

import (
	"encoding/json"
	"fmt"

	"github.com/ae-base-server/pkg/settings/entities"
	"github.com/ae-base-server/pkg/settings/repository"
	"github.com/unburdy/unburdy-server-api/modules/client_management/settings"
	"gorm.io/gorm"
)

// SettingsHelper provides convenient access to billing settings
type SettingsHelper struct {
	repo *repository.SettingsRepository
}

// NewSettingsHelper creates a new settings helper
func NewSettingsHelper(db *gorm.DB) *SettingsHelper {
	return &SettingsHelper{
		repo: repository.NewSettingsRepository(db),
	}
}

// GetBillingTax retrieves tax settings for a tenant
func (h *SettingsHelper) GetBillingTax(tenantID uint) (*settings.BillingTaxSettings, error) {
	setting, err := h.repo.GetSetting(tenantID, settings.DomainBilling, settings.KeyTax)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing.tax setting: %w", err)
	}

	if setting == nil {
		// Return defaults
		return settings.GetDefaultBillingTax(), nil
	}

	var tax settings.BillingTaxSettings
	if err := json.Unmarshal(setting.Data, &tax); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tax settings: %w", err)
	}

	return &tax, nil
}

// GetInvoiceNumber retrieves invoice number settings for a tenant
func (h *SettingsHelper) GetInvoiceNumber(tenantID uint) (*settings.InvoiceNumberSettings, error) {
	setting, err := h.repo.GetSetting(tenantID, settings.DomainBilling, settings.KeyInvoiceNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing.invoice_number setting: %w", err)
	}

	if setting == nil {
		// Return defaults
		return settings.GetDefaultInvoiceNumber(), nil
	}

	var number settings.InvoiceNumberSettings
	if err := json.Unmarshal(setting.Data, &number); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invoice number settings: %w", err)
	}

	return &number, nil
}

// GetPaymentTerms retrieves payment terms settings for a tenant
func (h *SettingsHelper) GetPaymentTerms(tenantID uint) (*settings.PaymentTermsSettings, error) {
	setting, err := h.repo.GetSetting(tenantID, settings.DomainBilling, settings.KeyPaymentTerms)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing.payment_terms setting: %w", err)
	}

	if setting == nil {
		// Return defaults
		return settings.GetDefaultPaymentTerms(), nil
	}

	var terms settings.PaymentTermsSettings
	if err := json.Unmarshal(setting.Data, &terms); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment terms settings: %w", err)
	}

	return &terms, nil
}

// GetInvoiceItems retrieves invoice items settings for a tenant
func (h *SettingsHelper) GetInvoiceItems(tenantID uint) (*settings.InvoiceItemsSettings, error) {
	setting, err := h.repo.GetSetting(tenantID, settings.DomainBilling, settings.KeyInvoiceItems)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing.invoice_items setting: %w", err)
	}

	if setting == nil {
		// Return defaults
		return settings.GetDefaultInvoiceItems(), nil
	}

	var items settings.InvoiceItemsSettings
	if err := json.Unmarshal(setting.Data, &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invoice items settings: %w", err)
	}

	return &items, nil
}

// GetBillingMode retrieves billing mode settings for a tenant
func (h *SettingsHelper) GetBillingMode(tenantID uint) (*settings.BillingModeSettings, error) {
	setting, err := h.repo.GetSetting(tenantID, settings.DomainBilling, settings.KeyBillingMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing.mode setting: %w", err)
	}

	if setting == nil {
		// Return defaults
		return settings.GetDefaultBillingMode(), nil
	}

	var mode settings.BillingModeSettings
	if err := json.Unmarshal(setting.Data, &mode); err != nil {
		return nil, fmt.Errorf("failed to unmarshal billing mode settings: %w", err)
	}

	return &mode, nil
}

// SetBillingTax saves tax settings for a tenant
func (h *SettingsHelper) SetBillingTax(tenantID uint, tax *settings.BillingTaxSettings) error {
	data, err := json.Marshal(tax)
	if err != nil {
		return fmt.Errorf("failed to marshal tax settings: %w", err)
	}

	setting := &entities.Setting{
		TenantID: tenantID,
		Domain:   "billing",
		Key:      "tax",
		Version:  1,
		Data:     data,
	}

	return h.repo.SetSetting(setting)
}

// SetInvoiceNumber saves invoice number settings for a tenant
func (h *SettingsHelper) SetInvoiceNumber(tenantID uint, number *settings.InvoiceNumberSettings) error {
	data, err := json.Marshal(number)
	if err != nil {
		return fmt.Errorf("failed to marshal invoice number settings: %w", err)
	}

	setting := &entities.Setting{
		TenantID: tenantID,
		Domain:   "billing",
		Key:      "invoice_number",
		Version:  1,
		Data:     data,
	}

	return h.repo.SetSetting(setting)
}

// SetPaymentTerms saves payment terms settings for a tenant
func (h *SettingsHelper) SetPaymentTerms(tenantID uint, terms *settings.PaymentTermsSettings) error {
	data, err := json.Marshal(terms)
	if err != nil {
		return fmt.Errorf("failed to marshal payment terms settings: %w", err)
	}

	setting := &entities.Setting{
		TenantID: tenantID,
		Domain:   "billing",
		Key:      "payment_terms",
		Version:  1,
		Data:     data,
	}

	return h.repo.SetSetting(setting)
}
