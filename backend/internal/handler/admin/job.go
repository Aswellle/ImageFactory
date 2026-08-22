package admin

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// JobHandler handles admin image-generation job management.
type JobHandler struct {
	admin service.AdminService
}

// NewJobHandler builds a JobHandler.
func NewJobHandler(admin service.AdminService) *JobHandler {
	return &JobHandler{admin: admin}
}

// List returns all image-generation jobs across users, filterable by status.
// GET /v1/admin/jobs?page=1&page_size=20&status=failed&search=foo
func (h *JobHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	search := c.Query("search")

	jobs, total, err := h.admin.JobList(c.Request.Context(), page, pageSize, status, search)
	if err != nil {
		response.Error(c, 500, "INTERNAL_ERROR", "failed to list jobs", rid(c))
		return
	}
	response.Paginated(c, jobs, total, page, pageSize)
}

// Retry re-arms a failed job for another attempt.
// POST /v1/admin/jobs/:id/retry
func (h *JobHandler) Retry(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, "INVALID_REQUEST", "invalid job id", rid(c))
		return
	}
	if err := h.admin.RetryJob(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrNotRetryable) {
			response.Error(c, 409, "INVALID_REQUEST", "only failed jobs can be retried", rid(c))
			return
		}
		response.Error(c, 500, "INTERNAL_ERROR", "failed to retry job", rid(c))
		return
	}
	response.NoContent(c)
}
