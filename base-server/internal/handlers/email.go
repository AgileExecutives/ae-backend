package handlers

import (
	"net/http"

	"github.com/ae-base-server/internal/models"
	"github.com/ae-base-server/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EmailHandler struct {
	db *gorm.DB
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(db *gorm.DB) *EmailHandler {
	return &EmailHandler{db: db}
}

// GetEmails retrieves all emails with pagination
// DISABLED-SWAGGER: @Summary Get all emails
// DISABLED-SWAGGER: @Description Get a paginated list of all emails
// DISABLED-SWAGGER: @Tags emails
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Security BearerAuth
// DISABLED-SWAGGER: @Param page query int false "Page number" default(1)
// DISABLED-SWAGGER: @Param limit query int false "Items per page" default(10)
// DISABLED-SWAGGER: @Param status query string false "Filter by email status"
// DISABLED-SWAGGER: @Success 200 {object} models.APIResponse{data=models.ListResponse}
// DISABLED-SWAGGER: @Failure 500 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Router /emails [get]
func (h *EmailHandler) GetEmails(c *gin.Context) {
	page, limit := utils.GetPaginationParams(c)
	offset := utils.GetOffset(page, limit)

	var emails []models.Email
	var total int64

	query := h.db.Model(&models.Email{})

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to count emails", err.Error()))
		return
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&emails).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve emails", err.Error()))
		return
	}

	// Convert to response format
	var responses []models.EmailResponse
	for _, email := range emails {
		responses = append(responses, email.ToResponse())
	}

	response := models.ListResponse{
		Data: responses,
		Pagination: models.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      int(total),
			TotalPages: utils.CalculateTotalPages(int(total), limit),
		},
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Emails retrieved successfully", response))
}

// GetEmail retrieves a specific email by ID
// DISABLED-SWAGGER: @Summary Get email by ID
// DISABLED-SWAGGER: @Description Get a specific email by its ID
// DISABLED-SWAGGER: @Tags emails
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Security BearerAuth
// DISABLED-SWAGGER: @Param id path int true "Email ID"
// DISABLED-SWAGGER: @Success 200 {object} models.APIResponse{data=models.EmailResponse}
// DISABLED-SWAGGER: @Failure 400 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Failure 404 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Router /emails/{id} [get]
func (h *EmailHandler) GetEmail(c *gin.Context) {
	id, err := utils.ValidateID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid email ID", err.Error()))
		return
	}

	var email models.Email
	if err := h.db.First(&email, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Email not found", "Email with specified ID does not exist"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve email", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Email retrieved successfully", email.ToResponse()))
}

// SendEmail creates and queues an email for sending
// DISABLED-SWAGGER: @Summary Send an email
// DISABLED-SWAGGER: @Description Create and queue an email for sending
// DISABLED-SWAGGER: @Tags emails
// DISABLED-SWAGGER: @Accept json
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Security BearerAuth
// DISABLED-SWAGGER: @Param request body models.EmailSendRequest true "Email send data"
// DISABLED-SWAGGER: @Success 201 {object} models.APIResponse{data=models.EmailResponse}
// DISABLED-SWAGGER: @Failure 400 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Router /emails/send [post]
func (h *EmailHandler) SendEmail(c *gin.Context) {
	var req models.EmailSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	email := models.Email{
		To:       req.To,
		From:     req.From,
		Subject:  req.Subject,
		Body:     req.Body,
		HTMLBody: req.HTMLBody,
		Status:   "pending",
	}

	if err := h.db.Create(&email).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to create email", err.Error()))
		return
	}

	// TODO: Here you would typically queue the email for actual sending
	// For now, we'll just mark it as sent immediately
	// In a real implementation, you'd use a background job queue

	c.JSON(http.StatusCreated, models.SuccessResponse("Email queued for sending", email.ToResponse()))
}

// GetEmailStats retrieves email statistics
// DISABLED-SWAGGER: @Summary Get email statistics
// DISABLED-SWAGGER: @Description Get email statistics including counts by status
// DISABLED-SWAGGER: @Tags emails
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Security BearerAuth
// DISABLED-SWAGGER: @Success 200 {object} models.APIResponse{data=object}
// DISABLED-SWAGGER: @Failure 500 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Router /emails/stats [get]
func (h *EmailHandler) GetEmailStats(c *gin.Context) {
	type EmailStats struct {
		Total     int64 `json:"total"`
		Pending   int64 `json:"pending"`
		Sent      int64 `json:"sent"`
		Delivered int64 `json:"delivered"`
		Failed    int64 `json:"failed"`
	}

	var stats EmailStats

	// Count total emails
	h.db.Model(&models.Email{}).Count(&stats.Total)

	// Count by status
	h.db.Model(&models.Email{}).Where("status = ?", "pending").Count(&stats.Pending)
	h.db.Model(&models.Email{}).Where("status = ?", "sent").Count(&stats.Sent)
	h.db.Model(&models.Email{}).Where("status = ?", "delivered").Count(&stats.Delivered)
	h.db.Model(&models.Email{}).Where("status = ?", "failed").Count(&stats.Failed)

	c.JSON(http.StatusOK, models.SuccessResponse("Email statistics retrieved successfully", stats))
}
