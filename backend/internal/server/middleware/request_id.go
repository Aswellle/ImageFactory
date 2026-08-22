package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID attaches a unique request ID to every incoming request. It reuses
// a client-supplied X-Request-ID when present (for request correlation across
// services) and otherwise generates one. Handlers must log/carry this ID, never
// secrets.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = "req_" + uuid.New().String()
		}
		c.Set("if.request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}
