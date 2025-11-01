package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	baseAPI "github.com/ae-base-server/api"
	"github.com/unburdy/calendar-module/entities"
	"github.com/unburdy/calendar-module/services"
)

// CalendarHandler handles calendar-related HTTP requests
type CalendarHandler struct {
	service *services.CalendarService
}

// NewCalendarHandler creates a new calendar handler
func NewCalendarHandler(service *services.CalendarService) *CalendarHandler {
	return &CalendarHandler{
		service: service,
	}
}

// Calendar CRUD Handlers

// CreateCalendar creates a new calendar
// @Summary Create a new calendar
// @Description Create a new calendar for the authenticated user
// @Tags calendar
// @Accept json
// @Produce json
// @Param calendar body entities.CreateCalendarRequest true "Calendar data"
// @Success 201 {object} entities.CalendarResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar [post]
func (h *CalendarHandler) CreateCalendar(c *gin.Context) {
	var req entities.CreateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant ID and user ID from context (set by auth middleware)
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	calendar, err := h.service.CreateCalendar(req, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, calendar.ToResponse())
}

// GetCalendar retrieves a specific calendar
// @Summary Get calendar by ID
// @Description Retrieve a calendar by its ID
// @Tags calendar
// @Produce json
// @Param id path int true "Calendar ID"
// @Success 200 {object} entities.CalendarResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar/{id} [get]
func (h *CalendarHandler) GetCalendar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calendar ID"})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	calendar, err := h.service.GetCalendarByID(uint(id), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, calendar.ToResponse())
}

// GetCalendarsWithMetadata retrieves all calendars with 2-level deep preloading
// @Summary Get calendars with complete metadata
// @Description Retrieve all calendars for the authenticated user with 2-level deep preloading including entries with their series and series with their entries
// @Tags calendar
// @Produce json
// @Success 200 {object} map[string]interface{} "Returns calendars array with complete metadata including nested relationships"
// @Success 200 {array} entities.CalendarResponse "Array of calendars with preloaded entries, series, and external calendars"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 500 {object} map[string]interface{} "Internal server error during calendar retrieval"
// @Router /calendar [get]
// @Security BearerAuth
func (h *CalendarHandler) GetCalendarsWithMetadata(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	calendars, err := h.service.GetCalendarsWithDeepPreload(tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []entities.CalendarResponse
	for _, calendar := range calendars {
		responses = append(responses, calendar.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"calendars": responses,
	})
}

// UpdateCalendar updates an existing calendar
// @Summary Update calendar
// @Description Update an existing calendar
// @Tags calendar
// @Accept json
// @Produce json
// @Param id path int true "Calendar ID"
// @Param calendar body entities.UpdateCalendarRequest true "Updated calendar data"
// @Success 200 {object} entities.CalendarResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar/{id} [put]
func (h *CalendarHandler) UpdateCalendar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calendar ID"})
		return
	}

	var req entities.UpdateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	calendar, err := h.service.UpdateCalendar(uint(id), tenantID, userID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, calendar.ToResponse())
}

// DeleteCalendar deletes a calendar
// @Summary Delete calendar
// @Description Delete a calendar by ID
// @Tags calendar
// @Produce json
// @Param id path int true "Calendar ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar/{id} [delete]
func (h *CalendarHandler) DeleteCalendar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calendar ID"})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	err = h.service.DeleteCalendar(uint(id), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Calendar deleted successfully"})
}

// Calendar Entry CRUD Handlers

// CreateCalendarEntry creates a new calendar entry
// @Summary Create a new calendar entry
// @Description Create a new calendar entry
// @Tags calendar-entries
// @Accept json
// @Produce json
// @Param entry body entities.CreateCalendarEntryRequest true "Calendar entry data"
// @Success 201 {object} entities.CalendarEntryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar-entries [post]
func (h *CalendarHandler) CreateCalendarEntry(c *gin.Context) {
	var req entities.CreateCalendarEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	entry, err := h.service.CreateCalendarEntry(req, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry.ToResponse())
}

// GetCalendarEntry retrieves a specific calendar entry
// @Summary Get calendar entry by ID
// @Description Retrieve a calendar entry by its ID
// @Tags calendar-entries
// @Produce json
// @Param id path int true "Calendar Entry ID"
// @Success 200 {object} entities.CalendarEntryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar-entries/{id} [get]
func (h *CalendarHandler) GetCalendarEntry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calendar entry ID"})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	entry, err := h.service.GetCalendarEntryByID(uint(id), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry.ToResponse())
}

// GetAllCalendarEntries retrieves all calendar entries with pagination
// @Summary Get all calendar entries
// @Description Retrieve all calendar entries for the authenticated user
// @Tags calendar-entries
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar-entries [get]
func (h *CalendarHandler) GetAllCalendarEntries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	entries, total, err := h.service.GetAllCalendarEntries(page, limit, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []entities.CalendarEntryResponse
	for _, entry := range entries {
		responses = append(responses, entry.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": responses,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// UpdateCalendarEntry updates an existing calendar entry
// @Summary Update calendar entry
// @Description Update an existing calendar entry
// @Tags calendar-entries
// @Accept json
// @Produce json
// @Param id path int true "Calendar Entry ID"
// @Param entry body entities.UpdateCalendarEntryRequest true "Updated calendar entry data"
// @Success 200 {object} entities.CalendarEntryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar-entries/{id} [put]
func (h *CalendarHandler) UpdateCalendarEntry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calendar entry ID"})
		return
	}

	var req entities.UpdateCalendarEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	entry, err := h.service.UpdateCalendarEntry(uint(id), tenantID, userID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry.ToResponse())
}

// DeleteCalendarEntry deletes a calendar entry
// @Summary Delete calendar entry
// @Description Delete a calendar entry by ID
// @Tags calendar-entries
// @Produce json
// @Param id path int true "Calendar Entry ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar-entries/{id} [delete]
func (h *CalendarHandler) DeleteCalendarEntry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calendar entry ID"})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	err = h.service.DeleteCalendarEntry(uint(id), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Calendar entry deleted successfully"})
}

// Calendar Series CRUD Handlers

// CreateCalendarSeries creates a new calendar series
// @Summary Create a new calendar series
// @Description Create a new calendar series for recurring events
// @Tags calendar-series
// @Accept json
// @Produce json
// @Param series body entities.CreateCalendarSeriesRequest true "Calendar series data"
// @Success 201 {object} entities.CalendarSeriesResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar-series [post]
func (h *CalendarHandler) CreateCalendarSeries(c *gin.Context) {
	var req entities.CreateCalendarSeriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	series, err := h.service.CreateCalendarSeries(req, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, series.ToResponse())
}

// GetCalendarSeries retrieves a specific calendar series
// @Summary Get calendar series by ID
// @Description Retrieve a calendar series by its ID
// @Tags calendar-series
// @Produce json
// @Param id path int true "Calendar Series ID"
// @Success 200 {object} entities.CalendarSeriesResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar-series/{id} [get]
func (h *CalendarHandler) GetCalendarSeries(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calendar series ID"})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	series, err := h.service.GetCalendarSeriesByID(uint(id), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, series.ToResponse())
}

// GetAllCalendarSeries retrieves all calendar series with pagination
// @Summary Get all calendar series
// @Description Retrieve all calendar series for the authenticated user
// @Tags calendar-series
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar-series [get]
func (h *CalendarHandler) GetAllCalendarSeries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	series, total, err := h.service.GetAllCalendarSeries(page, limit, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []entities.CalendarSeriesResponse
	for _, s := range series {
		responses = append(responses, s.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"series": responses,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// UpdateCalendarSeries updates an existing calendar series
// @Summary Update calendar series
// @Description Update an existing calendar series
// @Tags calendar-series
// @Accept json
// @Produce json
// @Param id path int true "Calendar Series ID"
// @Param series body entities.UpdateCalendarSeriesRequest true "Updated calendar series data"
// @Success 200 {object} entities.CalendarSeriesResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar-series/{id} [put]
func (h *CalendarHandler) UpdateCalendarSeries(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calendar series ID"})
		return
	}

	var req entities.UpdateCalendarSeriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	series, err := h.service.UpdateCalendarSeries(uint(id), tenantID, userID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, series.ToResponse())
}

// DeleteCalendarSeries deletes a calendar series
// @Summary Delete calendar series
// @Description Delete a calendar series by ID
// @Tags calendar-series
// @Produce json
// @Param id path int true "Calendar Series ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar-series/{id} [delete]
func (h *CalendarHandler) DeleteCalendarSeries(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calendar series ID"})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	err = h.service.DeleteCalendarSeries(uint(id), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Calendar series deleted successfully"})
}

// External Calendar CRUD Handlers

// CreateExternalCalendar creates a new external calendar
// @Summary Create a new external calendar
// @Description Create a new external calendar
// @Tags external-calendars
// @Accept json
// @Produce json
// @Param external body entities.CreateExternalCalendarRequest true "External calendar data"
// @Success 201 {object} entities.ExternalCalendarResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /external-calendars [post]
func (h *CalendarHandler) CreateExternalCalendar(c *gin.Context) {
	var req entities.CreateExternalCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	external, err := h.service.CreateExternalCalendar(req, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, external.ToResponse())
}

// GetExternalCalendar retrieves a specific external calendar
// @Summary Get external calendar by ID
// @Description Retrieve an external calendar by its ID
// @Tags external-calendars
// @Produce json
// @Param id path int true "External Calendar ID"
// @Success 200 {object} entities.ExternalCalendarResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /external-calendars/{id} [get]
func (h *CalendarHandler) GetExternalCalendar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid external calendar ID"})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	external, err := h.service.GetExternalCalendarByID(uint(id), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, external.ToResponse())
}

// GetAllExternalCalendars retrieves all external calendars with pagination
// @Summary Get all external calendars
// @Description Retrieve all external calendars for the authenticated user
// @Tags external-calendars
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /external-calendars [get]
func (h *CalendarHandler) GetAllExternalCalendars(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	externals, total, err := h.service.GetAllExternalCalendars(page, limit, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []entities.ExternalCalendarResponse
	for _, external := range externals {
		responses = append(responses, external.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"external_calendars": responses,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// UpdateExternalCalendar updates an existing external calendar
// @Summary Update external calendar
// @Description Update an existing external calendar
// @Tags external-calendars
// @Accept json
// @Produce json
// @Param id path int true "External Calendar ID"
// @Param external body entities.UpdateExternalCalendarRequest true "Updated external calendar data"
// @Success 200 {object} entities.ExternalCalendarResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /external-calendars/{id} [put]
func (h *CalendarHandler) UpdateExternalCalendar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid external calendar ID"})
		return
	}

	var req entities.UpdateExternalCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	external, err := h.service.UpdateExternalCalendar(uint(id), tenantID, userID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, external.ToResponse())
}

// DeleteExternalCalendar deletes an external calendar
// @Summary Delete external calendar
// @Description Delete an external calendar by ID
// @Tags external-calendars
// @Produce json
// @Param id path int true "External Calendar ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /external-calendars/{id} [delete]
func (h *CalendarHandler) DeleteExternalCalendar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid external calendar ID"})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	err = h.service.DeleteExternalCalendar(uint(id), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "External calendar deleted successfully"})
}

// Specialized Handlers

// GetCalendarWeekView retrieves calendar entries for a specific week
// @Summary Get calendar week view
// @Description Retrieve calendar entries for a specific week
// @Tags calendar-views
// @Produce json
// @Param date query string true "Date in YYYY-MM-DD format" example:"2025-01-15"
// @Success 200 {array} entities.CalendarEntryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar/week [get]
func (h *CalendarHandler) GetCalendarWeekView(c *gin.Context) {
	var req entities.WeekViewRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}

	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	entries, err := h.service.GetCalendarWeekView(date, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []entities.CalendarEntryResponse
	for _, entry := range entries {
		responses = append(responses, entry.ToResponse())
	}

	c.JSON(http.StatusOK, responses)
}

// GetCalendarYearView retrieves calendar entries for a specific year
// @Summary Get calendar year view
// @Description Retrieve calendar entries for a specific year
// @Tags calendar-views
// @Produce json
// @Param year query int true "Year" example:2025
// @Success 200 {array} entities.CalendarEntryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar/year [get]
func (h *CalendarHandler) GetCalendarYearView(c *gin.Context) {
	var req entities.YearViewRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	entries, err := h.service.GetCalendarYearView(req.Year, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []entities.CalendarEntryResponse
	for _, entry := range entries {
		responses = append(responses, entry.ToResponse())
	}

	c.JSON(http.StatusOK, responses)
}

// GetFreeSlots finds available time slots
// @Summary Get free time slots
// @Description Find available time slots based on duration, interval, and maximum number
// @Tags calendar-availability
// @Produce json
// @Param duration query int true "Duration in minutes" example:60
// @Param interval query int true "Interval between slots in minutes" example:30
// @Param number_max query int true "Maximum number of slots to return" example:10
// @Success 200 {array} entities.FreeSlot
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar/free-slots [get]
func (h *CalendarHandler) GetFreeSlots(c *gin.Context) {
	var req entities.FreeSlotRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	slots, err := h.service.GetFreeSlots(req, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, slots)
}

// ImportHolidays imports holidays into a specific calendar using unburdy format
// @Summary Import holidays into calendar
// @Description Import school holidays and public holidays into a specific calendar from unburdy format data
// @Tags calendar-utilities
// @Accept json
// @Produce json
// @Param id path int true "Calendar ID"
// @Param holidays body entities.ImportHolidaysRequest true "Import holidays request with state, year range, and holidays data"
// @Success 200 {object} entities.HolidayImportResult
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /calendar/{id}/import_holidays [post]
// @Security BearerAuth
func (h *CalendarHandler) ImportHolidays(c *gin.Context) {
	// Get calendar ID from path parameter
	idStr := c.Param("id")
	calendarID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calendar ID"})
		return
	}

	var req entities.ImportHolidaysRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate year range
	if req.YearTo < req.YearFrom {
		c.JSON(http.StatusBadRequest, gin.H{"error": "year_to must be greater than or equal to year_from"})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get tenant ID: " + err.Error()})
		return
	}
	userID, err := baseAPI.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get user ID: " + err.Error()})
		return
	}

	result, err := h.service.ImportHolidaysToCalendar(uint(calendarID), req, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
