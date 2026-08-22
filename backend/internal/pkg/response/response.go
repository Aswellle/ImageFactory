// Package response provides uniform JSON envelope helpers.
package response

import (
	"github.com/gin-gonic/gin"
)

// OK returns a standard success envelope.
func OK(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{"data": data})
}

// Created returns a 201 success envelope.
func Created(c *gin.Context, data interface{}) {
	c.JSON(201, gin.H{"data": data})
}

// NoContent returns 204.
func NoContent(c *gin.Context) {
	c.Status(204)
}

// Error returns a uniform error envelope. The request_id is always included
// when available so client errors can be correlated server-side.
func Error(c *gin.Context, status int, code, message, requestID string) {
	body := gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	}
	if requestID != "" {
		body["error"].(gin.H)["request_id"] = requestID
	}
	c.JSON(status, body)
}

// Paginated returns a paginated list envelope.
func Paginated(c *gin.Context, items interface{}, total, page, pageSize int) {
	c.JSON(200, gin.H{
		"data": items,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
