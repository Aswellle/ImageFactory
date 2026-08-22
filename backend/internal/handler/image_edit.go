// Package handler (image_edit): HTTP handler for the OpenAI-compatible image
// edit endpoint.
//
// ImageEditHandler delegates to service.ImageEditService for parsing,
// validation, orchestration, and AssetVersion creation.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
	"github.com/imageforge/imageforge/internal/service"
)

// ImageEditHandler handles image edit requests.
type ImageEditHandler struct {
	edit *service.ImageEditService
}

// NewImageEditHandler builds an ImageEditHandler.
func NewImageEditHandler(edit *service.ImageEditService) *ImageEditHandler {
	return &ImageEditHandler{edit: edit}
}

// Edit handles POST /v1/images/edits.
//
// The endpoint is authenticated: it requires a logged-in user (set by the
// auth middleware) so the edited result can be stored as an AssetVersion owned
// by that user. It accepts both JSON (base64/URL sources) and multipart
// uploads; the service layer normalizes both.
func (h *ImageEditHandler) Edit(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
		return
	}

	req, err := h.edit.ParseImageEditRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}

	result, err := h.edit.Edit(c.Request.Context(), uid, req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
