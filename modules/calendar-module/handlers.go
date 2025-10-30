package calendar

// Package calendar provides calendar event handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ae-base-server/api"
	"github.com/ae-base-server/pkg/utils"
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

// CreateCalendar creates a new calendar
// @Summary Create calendar
// @Description Create a new calendar for the authenticated user
// @Tags calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param calendar body CalendarCreateRequest true "Calendar data"
// @Success 201 {object} api.APIResponse{data=CalendarResponse}
// @Failure 400 {object} api.ErrorResponse
// @Failure 401 {object} api.ErrorResponse
// @Router /calendar/calendars [post]
func (h *Handler) CreateCalendar(c *gin.Context) {
	var req CalendarCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	user, err := api.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponseFunc("User not authenticated", err.Error()))
		return
	}

	calendar := Calendar{
		Name:               req.Name,
		ScheduleDaysAhead:  req.ScheduleDaysAhead,
		ScheduleDaysNotice: req.ScheduleDaysNotice,
		ScheduleUnit:       req.ScheduleUnit,
		UserID:             user.ID,
		TenantID:           user.TenantID,
	}

	if err := h.db.Create(&calendar).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Failed to create calendar", err.Error()))
		return
	}

	response := CalendarResponse{
		Calendar: calendar,
		User:     &api.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email},
	}

	c.JSON(http.StatusCreated, api.SuccessResponse("Calendar created successfully", response))
}

// GetCalendars retrieves calendars for the authenticated user
// @Summary Get calendars
// @Description Get paginated list of calendars for authenticated user
// @Tags calendar
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} api.APIResponse{data=CalendarListResponse}
// @Failure 401 {object} api.ErrorResponse
// @Router /calendar/calendars [get]
func (h *Handler) GetCalendars(c *gin.Context) {
	page, limit := utils.GetPaginationParams(c)

	user, err := api.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponseFunc("User not authenticated", err.Error()))
		return
	}

	query := h.db.Where("tenant_id = ?", user.TenantID)

	var total int64
	query.Model(&Calendar{}).Count(&total)

	var calendars []Calendar
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&calendars).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Failed to retrieve calendars", err.Error()))
		return
	}

	calendarResponses := make([]CalendarResponse, len(calendars))
	for i, calendar := range calendars {
		// Count events and recurring events for each calendar
		var eventCount, recurringCount int64
		h.db.Model(&Event{}).Where("calendar_id = ?", calendar.ID).Count(&eventCount)
		h.db.Model(&RecurringEvent{}).Where("calendar_id = ?", calendar.ID).Count(&recurringCount)

		calendarResponses[i] = CalendarResponse{
			Calendar:       calendar,
			User:           &api.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email},
			EventCount:     eventCount,
			RecurringCount: recurringCount,
		}
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	response := CalendarListResponse{
		Calendars: calendarResponses,
	}
	response.Pagination.Page = page
	response.Pagination.Limit = limit
	response.Pagination.Total = total
	response.Pagination.TotalPages = totalPages

	c.JSON(http.StatusOK, api.SuccessResponse("Calendars retrieved successfully", response))
}

// GetCalendar retrieves a specific calendar by ID
// @Summary Get calendar by ID
// @Description Get a specific calendar by its ID
// @Tags calendar
// @Produce json
// @Security BearerAuth
// @Param id path int true "Calendar ID"
// @Success 200 {object} api.APIResponse{data=CalendarResponse}
// @Failure 404 {object} api.ErrorResponse
// @Failure 401 {object} api.ErrorResponse
// @Router /calendar/calendars/{id} [get]
func (h *Handler) GetCalendar(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponseFunc("Invalid ID", "Calendar ID must be a number"))
		return
	}

	user, err := api.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponseFunc("User not authenticated", err.Error()))
		return
	}

	var calendar Calendar
	if err := h.db.Where("id = ? AND tenant_id = ?", id, user.TenantID).First(&calendar).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, api.ErrorResponseFunc("Calendar not found", "Calendar does not exist or access denied"))
		} else {
			c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Database error", err.Error()))
		}
		return
	}

	// Count events and recurring events
	var eventCount, recurringCount int64
	h.db.Model(&Event{}).Where("calendar_id = ?", calendar.ID).Count(&eventCount)
	h.db.Model(&RecurringEvent{}).Where("calendar_id = ?", calendar.ID).Count(&recurringCount)

	response := CalendarResponse{
		Calendar:       calendar,
		User:           &api.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email},
		EventCount:     eventCount,
		RecurringCount: recurringCount,
	}

	c.JSON(http.StatusOK, api.SuccessResponse("Calendar retrieved successfully", response))
}

// UpdateCalendar updates a calendar
// @Summary Update calendar
// @Description Update a calendar by its ID
// @Tags calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Calendar ID"
// @Param calendar body CalendarUpdateRequest true "Calendar update data"
// @Success 200 {object} api.APIResponse{data=CalendarResponse}
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 401 {object} api.ErrorResponse
// @Router /calendar/calendars/{id} [put]
func (h *Handler) UpdateCalendar(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponseFunc("Invalid ID", "Calendar ID must be a number"))
		return
	}

	var req CalendarUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	user, err := api.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponseFunc("User not authenticated", err.Error()))
		return
	}

	var calendar Calendar
	if err := h.db.Where("id = ? AND tenant_id = ?", id, user.TenantID).First(&calendar).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, api.ErrorResponseFunc("Calendar not found", "Calendar does not exist or access denied"))
		} else {
			c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Database error", err.Error()))
		}
		return
	}

	// Update fields
	if req.Name != "" {
		calendar.Name = req.Name
	}
	if req.ScheduleDaysAhead > 0 {
		calendar.ScheduleDaysAhead = req.ScheduleDaysAhead
	}
	if req.ScheduleDaysNotice > 0 {
		calendar.ScheduleDaysNotice = req.ScheduleDaysNotice
	}
	if req.ScheduleUnit != "" {
		calendar.ScheduleUnit = req.ScheduleUnit
	}

	if err := h.db.Save(&calendar).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Failed to update calendar", err.Error()))
		return
	}

	response := CalendarResponse{
		Calendar: calendar,
		User:     &api.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email},
	}

	c.JSON(http.StatusOK, api.SuccessResponse("Calendar updated successfully", response))
}

// DeleteCalendar deletes a calendar
// @Summary Delete calendar
// @Description Delete a calendar by its ID
// @Tags calendar
// @Produce json
// @Security BearerAuth
// @Param id path int true "Calendar ID"
// @Success 200 {object} api.APIResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 401 {object} api.ErrorResponse
// @Router /calendar/calendars/{id} [delete]
func (h *Handler) DeleteCalendar(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponseFunc("Invalid ID", "Calendar ID must be a number"))
		return
	}

	user, err := api.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponseFunc("User not authenticated", err.Error()))
		return
	}

	var calendar Calendar
	if err := h.db.Where("id = ? AND tenant_id = ?", id, user.TenantID).First(&calendar).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, api.ErrorResponseFunc("Calendar not found", "Calendar does not exist or access denied"))
		} else {
			c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Database error", err.Error()))
		}
		return
	}

	if err := h.db.Delete(&calendar).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Failed to delete calendar", err.Error()))
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse("Calendar deleted successfully", nil))
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
		Title:       req.Title,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Location:    req.Location,
		IsAllDay:    req.IsAllDay,
		EventType:   req.EventType,
		Priority:    req.Priority,
		UserID:      user.ID,
		TenantID:    user.TenantID, // Automatic tenant isolation
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
	page, limit := utils.GetPaginationParams(c)

	// Get user for tenant isolation
	user, err := api.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponseFunc("User not authenticated", err.Error()))
		return
	}

	// Build query with tenant isolation
	query := h.db.Where("tenant_id = ?", user.TenantID)

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
	if err := h.db.Where("id = ? AND tenant_id = ?", id, user.TenantID).First(&event).Error; err != nil {
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

// CreateRecurringEvent creates a new recurring event
// @Summary Create recurring event
// @Description Create a new recurring event pattern
// @Tags calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param recurring_event body RecurringEventCreateRequest true "Recurring event data"
// @Success 201 {object} api.APIResponse{data=RecurringEventResponse}
// @Failure 400 {object} api.ErrorResponse
// @Failure 401 {object} api.ErrorResponse
// @Router /calendar/recurring-events [post]
func (h *Handler) CreateRecurringEvent(c *gin.Context) {
	var req RecurringEventCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	user, err := api.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponseFunc("User not authenticated", err.Error()))
		return
	}

	// Verify calendar exists and belongs to user's tenant
	var calendar Calendar
	if err := h.db.Where("id = ? AND tenant_id = ?", req.CalendarID, user.TenantID).First(&calendar).Error; err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponseFunc("Invalid calendar", "Calendar not found or access denied"))
		return
	}

	recurringEvent := RecurringEvent{
		Title:        req.Title,
		Description:  req.Description,
		Location:     req.Location,
		Participants: req.Participants,
		Weekday:      req.Weekday,
		Interval:     req.Interval,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		CalendarID:   req.CalendarID,
		UserID:       user.ID,
		TenantID:     user.TenantID,
		IsActive:     true,
	}

	if err := h.db.Create(&recurringEvent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Failed to create recurring event", err.Error()))
		return
	}

	response := RecurringEventResponse{
		RecurringEvent: recurringEvent,
		User:           &api.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email},
		Calendar:       &CalendarResponse{Calendar: calendar},
	}

	c.JSON(http.StatusCreated, api.SuccessResponse("Recurring event created successfully", response))
}

// GetRecurringEvents retrieves recurring events for the authenticated user
// @Summary Get recurring events
// @Description Get paginated list of recurring events for authenticated user
// @Tags calendar
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param calendar_id query int false "Filter by calendar ID"
// @Success 200 {object} api.APIResponse{data=RecurringEventListResponse}
// @Failure 401 {object} api.ErrorResponse
// @Router /calendar/recurring-events [get]
func (h *Handler) GetRecurringEvents(c *gin.Context) {
	page, limit := utils.GetPaginationParams(c)

	user, err := api.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponseFunc("User not authenticated", err.Error()))
		return
	}

	query := h.db.Where("tenant_id = ?", user.TenantID)

	// Add calendar filter if provided
	if calendarID := c.Query("calendar_id"); calendarID != "" {
		query = query.Where("calendar_id = ?", calendarID)
	}

	var total int64
	query.Model(&RecurringEvent{}).Count(&total)

	var recurringEvents []RecurringEvent
	offset := (page - 1) * limit
	if err := query.Preload("Calendar").Offset(offset).Limit(limit).Order("created_at DESC").Find(&recurringEvents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Failed to retrieve recurring events", err.Error()))
		return
	}

	recurringEventResponses := make([]RecurringEventResponse, len(recurringEvents))
	for i, recurringEvent := range recurringEvents {
		// Count generated events for each recurring event
		var eventCount int64
		h.db.Model(&Event{}).Where("recurring_event_id = ?", recurringEvent.ID).Count(&eventCount)

		calendarResp := &CalendarResponse{Calendar: *recurringEvent.Calendar}
		recurringEventResponses[i] = RecurringEventResponse{
			RecurringEvent: recurringEvent,
			User:           &api.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email},
			Calendar:       calendarResp,
			EventCount:     eventCount,
		}
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	response := RecurringEventListResponse{
		RecurringEvents: recurringEventResponses,
	}
	response.Pagination.Page = page
	response.Pagination.Limit = limit
	response.Pagination.Total = total
	response.Pagination.TotalPages = totalPages

	c.JSON(http.StatusOK, api.SuccessResponse("Recurring events retrieved successfully", response))
}

// GetRecurringEvent retrieves a specific recurring event by ID
// @Summary Get recurring event by ID
// @Description Get a specific recurring event by its ID
// @Tags calendar
// @Produce json
// @Security BearerAuth
// @Param id path int true "Recurring Event ID"
// @Success 200 {object} api.APIResponse{data=RecurringEventResponse}
// @Failure 404 {object} api.ErrorResponse
// @Failure 401 {object} api.ErrorResponse
// @Router /calendar/recurring-events/{id} [get]
func (h *Handler) GetRecurringEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponseFunc("Invalid ID", "Recurring Event ID must be a number"))
		return
	}

	user, err := api.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponseFunc("User not authenticated", err.Error()))
		return
	}

	var recurringEvent RecurringEvent
	if err := h.db.Preload("Calendar").Where("id = ? AND tenant_id = ?", id, user.TenantID).First(&recurringEvent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, api.ErrorResponseFunc("Recurring event not found", "Recurring event does not exist or access denied"))
		} else {
			c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Database error", err.Error()))
		}
		return
	}

	// Count generated events
	var eventCount int64
	h.db.Model(&Event{}).Where("recurring_event_id = ?", recurringEvent.ID).Count(&eventCount)

	calendarResp := &CalendarResponse{Calendar: *recurringEvent.Calendar}
	response := RecurringEventResponse{
		RecurringEvent: recurringEvent,
		User:           &api.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email},
		Calendar:       calendarResp,
		EventCount:     eventCount,
	}

	c.JSON(http.StatusOK, api.SuccessResponse("Recurring event retrieved successfully", response))
}

// UpdateRecurringEvent updates a recurring event
// @Summary Update recurring event
// @Description Update a recurring event by its ID
// @Tags calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Recurring Event ID"
// @Param recurring_event body RecurringEventUpdateRequest true "Recurring event update data"
// @Success 200 {object} api.APIResponse{data=RecurringEventResponse}
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 401 {object} api.ErrorResponse
// @Router /calendar/recurring-events/{id} [put]
func (h *Handler) UpdateRecurringEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponseFunc("Invalid ID", "Recurring Event ID must be a number"))
		return
	}

	var req RecurringEventUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	user, err := api.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponseFunc("User not authenticated", err.Error()))
		return
	}

	var recurringEvent RecurringEvent
	if err := h.db.Preload("Calendar").Where("id = ? AND tenant_id = ?", id, user.TenantID).First(&recurringEvent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, api.ErrorResponseFunc("Recurring event not found", "Recurring event does not exist or access denied"))
		} else {
			c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Database error", err.Error()))
		}
		return
	}

	// Update fields
	if req.Title != "" {
		recurringEvent.Title = req.Title
	}
	if req.Description != "" {
		recurringEvent.Description = req.Description
	}
	if req.Location != "" {
		recurringEvent.Location = req.Location
	}
	if req.Participants != "" {
		recurringEvent.Participants = req.Participants
	}
	if req.Weekday >= 0 && req.Weekday <= 6 {
		recurringEvent.Weekday = req.Weekday
	}
	if req.Interval > 0 {
		recurringEvent.Interval = req.Interval
	}
	if req.StartTime != "" {
		recurringEvent.StartTime = req.StartTime
	}
	if req.EndTime != "" {
		recurringEvent.EndTime = req.EndTime
	}
	if req.IsActive != nil {
		recurringEvent.IsActive = *req.IsActive
	}

	if err := h.db.Save(&recurringEvent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Failed to update recurring event", err.Error()))
		return
	}

	calendarResp := &CalendarResponse{Calendar: *recurringEvent.Calendar}
	response := RecurringEventResponse{
		RecurringEvent: recurringEvent,
		User:           &api.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email},
		Calendar:       calendarResp,
	}

	c.JSON(http.StatusOK, api.SuccessResponse("Recurring event updated successfully", response))
}

// DeleteRecurringEvent deletes a recurring event
// @Summary Delete recurring event
// @Description Delete a recurring event by its ID
// @Tags calendar
// @Produce json
// @Security BearerAuth
// @Param id path int true "Recurring Event ID"
// @Success 200 {object} api.APIResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 401 {object} api.ErrorResponse
// @Router /calendar/recurring-events/{id} [delete]
func (h *Handler) DeleteRecurringEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponseFunc("Invalid ID", "Recurring Event ID must be a number"))
		return
	}

	user, err := api.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.ErrorResponseFunc("User not authenticated", err.Error()))
		return
	}

	var recurringEvent RecurringEvent
	if err := h.db.Where("id = ? AND tenant_id = ?", id, user.TenantID).First(&recurringEvent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, api.ErrorResponseFunc("Recurring event not found", "Recurring event does not exist or access denied"))
		} else {
			c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Database error", err.Error()))
		}
		return
	}

	if err := h.db.Delete(&recurringEvent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponseFunc("Failed to delete recurring event", err.Error()))
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse("Recurring event deleted successfully", nil))
}
