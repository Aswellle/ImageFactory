package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
)

// Context keys for handler-layer values set by middleware. Using a typed key
// string avoids collisions across packages.
const (
	ctxKeyUserID    = "if.user_id"
	ctxKeyRole      = "if.role"
	ctxKeyRequestID = "if.request_id"
	ctxKeyAPIKeyID  = "if.api_key_id"
)

// requestID reads the request ID set by the request-id middleware.
func requestID(c *gin.Context) string {
	v, _ := c.Get(ctxKeyRequestID)
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// userID reads the authenticated user ID.
func userID(c *gin.Context) (int64, bool) {
	v, ok := c.Get(ctxKeyUserID)
	if !ok {
		return 0, false
	}
	n, ok := v.(int64)
	return n, ok
}

// writeError maps an error to an HTTP response using the error envelope.
func writeError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.Error); ok {
		response.Error(c, appErr.HTTPStatus, string(appErr.Code), appErr.Message, appErr.RequestID)
		return
	}
	response.Error(c, 500, string(errors.ErrorCodeInternal), "internal error", requestID(c))
}
