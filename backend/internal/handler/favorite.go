package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// FavoriteHandler handles favorite endpoints.
type FavoriteHandler struct {
	favorites *service.FavoriteService
}

// NewFavoriteHandler builds a FavoriteHandler.
func NewFavoriteHandler(favorites *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favorites: favorites}
}

// assetIDQuery is the query param for asset-based endpoints.

// Create handles POST /v1/favorites?asset_id=N.
func (h *FavoriteHandler) Create(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	assetID, err := strconv.ParseInt(c.Query("asset_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid asset_id", requestID(c))
		return
	}
	f, err := h.favorites.Add(c.Request.Context(), uid, assetID)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Created(c, f)
}

// List handles GET /v1/favorites.
func (h *FavoriteHandler) List(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	favs, err := h.favorites.List(c.Request.Context(), uid)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, favs)
}

// Delete handles DELETE /v1/favorites/:id.
func (h *FavoriteHandler) Delete(c *gin.Context) {
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
	if err := h.favorites.Remove(c.Request.Context(), uid, assetID); err != nil {
		writeError(c, err)
		return
	}
	response.NoContent(c)
}

// Check handles GET /v1/favorites/check?asset_id=N — returns {data: true/false}.
func (h *FavoriteHandler) Check(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}
	assetID, err := strconv.ParseInt(c.Query("asset_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, string(errors.ErrorCodeInvalidRequest), "invalid asset_id", requestID(c))
		return
	}
	isFav, err := h.favorites.IsFavorited(c.Request.Context(), uid, assetID)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, isFav)
}
