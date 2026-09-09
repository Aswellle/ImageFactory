package scheduling

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// mockStateHandler implements AccountStateHandler for testing.
type mockStateHandler struct {
	rateLimits   []time.Time
	cleared      bool
	overloads    []time.Time
	errors       []string
	tempUnscheds []struct {
		until  time.Time
		reason string
	}
	extraUpdates  []map[string]any
	windowUpdates []struct {
		start, end *time.Time
		util       float64
	}
}

func (m *mockStateHandler) UpdateRateLimit(_ context.Context, _ int64, resetAt time.Time) error {
	m.rateLimits = append(m.rateLimits, resetAt)
	return nil
}

func (m *mockStateHandler) ClearRateLimit(_ context.Context, _ int64) error {
	m.cleared = true
	return nil
}

func (m *mockStateHandler) SetOverloaded(_ context.Context, _ int64, until time.Time) error {
	m.overloads = append(m.overloads, until)
	return nil
}

func (m *mockStateHandler) SetError(_ context.Context, _ int64, errMsg string) error {
	m.errors = append(m.errors, errMsg)
	return nil
}

func (m *mockStateHandler) SetTempUnschedulable(_ context.Context, _ int64, until time.Time, reason string) error {
	m.tempUnscheds = append(m.tempUnscheds, struct {
		until  time.Time
		reason string
	}{until, reason})
	return nil
}

func (m *mockStateHandler) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	m.extraUpdates = append(m.extraUpdates, updates)
	return nil
}

func (m *mockStateHandler) UpdateSessionWindow(_ context.Context, _ int64, start, end *time.Time, util float64) error {
	m.windowUpdates = append(m.windowUpdates, struct {
		start, end *time.Time
		util       float64
	}{start, end, util})
	return nil
}

func TestPoolSelectAvailable(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(time.Hour)
	pool := NewPool()

	active := NewAccount(1, "active", PlatformOpenAI, "oauth", nil, nil, 50, StatusActive, nil, nil, nil, nil, nil, nil, "")
	rateLimited := NewAccount(2, "limited", PlatformOpenAI, "oauth", nil, nil, 50, StatusActive, nil, &futureTime, nil, nil, nil, nil, "")
	overloaded := NewAccount(3, "overloaded", PlatformOpenAI, "oauth", nil, nil, 50, StatusActive, nil, nil, &futureTime, nil, nil, nil, "")
	errored := NewAccount(4, "errored", PlatformOpenAI, "oauth", nil, nil, 50, StatusError, nil, nil, nil, nil, nil, nil, "")

	pool.AddAccount(active)
	pool.AddAccount(rateLimited)
	pool.AddAccount(overloaded)
	pool.AddAccount(errored)

	available := pool.SelectAvailable(now)
	if len(available) != 1 {
		t.Errorf("expected 1 available account, got %d", len(available))
	}
	if len(available) > 0 && available[0].ID != 1 {
		t.Errorf("expected account ID 1, got %d", available[0].ID)
	}
}

func TestPoolSelectByPriority(t *testing.T) {
	now := time.Now()
	pool := NewPool()

	lowPriority := NewAccount(1, "low", PlatformOpenAI, "oauth", nil, nil, 90, StatusActive, nil, nil, nil, nil, nil, nil, "")
	midPriority := NewAccount(2, "mid", PlatformOpenAI, "oauth", nil, nil, 50, StatusActive, nil, nil, nil, nil, nil, nil, "")
	highPriority := NewAccount(3, "high", PlatformOpenAI, "oauth", nil, nil, 10, StatusActive, nil, nil, nil, nil, nil, nil, "")

	pool.AddAccount(lowPriority)
	pool.AddAccount(midPriority)
	pool.AddAccount(highPriority)

	selected := pool.SelectByPriority(now)
	if selected == nil {
		t.Fatal("expected non-nil selection")
	}
	if selected.Priority != 10 {
		t.Errorf("expected highest priority (10), got %d", selected.Priority)
	}
}

func TestPoolSelectByPriority_AllSame(t *testing.T) {
	now := time.Now()
	pool := NewPool()

	for i := 1; i <= 5; i++ {
		pool.AddAccount(NewAccount(int64(i), "test", PlatformOpenAI, "oauth", nil, nil, 50, StatusActive, nil, nil, nil, nil, nil, nil, ""))
	}

	selected := pool.SelectByPriority(now)
	if selected == nil {
		t.Fatal("expected non-nil selection")
	}
	if selected.ID < 1 || selected.ID > 5 {
		t.Errorf("expected ID 1-5, got %d", selected.ID)
	}
}

func TestPoolSelectWithThreshold(t *testing.T) {
	now := time.Now()
	resetAt := now.Add(time.Hour)

	pool := NewPool()

	thresholdExceeded := NewAccount(1, "exceed", PlatformOpenAI, "oauth",
		map[string]any{"chatgpt_account_id": "acc_1"},
		map[string]any{
			"codex_5h_used_percent": 90,
			"codex_5h_reset_at":     resetAt.Format(time.RFC3339),
			"chatgpt_account_id":    "acc_1",
		},
		50, StatusActive, nil, nil, nil, nil, nil, nil, "")

	normal := NewAccount(2, "normal", PlatformOpenAI, "oauth",
		map[string]any{"chatgpt_account_id": "acc_2"},
		map[string]any{
			"codex_5h_used_percent": 50,
			"codex_5h_reset_at":     resetAt.Format(time.RFC3339),
			"chatgpt_account_id":    "acc_2",
		},
		50, StatusActive, nil, nil, nil, nil, nil, nil, "")

	pool.AddAccount(thresholdExceeded)
	pool.AddAccount(normal)

	thresholds := map[string]int{PlatformOpenAI: 80}
	selected := pool.SelectWithThreshold(now, thresholds)
	if selected == nil {
		t.Fatal("expected non-nil selection")
	}
	if selected.ID != 2 {
		t.Errorf("expected account 2 (within threshold), got %d", selected.ID)
	}
}

func TestPoolSelectLoadAware(t *testing.T) {
	now := time.Now()
	pool := NewPool()

	for i := 1; i <= 5; i++ {
		pool.AddAccount(NewAccount(int64(i), "test", PlatformOpenAI, "oauth", nil, nil, 50, StatusActive, nil, nil, nil, nil, nil, nil, ""))
	}

	weights := DefaultScoreWeights()
	selected := pool.SelectLoadAware(weights, nil, now)
	if selected == nil {
		t.Fatal("expected non-nil selection")
	}
	if selected.ID < 1 || selected.ID > 5 {
		t.Errorf("expected ID 1-5, got %d", selected.ID)
	}
}

func TestPoolSelectLoadAware_StickySession(t *testing.T) {
	now := time.Now()
	pool := NewPool()

	for i := 1; i <= 5; i++ {
		pool.AddAccount(NewAccount(int64(i), "test", PlatformOpenAI, "oauth", nil, nil, 50, StatusActive, nil, nil, nil, nil, nil, nil, ""))
	}

	weights := DefaultScoreWeights()
	stickyID := int64(3)
	selected := pool.SelectLoadAware(weights, &stickyID, now)
	if selected == nil {
		t.Fatal("expected non-nil selection")
	}
	if selected.ID != 3 {
		t.Errorf("expected sticky account 3, got %d", selected.ID)
	}
}

func TestSelectTopK(t *testing.T) {
	candidates := []CandidateScore{
		{Account: &Account{ID: 1, Priority: 50}, Score: 0.9},
		{Account: &Account{ID: 2, Priority: 50}, Score: 0.7},
		{Account: &Account{ID: 3, Priority: 50}, Score: 0.95},
		{Account: &Account{ID: 4, Priority: 50}, Score: 0.5},
		{Account: &Account{ID: 5, Priority: 50}, Score: 0.85},
	}

	top3 := SelectTopK(candidates, 3)
	if len(top3) != 3 {
		t.Errorf("expected 3 candidates, got %d", len(top3))
	}

	if top3[0].Score < top3[1].Score || top3[1].Score < top3[2].Score {
		t.Error("top-K should be in descending score order")
	}

	if top3[0].Account.ID != 3 {
		t.Errorf("expected highest scoring account (ID 3), got %d", top3[0].Account.ID)
	}
}

func TestSelectTopK_LessThanK(t *testing.T) {
	candidates := []CandidateScore{
		{Account: &Account{ID: 1, Priority: 50}, Score: 0.5},
	}

	result := SelectTopK(candidates, 3)
	if len(result) != 1 {
		t.Errorf("expected 1 candidate, got %d", len(result))
	}
}

func TestScoreCandidates(t *testing.T) {
	accounts := []*Account{
		NewAccount(1, "good", PlatformOpenAI, "oauth", nil, nil, 10, StatusActive, nil, nil, nil, nil, nil, nil, ""),
		NewAccount(2, "bad", PlatformOpenAI, "oauth", nil, nil, 90, StatusError, nil, nil, nil, nil, nil, nil, ""),
	}

	weights := DefaultScoreWeights()
	candidates := ScoreCandidates(accounts, weights, nil)

	if len(candidates) != 2 {
		t.Errorf("expected 2 candidates, got %d", len(candidates))
	}

	var score1, score2 float64
	for _, c := range candidates {
		if c.Account.ID == 1 {
			score1 = c.Score
		}
		if c.Account.ID == 2 {
			score2 = c.Score
		}
	}

	if score1 <= score2 {
		t.Errorf("account 1 (active, priority 10) should score higher than account 2 (error, priority 90): %f vs %f", score1, score2)
	}
}

func TestScoreCandidates_StickyBonus(t *testing.T) {
	accounts := []*Account{
		NewAccount(1, "other", PlatformOpenAI, "oauth", nil, nil, 10, StatusActive, nil, nil, nil, nil, nil, nil, ""),
		NewAccount(2, "sticky", PlatformOpenAI, "oauth", nil, nil, 10, StatusActive, nil, nil, nil, nil, nil, nil, ""),
	}

	stickyID := int64(2)
	weights := DefaultScoreWeights()
	candidates := ScoreCandidates(accounts, weights, &stickyID)

	var score1, score2 float64
	for _, c := range candidates {
		if c.Account.ID == 1 {
			score1 = c.Score
		}
		if c.Account.ID == 2 {
			score2 = c.Score
		}
	}

	if score2 <= score1 {
		t.Errorf("sticky account should score higher: %f vs %f", score2, score1)
	}
}

func TestHandleUpstreamError_401(t *testing.T) {
	handler := &mockStateHandler{}
	strategy := NewRateLimitStrategy(handler)
	acc := NewAccount(1, "test", PlatformOpenAI, "oauth", nil, nil, 50, StatusActive, nil, nil, nil, nil, nil, nil, "")

	result := strategy.HandleUpstreamError(context.Background(), acc, 401, nil, nil)
	if !result.IsAuthError {
		t.Error("401 should be auth error")
	}
	if !result.ShouldDisable {
		t.Error("401 should disable account")
	}
	if len(handler.errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(handler.errors))
	}
}

func TestHandleUpstreamError_429(t *testing.T) {
	handler := &mockStateHandler{}
	strategy := NewRateLimitStrategy(handler)
	acc := NewAccount(1, "test", PlatformOpenAI, "oauth", nil, nil, 50, StatusActive, nil, nil, nil, nil, nil, nil, "")

	headers := http.Header{}
	headers.Set("retry-after", "120")

	result := strategy.HandleUpstreamError(context.Background(), acc, 429, headers, nil)
	if !result.ShouldRateLimit {
		t.Error("429 should trigger rate limit")
	}
	if result.RateLimitReset.IsZero() {
		t.Error("429 should have reset time")
	}
	if len(handler.rateLimits) != 1 {
		t.Errorf("expected 1 rate limit update, got %d", len(handler.rateLimits))
	}
}

func TestHandleUpstreamError_500(t *testing.T) {
	handler := &mockStateHandler{}
	strategy := NewRateLimitStrategy(handler)
	acc := NewAccount(1, "test", PlatformOpenAI, "oauth", nil, nil, 50, StatusActive, nil, nil, nil, nil, nil, nil, "")

	result := strategy.HandleUpstreamError(context.Background(), acc, 500, nil, nil)
	if result.ShouldRateLimit {
		t.Error("500 should not trigger rate limit")
	}
	if len(handler.tempUnscheds) != 1 {
		t.Errorf("expected 1 temp-unsched, got %d", len(handler.tempUnscheds))
	}
}

func TestHandleUpstreamError_NilAccount(t *testing.T) {
	handler := &mockStateHandler{}
	strategy := NewRateLimitStrategy(handler)

	result := strategy.HandleUpstreamError(context.Background(), nil, 500, nil, nil)
	if result.ShouldDisable || result.ShouldRateLimit || result.IsAuthError {
		t.Error("nil account should produce empty result")
	}
}

func TestHandleAnthropic429(t *testing.T) {
	handler := &mockStateHandler{}
	strategy := NewRateLimitStrategy(handler)
	acc := NewAccount(1, "test", PlatformAnthropic, "oauth", nil, nil, 50, StatusActive, nil, nil, nil, nil, nil, nil, "")

	resetAt := time.Now().Add(time.Hour)
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-utilization", "100")
	headers.Set("anthropic-ratelimit-unified-5h-status", "rejected")
	headers.Set("anthropic-ratelimit-unified-5h-reset", resetAt.Format(time.RFC3339))

	result := strategy.HandleAnthropic429(context.Background(), acc, headers)
	if !result.ShouldRateLimit {
		t.Error("Anthropic 429 with rejected window should rate limit")
	}
	if len(handler.rateLimits) != 1 {
		t.Errorf("expected 1 rate limit update, got %d", len(handler.rateLimits))
	}
}

func TestHandleOpenAIImageRateLimit(t *testing.T) {
	handler := &mockStateHandler{}
	strategy := NewRateLimitStrategy(handler)
	acc := NewAccount(1, "test", PlatformOpenAI, "oauth", nil, nil, 50, StatusActive, nil, nil, nil, nil, nil, nil, "")

	body := []byte(`{"error":{"message":"image generation rate limit, try again in 30s"}}`)
	headers := http.Header{}

	handled := strategy.HandleOpenAIImageRateLimit(context.Background(), acc, 429, headers, body)
	if !handled {
		t.Error("should handle OpenAI image rate limit")
	}
	if len(handler.rateLimits) != 1 {
		t.Errorf("expected 1 rate limit, got %d", len(handler.rateLimits))
	}
}

func TestClearRateLimit(t *testing.T) {
	handler := &mockStateHandler{}
	strategy := NewRateLimitStrategy(handler)

	err := strategy.ClearRateLimit(context.Background(), 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !handler.cleared {
		t.Error("expected rate limit to be cleared")
	}
}

func TestUpdateSessionWindow_Anthropic(t *testing.T) {
	handler := &mockStateHandler{}
	strategy := NewRateLimitStrategy(handler)
	acc := NewAccount(1, "test", PlatformAnthropic, "oauth", nil, nil, 50, StatusActive, nil, nil, nil, nil, nil, nil, "")

	resetAt := time.Now().Add(time.Hour)
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-utilization", "0.75")
	headers.Set("anthropic-ratelimit-unified-5h-reset", resetAt.Format(time.RFC3339))

	strategy.UpdateSessionWindow(context.Background(), acc, headers)

	if len(handler.windowUpdates) != 1 {
		t.Errorf("expected 1 window update, got %d", len(handler.windowUpdates))
	}
	if len(handler.extraUpdates) != 1 {
		t.Errorf("expected 1 extra update, got %d", len(handler.extraUpdates))
	}
}

func TestExtractForbiddenMessage(t *testing.T) {
	body := []byte(`{"error":{"message":"access denied for this resource"}}`)
	msg := extractForbiddenMessage(body)
	if msg != "access denied for this resource" {
		t.Errorf("expected parsed message, got %q", msg)
	}

	msg2 := extractForbiddenMessage([]byte("raw error text"))
	if msg2 != "raw error text" {
		t.Errorf("expected raw body, got %q", msg2)
	}

	msg3 := extractForbiddenMessage(nil)
	if msg3 != "no details" {
		t.Error("expected 'no details' for nil body")
	}
}
