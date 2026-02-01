package handlers

import (
	"net/http"

	"github.com/ae-base-server/internal/models"
	"github.com/ae-base-server/pkg/auth"
	"github.com/gin-gonic/gin"
)

// CheckResetToken validates a password reset token without using it
// @Summary Check password reset token
// @ID checkResetToken
// @Description Validate a password reset token to check if it's valid and not expired
// @Tags authentication
// @Produce json
// @Param token path string true "Reset token"
// @Success 200 {object} models.APIResponse{data=object{email=string,valid=bool}}
// @Failure 400 {object} models.ErrorResponse
// @Router /auth/check-reset-token/{token} [get]
func (h *AuthHandlers) CheckResetToken(c *gin.Context) {
	token := c.Param("token")

	// Validate reset token
	email, err := auth.ValidateResetToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid or expired reset token", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Token is valid", gin.H{
		"valid": true,
		"email": email,
	}))
}
