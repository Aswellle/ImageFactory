package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// APIKeyHandler handles admin API key management.
type APIKeyHandler struct {
	admin service.AdminService
}

// NewAPIKeyHandler builds an APIKeyHandler.
func NewAPIKeyHandler(admin service.AdminService) *APIKeyHandler {
	return &APIKeyHandler{admin: admin}
}

// List returns all API keys across users, filterable by status.
// GET /v1/admin/api-keys?page=1&page_size=20&status=active&search=foo
func (h *APIKeyHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	search := c.Query("search")

	keys, total, err := h.admin.APIKeyList(c.Request.Context(), page, pageSize, status, search)
	if err != nil {
		response.Error(c, 500, "INTERNAL_ERROR", "failed to list api keys", rid(c))
		return
	}
	response.Paginated(c, keys, total, page, pageSize)
}

// Revoke deactivates an API key.
// POST /v1/admin/api-keys/:id/revoke
func (h *APIKeyHandler) Revoke(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, "INVALID_REQUEST", "invalid api key id", rid(c))
		return
	}
	if err := h.admin.RevokeAPIKey(c.Request.Context(), id); err != nil {
		response.Error(c, 500, "INTERNAL_ERROR", "failed to revoke api key", rid(c))
		return
	}
	response.NoContent(c)
}
