package handlers

import (
	"net/http"
	"os"
	"strconv"
	"time"

	baseAPI "github.com/ae-base-server/api"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/booking-module/entities"
	"github.com/unburdy/booking-module/services"
)

type BookingHandler struct {
	service        *services.BookingService
	bookingLinkSvc *services.BookingLinkService
	freeSlotsSvc   *services.FreeSlotsService
}

func NewBookingHandler(service *services.BookingService, bookingLinkSvc *services.BookingLinkService, freeSlotsSvc *services.FreeSlotsService) *BookingHandler {
	return &BookingHandler{
		service:        service,
		bookingLinkSvc: bookingLinkSvc,
		freeSlotsSvc:   freeSlotsSvc,
	}
}

// CreateConfiguration godoc
// @Summary Create a new booking configuration
// @Description Create a new booking configuration/template for a user's calendar
// @Tags booking-templates
// @Accept json
// @Produce json
// @Param configuration body entities.CreateBookingTemplateRequest true "Booking configuration data"
// @Success 201 {object} baseAPI.APIResponse{data=entities.BookingTemplateResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates [post]
// @ID createBookingTemplate
func (h *BookingHandler) CreateConfiguration(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("", "Tenant information required"))
		return
	}

	var req entities.CreateBookingTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request data", err.Error()))
		return
	}

	config, err := h.service.CreateConfiguration(req, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, baseAPI.SuccessResponse("Booking configuration created successfully", config.ToResponse()))
}

// GetConfiguration godoc
// @Summary Get a booking configuration by ID
// @Description Retrieve a specific booking configuration by ID
// @Tags booking-templates
// @Produce json
// @Param id path int true "Configuration ID"
// @Success 200 {object} baseAPI.APIResponse{data=entities.BookingTemplateResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 404 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates/{id} [get]
// @ID getBookingTemplate
func (h *BookingHandler) GetConfiguration(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("", "Tenant information required"))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("", "Invalid configuration ID"))
		return
	}

	config, err := h.service.GetConfiguration(uint(id), tenantID)
	if err != nil {
		if err.Error() == "booking configuration not found" {
			c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("", err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("", err.Error()))
		return
	}

	c.JSON(http.StatusOK, baseAPI.SuccessResponse("", config.ToResponse()))
}

// GetAllConfigurations godoc
// @Summary Get all booking configurations
// @Description Retrieve all booking configurations for the tenant
// @Tags booking-templates
// @Produce json
// @Success 200 {object} baseAPI.APIResponse{data=[]entities.BookingTemplateResponse}
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates [get]
// @ID listBookingTemplates
func (h *BookingHandler) GetAllConfigurations(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("", "Tenant information required"))
		return
	}

	// Get all configurations without pagination (limit = -1)
	configs, _, err := h.service.GetAllConfigurations(tenantID, 1, -1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("", err.Error()))
		return
	}

	responses := make([]entities.BookingTemplateResponse, len(configs))
	for i, config := range configs {
		responses[i] = config.ToResponse()
	}

	c.JSON(http.StatusOK, baseAPI.SuccessResponse("", responses))
}

// GetConfigurationsByUser godoc
// @Summary Get booking configurations by user ID
// @Description Retrieve all booking configurations for a specific user
// @Tags booking-templates
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {object} baseAPI.APIResponse{data=[]entities.BookingTemplateResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates/by-user [get]
// @ID listBookingTemplatesByUser
func (h *BookingHandler) GetConfigurationsByUser(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("", "Tenant information required"))
		return
	}

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Missing parameter", "user_id query parameter is required"))
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid parameter", "user_id must be a valid positive integer"))
		return
	}

	configs, err := h.service.GetConfigurationsByUser(uint(userID), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("", err.Error()))
		return
	}

	responses := make([]entities.BookingTemplateResponse, len(configs))
	for i, config := range configs {
		responses[i] = config.ToResponse()
	}

	c.JSON(http.StatusOK, baseAPI.SuccessResponse("", responses))
}

// GetConfigurationsByCalendar godoc
// @Summary Get booking configurations by calendar ID
// @Description Retrieve all booking configurations for a specific calendar
// @Tags booking-templates
// @Produce json
// @Param calendar_id query int true "Calendar ID"
// @Success 200 {object} baseAPI.APIResponse{data=[]entities.BookingTemplateResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates/by-calendar [get]
// @ID listBookingTemplatesByCalendar
func (h *BookingHandler) GetConfigurationsByCalendar(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("", "Tenant information required"))
		return
	}

	calendarIDStr := c.Query("calendar_id")
	if calendarIDStr == "" {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Missing parameter", "calendar_id query parameter is required"))
		return
	}

	calendarID, err := strconv.ParseUint(calendarIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid parameter", "calendar_id must be a valid positive integer"))
		return
	}

	configs, err := h.service.GetConfigurationsByCalendar(uint(calendarID), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("", err.Error()))
		return
	}

	responses := make([]entities.BookingTemplateResponse, len(configs))
	for i, config := range configs {
		responses[i] = config.ToResponse()
	}

	c.JSON(http.StatusOK, baseAPI.SuccessResponse("", responses))
}

// UpdateConfiguration godoc
// @Summary Update a booking configuration
// @Description Update an existing booking configuration
// @Tags booking-templates
// @Accept json
// @Produce json
// @Param id path int true "Configuration ID"
// @Param configuration body entities.UpdateBookingTemplateRequest true "Updated configuration data"
// @Success 200 {object} baseAPI.APIResponse{data=entities.BookingTemplateResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 404 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates/{id} [put]
// @ID updateBookingTemplate
func (h *BookingHandler) UpdateConfiguration(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("", "Tenant information required"))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("", "Invalid configuration ID"))
		return
	}

	var req entities.UpdateBookingTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("", err.Error()))
		return
	}

	config, err := h.service.UpdateConfiguration(uint(id), tenantID, req)
	if err != nil {
		if err.Error() == "booking configuration not found" {
			c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("", err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("", err.Error()))
		return
	}

	c.JSON(http.StatusOK, baseAPI.SuccessResponse("Booking configuration updated successfully", config.ToResponse()))
}

// DeleteConfiguration godoc
// @Summary Delete a booking configuration
// @Description Soft delete a booking configuration by ID
// @Tags booking-templates
// @Produce json
// @Param id path int true "Configuration ID"
// @Success 200 {object} baseAPI.APIResponse
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 404 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates/{id} [delete]
// @ID deleteBookingTemplate
func (h *BookingHandler) DeleteConfiguration(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("", "Tenant information required"))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("", "Invalid configuration ID"))
		return
	}

	err = h.service.DeleteConfiguration(uint(id), tenantID)
	if err != nil {
		if err.Error() == "booking configuration not found" {
			c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("", err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("", err.Error()))
		return
	}

	c.JSON(http.StatusOK, baseAPI.SuccessMessageResponse("Booking configuration deleted successfully"))
}

// CreateBookingLink godoc
// @Summary Create a booking link
// @Description Generate a booking link token for a client to book appointments
// @Tags booking-templates
// @Accept json
// @Produce json
// @Param link body entities.CreateBookingLinkRequest true "Booking link data"
// @Success 201 {object} baseAPI.APIResponse{data=entities.BookingLinkResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 404 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/link [post]
// @ID createBookingLink
func (h *BookingHandler) CreateBookingLink(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("", "Tenant information required"))
		return
	}

	var req entities.CreateBookingLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request data", err.Error()))
		return
	}

	token, err := h.bookingLinkSvc.GenerateBookingLink(req.TemplateID, req.ClientID, tenantID, req.Purpose)
	if err != nil {
		if err.Error() == "booking template not found" {
			c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("", err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("", err.Error()))
		return
	}

	// Construct the booking URL with frontend URL from environment
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3005" // Default fallback
	}
	bookingURL := frontendURL + "/booking/" + token

	// Calculate expiry for one-time links
	var expiresAt *time.Time
	if req.Purpose == entities.OneTimeBookingLink {
		expiry := time.Now().Add(24 * time.Hour)
		expiresAt = &expiry
	}

	response := entities.BookingLinkResponse{
		Token:     token,
		URL:       bookingURL,
		Purpose:   req.Purpose,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	c.JSON(http.StatusCreated, baseAPI.SuccessResponse("Booking link created successfully", response))
}

// GetFreeSlots godoc
// @Summary Get available time slots for booking
// @Description Retrieve available time slots based on a booking link token. Token is validated and must not be blacklisted.
// @Tags booking-slots
// @Produce json
// @Param token path string true "Booking link token"
// @Param start query string false "Start date for slot search (YYYY-MM-DD)" example="2025-11-01"
// @Param end query string false "End date for slot search (YYYY-MM-DD)" example="2025-11-30"
// @Success 200 {object} baseAPI.APIResponse{data=entities.FreeSlotsResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 404 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Router /booking/freeslots/{token} [get]
// @ID getBookingFreeSlots
func (h *BookingHandler) GetFreeSlots(c *gin.Context) {
	// Get booking claims from middleware (already validated)
	claimsInterface, exists := c.Get("booking_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Invalid booking token"))
		return
	}

	claims, ok := claimsInterface.(*entities.BookingLinkClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("", "Failed to parse booking claims"))
		return
	}

	// Get the booking template
	template, err := h.service.GetConfiguration(claims.TemplateID, claims.TenantID)
	if err != nil {
		if err.Error() == "booking configuration not found" {
			c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("", "Booking template not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("", err.Error()))
		return
	}

	// Parse date range from query params (default to current month)
	now := time.Now()
	var startDate, endDate time.Time

	startParam := c.Query("start")
	if startParam != "" {
		startDate, err = time.Parse("2006-01-02", startParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid parameter", "start date must be in YYYY-MM-DD format"))
			return
		}
	} else {
		// Default to start of current month
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	}

	endParam := c.Query("end")
	if endParam != "" {
		endDate, err = time.Parse("2006-01-02", endParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid parameter", "end date must be in YYYY-MM-DD format"))
			return
		}
	} else {
		// Default to end of current month
		endDate = startDate.AddDate(0, 1, -1)
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.UTC)
	}

	// Calculate free slots
	req := services.FreeSlotsRequest{
		TemplateID: claims.TemplateID,
		TenantID:   claims.TenantID,
		CalendarID: claims.CalendarID,
		StartDate:  startDate,
		EndDate:    endDate,
		Timezone:   template.Timezone,
	}

	freeSlots, err := h.freeSlotsSvc.CalculateFreeSlots(req, template)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("", err.Error()))
		return
	}

	c.JSON(http.StatusOK, baseAPI.SuccessResponse("Free slots retrieved successfully", freeSlots))
}
