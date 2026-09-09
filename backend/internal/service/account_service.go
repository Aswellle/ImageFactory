package service

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/account"
	"github.com/imageforge/imageforge/internal/pkg/crypto"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/repository"
)

// AccountService manages AI provider accounts and selects which account to use
// for each generation request.
//
// Core responsibilities:
// 1. Account CRUD (create, read, update, delete)
// 2. Account selection: pick an available account by priority + round-robin
// 3. State management: rate limiting, overload protection, error tracking
type AccountService struct {
	repo *repository.AccountRepository
}

// NewAccountService builds an AccountService.
func NewAccountService(repo *repository.AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

// CreateInput holds fields for creating an account.
type CreateInput struct {
	Name         string
	Platform     string
	Type         string
	Credentials  map[string]any
	Priority     int
	ExpiresAt    *time.Time
	RateLimited  bool
	Overloaded   bool
	Schedulable  bool
	ErrorMessage string
}

// Create adds a new account.
func (s *AccountService) Create(ctx context.Context, input CreateInput) (*ent.Account, error) {
	acc, err := s.repo.Create(ctx, input.Name, input.Platform, input.Type, input.Credentials, input.Priority)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to create account", err)
	}
	return acc, nil
}

// GetByID fetches an account by ID.
func (s *AccountService) GetByID(ctx context.Context, id int64) (*ent.Account, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns all accounts.
func (s *AccountService) List(ctx context.Context) ([]*ent.Account, error) {
	return s.repo.List(ctx)
}

// Update modifies an account.
func (s *AccountService) Update(ctx context.Context, id int64, name string, credentials map[string]any, priority int) (*ent.Account, error) {
	encrypted, err := crypto.EncryptCredentials(credentials)
	if err != nil {
		return nil, fmt.Errorf("encrypt credentials: %w", err)
	}
	return s.repo.Update(ctx, id, func(tx *ent.AccountUpdateOne) *ent.AccountUpdateOne {
		return tx.SetName(name).SetCredentials(encrypted).SetPriority(priority)
	})
}

// Delete removes an account.
func (s *AccountService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// SetStatus updates an account's status.
func (s *AccountService) SetStatus(ctx context.Context, id int64, status string) error {
	st := accountStatus(status)
	if st == "" {
		return errors.New(errors.ErrInvalidRequest, "invalid account status: "+status)
	}
	_, err := s.repo.Update(ctx, id, func(tx *ent.AccountUpdateOne) *ent.AccountUpdateOne {
		return tx.SetStatus(st)
	})
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to set account status", err)
	}
	return nil
}

// SetSchedulable toggles whether an account can be selected for work.
func (s *AccountService) SetSchedulable(ctx context.Context, id int64, schedulable bool) error {
	return s.repo.SetSchedulable(ctx, id, schedulable)
}

// SelectAccount picks an available account for a generation request.
//
// Selection strategy (simplified, ported from Sub2API):
// 1. Filter: only active, schedulable accounts
// 2. Exclude: skip rate-limited or overloaded accounts
// 3. Priority: sort by priority (lower number = higher priority)
// 4. Round-robin: random pick within the same priority tier
//
// All filtering happens at the database level (ListAvailableByPlatform)
// to avoid loading the full table into memory.
func (s *AccountService) SelectAccount(ctx context.Context, platform string) (*ent.Account, error) {
	available, err := s.repo.ListAvailableByPlatform(ctx, platform)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to list accounts", err)
	}
	if len(available) == 0 {
		return nil, errors.New(errors.ErrNotFound, "no available account for platform: "+platform)
	}
	return s.pickByPriority(available), nil
}

// pickByPriority selects the highest-priority account, randomizing within the
// same priority tier to avoid always picking the same one.
func (s *AccountService) pickByPriority(accounts []*ent.Account) *ent.Account {
	if len(accounts) == 0 {
		return nil
	}
	if len(accounts) == 1 {
		return accounts[0]
	}

	// Find the highest priority (lowest number).
	bestPriority := accounts[0].Priority
	for _, acc := range accounts {
		if acc.Priority < bestPriority {
			bestPriority = acc.Priority
		}
	}

	// Collect all accounts at the highest priority.
	topTier := make([]*ent.Account, 0)
	for _, acc := range accounts {
		if acc.Priority == bestPriority {
			topTier = append(topTier, acc)
		}
	}

	// Random pick within the tier (round-robin).
	return topTier[rand.N(len(topTier))]
}

// MarkRateLimited flags an account as rate-limited until retryAfter.
func (s *AccountService) MarkRateLimited(ctx context.Context, id int64, retryAfter time.Duration) error {
	resetAt := time.Now().Add(retryAfter)
	return s.repo.MarkRateLimited(ctx, id, resetAt)
}

// MarkOverloaded flags an account as overloaded for the given duration.
func (s *AccountService) MarkOverloaded(ctx context.Context, id int64, duration time.Duration) error {
	until := time.Now().Add(duration)
	return s.repo.MarkOverloaded(ctx, id, until)
}

// MarkError records an error message on an account.
func (s *AccountService) MarkError(ctx context.Context, id int64, errMsg string) error {
	return s.repo.MarkError(ctx, id, errMsg)
}

// ClearError clears any recorded error on an account.
func (s *AccountService) ClearError(ctx context.Context, id int64) error {
	return s.repo.ClearError(ctx, id)
}

// UpdateLastUsed refreshes the account's last-used timestamp.
func (s *AccountService) UpdateLastUsed(ctx context.Context, id int64) error {
	return s.repo.UpdateLastUsed(ctx, id)
}

// GetAPIKey extracts the plaintext API key from an account's credentials.
func (s *AccountService) GetAPIKey(account *ent.Account) string {
	if account.Credentials == nil {
		return ""
	}
	if v, ok := account.Credentials["api_key"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// accountStatus validates and returns a valid account status.
func accountStatus(status string) account.Status {
	switch status {
	case string(account.StatusActive):
		return account.StatusActive
	case string(account.StatusDisabled):
		return account.StatusDisabled
	case string(account.StatusError):
		return account.StatusError
	default:
		return ""
	}
}

// Ensure AccountService satisfies the expected interface.
var _ = (*AccountService)(nil)
