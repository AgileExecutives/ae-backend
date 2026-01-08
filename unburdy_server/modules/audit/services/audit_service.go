package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/unburdy/unburdy-server-api/modules/audit/entities"
	"gorm.io/gorm"
)

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

type LogEventRequest struct {
	TenantID   uint
	UserID     uint
	EntityType entities.EntityType
	EntityID   uint
	Action     entities.AuditAction
	Metadata   *entities.AuditLogMetadata
	IPAddress  string
	UserAgent  string
}

func (s *AuditService) LogEvent(req LogEventRequest) error {
	auditLog := entities.AuditLog{
		TenantID:   req.TenantID,
		UserID:     req.UserID,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		Action:     req.Action,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
		CreatedAt:  time.Now(),
	}

	if req.Metadata != nil {
		metadataJSON, err := req.Metadata.ToJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		auditLog.Metadata = metadataJSON
	}

	if err := s.db.Create(&auditLog).Error; err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

func (s *AuditService) GetAuditLogs(filter entities.AuditLogFilter) ([]entities.AuditLog, int64, error) {
	var logs []entities.AuditLog
	var total int64

	query := s.db.Model(&entities.AuditLog{}).Where("tenant_id = ?", filter.TenantID)

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}

	if filter.EntityType != nil {
		query = query.Where("entity_type = ?", *filter.EntityType)
	}

	if filter.EntityID != nil {
		query = query.Where("entity_id = ?", *filter.EntityID)
	}

	if filter.Action != nil {
		query = query.Where("action = ?", *filter.Action)
	}

	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}

	limit := filter.Limit
	if limit < 1 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	offset := (page - 1) * limit

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch audit logs: %w", err)
	}

	return logs, total, nil
}

func (s *AuditService) GetAuditLogsByEntity(tenantID, entityID uint, entityType entities.EntityType) ([]entities.AuditLog, error) {
	var logs []entities.AuditLog

	if err := s.db.Where("tenant_id = ? AND entity_type = ? AND entity_id = ?", tenantID, entityType, entityID).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch entity audit logs: %w", err)
	}

	return logs, nil
}

func (s *AuditService) ExportAuditLogsToCSV(filter entities.AuditLogFilter) (string, error) {
	filter.Limit = 0
	filter.Page = 0

	var logs []entities.AuditLog

	query := s.db.Model(&entities.AuditLog{}).Where("tenant_id = ?", filter.TenantID)

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}

	if filter.EntityType != nil {
		query = query.Where("entity_type = ?", *filter.EntityType)
	}

	if filter.EntityID != nil {
		query = query.Where("entity_id = ?", *filter.EntityID)
	}

	if filter.Action != nil {
		query = query.Where("action = ?", *filter.Action)
	}

	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	if err := query.Order("created_at ASC").Find(&logs).Error; err != nil {
		return "", fmt.Errorf("failed to fetch audit logs for export: %w", err)
	}

	csv := "ID,Tenant ID,User ID,Entity Type,Entity ID,Action,Metadata,IP Address,User Agent,Created At\n"

	for _, log := range logs {
		metadataStr := ""
		if len(log.Metadata) > 0 {
			var metadata map[string]interface{}
			if err := json.Unmarshal(log.Metadata, &metadata); err == nil {
				metadataBytes, _ := json.Marshal(metadata)
				metadataStr = string(metadataBytes)
			}
		}

		csv += fmt.Sprintf("%d,%d,%d,%s,%d,%s,\"%s\",\"%s\",\"%s\",%s\n",
			log.ID,
			log.TenantID,
			log.UserID,
			log.EntityType,
			log.EntityID,
			log.Action,
			escapeCSV(metadataStr),
			escapeCSV(log.IPAddress),
			escapeCSV(log.UserAgent),
			log.CreatedAt.Format(time.RFC3339),
		)
	}

	return csv, nil
}

func escapeCSV(s string) string {
	escaped := ""
	for _, ch := range s {
		if ch == '"' {
			escaped += "\"\""
		} else {
			escaped += string(ch)
		}
	}
	return escaped
}

func (s *AuditService) GetAuditStatistics(tenantID uint, startDate, endDate *time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := s.db.Model(&entities.AuditLog{}).Where("tenant_id = ?", tenantID)

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total logs: %w", err)
	}
	stats["total_logs"] = total

	var actionCounts []struct {
		Action entities.AuditAction
		Count  int64
	}
	if err := query.Select("action, COUNT(*) as count").Group("action").Scan(&actionCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to count by action: %w", err)
	}

	actionStats := make(map[string]int64)
	for _, ac := range actionCounts {
		actionStats[string(ac.Action)] = ac.Count
	}
	stats["by_action"] = actionStats

	var entityCounts []struct {
		EntityType entities.EntityType
		Count      int64
	}
	if err := query.Select("entity_type, COUNT(*) as count").Group("entity_type").Scan(&entityCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to count by entity type: %w", err)
	}

	entityStats := make(map[string]int64)
	for _, ec := range entityCounts {
		entityStats[string(ec.EntityType)] = ec.Count
	}
	stats["by_entity_type"] = entityStats

	return stats, nil
}
