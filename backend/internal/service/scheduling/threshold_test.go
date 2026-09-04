package scheduling

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func TestEvaluateAccountSchedulingThreshold_NilAccount(t *testing.T) {
	now := time.Now()
	result := EvaluateAccountSchedulingThreshold(nil, map[string]int{"openai": 80}, now)
	if result.ShouldPause {
		t.Error("nil account should not trigger pause")
	}
}

func TestEvaluateAccountSchedulingThreshold_EmptyPlatform(t *testing.T) {
	now := time.Now()
	acc := NewAccount(1, "test", "", "oauth", nil, nil, 50, StatusActive, nil, nil, nil, nil, nil, nil, "")
	result := EvaluateAccountSchedulingThreshold(acc, map[string]int{"openai": 80}, now)
	if result.ShouldPause {
		t.Error("empty platform should not trigger pause")
	}
}

func TestEvaluateAccountSchedulingThreshold_UnsupportedPlatform(t *testing.T) {
	now := time.Now()
	acc := NewAccount(1, "test", PlatformGemini, "api_key", nil, nil, 50, StatusActive, nil, nil, nil, nil, nil, nil, "")
	result := EvaluateAccountSchedulingThreshold(acc, map[string]int{PlatformGemini: 80}, now)
	if result.ShouldPause {
		t.Error("unsupported platform (gemini) should not trigger pause")
	}
}

func TestEvaluateAccountSchedulingThreshold_Threshold100NoPause(t *testing.T) {
	now := time.Now()
	acc := NewAccount(1, "test", PlatformOpenAI, "oauth", nil,
		map[string]any{
			"codex_5h_used_percent": 95,
			"codex_5h_reset_at":     now.Add(time.Hour).Format(time.RFC3339),
		},
		50, StatusActive, nil, nil, nil, nil, nil, nil, "")
	result := EvaluateAccountSchedulingThreshold(acc, map[string]int{PlatformOpenAI: 100}, now)
	if result.ShouldPause {
		t.Error("threshold >= 100 should never pause")
	}
}

func TestEvaluateAccountSchedulingThreshold_OpenAIExceedsThreshold(t *testing.T) {
	now := time.Now()
	resetAt := now.Add(2 * time.Hour)
	acc := NewAccount(1, "test", PlatformOpenAI, "oauth",
		map[string]any{"chatgpt_account_id": "acc_123"},
		map[string]any{
			"codex_5h_used_percent": 90,
			"codex_5h_reset_at":     resetAt.Format(time.RFC3339),
			"chatgpt_account_id":    "acc_123",
		},
		50, StatusActive, nil, nil, nil, nil, nil, nil, "")

	result := EvaluateAccountSchedulingThreshold(acc, map[string]int{PlatformOpenAI: 80}, now)
	if !result.ShouldPause {
		t.Error("should pause when usage 90% >= threshold 80%")
	}
	if result.Window != "5h" {
		t.Errorf("expected window 5h, got %s", result.Window)
	}
	if result.UsedPercent != 90.0 {
		t.Errorf("expected used percent 90, got %f", result.UsedPercent)
	}
	if result.Until == nil || result.Until.Truncate(time.Second) != resetAt.Truncate(time.Second) {
		t.Errorf("expected until %v, got %v", resetAt.Truncate(time.Second), result.Until)
	}
}

func TestEvaluateAccountSchedulingThreshold_OpenAIBelowThreshold(t *testing.T) {
	now := time.Now()
	acc := NewAccount(1, "test", PlatformOpenAI, "oauth",
		map[string]any{"chatgpt_account_id": "acc_123"},
		map[string]any{
			"codex_5h_used_percent": 50,
			"codex_5h_reset_at":     now.Add(time.Hour).Format(time.RFC3339),
			"chatgpt_account_id":    "acc_123",
		},
		50, StatusActive, nil, nil, nil, nil, nil, nil, "")

	result := EvaluateAccountSchedulingThreshold(acc, map[string]int{PlatformOpenAI: 80}, now)
	if result.ShouldPause {
		t.Error("should not pause when usage 50% < threshold 80%")
	}
}

func TestEvaluateAccountSchedulingThreshold_OpenAIWindowReset(t *testing.T) {
	now := time.Now()
	acc := NewAccount(1, "test", PlatformOpenAI, "oauth",
		map[string]any{"chatgpt_account_id": "acc_123"},
		map[string]any{
			"codex_5h_used_percent": 95,
			"codex_5h_reset_at":     now.Add(-time.Hour).Format(time.RFC3339),
			"chatgpt_account_id":    "acc_123",
		},
		50, StatusActive, nil, nil, nil, nil, nil, nil, "")

	result := EvaluateAccountSchedulingThreshold(acc, map[string]int{PlatformOpenAI: 80}, now)
	if result.ShouldPause {
		t.Error("should not pause when window has already reset")
	}
}

func TestEvaluateAccountSchedulingThreshold_IdentityConflict(t *testing.T) {
	now := time.Now()
	acc := NewAccount(1, "test", PlatformOpenAI, "oauth",
		map[string]any{"chatgpt_account_id": "acc_123"},
		map[string]any{
			"codex_5h_used_percent": 90,
			"codex_5h_reset_at":     now.Add(time.Hour).Format(time.RFC3339),
			"chatgpt_account_id":    "acc_different",
		},
		50, StatusActive, nil, nil, nil, nil, nil, nil, "")

	result := EvaluateAccountSchedulingThreshold(acc, map[string]int{PlatformOpenAI: 80}, now)
	if result.ShouldPause {
		t.Error("should not pause when identity conflict (untrusted snapshot)")
	}
}

func TestEvaluateAccountSchedulingThreshold_AnthropicExceedsThreshold(t *testing.T) {
	now := time.Now()
	resetAt := now.Add(time.Hour)
	acc := NewAccount(1, "test", PlatformAnthropic, "oauth", nil, nil,
		50, StatusActive, nil, nil, nil, nil, nil, &resetAt, "")
	acc.Extra = map[string]any{
		"session_window_utilization": 0.95,
	}

	result := EvaluateAccountSchedulingThreshold(acc, map[string]int{PlatformAnthropic: 80}, now)
	if !result.ShouldPause {
		t.Error("should pause when Anthropic session_window_utilization 95% >= 80%")
	}
	if result.UsedPercent != 95.0 {
		t.Errorf("expected used percent 95, got %f", result.UsedPercent)
	}
}

func TestEvaluateAccountSchedulingThreshold_AccountOverride(t *testing.T) {
	now := time.Now()
	acc := NewAccount(1, "test", PlatformOpenAI, "oauth",
		map[string]any{
			"chatgpt_account_id":                    "acc_123",
			accountSchedulingThresholdCredentialKey: 60,
		},
		map[string]any{
			"codex_5h_used_percent": 65,
			"codex_5h_reset_at":     now.Add(time.Hour).Format(time.RFC3339),
			"chatgpt_account_id":    "acc_123",
		},
		50, StatusActive, nil, nil, nil, nil, nil, nil, "")

	result := EvaluateAccountSchedulingThreshold(acc, map[string]int{PlatformOpenAI: 80}, now)
	if !result.ShouldPause {
		t.Error("should pause: usage 65% >= account override threshold 60%")
	}
	if result.ThresholdPercent != 60 {
		t.Errorf("expected threshold 60, got %d", result.ThresholdPercent)
	}
}

func TestParseAccountSchedulingThresholdValue(t *testing.T) {
	tests := []struct {
		input  any
		want   int
		wantOK bool
	}{
		{80, 80, true},
		{int64(90), 90, true},
		{float64(75.4), 75, true},
		{float64(75.6), 76, true},
		{float32(85), 85, true},
		{json.Number("95"), 95, true},
		{"70", 70, true},
		{"  80  ", 80, true},
		{0, 0, false},
		{101, 0, false},
		{-5, 0, false},
		{"abc", 0, false},
		{nil, 0, false},
		{true, 0, false},
	}

	for _, tt := range tests {
		got, ok := parseAccountSchedulingThresholdValue(tt.input)
		if ok != tt.wantOK {
			t.Errorf("parseAccountSchedulingThresholdValue(%v) ok = %v, want %v", tt.input, ok, tt.wantOK)
			continue
		}
		if ok && got != tt.want {
			t.Errorf("parseAccountSchedulingThresholdValue(%v) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseSchedulingResetAt(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		raw  any
		want bool
	}{
		{"nil", nil, false},
		{"empty string", "", false},
		{"RFC3339", now.Format(time.RFC3339), true},
		{"RFC3339Nano", now.Format(time.RFC3339Nano), true},
		{"unix seconds", now.Unix(), true},
		{"json number", json.Number("1735689600"), true},
		{"zero int", 0, false},
		{"negative int", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSchedulingResetAt(tt.raw)
			if (result != nil) != tt.want {
				t.Errorf("parseSchedulingResetAt(%v) nil=%v, want nil=%v", tt.raw, result == nil, !tt.want)
			}
		})
	}
}

func TestUtilizationAsPercent(t *testing.T) {
	tests := []struct {
		input any
		want  float64
	}{
		{0.5, 50.0},
		{50.0, 50.0},
		{1.0, 100.0},
		{0.0, 0.0},
		{100.0, 100.0},
		{"75", 75.0},
		{"0.8", 80.0},
		{int64(60), 60.0},
		{nil, 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		got := utilizationAsPercent(tt.input)
		if math.Abs(got-tt.want) > 0.001 {
			t.Errorf("utilizationAsPercent(%v) = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestParseTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"RFC3339", now.Format(time.RFC3339), false},
		{"RFC3339Nano", now.Format(time.RFC3339Nano), false},
		{"custom Z", "2025-01-01T00:00:00Z", false},
		{"custom millis", "2025-01-01T00:00:00.000Z", false},
		{"datetime", "2025-01-01 15:04:05", false},
		{"invalid", "not-a-date", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTime(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestAccountMethods(t *testing.T) {
	now := time.Now()
	resetAt := now.Add(time.Hour)

	acc := NewAccount(1, "test", PlatformOpenAI, AccountTypeOAuth,
		map[string]any{"api_key": "sk-test"},
		map[string]any{"extra_key": "extra_val"},
		50, StatusActive, nil, &resetAt, nil, nil, nil, nil, "")

	if !acc.IsActive() {
		t.Error("account should be active")
	}
	if !acc.IsOAuth() {
		t.Error("OAuth type account should be OAuth")
	}
	if !acc.IsOpenAIOAuth() {
		t.Error("OpenAI OAuth account should be OpenAI OAuth")
	}
	if !acc.IsRateLimited() {
		t.Error("account with future RateLimitResetAt should be rate limited")
	}
	if acc.GetCredential("api_key") != "sk-test" {
		t.Error("should return correct credential")
	}
	if acc.GetCredential("missing") != "" {
		t.Error("missing credential should return empty string")
	}
	if acc.GetExtraString("extra_key") != "extra_val" {
		t.Error("should return correct extra value")
	}

	var nilAcc *Account
	if nilAcc.IsActive() {
		t.Error("nil account should not be active")
	}
	if nilAcc.IsOAuth() {
		t.Error("nil account should not be OAuth")
	}
	if nilAcc.IsRateLimited() {
		t.Error("nil account should not be rate limited")
	}
}

func TestIsAllowedSchedulingThresholdPlatform(t *testing.T) {
	supported := []string{PlatformOpenAI, PlatformAnthropic, PlatformGrok, PlatformKimi, PlatformZhipu}
	unsupported := []string{PlatformGemini, PlatformAntigravity, PlatformDeepseek, PlatformComposite, ""}

	for _, p := range supported {
		if !isAllowedSchedulingThresholdPlatform(p) {
			t.Errorf("platform %s should be allowed", p)
		}
	}
	for _, p := range unsupported {
		if isAllowedSchedulingThresholdPlatform(p) {
			t.Errorf("platform %s should not be allowed", p)
		}
	}
}

func TestOpenAIQuotaWindowReset(t *testing.T) {
	now := time.Now()

	extra := map[string]any{
		"codex_5h_reset_at": now.Add(time.Hour).Format(time.RFC3339),
	}
	if openAIQuotaWindowReset(extra, "5h", now) {
		t.Error("window should not be reset when reset_at is in the future")
	}

	extra = map[string]any{
		"codex_5h_reset_at": now.Add(-time.Hour).Format(time.RFC3339),
	}
	if !openAIQuotaWindowReset(extra, "5h", now) {
		t.Error("window should be reset when reset_at is in the past")
	}

	extra = map[string]any{
		"codex_usage_updated_at":       now.Add(-30 * time.Minute).Format(time.RFC3339),
		"codex_5h_reset_after_seconds": 1800,
	}
	if !openAIQuotaWindowReset(extra, "5h", now) {
		t.Error("window should be reset when reset_after_seconds has elapsed")
	}

	if openAIQuotaWindowReset(nil, "5h", now) {
		t.Error("nil extra should not trigger reset")
	}
	if openAIQuotaWindowReset(map[string]any{}, "5h", now) {
		t.Error("empty extra should not trigger reset")
	}
}

func TestPickSooner(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)

	a := now
	b := later

	result := PickSooner(&a, &b)
	if !result.Equal(now) {
		t.Error("PickSooner should return earlier time")
	}
	if r2 := PickSooner(nil, &b); !r2.Equal(later) {
		t.Error("PickSooner(nil, b) should return b")
	}
	if r3 := PickSooner(&a, nil); !r3.Equal(now) {
		t.Error("PickSooner(a, nil) should return a")
	}
	if r4 := PickSooner(nil, nil); r4 != nil {
		t.Error("PickSooner(nil, nil) should return nil")
	}
}

func TestClamp01(t *testing.T) {
	if Clamp01(-0.5) != 0 {
		t.Error("Clamp01 should floor at 0")
	}
	if Clamp01(0.5) != 0.5 {
		t.Error("Clamp01 should pass through [0,1]")
	}
	if Clamp01(1.5) != 1 {
		t.Error("Clamp01 should cap at 1")
	}
}
