// Package handler exposes asset endpoints over HTTP.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// AssetHandler handles asset CRUD, versioning, and content serving.
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

// Content serves the actual image file for an asset.
// GET /v1/assets/:id/content?version=N — version optional; omitted = current.
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

	var version *int
	if v := c.Query("version"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid version", requestID(c))
			return
		}
		version = &n
	}

	data, contentType, err := h.assets.GetContent(c.Request.Context(), id, uid, version)
	if err != nil {
		writeError(c, err)
		return
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Data(http.StatusOK, contentType, data)
}

// Versions handles GET /v1/assets/:id/versions — list version history.
func (h *AssetHandler) Versions(c *gin.Context) {
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
	versions, err := h.assets.ListVersions(c.Request.Context(), id, uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, versions)
}

// GetVersion handles GET /v1/assets/:id/versions/:vid — get a specific version.
func (h *AssetHandler) GetVersion(c *gin.Context) {
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
	vid, err := strconv.Atoi(c.Param("vid"))
	if err != nil || vid <= 0 {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid version", requestID(c))
		return
	}
	ver, err := h.assets.GetVersion(c.Request.Context(), id, vid, uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, ver)
}

// suppress unused import warnings for ent (kept for symmetry with other handlers).
var _ = ent.Asset{}
