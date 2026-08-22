package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/imageforge/imageforge/internal/service"
)

// OpenAIImagesHandler handles OpenAI-compatible image endpoints.
type OpenAIImagesHandler struct {
	images *OpenAIImagesService
}

// NewOpenAIImagesHandler builds an OpenAIImagesHandler.
func NewOpenAIImagesHandler(images *OpenAIImagesService) *OpenAIImagesHandler {
	return &OpenAIImagesHandler{images: images}
}

// Generations handles POST /v1/images/generations.
func (h *OpenAIImagesHandler) Generations(c *gin.Context) {
	var req service.OpenAIImagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	// In the full implementation: delegate to GenerationService
	// For now, return a placeholder response
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "image generation service pending full integration",
			"type":    "api_error",
		},
	})
}

// Edits handles POST /v1/images/edits.
func (h *OpenAIImagesHandler) Edits(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "image editing not yet available",
			"type":    "api_error",
		},
	})
}

// Models handles GET /v1/models.
func (h *OpenAIImagesHandler) Models(c *gin.Context) {
	models := service.AvailableModels()
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   models,
	})
}

// OpenAIImagesService is a placeholder for the service that will handle
// the actual image generation logic.
type OpenAIImagesService struct{}

// NewOpenAIImagesService builds an OpenAIImagesService.
func NewOpenAIImagesService() *OpenAIImagesService {
	return &OpenAIImagesService{}
}

// Ensure time is used.
var _ = time.Now
