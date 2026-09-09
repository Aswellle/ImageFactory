// Copyright 2024 ImageForge
// Token 刷新和吊销处理器。

package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	appErrors "github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// RefreshHandler 处理 Token 刷新和吊销。
type RefreshHandler struct {
	jwt         *service.JWTService
	refreshStore *service.RefreshTokenStore
}

// NewRefreshHandler 创建 RefreshHandler。
func NewRefreshHandler(jwt *service.JWTService, refreshStore *service.RefreshTokenStore) *RefreshHandler {
	return &RefreshHandler{jwt: jwt, refreshStore: refreshStore}
}

// RefreshRequest 刷新请求体。
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh 处理 POST /v1/auth/refresh。
// 验证 refresh token，返回新的 access token 和 refresh token。
func (h *RefreshHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, string(appErrors.ErrorCodeInvalidRequest), "refresh_token required", requestID(c))
		return
	}

	userID, role, tokenVersion, _, err := h.refreshStore.Validate(req.RefreshToken)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			response.Error(c, http.StatusInternalServerError, string(appErrors.ErrorCodeInternal), "service unavailable", requestID(c))
			return
		}
		response.Error(c, http.StatusUnauthorized, string(appErrors.ErrorCodeAuthExpired), "invalid or expired refresh token", requestID(c))
		return
	}

	// 生成新的 access token
	accessToken, err := h.jwt.Generate(userID, role, tokenVersion)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, string(appErrors.ErrorCodeInternal), "failed to issue token", requestID(c))
		return
	}

	// 轮换 refresh token
	newRefreshToken, _, err := h.refreshStore.Generate(userID, role, tokenVersion, 7*24*60*60*1000000000)
	if err == nil {
		_ = h.refreshStore.Revoke(req.RefreshToken)
	}

	result := gin.H{
		"token":      accessToken,
		"expires_in": 1440 * 60, // 24h in seconds (default)
	}
	if newRefreshToken != "" {
		result["refresh_token"] = newRefreshToken
	}
	response.OK(c, result)
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Logout 处理 POST /v1/auth/logout。
// 吊销 refresh token，使会话失效。
func (h *RefreshHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, string(appErrors.ErrorCodeInvalidRequest), "refresh_token required", requestID(c))
		return
	}

	if err := h.refreshStore.Revoke(req.RefreshToken); err != nil {
		response.Error(c, http.StatusInternalServerError, string(appErrors.ErrorCodeInternal), "failed to logout", requestID(c))
		return
	}

	response.NoContent(c)
}

// 确保 jwt 包被引用。
var _ = jwt.SigningMethodHS256
