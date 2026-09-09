// Package job defines the image-generation job model, queue interface, and
// worker lifecycle. ImageForge treats every image operation as an async job so
// the UI can show progress, retry, and cancel.
//
// The queue is intentionally abstract: Phase 1 uses an in-memory channel
// queue; Phase 2 can swap in Redis without touching job logic.
package job

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/imageforge/imageforge/internal/domain"
)

// Job is the internal runtime representation of an image job. It wraps the DB
// entity (generation_jobs) with execution metadata.
type Job struct {
	ID        string
	UserID    int64
	ProjectID *int64
	Type      string // domain.JobTypeGeneration | JobTypeEdit
	Status    string // domain.JobStatus*
	Model     string
	Prompt    string

	// Retry tracking.
	Attempt     int
	MaxAttempts int

	// Provider linkage.
	Sub2APITaskID string

	// Timing.
	CreatedAt   time.Time
	ScheduledAt *time.Time

	// Inputs for edits (source asset IDs).
	InputAssetIDs []int64
}

// New creates a job with sensible defaults.
func New(userID int64, typ string) *Job {
	return &Job{
		ID:          "job_" + uuid.New().String(),
		UserID:      userID,
		Type:        typ,
		Status:      domain.JobStatusPending,
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
	}
}

// Task is the unit of work submitted to a worker. It carries a handle to
// update job status and persist the resulting asset.
type Task struct {
	Job *Job
	// Run executes the task. On success it must register the output asset.
	// On failure it returns a stable ImageForge error.
	Run func(ctx context.Context) error
}

// Queue is the job dispatch abstraction.
type Queue interface {
	// Submit enqueues a task for processing.
	Submit(ctx context.Context, t *Task) error
	// Stop gracefully shuts down the queue.
	Stop() error
}

// Processor processes tasks from the queue. One or more Processor instances
// may run concurrently.
type Processor interface {
	Start(ctx context.Context) error
	Stop() error
}
