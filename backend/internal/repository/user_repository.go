package repository

import (
	"context"
	"time"

	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/user"
)

// CreateUserParams holds fields for user creation.
type CreateUserParams struct {
	Email        string
	PasswordHash string
	Name         string
	Role         user.Role
}

// UserRepository wraps Ent user operations.
type UserRepository struct {
	db *ent.Client
}

// NewUserRepository builds a UserRepository.
func NewUserRepository(db *ent.Client) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user.
func (r *UserRepository) Create(ctx context.Context, p CreateUserParams) (*ent.User, error) {
	return r.db.User.Create().
		SetEmail(p.Email).
		SetPasswordHash(p.PasswordHash).
		SetName(p.Name).
		SetRole(p.Role).
		SetStatus(user.StatusActive).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)
}

// GetByID fetches a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*ent.User, error) {
	return r.db.User.Query().Where(user.ID(id)).Only(ctx)
}

// GetByEmail fetches a user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*ent.User, error) {
	return r.db.User.Query().Where(user.Email(email)).Only(ctx)
}

// UpdateLastLogin updates the last login timestamp.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id int64) error {
	_, err := r.db.User.UpdateOneID(id).SetLastLoginAt(time.Now()).Save(ctx)
	return err
}

// UpdatePassword updates a user's password hash.
func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, hash string) error {
	_, err := r.db.User.UpdateOneID(id).SetPasswordHash(hash).SetUpdatedAt(time.Now()).Save(ctx)
	return err
}

// IncrementTokenVersion increments the user's token version, invalidating all existing tokens.
func (r *UserRepository) IncrementTokenVersion(ctx context.Context, id int64) error {
	_, err := r.db.User.UpdateOneID(id).AddTokenVersion(1).SetUpdatedAt(time.Now()).Save(ctx)
	return err
}

// EnableTOTP 启用 2FA。
func (r *UserRepository) EnableTOTP(ctx context.Context, id int64, secret string) error {
	_, err := r.db.User.UpdateOneID(id).
		SetTotpSecret(secret).
		SetTotpEnabled(true).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

// DisableTOTP 禁用 2FA。
func (r *UserRepository) DisableTOTP(ctx context.Context, id int64) error {
	_, err := r.db.User.UpdateOneID(id).
		ClearTotpSecret().
		SetTotpEnabled(false).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

// UpdateLoginInfo 更新登录信息（IP、时间、清除失败计数）。
func (r *UserRepository) UpdateLoginInfo(ctx context.Context, id int64, ip string) error {
	_, err := r.db.User.UpdateOneID(id).
		SetLastLoginIP(ip).
		SetLastLoginAt(time.Now()).
		SetFailedLoginCount(0).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

// IncrementFailedLogin 增加失败登录计数。
func (r *UserRepository) IncrementFailedLogin(ctx context.Context, id int64) error {
	_, err := r.db.User.UpdateOneID(id).
		AddFailedLoginCount(1).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

