package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
)

// MaxRequestBodySize limits the size of incoming request bodies to prevent
// memory exhaustion attacks. Default limit is 1 MB.
func MaxRequestBodySize(maxBytes int64) gin.HandlerFunc {
	if maxBytes <= 0 {
		maxBytes = 1 << 20 // 1 MB default
	}

	return func(c *gin.Context) {
		// Only limit requests with a body
		if c.Request.Body == nil || c.Request.ContentLength == 0 {
			c.Next()
			return
		}

		// Check Content-Length header first (fast path)
		if c.Request.ContentLength > maxBytes {
			response.Error(c, http.StatusRequestEntityTooLarge, string(errors.ErrorCodeInvalidRequest),
				"request body too large", rid(c))
			c.Abort()
			return
		}

		// Limit the body reader to prevent reading more than maxBytes
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
