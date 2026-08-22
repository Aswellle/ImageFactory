package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// UserHandler handles admin user management.
type UserHandler struct {
	admin service.AdminService
}

// NewUserHandler builds a UserHandler.
func NewUserHandler(admin service.AdminService) *UserHandler {
	return &UserHandler{admin: admin}
}

// List returns a paginated list of users with filters.
// GET /v1/admin/users?page=1&page_size=20&status=active&role=user&search=foo
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	role := c.Query("role")
	search := c.Query("search")

	users, total, err := h.admin.UserList(c.Request.Context(), page, pageSize, status, role, search)
	if err != nil {
		response.Error(c, 500, "INTERNAL_ERROR", "failed to list users", rid(c))
		return
	}
	response.Paginated(c, users, total, page, pageSize)
}

// SetStatus enables or disables a user account.
// POST /v1/admin/users/:id/status  body: { "status": "active" | "suspended" }
func (h *UserHandler) SetStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, string(errors.ErrorCodeInvalidRequest), "invalid user id", rid(c))
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, string(errors.ErrorCodeInvalidRequest), "status is required", rid(c))
		return
	}
	if req.Status != "active" && req.Status != "suspended" {
		response.Error(c, 400, string(errors.ErrorCodeInvalidRequest), "status must be active or suspended", rid(c))
		return
	}
	if err := h.admin.SetUserStatus(c.Request.Context(), id, req.Status); err != nil {
		response.Error(c, 500, "INTERNAL_ERROR", "failed to update user status", rid(c))
		return
	}
	response.NoContent(c)
}
