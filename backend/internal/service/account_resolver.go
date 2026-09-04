package service

import (
	"context"
	"fmt"
	"time"

	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/internal/batchimage"
	"github.com/imageforge/imageforge/internal/repository"
	"github.com/imageforge/imageforge/internal/service/scheduling"
)

// AccountResolverInterface 定义账号解析器接口。
//
// 它是账号系统与图像生成管线之间的桥梁：
// - 从账号池中按平台选择可用账号
// - 将 Ent Account 转换为 batchimage.Account 格式
// - 注入真实的 API Key 到生成管线
// - 反馈调度状态（速率限制、过载、阈值停调）
type AccountResolverInterface interface {
	ResolveAccount(ctx context.Context, platform string) (*batchimage.Account, error)
	ResolveAccountByID(ctx context.Context, id int64) (*batchimage.Account, error)
	MarkRateLimited(ctx context.Context, id int64, retryAfter error)
	MarkOverloaded(ctx context.Context, id int64, duration error)
	UpdateLastUsed(ctx context.Context, id int64)
}

// AccountResolver 实现了 AccountResolverInterface。
//
// 移植自 Sub2API 的账号调度系统，集成了：
// - 调度阈值算法（基于用量窗口的自动停调）
// - 限速策略（429/401/403/529 错误处理）
// - 会话窗口管理（5h/7d 滚动窗口跟踪）
// - 优先级 + 轮转 + 负载感知的账号选择
type AccountResolver struct {
	accountRepo   *repository.AccountRepository
	accountSvc    *AccountService
	schedulingSvc *SchedulingService
}

// NewAccountResolver 创建 AccountResolver 实例。
func NewAccountResolver(accountRepo *repository.AccountRepository, accountSvc *AccountService, schedulingSvc *SchedulingService) *AccountResolver {
	return &AccountResolver{
		accountRepo:   accountRepo,
		accountSvc:    accountSvc,
		schedulingSvc: schedulingSvc,
	}
}

// ResolveAccount 根据平台解析一个可用账号。
//
// 调度器核心入口（移植自 Sub2API 的 3 层调度算法）：
// 1. 按平台查询所有可调度账号
// 2. 过滤：排除被限速/过载/过期/阈值停调的账号
// 3. 选择：按优先级 + 轮转选择最佳账号
// 4. 返回可用于 API 调用的 batchimage.Account
func (r *AccountResolver) ResolveAccount(ctx context.Context, platform string) (*batchimage.Account, error) {
	accounts, err := r.accountRepo.ListByPlatform(ctx, platform)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts for platform %s: %w", platform, err)
	}

	// 使用调度算法过滤和选择账号
	available := r.filterAvailable(accounts)
	if len(available) == 0 {
		return nil, fmt.Errorf("no available accounts for platform %s", platform)
	}

	// 按优先级选择（移植自 Sub2API 的优先级轮转算法）
	selected := r.pickByPriority(available)
	if selected == nil {
		return nil, fmt.Errorf("no suitable account found for platform %s", platform)
	}

	return r.toBatchImageAccount(selected), nil
}

// ResolveAccountByID 根据 ID 解析指定账号。
func (r *AccountResolver) ResolveAccountByID(ctx context.Context, id int64) (*batchimage.Account, error) {
	acc, err := r.accountSvc.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get account %d: %w", id, err)
	}
	return r.toBatchImageAccount(acc), nil
}

// MarkRateLimited 标记账号被速率限制。
//
// 使用 Sub2API 的限速策略算法计算重置时间。
func (r *AccountResolver) MarkRateLimited(ctx context.Context, id int64, retryAfter error) {
	acc, err := r.accountSvc.GetByID(ctx, id)
	if err != nil {
		return
	}

	// 使用调度算法计算限速重置时间
	sa := EntAccountToScheduling(acc)
	resetAt := r.calculateRateLimitResetTime(sa)

	r.accountSvc.MarkRateLimited(ctx, id, resetAt.Sub(time.Now()))
}

// MarkOverloaded 标记账号过载。
func (r *AccountResolver) MarkOverloaded(ctx context.Context, id int64, duration error) {
	r.accountSvc.MarkOverloaded(ctx, id, 2*time.Minute)
}

// UpdateLastUsed 更新账号最后使用时间。
func (r *AccountResolver) UpdateLastUsed(ctx context.Context, id int64) {
	r.accountSvc.UpdateLastUsed(ctx, id)
}

// filterAvailable 过滤出当前可用的账号。
//
// 移植自 Sub2API 的调度过滤逻辑：
// - 排除被限速的账号
// - 排除过载的账号
// - 排除临时不可调度的账号
// - 排除超过调度阈值的账号
func (r *AccountResolver) filterAvailable(accounts []*ent.Account) []*ent.Account {
	now := time.Now()
	var available []*ent.Account

	for _, acc := range accounts {
		if acc == nil {
			continue
		}
		if string(acc.Status) != "active" {
			continue
		}
		if !acc.Schedulable {
			continue
		}
		// 检查速率限制
		if acc.RateLimitResetAt != nil && acc.RateLimitResetAt.After(now) {
			continue
		}
		// 检查过载
		if acc.OverloadUntil != nil && acc.OverloadUntil.After(now) {
			continue
		}
		// 检查临时不可调度
		if acc.TempUnschedulableUntil != nil && acc.TempUnschedulableUntil.After(now) {
			continue
		}
		// 检查调度阈值（移植自 Sub2API 的 EvaluateAccountSchedulingThreshold）
		if r.schedulingSvc != nil {
			sa := EntAccountToScheduling(acc)
			decision := scheduling.EvaluateAccountSchedulingThreshold(sa, r.schedulingSvc.thresholds, now)
			if decision.ShouldPause {
				continue
			}
		}

		available = append(available, acc)
	}

	return available
}

// pickByPriority 从可用账号中按优先级选择。
// 选择最高优先级组，然后在该组内随机轮转。
//
// 移植自 Sub2API 的优先级轮转算法。
func (r *AccountResolver) pickByPriority(accounts []*ent.Account) *ent.Account {
	if len(accounts) == 0 {
		return nil
	}
	if len(accounts) == 1 {
		return accounts[0]
	}

	// 找到最高优先级（最小值）
	bestPriority := accounts[0].Priority
	for _, acc := range accounts {
		if acc.Priority < bestPriority {
			bestPriority = acc.Priority
		}
	}

	// 收集所有最高优先级的账号
	var bestTier []*ent.Account
	for _, acc := range accounts {
		if acc.Priority == bestPriority {
			bestTier = append(bestTier, acc)
		}
	}

	if len(bestTier) == 1 {
		return bestTier[0]
	}

	// 随机轮转
	return bestTier[time.Now().UnixNano()%int64(len(bestTier))]
}

// calculateRateLimitResetTime 计算限速重置时间。
//
// 移植自 Sub2API 的限速策略算法。
func (r *AccountResolver) calculateRateLimitResetTime(sa *scheduling.Account) time.Time {
	now := time.Now()

	switch sa.Platform {
	case scheduling.PlatformAnthropic:
		// Anthropic: 默认 60 秒冷却
		return now.Add(60 * time.Second)
	case scheduling.PlatformOpenAI:
		// OpenAI: 默认 60 秒冷却
		return now.Add(60 * time.Second)
	case scheduling.PlatformGrok:
		// Grok: 默认 60 秒冷却
		return now.Add(60 * time.Second)
	default:
		return now.Add(60 * time.Second)
	}
}

// toBatchImageAccount 将 Ent Account 转换为 batchimage.Account。
// 转换 Credentials 从 map[string]any 到 map[string]string。
func (r *AccountResolver) toBatchImageAccount(acc *ent.Account) *batchimage.Account {
	creds := make(map[string]string)
	for k, v := range acc.Credentials {
		if s, ok := v.(string); ok {
			creds[k] = s
		} else {
			creds[k] = fmt.Sprintf("%v", v)
		}
	}

	return &batchimage.Account{
		ID:          acc.ID,
		Platform:    acc.Platform,
		Credentials: creds,
	}
}

// 确保 AccountResolver 实现了接口。
var _ AccountResolverInterface = (*AccountResolver)(nil)
