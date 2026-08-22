package batchimage

import (
	"context"
	"fmt"
	"time"
)

// SettlementService handles post-generation job settlement.
// It validates terminal states, captures actual costs, and updates job records.
type SettlementService struct {
	// db would be *ent.Client in the full implementation
}

// NewSettlementService builds a SettlementService.
func NewSettlementService() *SettlementService {
	return &SettlementService{}
}

// Settle validates a completed job and records the final cost.
// In the full implementation, this would:
// 1. Validate the job is in a terminal state (completed/failed)
// 2. Capture actual cost from provider response
// 3. Update job status and cost fields
// 4. Write usage record for billing
func (s *SettlementService) Settle(ctx context.Context, job *BatchImageJob) error {
	if job == nil {
		return ErrBatchImageJobNotFound
	}

	// Validate terminal status
	switch job.Status {
	case BatchImageJobStatusCompleted, BatchImageJobStatusFailed, BatchImageJobStatusCancelled:
		// Valid terminal state
	default:
		return fmt.Errorf("job %s is not in a terminal state: %s", job.BatchID, job.Status)
	}

	// In the full implementation: capture cost, write usage record, etc.
	now := time.Now()
	job.SettledAt = &now
	job.UpdatedAt = now

	return nil
}

// CanSettle reports whether a job is ready for settlement.
func CanSettle(job *BatchImageJob) bool {
	if job == nil {
		return false
	}
	switch job.Status {
	case BatchImageJobStatusCompleted, BatchImageJobStatusFailed:
		return job.SettledAt == nil
	}
	return false
}
