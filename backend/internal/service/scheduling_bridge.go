package service

import (
	"context"
	"time"

	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/internal/repository"
	"github.com/imageforge/imageforge/internal/service/scheduling"
)

// SchedulingService bridges the pure scheduling algorithms (package scheduling)
// to ImageForge's Ent-based persistence layer.
//
// It implements scheduling.AccountStateHandler so the rate-limit and session
// window algorithms can persist state changes through the existing repository.
//
// Sub2API equivalent: internal/service/ratelimit_service.go (RateLimitService)
// combined with the scheduler's runtime state management.
type SchedulingService struct {
	accountRepo *repository.AccountRepository
	thresholds  map[string]int
}

// NewSchedulingService creates a SchedulingService.
func NewSchedulingService(accountRepo *repository.AccountRepository) *SchedulingService {
	return &SchedulingService{
		accountRepo: accountRepo,
		thresholds:  make(map[string]int),
	}
}

// SetThresholds configures the per-platform scheduling thresholds.
func (s *SchedulingService) SetThresholds(thresholds map[string]int) {
	s.thresholds = thresholds
}

// EvaluateThreshold evaluates whether an account should be paused based on
// its platform's scheduling threshold.
//
// Ported from Sub2API: ratelimit_service.go ApplyAccountSchedulingThreshold
func (s *SchedulingService) EvaluateThreshold(ctx context.Context, acc *ent.Account) scheduling.AccountSchedulingThresholdDecision {
	sa := EntAccountToScheduling(acc)
	return scheduling.EvaluateAccountSchedulingThreshold(sa, s.thresholds, time.Now())
}

// EntAccountToScheduling converts an ent.Account to a scheduling.Account.
func EntAccountToScheduling(acc *ent.Account) *scheduling.Account {
	if acc == nil {
		return nil
	}
	return scheduling.NewAccount(
		acc.ID,
		acc.Name,
		acc.Platform,
		acc.Type,
		acc.Credentials,
		acc.Extra,
		acc.Priority,
		string(acc.Status),
		acc.RateLimitedAt,
		acc.RateLimitResetAt,
		acc.OverloadUntil,
		acc.TempUnschedulableUntil,
		acc.SessionWindowStart,
		acc.SessionWindowEnd,
		"",
	)
}

// UpdateRateLimit implements scheduling.AccountStateHandler.
func (s *SchedulingService) UpdateRateLimit(ctx context.Context, accountID int64, resetAt time.Time) error {
	return s.accountRepo.MarkRateLimited(ctx, accountID, resetAt)
}

// ClearRateLimit implements scheduling.AccountStateHandler.
func (s *SchedulingService) ClearRateLimit(ctx context.Context, accountID int64) error {
	return s.accountRepo.SetSchedulable(ctx, accountID, true)
}

// SetOverloaded implements scheduling.AccountStateHandler.
func (s *SchedulingService) SetOverloaded(ctx context.Context, accountID int64, until time.Time) error {
	return s.accountRepo.MarkOverloaded(ctx, accountID, until)
}

// SetError implements scheduling.AccountStateHandler.
func (s *SchedulingService) SetError(ctx context.Context, accountID int64, errMsg string) error {
	return s.accountRepo.MarkError(ctx, accountID, errMsg)
}

// SetTempUnschedulable implements scheduling.AccountStateHandler.
func (s *SchedulingService) SetTempUnschedulable(ctx context.Context, accountID int64, until time.Time, reason string) error {
	return s.accountRepo.SetSchedulable(ctx, accountID, false)
}

// UpdateExtra implements scheduling.AccountStateHandler.
func (s *SchedulingService) UpdateExtra(ctx context.Context, accountID int64, updates map[string]any) error {
	return s.accountRepo.UpdateExtra(ctx, accountID, updates)
}

// UpdateSessionWindow implements scheduling.AccountStateHandler.
func (s *SchedulingService) UpdateSessionWindow(ctx context.Context, accountID int64, start, end *time.Time, util float64) error {
	return s.accountRepo.UpdateSessionWindow(ctx, accountID, start, end, util)
}

// RateLimitStrategy returns a scheduling.RateLimitStrategy backed by this service.
func (s *SchedulingService) RateLimitStrategy() *scheduling.RateLimitStrategy {
	return scheduling.NewRateLimitStrategy(s)
}
