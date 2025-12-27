package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unburdy/unburdy-server-api/modules/documents/entities"
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

// InvoiceNumberService handles invoice number generation with Redis caching
type InvoiceNumberService struct {
	db          *gorm.DB
	redis       *redis.Client
	mutex       sync.Mutex
	cacheTTL    time.Duration
	lockTimeout time.Duration
}

// NewInvoiceNumberService creates a new invoice number service
func NewInvoiceNumberService(db *gorm.DB, redisClient *redis.Client) *InvoiceNumberService {
	return &InvoiceNumberService{
		db:          db,
		redis:       redisClient,
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

	// Acquire lock
	lockKey := s.getLockKey(tenantID, organizationID, year, month)
	locked, err := s.acquireLock(ctx, lockKey)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		return nil, fmt.Errorf("could not acquire lock")
	}
	defer s.releaseLock(ctx, lockKey)

	// Get next sequence
	sequence, err := s.getNextSequenceFromRedis(ctx, tenantID, organizationID, year, month)
	if err != nil {
		sequence, err = s.getNextSequenceFromDB(ctx, tenantID, organizationID, year, month, config)
		if err != nil {
			return nil, fmt.Errorf("failed to get next sequence: %w", err)
		}
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

	// Update cache
	s.updateSequenceCache(ctx, tenantID, organizationID, year, month, sequence)

	response := log.ToResponse()
	response.Format = s.getFormatString(config)
	return &response, nil
}

// getNextSequenceFromRedis gets sequence from Redis
func (s *InvoiceNumberService) getNextSequenceFromRedis(
	ctx context.Context,
	tenantID uint,
	organizationID uint,
	year int,
	month int,
) (int, error) {
	key := s.getSequenceKey(tenantID, organizationID, year, month)
	seq, err := s.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if seq == 1 {
		s.redis.Expire(ctx, key, s.cacheTTL)
	}
	return int(seq), nil
}

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

// Helper methods
func (s *InvoiceNumberService) getSequenceKey(tenantID uint, organizationID uint, year int, month int) string {
	return fmt.Sprintf("invoice_seq:t%d:o%d:y%d:m%d", tenantID, organizationID, year, month)
}

func (s *InvoiceNumberService) getLockKey(tenantID uint, organizationID uint, year int, month int) string {
	return fmt.Sprintf("invoice_lock:t%d:o%d:y%d:m%d", tenantID, organizationID, year, month)
}

func (s *InvoiceNumberService) acquireLock(ctx context.Context, key string) (bool, error) {
	return s.redis.SetNX(ctx, key, "1", s.lockTimeout).Result()
}

func (s *InvoiceNumberService) releaseLock(ctx context.Context, key string) error {
	return s.redis.Del(ctx, key).Err()
}

func (s *InvoiceNumberService) updateSequenceCache(
	ctx context.Context,
	tenantID uint,
	organizationID uint,
	year int,
	month int,
	sequence int,
) error {
	key := s.getSequenceKey(tenantID, organizationID, year, month)
	return s.redis.Set(ctx, key, sequence, s.cacheTTL).Err()
}

// GetCurrentSequence gets current sequence without incrementing
func (s *InvoiceNumberService) GetCurrentSequence(
	ctx context.Context,
	tenantID uint,
	organizationID uint,
	year int,
	month int,
) (int, error) {
	key := s.getSequenceKey(tenantID, organizationID, year, month)
	seq, err := s.redis.Get(ctx, key).Int()
	if err == nil {
		return seq, nil
	}

	var record entities.InvoiceNumber
	err = s.db.Where(
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
