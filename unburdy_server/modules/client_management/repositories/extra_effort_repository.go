package repositories

import (
	"time"

	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"gorm.io/gorm"
)

// ExtraEffortRepository handles database operations for extra efforts
type ExtraEffortRepository struct {
	db *gorm.DB
}

// NewExtraEffortRepository creates a new ExtraEffortRepository
func NewExtraEffortRepository(db *gorm.DB) *ExtraEffortRepository {
	return &ExtraEffortRepository{db: db}
}

// Create creates a new extra effort
func (r *ExtraEffortRepository) Create(effort *entities.ExtraEffort) error {
	return r.db.Create(effort).Error
}

// GetByID retrieves an extra effort by ID
func (r *ExtraEffortRepository) GetByID(id, tenantID uint) (*entities.ExtraEffort, error) {
	var effort entities.ExtraEffort
	err := r.db.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Client").
		Preload("Session").
		First(&effort).Error
	return &effort, err
}

// List retrieves extra efforts with optional filters
func (r *ExtraEffortRepository) List(tenantID uint, filters map[string]interface{}) ([]entities.ExtraEffort, error) {
	var efforts []entities.ExtraEffort
	query := r.db.Where("tenant_id = ?", tenantID).
		Preload("Client").
		Preload("Session")

	// Apply filters
	if clientID, ok := filters["client_id"].(uint); ok {
		query = query.Where("client_id = ?", clientID)
	}
	if sessionID, ok := filters["session_id"].(uint); ok {
		query = query.Where("session_id = ?", sessionID)
	}
	if billingStatus, ok := filters["billing_status"].(string); ok {
		query = query.Where("billing_status = ?", billingStatus)
	}
	if effortType, ok := filters["effort_type"].(string); ok {
		query = query.Where("effort_type = ?", effortType)
	}
	if fromDate, ok := filters["from_date"].(time.Time); ok {
		query = query.Where("effort_date >= ?", fromDate)
	}
	if toDate, ok := filters["to_date"].(time.Time); ok {
		query = query.Where("effort_date <= ?", toDate)
	}

	err := query.Order("effort_date DESC").Find(&efforts).Error
	return efforts, err
}

// GetUnbilledByClient retrieves unbilled extra efforts for a specific client
func (r *ExtraEffortRepository) GetUnbilledByClient(clientID, tenantID uint) ([]entities.ExtraEffort, error) {
	var efforts []entities.ExtraEffort
	err := r.db.Where("client_id = ? AND tenant_id = ? AND billing_status = ? AND billable = ?", clientID, tenantID, "unbilled", true).
		Preload("Session").
		Order("effort_date ASC").
		Find(&efforts).Error
	return efforts, err
}

// GetUnbilledBySession retrieves unbilled extra efforts linked to a specific session
func (r *ExtraEffortRepository) GetUnbilledBySession(sessionID, tenantID uint) ([]entities.ExtraEffort, error) {
	var efforts []entities.ExtraEffort
	err := r.db.Where("session_id = ? AND tenant_id = ? AND billing_status = ? AND billable = ?", sessionID, tenantID, "unbilled", true).
		Order("effort_date ASC").
		Find(&efforts).Error
	return efforts, err
}

// Update updates an extra effort
func (r *ExtraEffortRepository) Update(effort *entities.ExtraEffort) error {
	return r.db.Save(effort).Error
}

// Delete soft deletes an extra effort
func (r *ExtraEffortRepository) Delete(id, tenantID uint) error {
	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entities.ExtraEffort{}).Error
}

// MarkAsBilled marks extra efforts as billed and links them to an invoice item
func (r *ExtraEffortRepository) MarkAsBilled(effortIDs []uint, invoiceItemID uint) error {
	return r.db.Model(&entities.ExtraEffort{}).
		Where("id IN ?", effortIDs).
		Updates(map[string]interface{}{
			"billing_status":  "billed",
			"invoice_item_id": invoiceItemID,
		}).Error
}

// MarkAsUnbilled marks extra efforts as unbilled (for reverting draft invoices)
func (r *ExtraEffortRepository) MarkAsUnbilled(effortIDs []uint) error {
	return r.db.Model(&entities.ExtraEffort{}).
		Where("id IN ?", effortIDs).
		Updates(map[string]interface{}{
			"billing_status":  "unbilled",
			"invoice_item_id": nil,
		}).Error
}

// Count returns the total number of extra efforts matching filters
func (r *ExtraEffortRepository) Count(tenantID uint, filters map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.Model(&entities.ExtraEffort{}).Where("tenant_id = ?", tenantID)

	// Apply filters (same as List)
	if clientID, ok := filters["client_id"].(uint); ok {
		query = query.Where("client_id = ?", clientID)
	}
	if sessionID, ok := filters["session_id"].(uint); ok {
		query = query.Where("session_id = ?", sessionID)
	}
	if billingStatus, ok := filters["billing_status"].(string); ok {
		query = query.Where("billing_status = ?", billingStatus)
	}
	if effortType, ok := filters["effort_type"].(string); ok {
		query = query.Where("effort_type = ?", effortType)
	}

	err := query.Count(&count).Error
	return count, err
}
