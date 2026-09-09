package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/account"
	"github.com/imageforge/imageforge/internal/pkg/crypto"
)



// AccountRepository 定义账号数据访问接口。
// 封装 Ent 操作，提供领域语义明确的方法。
type AccountRepository struct {
	db *ent.Client
}

// NewAccountRepository 创建 AccountRepository 实例。
func NewAccountRepository(db *ent.Client) *AccountRepository {
	return &AccountRepository{db: db}
}
// Create 创建新账号。敏感凭证字段在存储前自动加密。
func (r *AccountRepository) Create(ctx context.Context, name, platform, typ string, credentials map[string]any, priority int) (*ent.Account, error) {
	encrypted, err := crypto.EncryptCredentials(credentials)
	if err != nil {
		return nil, fmt.Errorf("encrypt credentials: %w", err)
	}
	return r.db.Account.Create().
		SetName(name).
		SetPlatform(platform).
		SetType(typ).
		SetCredentials(encrypted).
		SetPriority(priority).
		SetStatus("active").
		SetSchedulable(true).
		Save(ctx)
}

// GetByID 根据 ID 获取账号。敏感凭证自动解密。
func (r *AccountRepository) GetByID(ctx context.Context, id int64) (*ent.Account, error) {
	acc, err := r.db.Account.Query().Where(account.ID(id)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return r.decryptAccount(acc)
}

// List 获取所有账号。敏感凭证自动解密。
func (r *AccountRepository) List(ctx context.Context) ([]*ent.Account, error) {
	accounts, err := r.db.Account.Query().
		Order(ent.Asc(account.FieldPriority)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return r.decryptAccounts(accounts)
}

// ListByPlatform 按平台查询可调度账号。敏感凭证自动解密。
func (r *AccountRepository) ListByPlatform(ctx context.Context, platform string) ([]*ent.Account, error) {
	accounts, err := r.db.Account.Query().
		Where(
			account.Platform(platform),
			account.Schedulable(true),
			account.StatusEQ(account.StatusActive),
		).
		Order(ent.Asc(account.FieldPriority)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return r.decryptAccounts(accounts)
}

// ListAvailableByPlatform returns accounts that are ready to accept work.
// 敏感凭证自动解密。
func (r *AccountRepository) ListAvailableByPlatform(ctx context.Context, platform string) ([]*ent.Account, error) {
	now := time.Now()
	accounts, err := r.db.Account.Query().
		Where(
			account.Platform(platform),
			account.Schedulable(true),
			account.StatusEQ(account.StatusActive),
			account.Or(
				account.RateLimitResetAtIsNil(),
				account.RateLimitResetAtLTE(now),
			),
			account.Or(
				account.OverloadUntilIsNil(),
				account.OverloadUntilLTE(now),
			),
			account.Or(
				account.TempUnschedulableUntilIsNil(),
				account.TempUnschedulableUntilLTE(now),
			),
			account.Or(
				account.ExpiresAtIsNil(),
				account.ExpiresAtGT(now),
			),
		).
		Order(ent.Asc(account.FieldPriority)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return r.decryptAccounts(accounts)
}

// decryptAccount 解密单个账号的敏感凭证。
func (r *AccountRepository) decryptAccount(acc *ent.Account) (*ent.Account, error) {
	if acc == nil || len(acc.Credentials) == 0 {
		return acc, nil
	}
	decrypted, err := crypto.DecryptCredentials(acc.Credentials)
	if err != nil {
		return nil, fmt.Errorf("decrypt credentials for account %d: %w", acc.ID, err)
	}
	acc.Credentials = decrypted
	return acc, nil
}

// decryptAccounts 解密多个账号的敏感凭证。
func (r *AccountRepository) decryptAccounts(accounts []*ent.Account) ([]*ent.Account, error) {
	if len(accounts) == 0 {
		return accounts, nil
	}
	for i, acc := range accounts {
		decrypted, err := r.decryptAccount(acc)
		if err != nil {
			return nil, err
		}
		accounts[i] = decrypted
	}
	return accounts, nil
}


// Update 更新账号。
func (r *AccountRepository) Update(ctx context.Context, id int64, updater func(tx *ent.AccountUpdateOne) *ent.AccountUpdateOne) (*ent.Account, error) {
	acc, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return updater(r.db.Account.UpdateOne(acc)).Save(ctx)
}

// Delete 删除账号。
func (r *AccountRepository) Delete(ctx context.Context, id int64) error {
	return r.db.Account.DeleteOneID(id).Exec(ctx)
}

// MarkRateLimited 标记账号被速率限制。
// 使用事务确保状态一致性。
func (r *AccountRepository) MarkRateLimited(ctx context.Context, id int64, resetAt time.Time) error {
	return r.withTx(ctx, func(tx *ent.Tx) error {
		return tx.Account.UpdateOneID(id).
			SetRateLimitedAt(time.Now()).
			SetRateLimitResetAt(resetAt).
			Exec(ctx)
	})
}

// MarkOverloaded 标记账号过载。
// 使用事务确保状态一致性。
func (r *AccountRepository) MarkOverloaded(ctx context.Context, id int64, until time.Time) error {
	return r.withTx(ctx, func(tx *ent.Tx) error {
		return tx.Account.UpdateOneID(id).
			SetOverloadUntil(until).
			Exec(ctx)
	})
}

// MarkError 标记账号错误。
// 使用事务确保状态一致性。
func (r *AccountRepository) MarkError(ctx context.Context, id int64, errMsg string) error {
	return r.withTx(ctx, func(tx *ent.Tx) error {
		return tx.Account.UpdateOneID(id).
			SetStatus(account.StatusError).
			SetErrorMessage(errMsg).
			Exec(ctx)
	})
}

// withTx 在事务中执行操作，失败时自动回滚。
func (r *AccountRepository) withTx(ctx context.Context, fn func(tx *ent.Tx) error) error {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}


// ClearError 清除账号错误状态。
func (r *AccountRepository) ClearError(ctx context.Context, id int64) error {
	_, err := r.db.Account.UpdateOneID(id).
		SetStatus(account.StatusActive).
		SetNillableErrorMessage(nil).
		Save(ctx)
	return err
}

// UpdateLastUsed 更新最后使用时间。
func (r *AccountRepository) UpdateLastUsed(ctx context.Context, id int64) error {
	_, err := r.db.Account.UpdateOneID(id).
		SetLastUsedAt(time.Now()).
		Save(ctx)
	return err
}

// SetSchedulable 设置账号是否可调度。
func (r *AccountRepository) SetSchedulable(ctx context.Context, id int64, schedulable bool) error {
	_, err := r.db.Account.UpdateOneID(id).
		SetSchedulable(schedulable).
		Save(ctx)
	return err
}

// UpdateExtra 更新账号的 extra 字段（合并式更新）。
// 使用事务避免 TOCTOU 竞态条件。
// 移植自 Sub2API 的 account repository UpdateExtra。
func (r *AccountRepository) UpdateExtra(ctx context.Context, id int64, updates map[string]any) error {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	acc, err := tx.Account.Query().Where(account.ID(id)).Only(ctx)
	if err != nil {
		return err
	}
	extra := acc.Extra
	if extra == nil {
		extra = make(map[string]any)
	}
	for k, v := range updates {
		extra[k] = v
	}
	_, err = tx.Account.UpdateOneID(id).
		SetExtra(extra).
		Save(ctx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// UpdateSessionWindow 更新账号的会话窗口状态。
// 使用事务避免 TOCTOU 竞态条件。
// 移植自 Sub2API 的 account repository UpdateSessionWindow。
func (r *AccountRepository) UpdateSessionWindow(ctx context.Context, id int64, start, end *time.Time, utilization float64) error {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	updater := tx.Account.UpdateOneID(id)
	if start != nil {
		updater.SetSessionWindowStart(*start)
	}
	if end != nil {
		updater.SetSessionWindowEnd(*end)
	}
	updates := map[string]any{
		"session_window_utilization": utilization,
	}
	acc, err := tx.Account.Query().Where(account.ID(id)).Only(ctx)
	if err != nil {
		return err
	}
	extra := acc.Extra
	if extra == nil {
		extra = make(map[string]any)
	}
	for k, v := range updates {
		extra[k] = v
	}
	updater.SetExtra(extra)
	_, err = updater.Save(ctx)
	if err != nil {
		return err
	}
	return tx.Commit()
}
