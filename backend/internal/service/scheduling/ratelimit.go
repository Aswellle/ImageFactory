package scheduling

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// RateLimitStrategy handles upstream error responses and determines account state changes.
//
// This is the algorithmic core ported from Sub2API's ratelimit_service.go,
// extracted as pure functions that operate on the scheduling.Account type.
// The actual DB persistence is left to the caller via the AccountStateHandler interface.
//
// Sub2API reference: internal/service/ratelimit_service.go

// AccountStateHandler is the interface for persisting account state changes.
type AccountStateHandler interface {
	UpdateRateLimit(ctx context.Context, accountID int64, resetAt time.Time) error
	ClearRateLimit(ctx context.Context, accountID int64) error
	SetOverloaded(ctx context.Context, accountID int64, until time.Time) error
	SetError(ctx context.Context, accountID int64, errMsg string) error
	SetTempUnschedulable(ctx context.Context, accountID int64, until time.Time, reason string) error
	UpdateExtra(ctx context.Context, accountID int64, updates map[string]any) error
	UpdateSessionWindow(ctx context.Context, accountID int64, start, end *time.Time, utilization float64) error
}

// RateLimitStrategy provides the algorithmic core for handling upstream errors.
type RateLimitStrategy struct {
	handler AccountStateHandler
}

// NewRateLimitStrategy creates a new RateLimitStrategy.
func NewRateLimitStrategy(handler AccountStateHandler) *RateLimitStrategy {
	return &RateLimitStrategy{handler: handler}
}

// HandleUpstreamErrorResult is the result of handling an upstream error.
type HandleUpstreamErrorResult struct {
	ShouldDisable   bool
	ShouldRateLimit bool
	RateLimitReset  time.Time
	IsAuthError     bool
	ErrorMessage    string
}

// HandleUpstreamError processes an upstream error response and determines
// what state changes should happen to the account.
func (s *RateLimitStrategy) HandleUpstreamError(ctx context.Context, account *Account, statusCode int, headers http.Header, responseBody []byte) HandleUpstreamErrorResult {
	result := HandleUpstreamErrorResult{}

	if account == nil {
		return result
	}

	switch {
	case statusCode == 401:
		result.IsAuthError = true
		result.ShouldDisable = true
		result.ErrorMessage = "authentication failed (401)"
		if s.handler != nil {
			s.handler.SetError(ctx, account.ID, result.ErrorMessage)
		}

	case statusCode == 403:
		result.IsAuthError = true
		msg := extractForbiddenMessage(responseBody)
		result.ErrorMessage = fmt.Sprintf("forbidden (403): %s", msg)
		if s.handler != nil {
			s.handler.SetError(ctx, account.ID, result.ErrorMessage)
		}

	case statusCode == 429:
		resetAt := s.calculate429ResetTime(account, headers, responseBody)
		result.ShouldRateLimit = true
		result.RateLimitReset = resetAt
		result.ErrorMessage = fmt.Sprintf("rate limited (429), reset at %s", resetAt.Format(time.RFC3339))
		if s.handler != nil {
			s.handler.UpdateRateLimit(ctx, account.ID, resetAt)
		}

	case statusCode == 529:
		overloadUntil := time.Now().Add(2 * time.Minute)
		result.ErrorMessage = fmt.Sprintf("overloaded (529), cooldown until %s", overloadUntil.Format(time.RFC3339))
		if s.handler != nil {
			s.handler.SetOverloaded(ctx, account.ID, overloadUntil)
		}

	case statusCode >= 500:
		cooldown := time.Now().Add(30 * time.Second)
		result.ErrorMessage = fmt.Sprintf("server error (%d), temp cooldown", statusCode)
		if s.handler != nil {
			s.handler.SetTempUnschedulable(ctx, account.ID, cooldown, fmt.Sprintf("upstream_%d", statusCode))
		}
	}

	return result
}

// calculate429ResetTime determines the appropriate reset time for a 429 response.
func (s *RateLimitStrategy) calculate429ResetTime(account *Account, headers http.Header, responseBody []byte) time.Time {
	now := nowFunc()

	switch account.Platform {
	case PlatformAnthropic:
		if result := CalculateAnthropic429ResetTime(headers, now); result != nil {
			return result.ResetAt
		}
		if reset, ok := parseAnthropicAggregateReset(headers, now); ok {
			return reset
		}

	case PlatformOpenAI:
		if ts := ParseOpenAIRateLimitResetTime(responseBody); ts != nil {
			return *ts
		}
		if ts := CalculateOpenAI429ResetTime(headers, now); ts != nil {
			return *ts
		}

	case PlatformGrok:
		if ts := ParseRetryAfterResetTime(headers, now); ts != nil {
			return *ts
		}
	}

	// Default: extend existing rate limit or use 60s cooldown
	if account.RateLimitResetAt != nil && account.RateLimitResetAt.After(now) {
		return account.RateLimitResetAt.Add(60 * time.Second)
	}
	return now.Add(60 * time.Second)
}

// HandleAnthropic429 specifically handles Anthropic 429 responses.
func (s *RateLimitStrategy) HandleAnthropic429(ctx context.Context, account *Account, headers http.Header) HandleUpstreamErrorResult {
	result := HandleUpstreamErrorResult{ShouldRateLimit: true}

	now := nowFunc()

	// Check for Fable 7d_oi window first (model-specific)
	if limit := selectAnthropicFableWindowLimit(headers, now); limit != nil {
		result.RateLimitReset = limit.resetAt
		result.ErrorMessage = fmt.Sprintf("Anthropic 7d_oi window exhausted, reset at %s", limit.resetAt.Format(time.RFC3339))
		result.ShouldRateLimit = false
		if s.handler != nil {
			s.handler.UpdateExtra(ctx, account.ID, map[string]any{
				"fable_7d_oi_reset_at": limit.resetAt.Format(time.RFC3339),
			})
		}
		return result
	}

	// Check 5h and 7d windows
	if limit := selectAnthropicExhaustedWindow(headers, now); limit != nil {
		result.RateLimitReset = limit.resetAt
		result.ErrorMessage = fmt.Sprintf("Anthropic %s window exhausted, reset at %s", limit.window, limit.resetAt.Format(time.RFC3339))
		if s.handler != nil {
			s.handler.UpdateRateLimit(ctx, account.ID, limit.resetAt)
			s.handler.UpdateExtra(ctx, account.ID, map[string]any{
				"session_window_utilization": limit.utilization,
				"session_window_status":      "rejected",
			})
			s.handler.UpdateSessionWindow(ctx, account.ID, nil, &limit.resetAt, limit.utilization)
		}
		return result
	}

	// Fall back to aggregated reset
	if reset, ok := parseAnthropicAggregateReset(headers, now); ok {
		result.RateLimitReset = reset
		result.ErrorMessage = fmt.Sprintf("Anthropic rate limited, reset at %s", reset.Format(time.RFC3339))
		if s.handler != nil {
			s.handler.UpdateRateLimit(ctx, account.ID, reset)
		}
		return result
	}

	result.RateLimitReset = now.Add(60 * time.Second)
	result.ErrorMessage = "Anthropic rate limited (no reset header), 60s cooldown"
	if s.handler != nil {
		s.handler.UpdateRateLimit(ctx, account.ID, result.RateLimitReset)
	}
	return result
}

// anthropicWindowLimit mirrors Sub2API's anthropicWindowLimit struct.
type anthropicWindowLimit struct {
	window      string
	resetAt     time.Time
	utilization float64
	status      string
}

// selectAnthropicExhaustedWindow finds the most restrictive exhausted window.
func selectAnthropicExhaustedWindow(headers http.Header, now time.Time) *anthropicWindowLimit {
	if headers == nil {
		return nil
	}

	var limits []*anthropicWindowLimit
	for _, window := range []string{"5h", "7d"} {
		if isAnthropicWindowExceeded(headers, window) {
			if reset, ok := parseAnthropicWindowReset(headers, window, now); ok {
				util := schedulingPercentValue(headers.Get("anthropic-ratelimit-unified-" + window + "-utilization"))
				status := headers.Get("anthropic-ratelimit-unified-" + window + "-status")
				limits = append(limits, &anthropicWindowLimit{
					window:      window,
					resetAt:     reset,
					utilization: util,
					status:      status,
				})
			}
		}
	}

	if len(limits) == 0 {
		return nil
	}

	var winner *anthropicWindowLimit
	for _, l := range limits {
		if winner == nil || l.resetAt.After(winner.resetAt) {
			winner = l
		}
	}
	return winner
}

// selectAnthropicFableWindowLimit parses the Anthropic 7d_oi per-model window headers.
func selectAnthropicFableWindowLimit(headers http.Header, now time.Time) *anthropicWindowLimit {
	if headers == nil {
		return nil
	}

	statusHeader := headers.Get("anthropic-ratelimit-unified-7d_oi-status")
	utilHeader := headers.Get("anthropic-ratelimit-unified-7d_oi-utilization")

	isRejected := strings.EqualFold(strings.TrimSpace(statusHeader), "rejected")
	util := schedulingPercentValue(utilHeader)
	if !isRejected && util < 1.0 {
		return nil
	}

	reset, ok := parseAnthropicResetTimestamp(headers.Get("anthropic-ratelimit-unified-7d_oi-reset"), now, 8*24*time.Hour)
	if !ok {
		reset, ok = parseAnthropicAggregateReset(headers, now)
		if !ok {
			return nil
		}
	}

	return &anthropicWindowLimit{
		window:      "7d_oi",
		resetAt:     reset,
		utilization: util * 100,
		status:      statusHeader,
	}
}

// UpdateSessionWindow updates the account's session window based on response headers.
func (s *RateLimitStrategy) UpdateSessionWindow(ctx context.Context, account *Account, headers http.Header) {
	if account == nil || headers == nil || s.handler == nil {
		return
	}

	now := nowFunc()

	switch account.Platform {
	case PlatformAnthropic:
		usage := SampleAnthropicPassiveUsage(headers, now)
		if !usage.HasData {
			return
		}

		var windowEnd *time.Time
		if usage.Window5hResetAt != nil {
			windowEnd = usage.Window5hResetAt
		}

		utilization := usage.Window5hUtilization
		if usage.Window7dUtilization > utilization {
			utilization = usage.Window7dUtilization
		}

		s.handler.UpdateSessionWindow(ctx, account.ID, nil, windowEnd, utilization)

		updates := map[string]any{}
		if usage.Window5hUtilization > 0 {
			updates["session_window_utilization"] = usage.Window5hUtilization
		}
		if usage.Window5hResetAt != nil {
			updates["session_window_reset"] = usage.Window5hResetAt.Format(time.RFC3339)
		}
		if usage.Window7dUtilization > 0 {
			updates["passive_usage_7d_utilization"] = usage.Window7dUtilization
		}
		if usage.Window7dResetAt != nil {
			updates["passive_usage_7d_reset"] = usage.Window7dResetAt.Format(time.RFC3339)
		}
		if len(updates) > 0 {
			s.handler.UpdateExtra(ctx, account.ID, updates)
		}

	case PlatformOpenAI:
		s.updateOpenAISessionWindow(ctx, account, headers, now)
	}
}

func (s *RateLimitStrategy) updateOpenAISessionWindow(ctx context.Context, account *Account, headers http.Header, now time.Time) {
	updates := map[string]any{}

	if v := headers.Get("x-ratelimit-usage-requests"); v != "" {
		updates["openai_usage_requests"] = v
	}
	if v := headers.Get("x-ratelimit-usage-tokens"); v != "" {
		updates["openai_usage_tokens"] = v
	}
	if v := headers.Get("x-ratelimit-reset-usage"); v != "" {
		updates["openai_reset_usage"] = v
	}

	if len(updates) > 0 {
		updates["codex_usage_updated_at"] = now.Format(time.RFC3339)
		s.handler.UpdateExtra(ctx, account.ID, updates)
	}
}

// ClearRateLimit clears the account's rate limit state.
func (s *RateLimitStrategy) ClearRateLimit(ctx context.Context, accountID int64) error {
	if s.handler == nil {
		return fmt.Errorf("no state handler configured")
	}
	return s.handler.ClearRateLimit(ctx, accountID)
}

// HandleOpenAIImageRateLimit handles OpenAI image generation rate limits.
func (s *RateLimitStrategy) HandleOpenAIImageRateLimit(ctx context.Context, account *Account, statusCode int, headers http.Header, responseBody []byte) bool {
	if account == nil || statusCode != 429 {
		return false
	}

	if !IsOpenAIImageRateLimitError(statusCode, responseBody) {
		return false
	}

	now := nowFunc()

	cooldown := ParseOpenAIImageTryAgainCooldown(responseBody)
	if cooldown > 0 {
		resetAt := now.Add(cooldown)
		if s.handler != nil {
			s.handler.UpdateRateLimit(ctx, account.ID, resetAt)
		}
		return true
	}

	if ts := ParseRetryAfterResetTime(headers, now); ts != nil {
		if s.handler != nil {
			s.handler.UpdateRateLimit(ctx, account.ID, *ts)
		}
		return true
	}

	resetAt := now.Add(5 * time.Minute)
	if s.handler != nil {
		s.handler.UpdateRateLimit(ctx, account.ID, resetAt)
	}
	return true
}

// extractForbiddenMessage extracts a human-readable message from a 403 response body.
func extractForbiddenMessage(body []byte) string {
	if len(body) == 0 {
		return "no details"
	}
	var parsed struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error.Message != "" {
		return parsed.Error.Message
	}
	msg := string(body)
	if len(msg) > 200 {
		msg = msg[:200] + "..."
	}
	return msg
}

// nowFunc is a variable so it can be overridden in tests.
var nowFunc = time.Now
