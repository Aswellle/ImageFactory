package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// PasswordResetHandler handles send-reset-code and reset-password.
type PasswordResetHandler struct {
	reset *service.PasswordResetService
}

// NewPasswordResetHandler builds a PasswordResetHandler.
func NewPasswordResetHandler(reset *service.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{reset: reset}
}

// sendResetCodeRequest is the request body for sending a reset code.
type sendResetCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// resetPasswordRequest is the request body for resetting a password.
type resetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// SendResetCode handles POST /v1/auth/send-reset-code.
// For privacy, it always returns 200 regardless of whether the email is registered.
func (h *PasswordResetHandler) SendResetCode(c *gin.Context) {
	var req sendResetCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, string(errors.ErrorCodeInvalidRequest), "invalid request", requestID(c))
		return
	}

	// Ignore errors to avoid email enumeration — always return success.
	_ = h.reset.SendResetCode(c.Request.Context(), service.SendResetCodeInput{
		Email: req.Email,
	})

	response.OK(c, gin.H{"message": "if the email is registered, a verification code has been sent"})
}

// ResetPassword handles POST /v1/auth/reset-password.
func (h *PasswordResetHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, string(errors.ErrorCodeInvalidRequest), "invalid request", requestID(c))
		return
	}

	err := h.reset.ResetPassword(c.Request.Context(), service.ResetPasswordInput{
		Email:       req.Email,
		Code:        req.Code,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "password has been reset"})
}
