// Copyright 2024 ImageForge
// Scheduling threshold management handler.

package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// SchedulingHandler handles scheduling configuration endpoints.
type SchedulingHandler struct {
	schedulingSvc *service.SchedulingService
}

// NewSchedulingHandler builds a SchedulingHandler.
func NewSchedulingHandler(schedulingSvc *service.SchedulingService) *SchedulingHandler {
	return &SchedulingHandler{schedulingSvc: schedulingSvc}
}

// GetThresholds returns the current per-platform scheduling thresholds.
// GET /v1/admin/scheduling/thresholds
func (h *SchedulingHandler) GetThresholds(c *gin.Context) {
	response.OK(c, gin.H{"thresholds": h.schedulingSvc.GetThresholds()})
}

// SetThresholdsRequest is the body for updating scheduling thresholds.
type SetThresholdsRequest struct {
	Thresholds map[string]int `json:"thresholds" binding:"required,min=1"`
}

// SetThresholds updates the per-platform scheduling thresholds.
// POST /v1/admin/scheduling/thresholds
func (h *SchedulingHandler) SetThresholds(c *gin.Context) {
	var req SetThresholdsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request: "+err.Error(), rid(c))
		return
	}
	for platform, pct := range req.Thresholds {
		if pct < 0 || pct > 100 {
			response.Error(c, http.StatusBadRequest, "INVALID_REQUEST",
				"threshold for "+platform+" must be 0-100", rid(c))
			return
		}
	}
	h.schedulingSvc.SetThresholds(req.Thresholds)
	response.OK(c, gin.H{"thresholds": h.schedulingSvc.GetThresholds()})
}
