package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	image_task "github.com/imageforge/imageforge/internal/domain/image_task"
	"github.com/imageforge/imageforge/internal/service"
)

// AsyncImageHandler handles the submit-then-poll async image generation pattern.
type AsyncImageHandler struct {
	tasks *service.ImageTaskService
}

// NewAsyncImageHandler builds an AsyncImageHandler.
func NewAsyncImageHandler(tasks *service.ImageTaskService) *AsyncImageHandler {
	return &AsyncImageHandler{tasks: tasks}
}

// Pollable reports whether task lookups can be served.
func (h *AsyncImageHandler) pollable() bool {
	return h != nil && h.tasks != nil && h.tasks.Pollable()
}

// Submit creates an async image generation task and returns immediately.
// The client polls GET /v1/images/tasks/:id for the result.
func (h *AsyncImageHandler) Submit(c *gin.Context) {
	if !h.pollable() {
		imageTaskJSONError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "task service unavailable")
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		imageTaskJSONError(c, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required")
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		imageTaskJSONError(c, http.StatusBadRequest, "INVALID_REQUEST", "failed to read request body")
		return
	}

	task, err := h.tasks.Create(c.Request.Context(), userID)
	if err != nil {
		imageTaskJSONError(c, http.StatusInternalServerError, "TASK_CREATE_FAILED", "failed to create task")
		return
	}

	// Run generation detached.
	go h.run(task.ID, body, c.Request.Context())

	c.JSON(http.StatusAccepted, gin.H{
		"data": task,
	})
}

// Get returns the current status of an async image task.
func (h *AsyncImageHandler) Get(c *gin.Context) {
	if !h.pollable() {
		imageTaskJSONError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "task service unavailable")
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		imageTaskJSONError(c, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required")
		return
	}

	id := c.Param("id")
	task, err := h.tasks.Get(c.Request.Context(), userID, id)
	if err != nil {
		imageTaskJSONError(c, http.StatusNotFound, "TASK_NOT_FOUND", "task not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": task})
}

// run executes the image generation synchronously in a goroutine, then
// marks the task as completed or failed in the store.
func (h *AsyncImageHandler) run(taskID string, body []byte, parentCtx context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), h.tasks.ExecutionTimeout())
	defer cancel()

	result, statusCode := h.executeGeneration(ctx, body)
	if statusCode >= 200 && statusCode < 300 {
		_ = h.tasks.Complete(ctx, taskID, statusCode, result)
	} else {
		_ = h.tasks.Fail(ctx, taskID, statusCode, extractImageTaskError(result))
	}
}

// executeGeneration performs the actual upstream call.
// Returns the response body and HTTP status code.
func (h *AsyncImageHandler) executeGeneration(ctx context.Context, body []byte) ([]byte, int) {
	// Placeholder: actual upstream wiring deferred to ImageGateway integration.
	// For now, return a 501 to indicate the gateway is not yet connected.
	errPayload := imageTaskErrorPayload("GATEWAY_NOT_CONNECTED", "image generation gateway not yet integrated")
	return errPayload, http.StatusNotImplemented
}

// failTask marks a task as failed.
func (h *AsyncImageHandler) failTask(taskID string, statusCode int, taskErr json.RawMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = h.tasks.Fail(ctx, taskID, statusCode, taskErr)
}

// extractImageTaskError extracts an error payload from an upstream response.
func extractImageTaskError(body []byte) json.RawMessage {
	if len(body) == 0 {
		return imageTaskErrorPayload("UNKNOWN_ERROR", "unknown error")
	}
	// Try to parse as JSON error.
	var upstream struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &upstream); err == nil && upstream.Error.Message != "" {
		return imageTaskErrorPayload(upstream.Error.Type, upstream.Error.Message)
	}
	return imageTaskErrorPayload("UPSTREAM_ERROR", string(body))
}

func imageTaskErrorPayload(errorType, message string) json.RawMessage {
	b, _ := json.Marshal(map[string]string{"type": errorType, "message": message})
	return b
}

func imageTaskJSONError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}

// getUserID extracts the user ID from the Gin context (set by auth middleware).
func getUserID(c *gin.Context) int64 {
	v, ok := c.Get("user_id")
	if !ok {
		return 0
	}
	switch id := v.(type) {
	case int64:
		return id
	case int:
		return int64(id)
	case float64:
		return int64(id)
	}
	return 0
}

// Ensure image_task types are referenced (silences unused import if needed).
var _ = image_task.StatusCompleted


// Ensure service types are referenced.
var _ = service.ErrImageTaskNotFound
