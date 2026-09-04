package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// AccountHandler handles admin AI account management.
//
// AI accounts are the credentials used to call upstream image generation APIs
// (Gemini, OpenAI DALL-E, etc.). The admin can add, edit, delete, and control
// the scheduling status of these accounts.
type AccountHandler struct {
	accountSvc *service.AccountService
}

// NewAccountHandler builds an AccountHandler.
func NewAccountHandler(accountSvc *service.AccountService) *AccountHandler {
	return &AccountHandler{accountSvc: accountSvc}
}

// List returns all AI accounts, filterable by platform and status.
// GET /v1/admin/accounts?page=1&page_size=20&platform=gemini&status=active
func (h *AccountHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	platform := c.Query("platform")
	status := c.Query("status")

	accounts, err := h.accountSvc.List(c.Request.Context())
	if err != nil {
		response.Error(c, 500, "INTERNAL_ERROR", "failed to list accounts", rid(c))
		return
	}

	// 过滤
	filtered := make([]map[string]any, 0, len(accounts))
	for _, acc := range accounts {
		if platform != "" && acc.Platform != platform {
			continue
		}
		if status != "" && string(acc.Status) != status {
			continue
		}
		// 注意：不返回 credentials 中的敏感信息
		filtered = append(filtered, sanitizeAccount(acc))
	}

	// 分页
	total := len(filtered)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	response.Paginated(c, filtered, total, page, pageSize)
}

// Get returns a single account by ID.
// GET /v1/admin/accounts/:id
func (h *AccountHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, "INVALID_REQUEST", "invalid account id", rid(c))
		return
	}

	acc, err := h.accountSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, 404, "NOT_FOUND", "account not found", rid(c))
		return
	}

	// 注意：不返回 credentials 中的敏感信息
	response.OK(c, sanitizeAccount(acc))
}

// CreateRequest is the body for creating a new AI account.
type CreateRequest struct {
	Name        string         `json:"name" binding:"required"`
	Platform    string         `json:"platform" binding:"required"`
	Type        string         `json:"type" binding:"required"`
	Credentials map[string]any `json:"credentials" binding:"required"`
	Priority    int            `json:"priority"`
}

// Create adds a new AI account.
// POST /v1/admin/accounts
func (h *AccountHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "INVALID_REQUEST", "invalid request", rid(c))
		return
	}

	acc, err := h.accountSvc.Create(c.Request.Context(), service.CreateInput{
		Name:        req.Name,
		Platform:    req.Platform,
		Type:        req.Type,
		Credentials: req.Credentials,
		Priority:    req.Priority,
	})
	if err != nil {
		response.Error(c, 500, "INTERNAL_ERROR", "failed to create account", rid(c))
		return
	}

	response.Created(c, sanitizeAccount(acc))
}

// UpdateRequest is the body for updating an AI account.
type UpdateRequest struct {
	Name        string         `json:"name"`
	Credentials map[string]any `json:"credentials"`
	Priority    int            `json:"priority"`
}

// Update modifies an existing AI account.
// PUT /v1/admin/accounts/:id
func (h *AccountHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, "INVALID_REQUEST", "invalid account id", rid(c))
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "INVALID_REQUEST", "invalid request", rid(c))
		return
	}

	acc, err := h.accountSvc.Update(c.Request.Context(), id, req.Name, req.Credentials, req.Priority)
	if err != nil {
		response.Error(c, 500, "INTERNAL_ERROR", "failed to update account", rid(c))
		return
	}

	response.OK(c, sanitizeAccount(acc))
}

// Delete removes an AI account.
// DELETE /v1/admin/accounts/:id
func (h *AccountHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, "INVALID_REQUEST", "invalid account id", rid(c))
		return
	}

	if err := h.accountSvc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, 500, "INTERNAL_ERROR", "failed to delete account", rid(c))
		return
	}

	response.NoContent(c)
}

// SetStatusRequest is the body for setting account status.
type SetStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active error disabled"`
}

// SetStatus sets the status of an AI account.
// POST /v1/admin/accounts/:id/status
func (h *AccountHandler) SetStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, "INVALID_REQUEST", "invalid account id", rid(c))
		return
	}

	var req SetStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "INVALID_REQUEST", "invalid request", rid(c))
		return
	}

	if err := h.accountSvc.SetStatus(c.Request.Context(), id, req.Status); err != nil {
		response.Error(c, 500, "INTERNAL_ERROR", "failed to set account status", rid(c))
		return
	}

	response.OK(c, gin.H{"id": id, "status": req.Status})
}

// SetSchedulableRequest is the body for setting account schedulable state.
type SetSchedulableRequest struct {
	Schedulable bool `json:"schedulable"`
}

// SetSchedulable sets whether an AI account can be selected by the scheduler.
// POST /v1/admin/accounts/:id/schedulable
func (h *AccountHandler) SetSchedulable(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, "INVALID_REQUEST", "invalid account id", rid(c))
		return
	}

	var req SetSchedulableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "INVALID_REQUEST", "invalid request", rid(c))
		return
	}

	if err := h.accountSvc.SetSchedulable(c.Request.Context(), id, req.Schedulable); err != nil {
		response.Error(c, 500, "INTERNAL_ERROR", "failed to set schedulable", rid(c))
		return
	}

	response.OK(c, gin.H{"id": id, "schedulable": req.Schedulable})
}

// sanitizeAccount 返回账号的安全视图（隐藏敏感凭证）。
func sanitizeAccount(acc *ent.Account) map[string]any {
	if acc == nil {
		return nil
	}
	return map[string]any{
		"id":            acc.ID,
		"name":          acc.Name,
		"platform":      acc.Platform,
		"type":          acc.Type,
		"priority":      acc.Priority,
		"status":        acc.Status,
		"error_message": acc.ErrorMessage,
		"last_used_at":  acc.LastUsedAt,
		"expires_at":    acc.ExpiresAt,
		"schedulable":   acc.Schedulable,
		"created_at":    acc.CreatedAt,
		"updated_at":    acc.UpdatedAt,
	}
}

