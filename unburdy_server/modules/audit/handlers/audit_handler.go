package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	baseAPI "github.com/ae-base-server/api"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/unburdy-server-api/internal/models"
	"github.com/unburdy/unburdy-server-api/modules/audit/entities"
	"github.com/unburdy/unburdy-server-api/modules/audit/services"
)

type AuditHandler struct {
	service *services.AuditService
}

func NewAuditHandler(service *services.AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

// GetAuditLogs retrieves audit logs with filtering and pagination
// @Summary Get audit logs
// @Description Retrieve audit logs with optional filtering by user, entity, action, and date range. Supports pagination.
// @Tags audit
// @ID getAuditLogs
// @Produce json
// @Param user_id query int false "Filter by user ID" example(1)
// @Param entity_type query string false "Filter by entity type" Enums(invoice, invoice_item, session, extra_effort) example(invoice)
// @Param entity_id query int false "Filter by entity ID" example(123)
// @Param action query string false "Filter by action" Enums(invoice_draft_created, invoice_draft_updated, invoice_draft_cancelled, invoice_finalized, invoice_sent, invoice_marked_paid, invoice_marked_overdue, reminder_sent, credit_note_created, xrechnung_exported) example(invoice_finalized)
// @Param start_date query string false "Filter by start date (RFC3339)" example(2026-01-01T00:00:00Z)
// @Param end_date query string false "Filter by end date (RFC3339)" example(2026-12-31T23:59:59Z)
// @Param page query int false "Page number (default: 1)" example(1)
// @Param limit query int false "Items per page (default: 50, max: 100)" example(50)
// @Success 200 {object} entities.AuditLogListResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /audit/logs [get]
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	filter := entities.AuditLogFilter{TenantID: tenantID}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err == nil {
			uid := uint(userID)
			filter.UserID = &uid
		}
	}

	if entityTypeStr := c.Query("entity_type"); entityTypeStr != "" {
		entityType := entities.EntityType(entityTypeStr)
		filter.EntityType = &entityType
	}

	if entityIDStr := c.Query("entity_id"); entityIDStr != "" {
		entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
		if err == nil {
			eid := uint(entityID)
			filter.EntityID = &eid
		}
	}

	if actionStr := c.Query("action"); actionStr != "" {
		action := entities.AuditAction(actionStr)
		filter.Action = &action
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err == nil {
			filter.EndDate = &endDate
		}
	}

	filter.Page = 1
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	filter.Limit = 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	logs, total, err := h.service.GetAuditLogs(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	responses := make([]entities.AuditLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = log.ToResponse()
	}

	c.JSON(http.StatusOK, entities.AuditLogListResponse{
		Success: true,
		Message: "Audit logs retrieved successfully",
		Data:    responses,
		Page:    filter.Page,
		Limit:   filter.Limit,
		Total:   total,
	})
}

// GetEntityAuditLogs retrieves audit logs for a specific entity
// @Summary Get audit logs for a specific entity
// @Description Retrieve complete audit trail for a specific entity (e.g., all changes to an invoice)
// @Tags audit
// @ID getEntityAuditLogs
// @Produce json
// @Param entity_type path string true "Entity type" Enums(invoice, invoice_item, session, extra_effort) example(invoice)
// @Param entity_id path int true "Entity ID" example(123)
// @Success 200 {object} entities.AuditLogListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /audit/entity/{entity_type}/{entity_id} [get]
func (h *AuditHandler) GetEntityAuditLogs(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	entityType := entities.EntityType(c.Param("entity_type"))
	entityID, err := strconv.ParseUint(c.Param("entity_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Invalid entity ID"))
		return
	}

	logs, err := h.service.GetAuditLogsByEntity(tenantID, uint(entityID), entityType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	responses := make([]entities.AuditLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = log.ToResponse()
	}

	c.JSON(http.StatusOK, entities.AuditLogListResponse{
		Success: true,
		Message: "Entity audit logs retrieved successfully",
		Data:    responses,
		Page:    1,
		Limit:   len(responses),
		Total:   int64(len(responses)),
	})
}

// ExportAuditLogs exports audit logs to CSV format for GoBD compliance
// @Summary Export audit logs to CSV
// @Description Export audit logs in CSV format for tax authority compliance (GoBD). Supports same filtering options as GET /audit/logs.
// @Tags audit
// @ID exportAuditLogs
// @Produce text/csv
// @Param user_id query int false "Filter by user ID" example(1)
// @Param entity_type query string false "Filter by entity type" Enums(invoice, invoice_item, session, extra_effort) example(invoice)
// @Param entity_id query int false "Filter by entity ID" example(123)
// @Param action query string false "Filter by action" example(invoice_finalized)
// @Param start_date query string false "Filter by start date (RFC3339)" example(2026-01-01T00:00:00Z)
// @Param end_date query string false "Filter by end date (RFC3339)" example(2026-12-31T23:59:59Z)
// @Success 200 {file} string "CSV file download"
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /audit/export [get]
func (h *AuditHandler) ExportAuditLogs(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	filter := entities.AuditLogFilter{TenantID: tenantID}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err == nil {
			uid := uint(userID)
			filter.UserID = &uid
		}
	}

	if entityTypeStr := c.Query("entity_type"); entityTypeStr != "" {
		entityType := entities.EntityType(entityTypeStr)
		filter.EntityType = &entityType
	}

	if entityIDStr := c.Query("entity_id"); entityIDStr != "" {
		entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
		if err == nil {
			eid := uint(entityID)
			filter.EntityID = &eid
		}
	}

	if actionStr := c.Query("action"); actionStr != "" {
		action := entities.AuditAction(actionStr)
		filter.Action = &action
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err == nil {
			filter.EndDate = &endDate
		}
	}

	csv, err := h.service.ExportAuditLogsToCSV(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	filename := fmt.Sprintf("audit_logs_%s.csv", time.Now().Format("20060102_150405"))

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.String(http.StatusOK, csv)
}

// GetAuditStatistics retrieves audit log statistics and analytics
// @Summary Get audit statistics
// @Description Retrieve aggregated statistics for audit logs (action counts, user activity, entity type distribution). Returns statistics including total logs, action counts, user activity, and entity type distribution.
// @Tags audit
// @ID getAuditStatistics
// @Produce json
// @Param start_date query string false "Statistics start date (RFC3339)" example(2026-01-01T00:00:00Z)
// @Param end_date query string false "Statistics end date (RFC3339)" example(2026-12-31T23:59:59Z)
// @Success 200 {object} map[string]interface{} "Audit statistics with action_counts, user_activity, entity_type_counts, total_logs, date_range"
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /audit/statistics [get]
func (h *AuditHandler) GetAuditStatistics(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	var startDate, endDate *time.Time

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		sd, err := time.Parse(time.RFC3339, startDateStr)
		if err == nil {
			startDate = &sd
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		ed, err := time.Parse(time.RFC3339, endDateStr)
		if err == nil {
			endDate = &ed
		}
	}

	stats, err := h.service.GetAuditStatistics(tenantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Audit statistics retrieved successfully", stats))
}
