// Package routes provides HTTP route registration for the admin panel.
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/handler/admin"
	"github.com/imageforge/imageforge/internal/server/middleware"
)

// AdminHandlers groups the admin sub-handlers for route registration.
type AdminHandlers struct {
	Dashboard *admin.DashboardHandler
	User      *admin.UserHandler
	Job       *admin.JobHandler
	APIKey    *admin.APIKeyHandler
}

// RegisterAdminRoutes mounts all admin routes under /v1/admin with the
// adminAuth middleware applied to the whole group.
func RegisterAdminRoutes(
	v1 *gin.RouterGroup,
	h *AdminHandlers,
	adminAuth *middleware.AdminAuth,
) {
	adminGroup := v1.Group("/admin")
	adminGroup.Use(adminAuth.Require())
	{
		registerDashboardRoutes(adminGroup, h)
		registerUserRoutes(adminGroup, h)
		registerJobRoutes(adminGroup, h)
		registerAPIKeyRoutes(adminGroup, h)
	}
}

func registerDashboardRoutes(admin *gin.RouterGroup, h *AdminHandlers) {
	d := admin.Group("/dashboard")
	{
		d.GET("", h.Dashboard.Get)
	}
}

func registerUserRoutes(admin *gin.RouterGroup, h *AdminHandlers) {
	u := admin.Group("/users")
	{
		u.GET("", h.User.List)
		u.POST("/:id/status", h.User.SetStatus)
	}
}

func registerJobRoutes(admin *gin.RouterGroup, h *AdminHandlers) {
	j := admin.Group("/jobs")
	{
		j.GET("", h.Job.List)
		j.POST("/:id/retry", h.Job.Retry)
	}
}

func registerAPIKeyRoutes(admin *gin.RouterGroup, h *AdminHandlers) {
	k := admin.Group("/api-keys")
	{
		k.GET("", h.APIKey.List)
		k.POST("/:id/revoke", h.APIKey.Revoke)
	}
}
