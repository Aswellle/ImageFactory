package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// PromptTemplateHandler handles prompt-template CRUD, apply, and built-in listing.
type PromptTemplateHandler struct {
	templates *service.PromptTemplateService
}

// NewPromptTemplateHandler builds a PromptTemplateHandler.
func NewPromptTemplateHandler(templates *service.PromptTemplateService) *PromptTemplateHandler {
	return &PromptTemplateHandler{templates: templates}
}

// templateRequest is the create/update body for a user-owned template.
type templateRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Content     string   `json:"content" binding:"required"`
	Variables   []string `json:"variables"`
	Category    string   `json:"category"`
}

// updateTemplateRequest is the update body; all fields optional (omitted = no change).
type updateTemplateRequest struct {
	Name           *string  `json:"name"`
	Description    *string  `json:"description"`
	Content        *string  `json:"content"`
	Variables      []string `json:"variables"`
	Category       *string  `json:"category"`
	ClearVariables bool     `json:"clear_variables"`
}

// applyRequest is the body for filling a template's variables.
type applyRequest struct {
	Variables map[string]string `json:"variables"`
}

// Create handles POST /v1/prompt-templates.
func (h *PromptTemplateHandler) Create(c *gin.Context) {
	var req templateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid request", requestID(c))
		return
	}
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	t, err := h.templates.Create(c.Request.Context(), service.TemplateCreateInput{
		UserID:      uid,
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		Content:     req.Content,
		Variables:   req.Variables,
		Category:    req.Category,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	response.Created(c, t)
}

// List handles GET /v1/prompt-templates. Pass built_in=true to fetch the
// system-provided templates instead of the user's own.
func (h *PromptTemplateHandler) List(c *gin.Context) {
	if c.Query("built_in") == "true" {
		response.OK(c, service.GetBuiltIn())
		return
	}
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	templates, err := h.templates.List(c.Request.Context(), uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, templates)
}

// Get handles GET /v1/prompt-templates/:id.
func (h *PromptTemplateHandler) Get(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid template id", requestID(c))
		return
	}
	t, err := h.templates.Get(c.Request.Context(), id, uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, t)
}

func (h *PromptTemplateHandler) Update(c *gin.Context) {
	var req updateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid request", requestID(c))
		return
	}
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid template id", requestID(c))
		return
	}
	t, err := h.templates.Update(c.Request.Context(), id, uid, service.TemplateUpdateInput{
		Name:           req.Name,
		Description:    req.Description,
		Content:        req.Content,
		Variables:      req.Variables,
		Category:       req.Category,
		ClearVariables: req.ClearVariables,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, t)
}

// Delete handles DELETE /v1/prompt-templates/:id.
func (h *PromptTemplateHandler) Delete(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid template id", requestID(c))
		return
	}
	if err := h.templates.Delete(c.Request.Context(), id, uid); err != nil {
		writeError(c, err)
		return
	}
	response.NoContent(c)
}

// Apply handles POST /v1/prompt-templates/:id/apply — fills variables, returns the final prompt.
func (h *PromptTemplateHandler) Apply(c *gin.Context) {
	var req applyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow an empty body — just render the template as-is.
		req = applyRequest{}
	}
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid template id", requestID(c))
		return
	}
	prompt, err := h.templates.Apply(c.Request.Context(), id, uid, req.Variables)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, gin.H{"prompt": prompt})
}
