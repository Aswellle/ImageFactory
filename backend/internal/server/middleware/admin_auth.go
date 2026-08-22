// Package middleware provides HTTP middleware for authentication, authorization, and request processing.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// AdminAuth wraps JWT auth and enforces the admin role. It also accepts an
// admin API key via the `x-api-key` header as an alternative auth path for
// service-to-service calls to the admin panel.
//
// Resolution order:
//  1. x-api-key header — if it matches the configured admin API key, sets
//     user context as a synthetic admin.
//  2. Bearer JWT — validated and checked for the admin role.
//
// On success it sets the standard context keys (if.user_id, if.role) and
// calls the next handler.
type AdminAuth struct {
	auth          *Auth
	adminAPIKey   string
}

// NewAdminAuth builds an AdminAuth middleware.
func NewAdminAuth(auth *Auth, adminAPIKey string) *AdminAuth {
	return &AdminAuth{auth: auth, adminAPIKey: adminAPIKey}
}

// Require enforces admin access. Aborts with 401/403 on failure.
func (a *AdminAuth) Require() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Path 1: admin API key.
		if a.adminAPIKey != "" {
			if key := c.GetHeader("x-api-key"); key != "" && key == a.adminAPIKey {
				c.Set("if.user_id", int64(0))
				c.Set("if.role", "admin")
				c.Set("if.auth_type", "admin_api_key")
				c.Next()
				return
			}
		}
		// Path 2: bearer JWT with admin role.
		role, _ := c.Get("if.role")
		if a.auth != nil {
			// RequireAdmin already validates + aborts on failure.
			a.auth.RequireAdmin()(c)
			if c.IsAborted() {
				return
			}
			c.Set("if.auth_type", "jwt")
			return
		}
		_ = role
		response.Error(c, 403, string(errors.ErrorCodeForbidden), "admin access required", rid(c))
		c.Abort()
	}
}

// HasAdmin reports whether the request carries an admin-role identity. Useful
// for routes that are shared between user and admin contexts.
func HasAdmin(role string) bool {
	return service.HasAdmin(role)
}
