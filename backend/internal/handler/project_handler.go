package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/asset"
	"github.com/imageforge/imageforge/ent/project"
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

// AssetHandler handles asset queries.
type AssetHandler struct {
	assets *service.AssetService
}

// NewAssetHandler builds an AssetHandler.
func NewAssetHandler(assets *service.AssetService) *AssetHandler {
	return &AssetHandler{assets: assets}
}

// listAssetsQuery is the query params for listing assets.
type listAssetsQuery struct {
	ProjectID string `form:"project_id"`
	Tag       string `form:"tag"`
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"page_size,default=20"`
}

// List handles GET /v1/assets.
func (h *AssetHandler) List(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	var q listAssetsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid query", requestID(c))
		return
	}
	filter := service.AssetFilter{
		ProjectID: q.ProjectID,
		Tag:       q.Tag,
		Page:      q.Page,
		PageSize:  q.PageSize,
	}
	assets, total, err := h.assets.List(c.Request.Context(), uid, filter)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Paginated(c, assets, total, q.Page, q.PageSize)
}

// Get handles GET /v1/assets/:id.
func (h *AssetHandler) Get(c *gin.Context) {
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
	a, err := h.assets.Get(c.Request.Context(), id, uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, a)
}

// Delete handles DELETE /v1/assets/:id.
func (h *AssetHandler) Delete(c *gin.Context) {
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
	if err := h.assets.Delete(c.Request.Context(), id, uid); err != nil {
		writeError(c, err)
		return
	}
	response.NoContent(c)
}

// helper to parse int from string.
func parseIntPtr(s string) *int64 {
	if s == "" {
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &n
}

// suppress unused import warnings.
var _ = ent.Project{}
var _ = asset.FieldStatus
var _ = project.FieldName

// Content serves the actual image file for an asset.
func (h *AssetHandler) Content(c *gin.Context) {
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
	data, contentType, err := h.assets.GetContent(c.Request.Context(), id, uid)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Data(http.StatusOK, contentType, data)
}
