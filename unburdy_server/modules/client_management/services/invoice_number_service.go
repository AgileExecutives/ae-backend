package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// InvoiceNumberService handles invoice number generation
type InvoiceNumberService struct {
	db *gorm.DB
}

// NewInvoiceNumberService creates a new invoice number service
func NewInvoiceNumberService(db *gorm.DB) *InvoiceNumberService {
	return &InvoiceNumberService{db: db}
}

// InvoiceNumberFormat represents the format for invoice numbers
type InvoiceNumberFormat string

const (
	InvoiceNumberFormatSequential InvoiceNumberFormat = "sequential"
	InvoiceNumberFormatYearPrefix InvoiceNumberFormat = "year_prefix"
	InvoiceNumberFormatYearMonth  InvoiceNumberFormat = "year_month_prefix"
)

// GenerateInvoiceNumber generates the next invoice number for an organization
// This method is thread-safe and uses database transactions to ensure uniqueness
func (s *InvoiceNumberService) GenerateInvoiceNumber(organizationID uint, invoiceDate time.Time) (string, error) {
	// Get organization settings
	var org struct {
		InvoiceNumberFormat string
		InvoiceNumberPrefix *string
	}

	if err := s.db.Table("organizations").
		Select("invoice_number_format, invoice_number_prefix").
		Where("id = ?", organizationID).
		First(&org).Error; err != nil {
		return "", fmt.Errorf("failed to fetch organization settings: %w", err)
	}

	format := InvoiceNumberFormat(org.InvoiceNumberFormat)
	if format == "" {
		format = InvoiceNumberFormatSequential // Default to sequential
	}

	prefix := ""
	if org.InvoiceNumberPrefix != nil {
		prefix = *org.InvoiceNumberPrefix
	}

	// Use a transaction to ensure thread-safe counter increment
	var invoiceNumber string
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// For SQLite, ensure we can see committed rows from other transactions
		tx.Exec("PRAGMA read_uncommitted=1")

		var lastNumber string
		var lastInvoice struct {
			InvoiceNumber string
		}

		// Build query based on format
		query := tx.Table("invoices").
			Select("invoice_number").
			Where("organization_id = ? AND status != ? AND invoice_number != ?",
				organizationID, "draft", "DRAFT")

		// For year-based formats, only consider invoices from the same year
		if format == InvoiceNumberFormatYearPrefix {
			yearStr := invoiceDate.Format("2006")
			if prefix != "" {
				query = query.Where("invoice_number LIKE ?", prefix+"-"+yearStr+"%")
			} else {
				query = query.Where("invoice_number LIKE ?", yearStr+"%")
			}
		} else if format == InvoiceNumberFormatYearMonth {
			yearMonthStr := invoiceDate.Format("2006-01")
			if prefix != "" {
				query = query.Where("invoice_number LIKE ?", prefix+"-"+yearMonthStr+"%")
			} else {
				query = query.Where("invoice_number LIKE ?", yearMonthStr+"%")
			}
		}

		// Get the last invoice number
		if err := query.Order("invoice_number DESC").First(&lastInvoice).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				lastNumber = ""
			} else {
				return fmt.Errorf("failed to query last invoice: %w", err)
			}
		} else {
			lastNumber = lastInvoice.InvoiceNumber
		}

		// Generate the next number based on format
		var err error
		invoiceNumber, err = s.generateNextNumber(format, prefix, lastNumber, invoiceDate)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return invoiceNumber, nil
}

// generateNextNumber generates the next invoice number based on the format
func (s *InvoiceNumberService) generateNextNumber(format InvoiceNumberFormat, prefix, lastNumber string, invoiceDate time.Time) (string, error) {
	switch format {
	case InvoiceNumberFormatSequential:
		return s.generateSequential(prefix, lastNumber)
	case InvoiceNumberFormatYearPrefix:
		return s.generateYearPrefix(prefix, lastNumber, invoiceDate)
	case InvoiceNumberFormatYearMonth:
		return s.generateYearMonthPrefix(prefix, lastNumber, invoiceDate)
	default:
		return s.generateSequential(prefix, lastNumber)
	}
}

// generateSequential generates a sequential number like: 1, 2, 3 or PREFIX-0001, PREFIX-0002
func (s *InvoiceNumberService) generateSequential(prefix, lastNumber string) (string, error) {
	nextNum := 1

	if lastNumber != "" {
		// Extract the numeric part
		numPart := lastNumber
		if prefix != "" {
			numPart = strings.TrimPrefix(lastNumber, prefix+"-")
		}

		num, err := strconv.Atoi(numPart)
		if err != nil {
			// If we can't parse, start from 1
			nextNum = 1
		} else {
			nextNum = num + 1
		}
	}

	if prefix != "" {
		return fmt.Sprintf("%s-%04d", prefix, nextNum), nil
	}
	return fmt.Sprintf("%d", nextNum), nil
}

// generateYearPrefix generates year-prefixed numbers like: 2026-0001, 2026-0002
func (s *InvoiceNumberService) generateYearPrefix(prefix, lastNumber string, invoiceDate time.Time) (string, error) {
	year := invoiceDate.Format("2006")
	nextNum := 1

	if lastNumber != "" {
		// Build the expected prefix-year combination
		expectedPrefix := year
		if prefix != "" {
			expectedPrefix = prefix + "-" + year
		}

		// Check if lastNumber starts with the expected prefix
		if strings.HasPrefix(lastNumber, expectedPrefix) {
			// Extract the numeric part after the year
			parts := strings.Split(lastNumber, "-")
			if len(parts) >= 1 {
				numPart := parts[len(parts)-1]
				num, err := strconv.Atoi(numPart)
				if err == nil {
					nextNum = num + 1
				}
			}
		}
	}

	if prefix != "" {
		return fmt.Sprintf("%s-%s-%04d", prefix, year, nextNum), nil
	}
	return fmt.Sprintf("%s-%04d", year, nextNum), nil
}

// generateYearMonthPrefix generates year-month-prefixed numbers like: 2026-01-0001, 2026-01-0002
func (s *InvoiceNumberService) generateYearMonthPrefix(prefix, lastNumber string, invoiceDate time.Time) (string, error) {
	yearMonth := invoiceDate.Format("2006-01")
	nextNum := 1

	if lastNumber != "" {
		// Build the expected prefix-yearMonth combination
		expectedPrefix := yearMonth
		if prefix != "" {
			expectedPrefix = prefix + "-" + yearMonth
		}

		// Check if lastNumber starts with the expected prefix
		if strings.HasPrefix(lastNumber, expectedPrefix) {
			// Extract the numeric part after the year-month
			parts := strings.Split(lastNumber, "-")
			if len(parts) >= 1 {
				numPart := parts[len(parts)-1]
				num, err := strconv.Atoi(numPart)
				if err == nil {
					nextNum = num + 1
				}
			}
		}
	}

	if prefix != "" {
		return fmt.Sprintf("%s-%s-%04d", prefix, yearMonth, nextNum), nil
	}
	return fmt.Sprintf("%s-%04d", yearMonth, nextNum), nil
}
