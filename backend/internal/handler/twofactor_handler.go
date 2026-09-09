// Copyright 2024 ImageForge
// 2FA 管理处理器 — TOTP 设置、验证、禁用。

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/ent"
	appErrors "github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/repository"
	"github.com/imageforge/imageforge/internal/service"
)

// TwoFactorHandler 处理 2FA 相关端点。
type TwoFactorHandler struct {
	totp      *service.TOTPService
	users     *repository.UserRepository
}

// NewTwoFactorHandler 创建 2FA handler。
func NewTwoFactorHandler(totp *service.TOTPService, users *repository.UserRepository) *TwoFactorHandler {
	return &TwoFactorHandler{totp: totp, users: users}
}

// userWithTOTP 包含 2FA 相关字段的用户信息。
type userWithTOTP struct {
	*ent.User
}

// SetupResponse 返回设置信息。
type SetupResponse struct {
	Secret      string   `json:"secret"`
	QRCode      string   `json:"qr_code"`
	BackupCodes []string `json:"backup_codes"`
}

// Setup 处理 POST /v1/auth/2fa/setup。
func (h *TwoFactorHandler) Setup(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(appErrors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}

	user, err := h.users.GetByID(c.Request.Context(), uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, string(appErrors.ErrorCodeInternal), "failed to get user", requestID(c))
		return
	}

	if user.TotpEnabled {
		response.Error(c, http.StatusConflict, string(appErrors.ErrorCodeConflict), "2FA already enabled, disable first", requestID(c))
		return
	}

	result, err := h.totp.Setup(uid, user.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, string(appErrors.ErrorCodeInternal), "failed to setup 2FA", requestID(c))
		return
	}

	response.OK(c, SetupResponse{
		Secret:      result.Secret,
		QRCode:      result.QRCode,
		BackupCodes: result.BackupCodes,
	})
}

// EnableRequest 启用 2FA 请求。
type EnableRequest struct {
	Code        string `json:"code" binding:"required,len=6"`
	TemporarySecret string `json:"secret"` // 设置时返回的 secret
}

// Enable 处理 POST /v1/auth/2fa/enable。
func (h *TwoFactorHandler) Enable(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(appErrors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}

	var req EnableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, string(appErrors.ErrorCodeInvalidRequest), "6-digit code required", requestID(c))
		return
	}

	user, err := h.users.GetByID(c.Request.Context(), uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, string(appErrors.ErrorCodeInternal), "failed to get user", requestID(c))
		return
	}

	if user.TotpEnabled {
		response.Error(c, http.StatusConflict, string(appErrors.ErrorCodeConflict), "2FA already enabled", requestID(c))
		return
	}

	// 验证 TOTP 码
	if !h.totp.Validate(req.TemporarySecret, req.Code) {
		response.Error(c, http.StatusBadRequest, string(appErrors.ErrorCodeInvalidRequest), "invalid verification code", requestID(c))
		return
	}

	// 启用 2FA
	if err := h.users.EnableTOTP(c.Request.Context(), uid, req.TemporarySecret); err != nil {
		response.Error(c, http.StatusInternalServerError, string(appErrors.ErrorCodeInternal), "failed to enable 2FA", requestID(c))
		return
	}

	response.OK(c, gin.H{"message": "2FA enabled successfully"})
}

// Disable 处理 POST /v1/auth/2fa/disable。
func (h *TwoFactorHandler) Disable(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(appErrors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}

	var req EnableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, string(appErrors.ErrorCodeInvalidRequest), "6-digit code required", requestID(c))
		return
	}

	user, err := h.users.GetByID(c.Request.Context(), uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, string(appErrors.ErrorCodeInternal), "failed to get user", requestID(c))
		return
	}

	if !user.TotpEnabled {
		response.Error(c, http.StatusConflict, string(appErrors.ErrorCodeConflict), "2FA not enabled", requestID(c))
		return
	}

	if !h.totp.Validate(user.TotpSecret, req.Code) {
		response.Error(c, http.StatusBadRequest, string(appErrors.ErrorCodeInvalidRequest), "invalid verification code", requestID(c))
		return
	}

	if err := h.users.DisableTOTP(c.Request.Context(), uid); err != nil {
		response.Error(c, http.StatusInternalServerError, string(appErrors.ErrorCodeInternal), "failed to disable 2FA", requestID(c))
		return
	}

	response.OK(c, gin.H{"message": "2FA disabled successfully"})
}

// Status 处理 GET /v1/auth/2fa/status。
func (h *TwoFactorHandler) Status(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(appErrors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}

	user, err := h.users.GetByID(c.Request.Context(), uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, string(appErrors.ErrorCodeInternal), "failed to get user", requestID(c))
		return
	}

	response.OK(c, gin.H{
		"totp_enabled": user.TotpEnabled,
	})
}
