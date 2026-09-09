package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// APIKeyAuth returns middleware that authenticates requests using an API key.
// It supports both "Authorization: Bearer <key>" and "x-api-key: <key>" headers.
// On success, it sets the user_id and auth_method on the context.
func APIKeyAuth(apiKeys *service.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := extractAPIKey(c)
		if key == "" {
			response.Error(c, 401, string(errors.ErrorCodeUnauthorized), "API key required", rid(c))
			c.Abort()
			return
		}

		userID, err := apiKeys.Validate(c.Request.Context(), key)
		if err != nil {
			response.Error(c, 401, string(errors.ErrorCodeUnauthorized), "invalid API key", rid(c))
			c.Abort()
			return
		}

		c.Set("if.user_id", userID)
		c.Set("if.role", "user")
		c.Set("if.auth_method", "api_key")
		c.Next()
	}
}

// extractAPIKey reads the API key from the request.
func extractAPIKey(c *gin.Context) string {
	// Check x-api-key header first.
	if key := c.GetHeader("x-api-key"); key != "" {
		return key
	}
	// Fall back to Authorization: Bearer <key>.
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}
