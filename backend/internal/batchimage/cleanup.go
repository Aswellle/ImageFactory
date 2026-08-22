package batchimage

import (
	"context"
	"log"
	"time"
)

// CleanupService handles TTL-based cleanup of expired job data.
type CleanupService struct {
	// db would be *ent.Client in the full implementation
	retention time.Duration
}

// NewCleanupService builds a CleanupService with the given retention period.
func NewCleanupService(retention time.Duration) *CleanupService {
	if retention <= 0 {
		retention = 72 * time.Hour // default 72h
	}
	return &CleanupService{retention: retention}
}

// RunOnce performs a single cleanup pass.
// In the full implementation, this would:
// 1. Find jobs older than retention period
// 2. Delete associated storage objects
// 3. Soft-delete or archive job records
func (s *CleanupService) RunOnce(ctx context.Context, now time.Time) error {
	cutoff := now.Add(-s.retention)
	log.Printf("batchimage: cleanup pass, cutoff=%s", cutoff.Format(time.RFC3339))
	// In the full implementation: query and delete expired records
	return nil
}

// DefaultCleanupRetention returns the default retention period.
func DefaultCleanupRetention() time.Duration {
	return 72 * time.Hour
}
