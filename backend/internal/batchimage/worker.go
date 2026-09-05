package batchimage

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// Worker polls provider status and drives job transitions.
// In the full implementation, this would run as a background goroutine.
type Worker struct {
	registry *BatchImageProviderRegistry
	settle   *SettlementService
	log      *zap.Logger
	interval time.Duration
}

func NewWorker(registry *BatchImageProviderRegistry, settle *SettlementService, log *zap.Logger) *Worker {
	if log == nil {
		log = zap.NewNop()
	}
	return &Worker{
		registry: registry,
		settle:   settle,
		log:      log,
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
	job.RLock()
	switch job.Status {
	case BatchImageJobStatusSubmitted, BatchImageJobStatusRunning:
		// Active
	default:
		job.RUnlock()
		return nil
	}
	job.RUnlock()
 provider, ok := w.registry.Get(job.Provider)
 if !ok {
	w.log.Info("batchimage: unknown provider for job", zap.String("provider", job.Provider), zap.String("batch_id", job.BatchID))
	return nil
}

 status, err := provider.Get(ctx, job, nil)
 if err != nil {
	w.log.Warn("batchimage: poll error", zap.String("batch_id", job.BatchID), zap.Error(err))
 }

	// Map provider state to job status
	job.Lock()
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
		job.Unlock()
		return nil // Pending or unknown, no change
	}

	now := time.Now()
	job.UpdatedAt = now
if job.Status == BatchImageJobStatusCompleted || job.Status == BatchImageJobStatusFailed {
    job.FinishedAt = &now
}
job.Unlock()

// Settle completed/failed jobs
if CanSettle(job) {
    _ = w.settle.Settle(ctx, job)
}

return nil
}
