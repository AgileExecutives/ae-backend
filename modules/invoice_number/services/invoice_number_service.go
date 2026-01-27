package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/unburdy/invoice-number-module/entities"
	// "github.com/ae-base-server/pkg/settings/manager"
	"gorm.io/gorm"
)

// InvoiceNumberConfig defines invoice number format configuration
type InvoiceNumberConfig struct {
	Prefix       string
	YearFormat   string
	MonthFormat  string
	Padding      int
	Separator    string
	ResetMonthly bool
}

// DefaultInvoiceConfig returns the default configuration
func DefaultInvoiceConfig() InvoiceNumberConfig {
	return InvoiceNumberConfig{
		Prefix:       "INV",
		YearFormat:   "YYYY",
		MonthFormat:  "MM",
		Padding:      4,
		Separator:    "-",
		ResetMonthly: true,
	}
}

// InvoiceNumberService handles invoice number generation with database
type InvoiceNumberService struct {
	db *gorm.DB
	// settingsManager *manager.SettingsManager // TODO: Integrate with new settings system
	mutex       sync.Mutex
	cacheTTL    time.Duration
	lockTimeout time.Duration
}

// NewInvoiceNumberService creates a new invoice number service
func NewInvoiceNumberService(db *gorm.DB) *InvoiceNumberService {
	return &InvoiceNumberService{
		db: db,
		// settingsManager: settingsManager, // TODO: Integrate with new settings system
		cacheTTL:    24 * time.Hour,
		lockTimeout: 5 * time.Second,
	}
}

// GenerateInvoiceNumber generates the next invoice number
func (s *InvoiceNumberService) GenerateInvoiceNumber(
	ctx context.Context,
	tenantID uint,
	organizationID uint,
	config InvoiceNumberConfig,
) (*entities.InvoiceNumberResponse, error) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	// Use mutex for concurrency control
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Get next sequence from database
	sequence, err := s.getNextSequenceFromDB(ctx, tenantID, organizationID, year, month, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get next sequence: %w", err)
	}

	// Format invoice number
	invoiceNumber := s.formatInvoiceNumber(year, month, sequence, config)

	// Save to log
	log := &entities.InvoiceNumberLog{
		TenantID:       tenantID,
		OrganizationID: organizationID,
		InvoiceNumber:  invoiceNumber,
		Year:           year,
		Month:          month,
		Sequence:       sequence,
		Status:         "active",
		GeneratedAt:    now,
	}

	if err := s.db.Create(log).Error; err != nil {
		return nil, fmt.Errorf("failed to save log: %w", err)
	}

	response := log.ToResponse()
	response.Format = s.getFormatString(config)
	return &response, nil
}

// GenerateNextInvoiceNumber generates the next invoice number using settings
func (s *InvoiceNumberService) GenerateNextInvoiceNumber(
	ctx context.Context,
	tenantID uint,
	organizationID uint,
) (*entities.InvoiceNumberResponse, error) {
	// TODO: Integrate with new settings system
	// For now, use default config
	config := DefaultInvoiceConfig()
	return s.GenerateInvoiceNumber(ctx, tenantID, organizationID, config)
}

// Original settings integration (commented out for now):
/*
func (s *InvoiceNumberService) GenerateNextInvoiceNumberWithSettings(
	ctx context.Context,
	tenantID uint,
	organizationID uint,
) (*entities.InvoiceNumberResponse, error) {
	// Get invoice accessor
	invoiceAccessor, err := s.settingsManager.GetModuleAccessor("invoice")
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice settings accessor: %w", err)
	}

	// Get settings from the accessor
	prefix, err := invoiceAccessor.GetString("invoice_prefix")
	if err != nil {
		prefix = "INV" // Fallback default
	}

	// Get next invoice number and auto-increment
	nextNumber, err := invoiceAccessor.GetInt("next_invoice_number")
	if err != nil {
		nextNumber = 1000 // Fallback default
	}

	// Increment the next invoice number
	if err := invoiceAccessor.SetInt("next_invoice_number", nextNumber+1); err != nil {
		return nil, fmt.Errorf("failed to increment invoice number: %w", err)
	}

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	// Create a simplified config using settings
	config := InvoiceNumberConfig{
		Prefix:       prefix,
		YearFormat:   "YYYY",
		MonthFormat:  "MM",
		Padding:      4,
		Separator:    "-",
		ResetMonthly: false, // Use continuous numbering by default
	}

	// Format the invoice number as: PREFIX-YYYY-NUMBER (e.g., INV-2025-1000)
	invoiceNumber := fmt.Sprintf("%s-%04d-%0*d", prefix, year, config.Padding, nextNumber)

	// Save to log
	log := &entities.InvoiceNumberLog{
		TenantID:       tenantID,
		OrganizationID: organizationID,
		InvoiceNumber:  invoiceNumber,
		Year:           year,
		Month:          month,
		Sequence:       nextNumber,
		Status:         "active",
		GeneratedAt:    now,
	}

	if err := s.db.Create(log).Error; err != nil {
		return nil, fmt.Errorf("failed to save invoice number log: %w", err)
	}

	response := log.ToResponse()
	response.Format = fmt.Sprintf("%s-YYYY-NNNN", prefix)
	return &response, nil
}
*/

// getNextSequenceFromDB gets sequence from database
func (s *InvoiceNumberService) getNextSequenceFromDB(
	ctx context.Context,
	tenantID uint,
	organizationID uint,
	year int,
	month int,
	config InvoiceNumberConfig,
) (int, error) {
	var record entities.InvoiceNumber

	err := s.db.Where(
		"tenant_id = ? AND organization_id = ? AND year = ? AND month = ?",
		tenantID, organizationID, year, month,
	).First(&record).Error

	if err == gorm.ErrRecordNotFound {
		record = entities.InvoiceNumber{
			TenantID:       tenantID,
			OrganizationID: organizationID,
			Year:           year,
			Month:          month,
			Sequence:       1,
			Format:         s.getFormatString(config),
			LastNumber:     s.formatInvoiceNumber(year, month, 1, config),
		}
		if err := s.db.Create(&record).Error; err != nil {
			return 0, err
		}
		return 1, nil
	} else if err != nil {
		return 0, err
	}

	record.Sequence++
	record.LastNumber = s.formatInvoiceNumber(year, month, record.Sequence, config)
	if err := s.db.Save(&record).Error; err != nil {
		return 0, err
	}

	return record.Sequence, nil
}

// formatInvoiceNumber formats the invoice number
func (s *InvoiceNumberService) formatInvoiceNumber(
	year int,
	month int,
	sequence int,
	config InvoiceNumberConfig,
) string {
	parts := []string{}
	if config.Prefix != "" {
		parts = append(parts, config.Prefix)
	}
	if config.YearFormat == "YYYY" {
		parts = append(parts, fmt.Sprintf("%04d", year))
	} else if config.YearFormat == "YY" {
		parts = append(parts, fmt.Sprintf("%02d", year%100))
	}
	if config.MonthFormat == "MM" {
		parts = append(parts, fmt.Sprintf("%02d", month))
	} else if config.MonthFormat == "M" {
		parts = append(parts, strconv.Itoa(month))
	}
	parts = append(parts, fmt.Sprintf("%0*d", config.Padding, sequence))
	return strings.Join(parts, config.Separator)
}

// getFormatString returns format string
func (s *InvoiceNumberService) getFormatString(config InvoiceNumberConfig) string {
	parts := []string{}
	if config.Prefix != "" {
		parts = append(parts, config.Prefix)
	}
	if config.YearFormat != "" {
		parts = append(parts, "{"+config.YearFormat+"}")
	}
	if config.MonthFormat != "" {
		parts = append(parts, "{"+config.MonthFormat+"}")
	}
	parts = append(parts, fmt.Sprintf("{SEQ:%d}", config.Padding))
	return strings.Join(parts, config.Separator)
}

// GetCurrentSequence gets current sequence without incrementing
func (s *InvoiceNumberService) GetCurrentSequence(
	ctx context.Context,
	tenantID uint,
	organizationID uint,
	year int,
	month int,
) (int, error) {
	var record entities.InvoiceNumber
	err := s.db.Where(
		"tenant_id = ? AND organization_id = ? AND year = ? AND month = ?",
		tenantID, organizationID, year, month,
	).First(&record).Error

	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return record.Sequence, nil
}

// VoidInvoiceNumber marks invoice number as voided
func (s *InvoiceNumberService) VoidInvoiceNumber(
	ctx context.Context,
	tenantID uint,
	invoiceNumber string,
) error {
	return s.db.Model(&entities.InvoiceNumberLog{}).
		Where("tenant_id = ? AND invoice_number = ?", tenantID, invoiceNumber).
		Update("status", "voided").Error
}

// GetInvoiceNumberHistory retrieves history
func (s *InvoiceNumberService) GetInvoiceNumberHistory(
	ctx context.Context,
	tenantID uint,
	organizationID uint,
	year int,
	month int,
	page int,
	pageSize int,
) ([]entities.InvoiceNumberLog, int64, error) {
	var logs []entities.InvoiceNumberLog
	var total int64

	query := s.db.Model(&entities.InvoiceNumberLog{}).Where("tenant_id = ?", tenantID)

	if organizationID > 0 {
		query = query.Where("organization_id = ?", organizationID)
	}
	if year > 0 {
		query = query.Where("year = ?", year)
	}
	if month > 0 {
		query = query.Where("month = ?", month)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("generated_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
