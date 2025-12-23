package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	baseAPI "github.com/ae-base-server/api"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"github.com/unburdy/unburdy-server-api/modules/client_management/services"
)

// SessionHandler handles session-related HTTP requests
type SessionHandler struct {
	sessionService *services.SessionService
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionService *services.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// CreateSession handles creating a new session
// @Summary Create a new session
// @Description Create a new therapy/appointment session linked to a calendar entry
// @Tags sessions
// @ID createSession
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param session body entities.CreateSessionRequest true "Session information"
// @Success 201 {object} baseAPI.APIResponse{data=entities.SessionResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Router /sessions [post]
func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req entities.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	session, err := h.sessionService.CreateSession(req, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, baseAPI.SuccessResponse("Session created successfully", session.ToResponse()))
}

// GetSession handles retrieving a specific session
// @Summary Get session by ID
// @Description Retrieve a session by its ID
// @Tags sessions
// @ID getSessionById
// @Produce json
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} baseAPI.APIResponse{data=entities.SessionResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 404 {object} baseAPI.APIResponse
// @Router /sessions/{id} [get]
func (h *SessionHandler) GetSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", "Invalid session ID"))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	session, err := h.sessionService.GetSessionByID(uint(id), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("Not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, baseAPI.SuccessResponse("Session retrieved successfully", session.ToResponse()))
}

// GetSessionByCalendarEntry handles retrieving a session by calendar entry ID
// @Summary Get session by calendar entry ID
// @Description Retrieve a session associated with a specific calendar entry
// @Tags sessions
// @ID getSessionByCalendarEntry
// @Produce json
// @Security BearerAuth
// @Param id path int true "Calendar Entry ID"
// @Success 200 {object} baseAPI.APIResponse{data=entities.SessionResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 404 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Router /sessions/by_entry/{id} [get]
func (h *SessionHandler) GetSessionByCalendarEntry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", "Invalid calendar entry ID"))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	session, err := h.sessionService.GetSessionByCalendarEntryID(uint(id), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("Not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, baseAPI.SuccessResponse("Session retrieved successfully", session.ToResponse()))
}

// GetAllSessions handles retrieving all sessions with pagination
// @Summary Get all sessions
// @Description Retrieve all sessions with pagination
// @Tags sessions
// @ID getAllSessions
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} baseAPI.ListResponse{data=[]entities.SessionResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Router /sessions [get]
func (h *SessionHandler) GetAllSessions(c *gin.Context) {
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
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	sessions, total, err := h.sessionService.GetAllSessions(page, limit, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	responses := make([]entities.SessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = session.ToResponse()
	}

	c.JSON(http.StatusOK, baseAPI.SuccessListResponse(responses, page, limit, total))
}

// GetSessionsByClient handles retrieving all sessions for a specific client
// @Summary Get sessions by client ID
// @Description Retrieve all sessions for a specific client with pagination
// @Tags sessions
// @ID getSessionsByClient
// @Produce json
// @Security BearerAuth
// @Param id path int true "Client ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} baseAPI.ListResponse{data=[]entities.SessionResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Router /clients/{id}/sessions [get]
func (h *SessionHandler) GetSessionsByClient(c *gin.Context) {
	clientIDStr := c.Param("id")
	clientID, err := strconv.ParseUint(clientIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", "Invalid client ID"))
		return
	}

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
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	sessions, total, err := h.sessionService.GetSessionsByClientID(uint(clientID), tenantID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	responses := make([]entities.SessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = session.ToResponse()
	}

	c.JSON(http.StatusOK, baseAPI.SuccessListResponse(responses, page, limit, total))
}

// UpdateSession handles updating an existing session
// @Summary Update session
// @Description Update an existing session
// @Tags sessions
// @ID updateSession
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Param session body entities.UpdateSessionRequest true "Session information"
// @Success 200 {object} baseAPI.APIResponse{data=entities.SessionResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 404 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Router /sessions/{id} [put]
func (h *SessionHandler) UpdateSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", "Invalid session ID"))
		return
	}

	var req entities.UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	session, err := h.sessionService.UpdateSession(uint(id), tenantID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, baseAPI.SuccessResponse("Session updated successfully", session.ToResponse()))
}

// DeleteSession handles deleting a session
// @Summary Delete session
// @Description Delete a session by ID
// @Tags sessions
// @ID deleteSession
// @Produce json
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} baseAPI.APIResponse
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 404 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Router /sessions/{id} [delete]
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", "Invalid session ID"))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	err = h.sessionService.DeleteSession(uint(id), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, baseAPI.SuccessResponse("Session deleted successfully", nil))
}

// GetDetailedSessionsUpcoming handles retrieving detailed sessions for the next 7 days
// @Summary Get detailed sessions for upcoming 7 days
// @Description Retrieve all sessions scheduled for 7 days starting from the specified date (or current date if not specified) with detailed client information including their previous and next sessions
// @Tags sessions
// @ID getDetailedSessionsUpcoming
// @Produce json
// @Security BearerAuth
// @Param date query string false "Start date (YYYY-MM-DD format, e.g., 2025-12-23). Defaults to current date if not provided."
// @Success 200 {object} baseAPI.APIResponse{data=[]entities.SessionDetailResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Router /sessions/detail [get]
func (h *SessionHandler) GetDetailedSessionsUpcoming(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, baseAPI.ErrorResponseFunc("Unauthorized", "Unable to get tenant ID: "+err.Error()))
		return
	}

	// Parse optional date parameter
	var startDate *time.Time
	if dateStr := c.Query("date"); dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", "Invalid date format. Use YYYY-MM-DD format (e.g., 2025-12-23)"))
			return
		}
		startDate = &parsedDate
	}

	detailedSessions, err := h.sessionService.GetDetailedSessionsUpcoming7Days(tenantID, startDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, baseAPI.SuccessResponse("Detailed sessions retrieved successfully", detailedSessions))
}

// BookSessions creates a calendar series or single entry with sessions for a client
// @Summary Book sessions for a client
// @Description Create a recurring calendar series (if interval_type is provided) or a single calendar entry and corresponding sessions for a client. For single entries, omit interval_type or set it to "none". For recurring series, provide interval_type (weekly/monthly-date/monthly-day/yearly), interval_value, and last_date.
// @Tags sessions
// @ID bookSessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param booking body entities.BookSessionsRequest true "Booking information"
// @Success 201 {object} baseAPI.APIResponse{data=entities.BookSessionsResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Router /sessions/book [post]
func (h *SessionHandler) BookSessions(c *gin.Context) {
	var req entities.BookSessionsRequest
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

	seriesID, sessions, err := h.sessionService.BookSessions(req, tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	responses := make([]entities.SessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = session.ToResponse()
	}

	response := entities.BookSessionsResponse{
		SeriesID: seriesID,
		Sessions: responses,
	}

	c.JSON(http.StatusCreated, baseAPI.SuccessResponse("Sessions booked successfully", response))
}

// BookSessionsWithToken creates sessions using a booking token
// @Summary Book sessions with token (public endpoint)
// @Description Create sessions for a client using a booking token. This endpoint does NOT require authentication - the token itself is the authorization. The token contains client_id, calendar_id, tenant_id, and user_id.
// @Tags sessions
// @ID bookSessionsWithToken
// @Accept json
// @Produce json
// @Param token path string true "Booking token (acts as authorization)"
// @Param booking body entities.BookSessionsWithTokenRequest true "Booking information (without client_id and calendar_id)"
// @Success 201 {object} baseAPI.APIResponse{data=entities.BookSessionsResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 404 {object} baseAPI.APIResponse "Invalid or expired token"
// @Failure 500 {object} baseAPI.APIResponse
// @Router /sessions/book/{token} [post]
func (h *SessionHandler) BookSessionsWithToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", "Token is required"))
		return
	}

	var req entities.BookSessionsWithTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, baseAPI.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	seriesID, sessions, err := h.sessionService.BookSessionsWithToken(token, req)
	if err != nil {
		// Log the full error for debugging
		fmt.Printf("Error booking sessions with token: %v\n", err)

		if err.Error() == "invalid or expired booking token" {
			c.JSON(http.StatusNotFound, baseAPI.ErrorResponseFunc("Invalid token", err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	responses := make([]entities.SessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = session.ToResponse()
	}

	response := entities.BookSessionsResponse{
		SeriesID: seriesID,
		Sessions: responses,
	}

	c.JSON(http.StatusCreated, baseAPI.SuccessResponse("Sessions booked successfully", response))
}
