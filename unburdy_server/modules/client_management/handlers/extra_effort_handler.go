package handlers

import (
	"net/http"
	"strconv"
	"time"

	baseAPI "github.com/ae-base-server/api"
	"github.com/gin-gonic/gin"
	_ "github.com/unburdy/unburdy-server-api/internal/models" // imported for swagger documentation
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"github.com/unburdy/unburdy-server-api/modules/client_management/services"
)

// ExtraEffortHandler handles extra effort-related HTTP requests
type ExtraEffortHandler struct {
	service *services.ExtraEffortService
}

// NewExtraEffortHandler creates a new extra effort handler
func NewExtraEffortHandler(service *services.ExtraEffortService) *ExtraEffortHandler {
	return &ExtraEffortHandler{
		service: service,
	}
}

// CreateExtraEffort handles creating a new extra effort
// @Summary Create extra effort
// @Description Record extra therapeutic work (preparation, consultation, meeting, etc.)
// @Tags extra-efforts
// @ID createExtraEffort
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param effort body entities.CreateExtraEffortRequest true "Extra effort information"
// @Success 201 {object} entities.ExtraEffortAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /extra-efforts [post]
func (h *ExtraEffortHandler) CreateExtraEffort(c *gin.Context) {
	var req entities.CreateExtraEffortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get user ID: "+err.Error()))
		return
	}

	effort, err := h.service.CreateExtraEffort(&req, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("Failed to create extra effort", err.Error()))
		return
	}

	response := entities.ExtraEffortResponse{
		ID:            effort.ID,
		ClientID:      effort.ClientID,
		SessionID:     effort.SessionID,
		EffortType:    effort.EffortType,
		EffortDate:    effort.EffortDate,
		DurationMin:   effort.DurationMin,
		Description:   effort.Description,
		Billable:      effort.Billable,
		BillingStatus: effort.BillingStatus,
		CreatedAt:     effort.CreatedAt,
	}

	c.JSON(http.StatusCreated, entities.ExtraEffortAPIResponse{
		Success: true,
		Message: "Extra effort created successfully",
		Data:    response,
	})
}

// ListExtraEfforts handles listing extra efforts with filters
// @Summary List extra efforts
// @Description Retrieve extra efforts with optional filters
// @Tags extra-efforts
// @ID listExtraEfforts
// @Produce json
// @Security BearerAuth
// @Param client_id query int false "Filter by client ID"
// @Param session_id query int false "Filter by session ID"
// @Param billing_status query string false "Filter by billing status (unbilled, billed, excluded)"
// @Param effort_type query string false "Filter by effort type"
// @Param from_date query string false "Filter from date (YYYY-MM-DD)"
// @Param to_date query string false "Filter to date (YYYY-MM-DD)"
// @Success 200 {object} entities.ExtraEffortListAPIResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /extra-efforts [get]
func (h *ExtraEffortHandler) ListExtraEfforts(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	// Build filters
	filters := make(map[string]interface{})

	if clientIDStr := c.Query("client_id"); clientIDStr != "" {
		clientID, err := strconv.ParseUint(clientIDStr, 10, 32)
		if err == nil {
			filters["client_id"] = uint(clientID)
		}
	}

	if sessionIDStr := c.Query("session_id"); sessionIDStr != "" {
		sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
		if err == nil {
			filters["session_id"] = uint(sessionID)
		}
	}

	if billingStatus := c.Query("billing_status"); billingStatus != "" {
		filters["billing_status"] = billingStatus
	}

	if effortType := c.Query("effort_type"); effortType != "" {
		filters["effort_type"] = effortType
	}

	if fromDateStr := c.Query("from_date"); fromDateStr != "" {
		fromDate, err := time.Parse("2006-01-02", fromDateStr)
		if err == nil {
			filters["from_date"] = fromDate
		}
	}

	if toDateStr := c.Query("to_date"); toDateStr != "" {
		toDate, err := time.Parse("2006-01-02", toDateStr)
		if err == nil {
			filters["to_date"] = toDate
		}
	}

	efforts, total, err := h.service.ListExtraEfforts(tenantID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("Failed to list extra efforts", err.Error()))
		return
	}

	// Convert to response format
	responses := make([]entities.ExtraEffortResponse, 0, len(efforts))
	for _, effort := range efforts {
		resp := entities.ExtraEffortResponse{
			ID:            effort.ID,
			ClientID:      effort.ClientID,
			SessionID:     effort.SessionID,
			EffortType:    effort.EffortType,
			EffortDate:    effort.EffortDate,
			DurationMin:   effort.DurationMin,
			Description:   effort.Description,
			Billable:      effort.Billable,
			BillingStatus: effort.BillingStatus,
			CreatedAt:     effort.CreatedAt,
		}
		responses = append(responses, resp)
	}

	c.JSON(http.StatusOK, entities.ExtraEffortListAPIResponse{
		Success: true,
		Message: "Extra efforts retrieved successfully",
		Data:    responses,
		Total:   int(total),
	})
}

// GetExtraEffort handles retrieving a specific extra effort
// @Summary Get extra effort by ID
// @Description Retrieve an extra effort by its ID
// @Tags extra-efforts
// @ID getExtraEffort
// @Produce json
// @Security BearerAuth
// @Param id path int true "Extra effort ID"
// @Success 200 {object} entities.ExtraEffortAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /extra-efforts/{id} [get]
func (h *ExtraEffortHandler) GetExtraEffort(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", "Invalid extra effort ID"))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	effort, err := h.service.GetExtraEffort(uint(id), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("Not found", "Extra effort not found"))
		return
	}

	response := entities.ExtraEffortResponse{
		ID:            effort.ID,
		ClientID:      effort.ClientID,
		SessionID:     effort.SessionID,
		EffortType:    effort.EffortType,
		EffortDate:    effort.EffortDate,
		DurationMin:   effort.DurationMin,
		Description:   effort.Description,
		Billable:      effort.Billable,
		BillingStatus: effort.BillingStatus,
		CreatedAt:     effort.CreatedAt,
	}

	c.JSON(http.StatusOK, entities.ExtraEffortAPIResponse{
		Success: true,
		Message: "Extra effort retrieved successfully",
		Data:    response,
	})
}

// UpdateExtraEffort handles updating an extra effort
// @Summary Update extra effort
// @Description Update an existing extra effort (only if unbilled)
// @Tags extra-efforts
// @ID updateExtraEffort
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Extra effort ID"
// @Param effort body entities.UpdateExtraEffortRequest true "Updated extra effort information"
// @Success 200 {object} entities.ExtraEffortAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /extra-efforts/{id} [put]
func (h *ExtraEffortHandler) UpdateExtraEffort(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", "Invalid extra effort ID"))
		return
	}

	var req entities.UpdateExtraEffortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	if err := h.service.UpdateExtraEffort(uint(id), tenantID, &req); err != nil {
		if err.Error() == "extra effort not found: record not found" {
			c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("Not found", "Extra effort not found"))
			return
		}
		if err.Error() == "cannot update billed extra efforts" {
			c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Cannot update", "Cannot update billed extra efforts"))
			return
		}
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("Failed to update extra effort", err.Error()))
		return
	}

	c.JSON(http.StatusOK, baseAPI.SuccessResponse("Extra effort updated successfully", nil))
}

// DeleteExtraEffort handles deleting an extra effort
// @Summary Delete extra effort
// @Description Delete an extra effort (only if unbilled)
// @Tags extra-efforts
// @ID deleteExtraEffort
// @Produce json
// @Security BearerAuth
// @Param id path int true "Extra effort ID"
// @Success 200 {object} entities.ExtraEffortDeleteResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /extra-efforts/{id} [delete]
func (h *ExtraEffortHandler) DeleteExtraEffort(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", "Invalid extra effort ID"))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	if err := h.service.DeleteExtraEffort(uint(id), tenantID); err != nil {
		if err.Error() == "extra effort not found: record not found" {
			c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("Not found", "Extra effort not found"))
			return
		}
		if err.Error() == "cannot delete billed extra efforts" {
			c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Cannot delete", "Cannot delete billed extra efforts"))
			return
		}
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("Failed to delete extra effort", err.Error()))
		return
	}

	c.JSON(http.StatusOK, entities.ExtraEffortDeleteResponse{
		Success: true,
		Message: "Extra effort deleted successfully",
	})
}
