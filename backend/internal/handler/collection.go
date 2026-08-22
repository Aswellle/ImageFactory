package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// CollectionHandler handles collection CRUD.
type CollectionHandler struct {
	collections *service.CollectionService
}

// NewCollectionHandler builds a CollectionHandler.
func NewCollectionHandler(collections *service.CollectionService) *CollectionHandler {
	return &CollectionHandler{collections: collections}
}

// collectionRequest is the create/update body.
type collectionRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// assetRequest is the body for adding/removing assets from collections.
type collectionAssetRequest struct {
	AssetID int64 `json:"asset_id" binding:"required"`
}

// Create handles POST /v1/collections.
func (h *CollectionHandler) Create(c *gin.Context) {
	var req collectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid request", requestID(c))
		return
	}
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	col, err := h.collections.Create(c.Request.Context(), service.CollectionCreateInput{
		UserID:      uid,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	response.Created(c, col)
}

// List handles GET /v1/collections.
func (h *CollectionHandler) List(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	cols, err := h.collections.List(c.Request.Context(), uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, cols)
}

// Get handles GET /v1/collections/:id.
func (h *CollectionHandler) Get(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid collection id", requestID(c))
		return
	}
	col, err := h.collections.Get(c.Request.Context(), id, uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, col)
}

// Update handles PUT /v1/collections/:id.
func (h *CollectionHandler) Update(c *gin.Context) {
	var req collectionRequest
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
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid collection id", requestID(c))
		return
	}
	col, err := h.collections.Update(c.Request.Context(), id, uid, service.CollectionUpdateInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, col)
}

// Delete handles DELETE /v1/collections/:id.
func (h *CollectionHandler) Delete(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid collection id", requestID(c))
		return
	}
	if err := h.collections.Delete(c.Request.Context(), id, uid); err != nil {
		writeError(c, err)
		return
	}
	response.NoContent(c)
}

// AddAsset handles POST /v1/collections/:id/assets.
func (h *CollectionHandler) AddAsset(c *gin.Context) {
	var req collectionAssetRequest
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
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid collection id", requestID(c))
		return
	}
	if err := h.collections.AddAsset(c.Request.Context(), id, req.AssetID, uid); err != nil {
		writeError(c, err)
		return
	}
	response.NoContent(c)
}

// RemoveAsset handles DELETE /v1/collections/:id/assets/:assetId.
func (h *CollectionHandler) RemoveAsset(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid collection id", requestID(c))
		return
	}
	assetID, err := strconv.ParseInt(c.Param("assetId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid asset id", requestID(c))
		return
	}
	if err := h.collections.RemoveAsset(c.Request.Context(), id, assetID, uid); err != nil {
		writeError(c, err)
		return
	}
	response.NoContent(c)
}

// ListAssets handles GET /v1/collections/:id/assets.
func (h *CollectionHandler) ListAssets(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid collection id", requestID(c))
		return
	}
	assets, err := h.collections.ListAssets(c.Request.Context(), id, uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, assets)
}
