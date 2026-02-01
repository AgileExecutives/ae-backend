package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/ae-base-server/internal/models"
	"github.com/gin-gonic/gin"
)

// GetLatestEmails retrieves all mock emails (only available when MOCK_EMAIL=true)
// @Summary Get latest mock emails
// @ID getLatestEmails
// @Description Get all emails sent in mock mode for testing purposes
// @Tags emails
// @Produce json
// @Success 200 {object} models.APIResponse{data=[]services.MockEmailRecord}
// @Failure 500 {object} models.ErrorResponse
// @Failure 503 {object} models.ErrorResponse
// @Router /emails/latest-emails [get]
func (h *EmailHandler) GetLatestEmails(c *gin.Context) {
	// Check if mock email is enabled
	mockEmail := os.Getenv("MOCK_EMAIL")
	if strings.ToLower(mockEmail) != "true" {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponseFunc(
			"Mock email not enabled",
			"This endpoint is only available when MOCK_EMAIL=true",
		))
		return
	}

	// Get emails from service
	emails, err := h.emailService.GetLatestEmails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc(
			"Failed to retrieve mock emails",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Mock emails retrieved successfully", emails))
}
