package scheduling

import (
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestParseAnthropicRateLimitHeaders(t *testing.T) {
	now := time.Now()
	resetAt := now.Add(time.Hour)

	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-utilization", "85")
	headers.Set("anthropic-ratelimit-unified-5h-reset", resetAt.Format(time.RFC3339))
	headers.Set("anthropic-ratelimit-unified-5h-status", "rejected")

	result := ParseAnthropicRateLimitHeaders(headers, now)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Window != "5h" {
		t.Errorf("expected window 5h, got %s", result.Window)
	}
	if result.Utilization != 85.0 {
		t.Errorf("expected utilization 85, got %f", result.Utilization)
	}
	if result.Status != "rejected" {
		t.Errorf("expected status rejected, got %s", result.Status)
	}
	if !result.ResetAt.Equal(resetAt.Truncate(time.Second)) {
		t.Errorf("expected reset at %v, got %v", resetAt, result.ResetAt)
	}
}

func TestParseAnthropicRateLimitHeaders_Nil(t *testing.T) {
	result := ParseAnthropicRateLimitHeaders(nil, time.Now())
	if result != nil {
		t.Error("nil headers should return nil")
	}
}

func TestIsAnthropicWindowRejected(t *testing.T) {
	headers := http.Header{}

	if isAnthropic5hRejected(headers) {
		t.Error("empty headers should not be rejected")
	}

	headers.Set("anthropic-ratelimit-unified-5h-status", "rejected")
	if !isAnthropic5hRejected(headers) {
		t.Error("should be rejected when status is 'rejected'")
	}

	headers.Set("anthropic-ratelimit-unified-5h-status", "ok")
	if isAnthropic5hRejected(headers) {
		t.Error("should not be rejected when status is 'ok'")
	}
}

func TestParseAnthropicResetTimestamp(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	ts, ok := parseAnthropicResetTimestamp(future.Format(time.RFC3339), now, 8*24*time.Hour)
	if !ok {
		t.Error("should parse RFC3339 future timestamp")
	}
	if !ts.Equal(future.Truncate(time.Second)) {
		t.Errorf("expected %v, got %v", future, ts)
	}

	_, ok = parseAnthropicResetTimestamp(past.Format(time.RFC3339), now, 8*24*time.Hour)
	if ok {
		t.Error("past timestamp should not be valid")
	}

	unixSec := future.Unix()
	_, ok = parseAnthropicResetTimestamp(strconv.FormatInt(unixSec, 10), now, 8*24*time.Hour)
	if !ok {
		t.Error("should parse unix seconds timestamp")
	}

	unixMs := future.UnixNano() / 1e6
	_, ok = parseAnthropicResetTimestamp(strconv.FormatInt(unixMs, 10), now, 8*24*time.Hour)
	if !ok {
		t.Error("should parse unix milliseconds timestamp")
	}


	_, ok = parseAnthropicResetTimestamp("", now, 8*24*time.Hour)
	if ok {
		t.Error("empty string should not be valid")
	}

	_, ok = parseAnthropicResetTimestamp(now.Add(30*24*time.Hour).Format(time.RFC3339), now, 8*24*time.Hour)
	if ok {
		t.Error("timestamp too far in future should not be valid")
	}
}

func TestSampleAnthropicPassiveUsage(t *testing.T) {
	now := time.Now()
	resetAt5h := now.Add(time.Hour)
	resetAt7d := now.Add(24 * time.Hour)

	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-utilization", "0.75")
	headers.Set("anthropic-ratelimit-unified-5h-reset", resetAt5h.Format(time.RFC3339))
	headers.Set("anthropic-ratelimit-unified-7d-utilization", "0.50")
	headers.Set("anthropic-ratelimit-unified-7d-reset", resetAt7d.Format(time.RFC3339))

	usage := SampleAnthropicPassiveUsage(headers, now)
	if !usage.HasData {
		t.Error("should have data")
	}
	if usage.Window5hUtilization != 75.0 {
		t.Errorf("expected 5h utilization 75, got %f", usage.Window5hUtilization)
	}
	if usage.Window7dUtilization != 50.0 {
		t.Errorf("expected 7d utilization 50, got %f", usage.Window7dUtilization)
	}
	if usage.Window5hResetAt == nil || !usage.Window5hResetAt.Equal(resetAt5h.Truncate(time.Second)) {
		t.Errorf("expected 5h reset at %v, got %v", resetAt5h, usage.Window5hResetAt)
	}
}

func TestSampleAnthropicPassiveUsage_Nil(t *testing.T) {
	usage := SampleAnthropicPassiveUsage(nil, time.Now())
	if usage.HasData {
		t.Error("nil headers should return no data")
	}
}

func TestCalculateWindowCostSchedulability(t *testing.T) {
	if CalculateWindowCostSchedulability(100, 0, 10) != WindowCostSchedulable {
		t.Error("no limit should be schedulable")
	}

	if CalculateWindowCostSchedulability(50, 100, 10) != WindowCostSchedulable {
		t.Error("below limit should be schedulable")
	}

	if CalculateWindowCostSchedulability(100, 100, 10) != WindowCostStickyOnly {
		t.Error("at limit should be sticky-only")
	}

	if CalculateWindowCostSchedulability(111, 100, 10) != WindowCostNotSchedulable {
		t.Error("above limit+reserve should not be schedulable")
	}
}

func TestCalculateAnthropic429ResetTime(t *testing.T) {
	now := time.Now()
	reset5h := now.Add(time.Hour)
	reset7d := now.Add(24 * time.Hour)

	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-utilization", "100")
	headers.Set("anthropic-ratelimit-unified-5h-status", "rejected")
	headers.Set("anthropic-ratelimit-unified-5h-reset", reset5h.Format(time.RFC3339))
	headers.Set("anthropic-ratelimit-unified-7d-utilization", "100")
	headers.Set("anthropic-ratelimit-unified-7d-status", "rejected")
	headers.Set("anthropic-ratelimit-unified-7d-reset", reset7d.Format(time.RFC3339))

	result := CalculateAnthropic429ResetTime(headers, now)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Window != "7d" {
		t.Errorf("expected window 7d, got %s", result.Window)
	}
	if !result.ResetAt.Equal(reset7d.Truncate(time.Second)) {
		t.Errorf("expected reset at %v, got %v", reset7d, result.ResetAt)
	}
}

func TestCalculateAnthropic429ResetTime_Nil(t *testing.T) {
	result := CalculateAnthropic429ResetTime(nil, time.Now())
	if result != nil {
		t.Error("nil headers should return nil")
	}
}

func TestParseOpenAIRateLimitResetTime(t *testing.T) {
	now := time.Now()

	unixTs := now.Add(time.Hour).Unix()
	body := []byte(`{"error":{"message":"rate limited","type":"usage_limit_reached","resets_at":` + strconv.FormatInt(unixTs, 10) + `}}`)
	result := ParseOpenAIRateLimitResetTime(body)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	expected := now.Add(time.Hour).Truncate(time.Second)
	if !result.Truncate(time.Second).Equal(expected) {
		t.Errorf("expected ~%v, got %v", expected, *result)
	}

	body2 := []byte(`{"error":{"message":"rate limited","resets_in_seconds":3600}}`)
	result2 := ParseOpenAIRateLimitResetTime(body2)
	if result2 == nil {
		t.Fatal("expected non-nil result for resets_in_seconds")
	}

	result3 := ParseOpenAIRateLimitResetTime(nil)
	if result3 != nil {
		t.Error("nil body should return nil")
	}

	result4 := ParseOpenAIRateLimitResetTime([]byte("not json"))
	if result4 != nil {
		t.Error("invalid JSON should return nil")
	}
}

func TestCalculateOpenAI429ResetTime(t *testing.T) {
	now := time.Now()
	headers := http.Header{}

	headers.Set("x-ratelimit-reset-usage", "30s")
	result := CalculateOpenAI429ResetTime(headers, now)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	expected := now.Add(30 * time.Second)
	if result.Truncate(time.Millisecond) != expected.Truncate(time.Millisecond) {
		t.Errorf("expected ~%v, got %v", expected, *result)
	}

	headers = http.Header{}
	headers.Set("retry-after", "120")
	result2 := CalculateOpenAI429ResetTime(headers, now)
	if result2 == nil {
		t.Fatal("expected non-nil result for retry-after")
	}
	expected2 := now.Add(120 * time.Second)
	if result2.Truncate(time.Millisecond) != expected2.Truncate(time.Millisecond) {
		t.Errorf("expected ~%v, got %v", expected2, *result2)
	}

	result3 := CalculateOpenAI429ResetTime(nil, now)
	if result3 != nil {
		t.Error("nil headers should return nil")
	}
}

func TestParseOpenAIImageTryAgainCooldown(t *testing.T) {
	tests := []struct {
		body string
		want time.Duration
	}{
		{`{"error":{"message":"try again in 5s"}}`, 5 * time.Second},
		{`{"error":{"message":"try again in 1.5 minutes"}}`, 90 * time.Second},
		{`{"error":{"message":"try again in 500ms"}}`, 500 * time.Millisecond},
		{`{"error":{"message":"no cooldown"}}`, 0},
		{`{}`, 0},
		{"", 0},
	}

	for _, tt := range tests {
		got := ParseOpenAIImageTryAgainCooldown([]byte(tt.body))
		if got != tt.want {
			t.Errorf("ParseOpenAIImageTryAgainCooldown(%q) = %v, want %v", tt.body, got, tt.want)
		}
	}
}

func TestIsOpenAIImageRateLimitError(t *testing.T) {
	if !IsOpenAIImageRateLimitError(429, []byte(`{"error":{"message":"image rate limit, try again"}}`)) {
		t.Error("should detect image rate limit error")
	}
	if IsOpenAIImageRateLimitError(429, []byte(`{"error":{"message":"generic error"}}`)) {
		t.Error("should not detect non-image error")
	}
	if IsOpenAIImageRateLimitError(400, []byte(`{"error":{"message":"image rate limit"}}`)) {
		t.Error("should not detect non-429 status")
	}
}

func TestParseRetryAfterResetTime(t *testing.T) {
	now := time.Now()
	headers := http.Header{}
	headers.Set("retry-after", "60")

	result := ParseRetryAfterResetTime(headers, now)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	expected := now.Add(60 * time.Second)
	if result.Truncate(time.Millisecond) != expected.Truncate(time.Millisecond) {
		t.Errorf("expected ~%v, got %v", expected, *result)
	}

	result2 := ParseRetryAfterResetTime(nil, now)
	if result2 != nil {
		t.Error("nil headers should return nil")
	}

	headers2 := http.Header{}
	headers2.Set("retry-after", "abc")
	result3 := ParseRetryAfterResetTime(headers2, now)
	if result3 != nil {
		t.Error("invalid retry-after should return nil")
	}
}
