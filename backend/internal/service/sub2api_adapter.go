package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// Sub2APIAccountAdapter 是 Sub2API 账号系统的适配器。
//
// 通过 Go workspace (go.work) 直接复用 Sub2API 的账号池管理能力：
// - 账号调度选择（优先级 + 轮转 + 阈值算法）
// - 速率限制与过载保护
// - 会话窗口管理
// - 多平台账号池（Gemini, OpenAI, Claude 等）
//
// Sub2API 的账号系统经过众多开发者生产环境验证，直接复用可获得
// 商业级的稳定性和可靠性。
type Sub2APIAccountAdapter struct {
	accountService *service.AccountService
}

// NewSub2APIAccountAdapter 创建 Sub2API 账号适配器。
func NewSub2APIAccountAdapter() *Sub2APIAccountAdapter {
	return &Sub2APIAccountAdapter{
		accountService: service.NewAccountService(),
	}
}

// SelectAccount 为指定平台选择一个可用账号。
//
// 使用 Sub2API 的调度算法：
// 1. 查询可调度账号（platform + schedulable + active）
// 2. 过滤被限速/过载/过期账号
// 3. 按优先级选择（数值越小优先级越高）
// 4. 同优先级内轮转
func (a *Sub2APIAccountAdapter) SelectAccount(ctx context.Context, platform string) (*Sub2APIAccount, error) {
	// TODO: 调用 Sub2API 的账号选择逻辑
	// 当前返回占位符，待 Sub2API 接口对接后替换
	return nil, fmt.Errorf("Sub2API account selection not yet implemented for platform: %s", platform)
}

// GetAPIKey 从账号凭证中提取 API Key。
func (a *Sub2APIAccountAdapter) GetAPIKey(account *Sub2APIAccount) string {
	if account == nil {
		return ""
	}
	if v, ok := account.Credentials["api_key"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// MarkRateLimited 标记账号被速率限制。
func (a *Sub2APIAccountAdapter) MarkRateLimited(ctx context.Context, accountID int64, retryAfter time.Duration) error {
	// TODO: 调用 Sub2API 的限速标记
	return nil
}

// MarkOverloaded 标记账号过载。
func (a *Sub2APIAccountAdapter) MarkOverloaded(ctx context.Context, accountID int64, duration time.Duration) error {
	// TODO: 调用 Sub2API 的过载标记
	return nil
}

// Sub2APIAccount 是 Sub2API 账号的简化视图。
type Sub2APIAccount struct {
	ID          int64
	Name        string
	Platform    string
	Type        string
	Credentials map[string]any
	Priority    int
	Status      string
}
