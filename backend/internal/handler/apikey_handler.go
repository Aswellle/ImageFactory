package handler

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// APIKeyHandler handles API key CRUD.
type APIKeyHandler struct {
	apiKeys *service.APIKeyService
}

// NewAPIKeyHandler builds an APIKeyHandler.
func NewAPIKeyHandler(apiKeys *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeys: apiKeys}
}

// createKeyRequest is the create body.
type createKeyRequest struct {
	Name      string `json:"name" binding:"required"`
	ExpiresIn string `json:"expires_in"` // e.g. "30d", "1y", "" = never
}

// expiresInPattern matches duration strings like "30d", "1y", "365d".
var expiresInPattern = regexp.MustCompile(`^[0-9]+[dmy]$`)

// Create handles POST /v1/api-keys.
func (h *APIKeyHandler) Create(c *gin.Context) {
	var req createKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid request", requestID(c))
		return
	}
	if req.ExpiresIn != "" {
		if !expiresInPattern.MatchString(req.ExpiresIn) {
			response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid expires_in format: use e.g. \"30d\", \"1y\", or \"\" for never", requestID(c))
			return
		}
	}
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}

	result, err := h.apiKeys.Create(c.Request.Context(), service.CreateKeyInput{
		UserID: uid,
		Name:   req.Name,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	response.Created(c, result)
}

// List handles GET /v1/api-keys.
func (h *APIKeyHandler) List(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	keys, err := h.apiKeys.List(c.Request.Context(), uid)
	if err != nil {
		writeError(c, err)
		return
	}
	items := make([]gin.H, 0, len(keys))
	for _, k := range keys {
		items = append(items, gin.H{
			"id":         k.ID,
			"name":       k.Name,
			"prefix":     k.KeyPrefix,
			"status":     k.Status,
			"last_used":  k.LastUsedAt,
			"expires_at": k.ExpiresAt,
			"created_at": k.CreatedAt,
		})
	}
	response.OK(c, items)
}

// Revoke handles DELETE /v1/api-keys/:id.
func (h *APIKeyHandler) Revoke(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid id", requestID(c))
		return
	}
	if err := h.apiKeys.Revoke(c.Request.Context(), id, uid); err != nil {
		writeError(c, err)
		return
	}
	response.NoContent(c)
}
