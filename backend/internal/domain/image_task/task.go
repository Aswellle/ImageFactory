// Package imagetask defines the domain types and store interface for async
// image generation tasks. It is a small, dependency-free package that both the
// service layer and the repository layer can import without creating cycles.
package imagetask

import "time"

import "errors"

// Domain errors for image tasks.
var (
	ErrImageTaskNotFound  = errors.New("image task not found")
	ErrImageTaskForbidden = errors.New("image task access denied")
)

// Status is the lifecycle state of an async image task.
type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

// Record is the internal representation persisted in the store.
// Ownership fields are intentionally excluded from the API view.
type Record struct {
	ID        string          `json:"id"`
	UserID    int64           `json:"user_id"`
	Status    Status          `json:"status"`
	StatusCode int            `json:"status_code,omitempty"`
	Result    []byte          `json:"result,omitempty"`
	Error     []byte          `json:"error,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Task is the API-safe representation returned to callers.
type Task struct {
	ID        string    `json:"id"`
	Status    Status    `json:"status"`
	StatusCode int      `json:"status_code,omitempty"`
	Result    []byte    `json:"result,omitempty"`
	Error     []byte    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToPublic converts an internal record to the API-safe view.
func ToPublic(r *Record) *Task {
	return &Task{
		ID:        r.ID,
		Status:    r.Status,
		StatusCode: r.StatusCode,
		Result:    r.Result,
		Error:     r.Error,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// Store defines the persistence operations for image tasks.
type Store interface {
	// Save persists a task record with the given TTL.
	Save(task *Record, ttl time.Duration) error
	// Get retrieves a task by its ID.
	Get(id string) (*Record, error)
}
