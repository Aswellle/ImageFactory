package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// TagHandler handles tag CRUD.
type TagHandler struct {
	tags *service.TagService
}

// NewTagHandler builds a TagHandler.
func NewTagHandler(tags *service.TagService) *TagHandler {
	return &TagHandler{tags: tags}
}

// tagRequest is the create/update body.
type tagRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

// tagAssetRequest is the body for tagging/untagging.
type tagAssetRequest struct {
	AssetID int64 `json:"asset_id" binding:"required"`
}

// Create handles POST /v1/tags.
func (h *TagHandler) Create(c *gin.Context) {
	var req tagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid request", requestID(c))
		return
	}
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	t, err := h.tags.Create(c.Request.Context(), service.TagCreateInput{
		UserID: uid,
		Name:   req.Name,
		Color:  req.Color,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	response.Created(c, t)
}

// List handles GET /v1/tags.
func (h *TagHandler) List(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	tags, err := h.tags.List(c.Request.Context(), uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, tags)
}

// Get handles GET /v1/tags/:id.
func (h *TagHandler) Get(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid tag id", requestID(c))
		return
	}
	t, err := h.tags.Get(c.Request.Context(), id, uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, t)
}

// Update handles PUT /v1/tags/:id.
func (h *TagHandler) Update(c *gin.Context) {
	var req tagRequest
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
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid tag id", requestID(c))
		return
	}
	t, err := h.tags.Update(c.Request.Context(), id, uid, service.TagUpdateInput{
		Name:  req.Name,
		Color: req.Color,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, t)
}

// Delete handles DELETE /v1/tags/:id.
func (h *TagHandler) Delete(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid tag id", requestID(c))
		return
	}
	if err := h.tags.Delete(c.Request.Context(), id, uid); err != nil {
		writeError(c, err)
		return
	}
	response.NoContent(c)
}

// TagAsset handles POST /v1/tags/:id/assets.
func (h *TagHandler) TagAsset(c *gin.Context) {
	var req tagAssetRequest
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
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid tag id", requestID(c))
		return
	}
	if err := h.tags.TagAsset(c.Request.Context(), id, req.AssetID, uid); err != nil {
		writeError(c, err)
		return
	}
	response.NoContent(c)
}

// UntagAsset handles DELETE /v1/tags/:id/assets/:assetId.
func (h *TagHandler) UntagAsset(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid tag id", requestID(c))
		return
	}
	assetID, err := strconv.ParseInt(c.Param("assetId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid asset id", requestID(c))
		return
	}
	if err := h.tags.UntagAsset(c.Request.Context(), id, assetID, uid); err != nil {
		writeError(c, err)
		return
	}
	response.NoContent(c)
}

// ListAssets handles GET /v1/tags/:id/assets — assets carrying this tag.
func (h *TagHandler) ListAssets(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid tag id", requestID(c))
		return
	}
	assets, err := h.tags.ListAssets(c.Request.Context(), id, uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, assets)
}

// ListAssetTags handles GET /v1/assets/:id/tags — tags for a specific asset.
func (h *TagHandler) ListAssetTags(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	assetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid asset id", requestID(c))
		return
	}
	tags, err := h.tags.ListTagsForAsset(c.Request.Context(), assetID, uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, tags)
}
