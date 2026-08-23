package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/apikey"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"go.uber.org/zap"
)
type APIKeyService struct {
	db  *ent.Client
	log *zap.Logger
}

func NewAPIKeyService(db *ent.Client) *APIKeyService {
	return &APIKeyService{db: db, log: zap.NewNop()}
}

// CreateKeyInput is the input for creating an API key.
type CreateKeyInput struct {
	UserID      int64
	Name        string
	Permissions int
	ExpiresAt   *time.Time
}

// CreateKeyResult returns the created key with its plaintext secret.
type CreateKeyResult struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Prefix    string `json:"prefix"`
	// Plaintext is ONLY returned once at creation.
	Plaintext string `json:"key"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// generateKey creates a new API key and returns (plaintext, hash, prefix).
func generateKey() (plaintext, hash, prefix string, err error) {
	raw := uuid.New().String() + uuid.New().String()
	plaintext = "sk-if-" + raw
	h := sha256.Sum256([]byte(plaintext))
	hash = hex.EncodeToString(h[:])
	prefix = plaintext[:8]
	return
}

// Create generates a new API key for a user. The plaintext is returned once.
func (s *APIKeyService) Create(ctx context.Context, input CreateKeyInput) (*CreateKeyResult, error) {
	plaintext, hash, prefix, err := generateKey()
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to generate key", err)
	}

	row, err := s.db.APIKey.Create().
		SetUserID(input.UserID).
		SetKeyHash(hash).
		SetKeyPrefix(prefix).
		SetName(input.Name).
		SetStatus(apikey.StatusActive).
		SetPermissions(input.Permissions).
		SetNillableExpiresAt(input.ExpiresAt).
		Save(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to save key", err)
	}

	return &CreateKeyResult{
		ID:        row.ID,
		Name:      row.Name,
		Prefix:    prefix,
		Plaintext: plaintext,
		Status:    string(row.Status),
		CreatedAt: row.CreatedAt.Format(time.RFC3339),
	}, nil
}

// List returns all API keys for a user (without plaintext).
func (s *APIKeyService) List(ctx context.Context, userID int64) ([]*ent.APIKey, error) {
	keys, err := s.db.APIKey.Query().
		Where(apikey.UserID(userID)).
		Order(ent.Desc(apikey.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to list keys", err)
	}
	return keys, nil
}

// Revoke deactivates an API key.
func (s *APIKeyService) Revoke(ctx context.Context, keyID, userID int64) error {
	key, err := s.db.APIKey.Query().
		Where(apikey.ID(keyID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New(errors.ErrNotFound, "key not found")
		}
		return errors.Wrap(errors.ErrInternal, "failed to fetch key", err)
	}
	if key.UserID != userID {
		return errors.New(errors.ErrForbidden, "access denied")
	}
	_, err = s.db.APIKey.UpdateOneID(keyID).
		SetStatus(apikey.StatusRevoked).
		Save(ctx)
	return err
}

// Validate checks a plaintext API key and returns the associated user ID.
func (s *APIKeyService) Validate(ctx context.Context, plaintext string) (int64, error) {
	h := sha256.Sum256([]byte(plaintext))
	hash := hex.EncodeToString(h[:])

	key, err := s.db.APIKey.Query().
		Where(apikey.KeyHash(hash)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, errors.New(errors.ErrUnauthorized, "invalid api key")
		}
		return 0, errors.Wrap(errors.ErrInternal, "failed to validate key", err)
	}
	if key.Status != apikey.StatusActive {
		return 0, errors.New(errors.ErrUnauthorized, "key is revoked")
	}
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return 0, errors.New(errors.ErrUnauthorized, "key expired")
	}
	// Best-effort last-used update.
if _, err := s.db.APIKey.UpdateOneID(key.ID).SetLastUsedAt(time.Now()).Save(ctx); err != nil {
	s.log.Debug("failed to update api key last_used", zap.Int64("key_id", key.ID), zap.Error(err))
}

return key.UserID, nil
}

// GetUsage returns usage statistics for a user.
func (s *APIKeyService) GetUsage(ctx context.Context, userID int64) (*UsageStats, error) {
	// Placeholder: full implementation queries usage_records.
	return &UsageStats{
		TotalRequests: 0,
		TotalImages:   0,
		TotalTokens:   0,
		PeriodStart:   time.Now().AddDate(0, 0, -30).Format("2006-01-02"),
		PeriodEnd:     time.Now().Format("2006-01-02"),
	}, nil
}

// UsageStats holds usage statistics.
type UsageStats struct {
	TotalRequests int    `json:"total_requests"`
	TotalImages   int    `json:"total_images"`
	TotalTokens   int    `json:"total_tokens"`
	PeriodStart   string `json:"period_start"`
	PeriodEnd     string `json:"period_end"`
}
