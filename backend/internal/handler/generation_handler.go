package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// GenerationHandler handles image generation requests.
type GenerationHandler struct {
	gen *service.GenerationService
}

// NewGenerationHandler builds a GenerationHandler.
func NewGenerationHandler(gen *service.GenerationService) *GenerationHandler {
	return &GenerationHandler{gen: gen}
}

// submitRequest is the generation request body.
type submitRequest struct {
	Prompt         string `json:"prompt" binding:"required"`
	NegativePrompt string `json:"negative_prompt"`
	Model          string `json:"model"`
	Size           string `json:"size"`
	Quality        string `json:"quality"`
	Style          string `json:"style"`
	ImageCount     int    `json:"n"`
	OutputFormat   string `json:"response_format"`
	ProjectID      *int64 `json:"project_id"`
}

// Create handles POST /v1/images/generations.
func (h *GenerationHandler) Create(c *gin.Context) {
	var req submitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid request", requestID(c))
		return
	}
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	result, err := h.gen.Submit(c.Request.Context(), service.SubmitRequest{
		UserID:         uid,
		ProjectID:      req.ProjectID,
		Type:           "generation",
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Model:          req.Model,
		Size:           req.Size,
		Quality:        req.Quality,
		Style:          req.Style,
		ImageCount:     req.ImageCount,
		OutputFormat:   req.OutputFormat,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	response.Created(c, result)
}

// Get handles GET /v1/images/jobs/:id.
func (h *GenerationHandler) Get(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	id := c.Param("id")
	j, err := h.gen.Get(c.Request.Context(), id, uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, gin.H{
		"job_id":          j.ExternalID,
		"status":          j.Status,
		"prompt":          j.Prompt,
		"model":           j.Model,
		"image_count":     j.ImageCount,
		"sub2api_task_id": j.Sub2apiTaskID,
		"error_code":      j.ErrorCode,
		"error_message":   j.ErrorMessage,
		"started_at":      j.StartedAt,
		"completed_at":    j.CompletedAt,
		"created_at":      j.CreatedAt,
	})
}

// List handles GET /v1/images/jobs.
func (h *GenerationHandler) List(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	jobs, err := h.gen.List(c.Request.Context(), uid, 20, 0)
	if err != nil {
		writeError(c, err)
		return
	}
	items := make([]gin.H, 0, len(jobs))
	for _, j := range jobs {
		items = append(items, gin.H{
			"job_id":        j.ExternalID,
			"status":        j.Status,
			"prompt":        j.Prompt,
			"model":         j.Model,
			"error_code":    j.ErrorCode,
			"created_at":    j.CreatedAt,
		})
	}
	response.OK(c, items)
}
