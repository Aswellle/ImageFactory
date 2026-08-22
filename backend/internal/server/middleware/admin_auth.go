// Package middleware provides HTTP middleware for authentication, authorization, and request processing.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// AdminAuth wraps JWT auth and enforces the admin role.
// It relies exclusively on Bearer JWT tokens validated by the auth middleware.
// No static API key bypass is supported.
//
// On success it sets the standard context keys (if.user_id, if.role) and
// calls the next handler.
type AdminAuth struct {
	auth *Auth
}

// NewAdminAuth builds an AdminAuth middleware.
func NewAdminAuth(auth *Auth) *AdminAuth {
	return &AdminAuth{auth: auth}
}

// Require enforces admin access via JWT. Aborts with 401/403 on failure.
func (a *AdminAuth) Require() gin.HandlerFunc {
	return func(c *gin.Context) {
		if a.auth == nil {
			response.Error(c, 403, string(errors.ErrorCodeForbidden), "admin access required", rid(c))
			c.Abort()
			return
		}
		// RequireAdmin already validates + aborts on failure.
		a.auth.RequireAdmin()(c)
		if c.IsAborted() {
			return
		}
		c.Set("if.auth_type", "jwt")
	}
}

// HasAdmin reports whether the request carries an admin-role identity. Useful
// for routes that are shared between user and admin contexts.
func HasAdmin(role string) bool {
	return service.HasAdmin(role)
}
