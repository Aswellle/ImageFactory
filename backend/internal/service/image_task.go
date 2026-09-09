package service

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand/v2"
	"time"

	"github.com/imageforge/imageforge/internal/domain/image_task"
)


const (
	defaultImageTaskTTL              = 30 * time.Minute
	defaultImageTaskExecutionTimeout = 5 * time.Minute
)

var (
	ErrImageTaskNotFound    = errors.New("image task not found")
	ErrImageTaskForbidden   = errors.New("image task access denied")
	ErrImageTaskUnavailable = errors.New("image task service unavailable")
)

// ImageTaskService manages the lifecycle of async image generation tasks.
type ImageTaskService struct {
	store            imagetask.Store
	ttl              time.Duration
	executionTimeout time.Duration
}

// NewImageTaskService builds an ImageTaskService with default TTL and timeout.
func NewImageTaskService(store imagetask.Store) *ImageTaskService {
	return NewImageTaskServiceWithOptions(store, defaultImageTaskTTL, defaultImageTaskExecutionTimeout)
}

// NewImageTaskServiceWithOptions builds an ImageTaskService with custom TTL and timeout.
func NewImageTaskServiceWithOptions(store imagetask.Store, ttl, executionTimeout time.Duration) *ImageTaskService {
	return &ImageTaskService{
		store:            store,
		ttl:              ttl,
		executionTimeout: executionTimeout,
	}
}

// ExecutionTimeout returns the max duration a task may run before being marked failed.
func (s *ImageTaskService) ExecutionTimeout() time.Duration {
	return s.executionTimeout
}

// Pollable reports whether the task store is reachable.
func (s *ImageTaskService) Pollable() bool {
	return s != nil && s.store != nil
}

// Create registers a new image task in "processing" state and returns its public view.
func (s *ImageTaskService) Create(ctx context.Context, userID int64) (*imagetask.Task, error) {
	if s.store == nil {
		return nil, ErrImageTaskUnavailable
	}
	now := time.Now()
	rec := &imagetask.Record{
		ID:        generateTaskID(),
		UserID:    userID,
		Status:    imagetask.StatusProcessing,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.Save(rec, s.ttl); err != nil {
		return nil, err
	}
	return imagetask.ToPublic(rec), nil
}

// Get retrieves a task by ID, enforcing user ownership.
func (s *ImageTaskService) Get(ctx context.Context, userID int64, id string) (*imagetask.Task, error) {
	if s.store == nil {
		return nil, ErrImageTaskUnavailable
	}
	rec, err := s.store.Get(id)
	if err != nil {
		return nil, ErrImageTaskNotFound
	}
	if rec.UserID != userID {
		return nil, ErrImageTaskForbidden
	}
	return imagetask.ToPublic(rec), nil
}

// Complete marks a task as completed with the upstream response.
func (s *ImageTaskService) Complete(ctx context.Context, id string, statusCode int, result json.RawMessage) error {
	return s.finish(ctx, id, string(imagetask.StatusCompleted), statusCode, result, nil)
}

// Fail marks a task as failed with an error payload.
func (s *ImageTaskService) Fail(ctx context.Context, id string, statusCode int, taskErr json.RawMessage) error {
	return s.finish(ctx, id, string(imagetask.StatusFailed), statusCode, nil, taskErr)
}

func (s *ImageTaskService) finish(ctx context.Context, id, status string, statusCode int, result, taskErr json.RawMessage) error {
	if s.store == nil {
		return ErrImageTaskUnavailable
	}
	rec, err := s.store.Get(id)
	if err != nil {
		return ErrImageTaskNotFound
	}
	rec.Status = imagetask.Status(status)
	rec.StatusCode = statusCode
	rec.Result = result
	rec.Error = taskErr
	rec.UpdatedAt = time.Now()
	return s.store.Save(rec, s.ttl)
}

func generateTaskID() string {
	return "iftask_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.N(len(letters))]
	}
	return string(b)
}


// ImageTask is an alias to the domain type for handler convenience.
type ImageTask = imagetask.Task

// ImageTaskRecord is an alias to the domain type for handler convenience.
type ImageTaskRecord = imagetask.Record
