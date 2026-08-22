package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// ProjectHandler handles project CRUD.
type ProjectHandler struct {
	projects *service.ProjectService
}

// NewProjectHandler builds a ProjectHandler.
func NewProjectHandler(projects *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projects: projects}
}

// projectRequest is the create/update body.
type projectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// Create handles POST /v1/projects.
func (h *ProjectHandler) Create(c *gin.Context) {
	var req projectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid request", requestID(c))
		return
	}
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	p, err := h.projects.Create(c.Request.Context(), service.ProjectCreateInput{
		UserID:      uid,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	response.Created(c, p)
}

// List handles GET /v1/projects.
func (h *ProjectHandler) List(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	projects, err := h.projects.List(c.Request.Context(), uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, projects)
}

// Get handles GET /v1/projects/:id.
func (h *ProjectHandler) Get(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	p, err := h.projects.Get(c.Request.Context(), c.Param("id"), uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, p)
}

