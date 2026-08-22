package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/domain"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// Auth validates the Bearer JWT and sets user_id + role on the context.
// Optional routes call AuthOptional; protected routes call Auth.
type Auth struct {
	jwt *service.JWTService
}

// NewAuth builds Auth middleware.
func NewAuth(jwt *service.JWTService) *Auth {
	return &Auth{jwt: jwt}
}

// Require enforces a valid token. It aborts with 401 on failure.
func (a *Auth) Require() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := a.parse(c)
		if err != nil {
			response.Error(c, 401, string(errors.ErrorCodeUnauthorized), "authentication required", rid(c))
			c.Abort()
			return
		}
		c.Set("if.user_id", claims.UserID)
		c.Set("if.role", claims.Role)
		c.Next()
	}
}

// RequireAdmin enforces admin role.
func (a *Auth) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := a.parse(c)
		if err != nil {
			response.Error(c, 401, string(errors.ErrorCodeUnauthorized), "authentication required", rid(c))
			c.Abort()
			return
		}
		if claims.Role != domain.RoleAdmin {
			response.Error(c, 403, string(errors.ErrorCodeForbidden), "admin access required", rid(c))
			c.Abort()
			return
		}
		c.Set("if.user_id", claims.UserID)
		c.Set("if.role", claims.Role)
		c.Next()
	}
}

// RequireOptional sets user context when a valid token is present but does not
// abort when it is absent.
func (a *Auth) RequireOptional() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := a.parse(c)
		if err == nil {
			c.Set("if.user_id", claims.UserID)
			c.Set("if.role", claims.Role)
		}
		c.Next()
	}
}

func (a *Auth) parse(c *gin.Context) (*service.Claims, error) {
	header := c.GetHeader("Authorization")
	if header == "" {
		return nil, errNoAuth
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, errNoAuth
	}
	return a.jwt.Parse(parts[1])
}

func rid(c *gin.Context) string {
	v, _ := c.Get("if.request_id")
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

var errNoAuth = errors.New(errors.ErrUnauthorized, "missing bearer token")
