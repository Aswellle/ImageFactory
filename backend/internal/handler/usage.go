// Package handler exposes usage dashboard endpoints over HTTP.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// UsageHandler serves usage statistics for the authenticated user.
type UsageHandler struct {
	usage *service.UsageService
}

// NewUsageHandler builds a UsageHandler.
func NewUsageHandler(usage *service.UsageService) *UsageHandler {
	return &UsageHandler{usage: usage}
}

// Get handles GET /v1/usage — returns the current period summary with daily breakdown.
func (h *UsageHandler) Get(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}

	summary, err := h.usage.GetUserUsage(c.Request.Context(), uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, summary)
}

// History handles GET /v1/usage/history?days=N — returns a per-day breakdown.
func (h *UsageHandler) History(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	history, err := h.usage.GetUsageHistory(c.Request.Context(), uid, days)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, history)
}
