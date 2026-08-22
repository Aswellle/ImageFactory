package batchimage

import (
	"context"
	"log"
	"time"
)

// Worker polls provider status and drives job transitions.
// In the full implementation, this would run as a background goroutine.
type Worker struct {
	registry *BatchImageProviderRegistry
	settle   *SettlementService
	interval time.Duration
}

// NewWorker builds a Worker.
func NewWorker(registry *BatchImageProviderRegistry, settle *SettlementService) *Worker {
	return &Worker{
		registry: registry,
		settle:   settle,
		interval: 5 * time.Second,
	}
}

// PollOnce performs a single poll cycle over all active jobs.
// In the full implementation, this would query the DB for active jobs.
func (w *Worker) PollOnce(ctx context.Context, jobs []*BatchImageJob) []error {
	var errs []error
	for _, job := range jobs {
		if err := w.pollJob(ctx, job); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (w *Worker) pollJob(ctx context.Context, job *BatchImageJob) error {
	// Only poll active jobs
	switch job.Status {
	case BatchImageJobStatusSubmitted, BatchImageJobStatusRunning:
		// Active
	default:
		return nil
	}

	provider, ok := w.registry.Get(job.Provider)
	if !ok {
		log.Printf("batchimage: unknown provider %s for job %s", job.Provider, job.BatchID)
		return nil
	}

	status, err := provider.Get(ctx, job, nil)
	if err != nil {
		log.Printf("batchimage: poll error for job %s: %v", job.BatchID, err)
		return nil // Don't fail the whole batch on single poll error
	}

	// Map provider state to job status
	switch status.InternalState {
	case BatchProviderStateSucceeded:
		job.Status = BatchImageJobStatusCompleted
	case BatchProviderStateFailed:
		job.Status = BatchImageJobStatusFailed
		if status.ErrorCode != "" {
			job.LastErrorCode = &status.ErrorCode
		}
		if status.ErrorMessage != "" {
			job.LastErrorMessage = &status.ErrorMessage
		}
	case BatchProviderStateCancelled:
		job.Status = BatchImageJobStatusCancelled
	case BatchProviderStateRunning:
		job.Status = BatchImageJobStatusRunning
	default:
		return nil // Pending or unknown, no change
	}

	now := time.Now()
	job.UpdatedAt = now
	if job.Status == BatchImageJobStatusCompleted || job.Status == BatchImageJobStatusFailed {
		job.FinishedAt = &now
	}

	// Settle completed/failed jobs
	if CanSettle(job) {
		_ = w.settle.Settle(ctx, job)
	}

	return nil
}
