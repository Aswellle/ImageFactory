// Package admin provides admin-only HTTP handlers. Every handler in this
// package expects the adminAuth middleware to have run and set if.user_id /
// if.role on the context.
package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// DashboardHandler serves the admin dashboard endpoints.
type DashboardHandler struct {
	admin service.AdminService
}

// NewDashboardHandler builds a DashboardHandler.
func NewDashboardHandler(admin service.AdminService) *DashboardHandler {
	return &DashboardHandler{admin: admin}
}

// Get returns the full dashboard payload (stats + recent activity).
// GET /v1/admin/dashboard
func (h *DashboardHandler) Get(c *gin.Context) {
	payload, err := h.admin.Dashboard(c.Request.Context())
	if err != nil {
		response.Error(c, 500, "INTERNAL_ERROR", "failed to load dashboard", rid(c))
		return
	}
	response.OK(c, payload)
}

// rid returns the request id from the context for error correlation.
func rid(c *gin.Context) string {
	v, _ := c.Get("if.request_id")
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
