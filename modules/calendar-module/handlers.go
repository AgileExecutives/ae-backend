package calendar
// Package calendar provides calendar event handlers
package calendar

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ae-saas-basic/ae-saas-basic/api"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler manages calendar operations
type Handler struct {
	db *gorm.DB
}

// NewHandler creates a new calendar handler
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// CreateEvent creates a new calendar event
// @Summary Create calendar event
// @Description Create a new calendar event for the authenticated user
// @Tags calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param event body EventCreateRequest true "Event data"
// @Success 201 {object} api.APIResponse{data=EventResponse}
// @Failure 400 {object} api.ErrorResponse
// @Failure 401 {object} api.ErrorResponse
// @Router /calendar/events [post]
func (h *Handler) CreateEvent(c *gin.Context) {
	var req EventCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	// Get user from auth middleware (provided by base-server)
	user, err := api.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponseFunc("User not authenticated", err.Error()))
		return
	}

	// Create event with tenant isolation
	event := Event{
		Title:          req.Title,
		Description:    req.Description,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Location:       req.Location,
		IsAllDay:       req.IsAllDay,
		EventType:      req.EventType,
		Priority:       req.Priority,
		UserID:         user.ID,
		OrganizationID: user.OrganizationID, // Automatic tenant isolation
	}

	// Validate event times
	if event.EndTime.Before(event.StartTime) {
		c.JSON(http.StatusBadRequest, api.ErrorResponseFunc("Invalid time", "End time must be after start time"))
		return
	}

	if err := h.db.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Failed to create event", err.Error()))
		return
	}

	// Prepare response
	response := EventResponse{
		Event: event,
		User:  &api.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email},
	}

	c.JSON(http.StatusCreated, api.SuccessResponse("Event created successfully", response))
}

// GetEvents retrieves events for the authenticated user
// @Summary Get calendar events
// @Description Get paginated list of calendar events for authenticated user's organization
// @Tags calendar
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param start_date query string false "Start date filter (RFC3339)"
// @Param end_date query string false "End date filter (RFC3339)"
// @Success 200 {object} api.APIResponse{data=EventListResponse}
// @Failure 401 {object} api.ErrorResponse
// @Router /calendar/events [get]
func (h *Handler) GetEvents(c *gin.Context) {
	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get user for tenant isolation
	user, err := api.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponseFunc("User not authenticated", err.Error()))
		return
	}

	// Build query with tenant isolation
	query := h.db.Where("organization_id = ?", user.OrganizationID)

	// Add date filters if provided
	if startDate := c.Query("start_date"); startDate != "" {
		if start, err := time.Parse(time.RFC3339, startDate); err == nil {
			query = query.Where("start_time >= ?", start)
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if end, err := time.Parse(time.RFC3339, endDate); err == nil {
			query = query.Where("end_time <= ?", end)
		}
	}

	// Get total count
	var total int64
	query.Model(&Event{}).Count(&total)

	// Get events with pagination
	var events []Event
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("start_time ASC").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Failed to retrieve events", err.Error()))
		return
	}

	// Convert to response format
	eventResponses := make([]EventResponse, len(events))
	for i, event := range events {
		eventResponses[i] = EventResponse{
			Event: event,
			User:  &api.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email},
		}
	}

	// Calculate pagination
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	response := EventListResponse{
		Events: eventResponses,
	}
	response.Pagination.Page = page
	response.Pagination.Limit = limit
	response.Pagination.Total = total
	response.Pagination.TotalPages = totalPages

	c.JSON(http.StatusOK, api.SuccessResponse("Events retrieved successfully", response))
}

// GetEvent retrieves a specific event by ID
// @Summary Get calendar event by ID
// @Description Get a specific calendar event by its ID
// @Tags calendar
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Success 200 {object} api.APIResponse{data=EventResponse}
// @Failure 404 {object} api.ErrorResponse
// @Failure 401 {object} api.ErrorResponse
// @Router /calendar/events/{id} [get]
func (h *Handler) GetEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponseFunc("Invalid ID", "Event ID must be a number"))
		return
	}

	user, err := api.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponseFunc("User not authenticated", err.Error()))
		return
	}

	var event Event
	// Enforce tenant isolation
	if err := h.db.Where("id = ? AND organization_id = ?", id, user.OrganizationID).First(&event).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, api.ErrorResponseFunc("Event not found", "Event does not exist or access denied"))
		} else {
			c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Database error", err.Error()))
		}
		return
	}

	response := EventResponse{
		Event: event,
		User:  &api.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email},
	}

	c.JSON(http.StatusOK, api.SuccessResponse("Event retrieved successfully", response))
}