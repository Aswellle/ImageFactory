package service

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/account"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/repository"
)

// AccountService 提供账号管理与调度选择功能。
//
// 核心职责：
// 1. 账号 CRUD（创建、读取、更新、删除）
// 2. 账号选择：根据优先级和轮转策略选择一个可用账号
// 3. 状态管理：速率限制、过载保护、错误标记
type AccountService struct {
	repo *repository.AccountRepository
}

// NewAccountService 创建 AccountService 实例。
func NewAccountService(repo *repository.AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

// CreateInput 创建账号的输入参数。
type CreateInput struct {
	Name        string
	Platform    string // "gemini", "openai" 等
	Type        string // "api_key", "oauth" 等
	Credentials map[string]any
	Priority    int
}

// Create 创建新账号。
func (s *AccountService) Create(ctx context.Context, input CreateInput) (*ent.Account, error) {
	if input.Name == "" {
		return nil, errors.New(errors.ErrInvalidRequest, "account name is required")
	}
	if input.Platform == "" {
		return nil, errors.New(errors.ErrInvalidRequest, "platform is required")
	}
	if input.Type == "" {
		input.Type = "api_key"
	}
	if input.Priority <= 0 {
		input.Priority = 50
	}
	return s.repo.Create(ctx, input.Name, input.Platform, input.Type, input.Credentials, input.Priority)
}

// GetByID 根据 ID 获取账号。
func (s *AccountService) GetByID(ctx context.Context, id int64) (*ent.Account, error) {
	acc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(errors.ErrNotFound, "account not found", err)
	}
	return acc, nil
}

// List 获取所有账号。
func (s *AccountService) List(ctx context.Context) ([]*ent.Account, error) {
	return s.repo.List(ctx)
}

// Update 更新账号。
func (s *AccountService) Update(ctx context.Context, id int64, name string, credentials map[string]any, priority int) (*ent.Account, error) {
	acc, err := s.repo.Update(ctx, id, func(tx *ent.AccountUpdateOne) *ent.AccountUpdateOne {
		tx.SetName(name).SetCredentials(credentials).SetPriority(priority)
		return tx
	})
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to update account", err)
	}
	return acc, nil
}

// Delete 删除账号。
func (s *AccountService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// SetStatus 设置账号状态。
func (s *AccountService) SetStatus(ctx context.Context, id int64, status string) error {
	acc, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	_, err = s.repo.Update(ctx, id, func(tx *ent.AccountUpdateOne) *ent.AccountUpdateOne {
		tx.SetStatus(accountStatus(status))
		return tx
	})
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to set account status", err)
	}
	_ = acc
	return nil
}

// SetSchedulable 设置账号是否可调度。
func (s *AccountService) SetSchedulable(ctx context.Context, id int64, schedulable bool) error {
	return s.repo.SetSchedulable(ctx, id, schedulable)
}

// SelectAccount 为图像生成请求选择一个可用账号。
//
// 选择策略（简化版，移植自 Sub2API）：
// 1. 过滤：只选择 active 状态、schedulable=true 的账号
// 2. 排除：跳过被速率限制或过载的账号
// 3. 优先级：按 priority 排序（数值越小优先级越高）
// 4. 轮转：同优先级内随机选择，避免总是用同一个账号
func (s *AccountService) SelectAccount(ctx context.Context, platform string) (*ent.Account, error) {
	accounts, err := s.repo.ListByPlatform(ctx, platform)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to list accounts", err)
	}

	// 过滤可用账号
	available := s.filterAvailable(accounts)
	if len(available) == 0 {
		return nil, errors.New(errors.ErrNotFound, "no available account for platform: "+platform)
	}

	// 按优先级分组，选择最高优先级的账号组
	return s.pickByPriority(available), nil
}

// filterAvailable 过滤出当前可用的账号。
func (s *AccountService) filterAvailable(accounts []*ent.Account) []*ent.Account {
	now := time.Now()
	available := make([]*ent.Account, 0, len(accounts))

	for _, acc := range accounts {
		// 跳过被速率限制的账号
		if acc.RateLimitResetAt != nil && now.Before(*acc.RateLimitResetAt) {
			continue
		}
		// 跳过过载的账号
		if acc.OverloadUntil != nil && now.Before(*acc.OverloadUntil) {
			continue
		}
		// 跳过已过期的账号
		if acc.ExpiresAt != nil && now.After(*acc.ExpiresAt) {
			continue
		}
		available = append(available, acc)
	}
	return available
}

// pickByPriority 从可用账号中按优先级选择。
// 选择最高优先级组，然后在该组内随机轮转。
func (s *AccountService) pickByPriority(accounts []*ent.Account) *ent.Account {
	if len(accounts) == 0 {
		return nil
	}
	if len(accounts) == 1 {
		return accounts[0]
	}

	// 找到最高优先级（最小数值）
	bestPriority := accounts[0].Priority
	for _, acc := range accounts {
		if acc.Priority < bestPriority {
			bestPriority = acc.Priority
		}
	}

	// 收集所有最高优先级的账号
	topTier := make([]*ent.Account, 0)
	for _, acc := range accounts {
		if acc.Priority == bestPriority {
			topTier = append(topTier, acc)
		}
	}

	// 同优先级内随机选择（轮转）
	return topTier[rand.IntN(len(topTier))]
}

// MarkRateLimited 标记账号被速率限制。
func (s *AccountService) MarkRateLimited(ctx context.Context, id int64, retryAfter time.Duration) error {
	resetAt := time.Now().Add(retryAfter)
	return s.repo.MarkRateLimited(ctx, id, resetAt)
}

// MarkOverloaded 标记账号过载。
func (s *AccountService) MarkOverloaded(ctx context.Context, id int64, duration time.Duration) error {
	until := time.Now().Add(duration)
	return s.repo.MarkOverloaded(ctx, id, until)
}

// MarkError 标记账号错误。
func (s *AccountService) MarkError(ctx context.Context, id int64, errMsg string) error {
	return s.repo.MarkError(ctx, id, errMsg)
}

// ClearError 清除账号错误状态。
func (s *AccountService) ClearError(ctx context.Context, id int64) error {
	return s.repo.ClearError(ctx, id)
}

// UpdateLastUsed 更新账号最后使用时间。
func (s *AccountService) UpdateLastUsed(ctx context.Context, id int64) error {
	return s.repo.UpdateLastUsed(ctx, id)
}

// GetAPIKey 从账号凭证中提取 API Key。
func (s *AccountService) GetAPIKey(account *ent.Account) string {
	if account == nil || account.Credentials == nil {
		return ""
	}
	if v, ok := account.Credentials["api_key"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// accountStatus 验证并返回有效的账号状态。
func accountStatus(status string) account.Status {
	switch status {
	case "active", "error", "disabled":
		return account.Status(status)
	default:
		return account.StatusActive
	}
}

// Ensure AccountService 实现了相关接口。
var _ = (*AccountService)(nil)
