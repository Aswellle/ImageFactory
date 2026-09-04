package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/internal/batchimage"
)

// AccountResolverInterface 定义账号解析器接口。
//
// 它是账号系统与图像生成管线之间的桥梁：
// - 从账号池中按平台选择可用账号
// - 将 Ent Account 转换为 batchimage.Account 格式
// - 注入真实的 API Key 到生成管线
// - 反馈调度状态（速率限制、过载）
type AccountResolverInterface interface {
	// ResolveAccount 根据平台解析一个可用账号
	ResolveAccount(ctx context.Context, platform string) (*batchimage.Account, error)
	// ResolveAccountByID 根据 ID 解析指定账号
	ResolveAccountByID(ctx context.Context, id int64) (*batchimage.Account, error)
	// MarkRateLimited 标记账号被速率限制
	MarkRateLimited(ctx context.Context, id int64, retryAfter error)
	// MarkOverloaded 标记账号过载
	MarkOverloaded(ctx context.Context, id int64, duration error)
	// UpdateLastUsed 更新账号最后使用时间
	UpdateLastUsed(ctx context.Context, id int64)
}

// AccountResolver 实现了 AccountResolverInterface。
type AccountResolver struct {
	accountSvc *AccountService
}

// NewAccountResolver 创建 AccountResolver 实例。
func NewAccountResolver(accountSvc *AccountService) *AccountResolver {
	return &AccountResolver{accountSvc: accountSvc}
}

// ResolveAccount 根据平台解析一个可用账号。
//
// 这是调度器的核心入口：
// 1. 按平台查询所有可调度账号
// 2. 过滤掉被限速/过载/过期的账号
// 3. 按优先级选择最佳账号
// 4. 返回可用于 API 调用的 batchimage.Account
func (r *AccountResolver) ResolveAccount(ctx context.Context, platform string) (*batchimage.Account, error) {
	acc, err := r.accountSvc.SelectAccount(ctx, platform)
	if err != nil {
		return nil, fmt.Errorf("select account for platform %s: %w", platform, err)
	}
	return r.toBatchImageAccount(acc), nil
}

// ResolveAccountByID 根据 ID 解析指定账号。
func (r *AccountResolver) ResolveAccountByID(ctx context.Context, id int64) (*batchimage.Account, error) {
	acc, err := r.accountSvc.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get account %d: %w", id, err)
	}
	return r.toBatchImageAccount(acc), nil
}

// MarkRateLimited 标记账号被速率限制。
func (r *AccountResolver) MarkRateLimited(ctx context.Context, id int64, retryAfter error) {
	// 默认 60 秒后重试
	_ = retryAfter
	if err := r.accountSvc.MarkRateLimited(ctx, id, 60); err != nil {
		// 标记失败不影响主流程
	}
}

// MarkOverloaded 标记账号过载。
func (r *AccountResolver) MarkOverloaded(ctx context.Context, id int64, duration error) {
	// 默认过载 5 分钟
	_ = duration
	if err := r.accountSvc.MarkOverloaded(ctx, id, 5); err != nil {
		// 标记失败不影响主流程
	}
}

// UpdateLastUsed 更新账号最后使用时间。
func (r *AccountResolver) UpdateLastUsed(ctx context.Context, id int64) {
	if err := r.accountSvc.UpdateLastUsed(ctx, id); err != nil {
		// 更新失败不影响主流程
	}
}

// toBatchImageAccount 将 Ent Account 转换为 batchimage.Account。
// 转换 Credentials 从 map[string]any 到 map[string]string。
func (r *AccountResolver) toBatchImageAccount(acc *ent.Account) *batchimage.Account {
	if acc == nil {
		return nil
	}
	// 转换凭证格式
	creds := make(map[string]string, len(acc.Credentials))
	for k, v := range acc.Credentials {
		creds[k] = fmt.Sprintf("%v", v)
	}
	return &batchimage.Account{
		ID:          acc.ID,
		Platform:    acc.Platform,
		Credentials: creds,
	}
}

// formatInt 将 int64 格式化为字符串。
func formatInt(n int64) string {
	return strconv.FormatInt(n, 10)
}

// 确保 AccountResolver 实现了接口。
var _ AccountResolverInterface = (*AccountResolver)(nil)
