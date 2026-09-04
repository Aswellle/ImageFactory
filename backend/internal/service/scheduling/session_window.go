package scheduling

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SessionWindowState tracks the rolling usage window for an account.
// Used by the session window management algorithm ported from Sub2API's
// ratelimit_service.go (UpdateSessionWindow, samplePassiveUsageFromHeaders).
type SessionWindowState struct {
	WindowStart  *time.Time
	WindowEnd    *time.Time
	Utilization  float64
	RequestCount int
	Cost         float64
}

// WindowCostSchedulability indicates whether an account can be scheduled
// based on its current window cost.
type WindowCostSchedulability int

const (
	WindowCostSchedulable WindowCostSchedulability = iota
	WindowCostStickyOnly
	WindowCostNotSchedulable
)

// AnthropicRateLimitHeaders holds parsed Anthropic rate-limit information
// from response headers.
type AnthropicRateLimitHeaders struct {
	Window     string
	ResetAt    time.Time
	Utilization float64
	Status     string
}

// ParseAnthropicRateLimitHeaders parses Anthropic rate-limit headers from an HTTP response.
func ParseAnthropicRateLimitHeaders(headers http.Header, now time.Time) *AnthropicRateLimitHeaders {
	if headers == nil {
		return nil
	}

	result := &AnthropicRateLimitHeaders{}

	for _, window := range []string{"5h", "7d", "7d_oi"} {
		if isAnthropicWindowRejected(headers, window) {
			result.Window = window
			break
		}
	}

	for _, window := range []string{"5h", "7d"} {
		key := "anthropic-ratelimit-unified-" + window + "-utilization"
		if v := headers.Get(key); v != "" {
			result.Utilization = utilizationAsPercent(v)
			break
		}
	}

	if reset, ok := parseAnthropicWindowReset(headers, "5h", now); ok {
		result.ResetAt = reset
	} else if reset, ok := parseAnthropicWindowReset(headers, "7d", now); ok {
		result.ResetAt = reset
	} else if reset, ok := parseAnthropicAggregateReset(headers, now); ok {
		result.ResetAt = reset
	}

	result.Status = headers.Get("anthropic-ratelimit-unified-5h-status")
	if result.Status == "" {
		result.Status = headers.Get("anthropic-ratelimit-unified-7d-status")
	}

	return result
}

func isAnthropicWindowRejected(headers http.Header, window string) bool {
	return strings.EqualFold(strings.TrimSpace(headers.Get("anthropic-ratelimit-unified-"+window+"-status")), "rejected")
}

func isAnthropic5hRejected(headers http.Header) bool {
	return isAnthropicWindowRejected(headers, "5h")
}

func parseAnthropicWindowReset(headers http.Header, window string, now time.Time) (time.Time, bool) {
	raw := headers.Get("anthropic-ratelimit-unified-" + window + "-reset")
	return parseAnthropicResetTimestamp(raw, now, 8*24*time.Hour)
}

func parseAnthropicAggregateReset(headers http.Header, now time.Time) (time.Time, bool) {
	return parseAnthropicResetTimestamp(headers.Get("anthropic-ratelimit-unified-reset"), now, 8*24*time.Hour)
}

func parseAnthropicResetTimestamp(raw string, now time.Time, maxAge time.Duration) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}

	var unixSec int64
	if strings.ContainsAny(raw, "T-:") {
		ts, err := parseTime(raw)
		if err != nil {
			return time.Time{}, false
		}
		unixSec = ts.Unix()
	} else {
		val, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return time.Time{}, false
		}
		if val > 1e12 {
			unixSec = int64(val / 1000)
		} else {
			unixSec = int64(val)
		}
	}

	resetAt := time.Unix(unixSec, 0)
	if !resetAt.After(now) {
		return time.Time{}, false
	}
	if resetAt.Sub(now) > maxAge {
		return time.Time{}, false
	}
	return resetAt, true
}

// AnthropicPassiveUsage holds passive usage sampling data from Anthropic
// response headers for 5h/7d/7d_oi windows.
type AnthropicPassiveUsage struct {
	Window5hUtilization   float64
	Window5hResetAt       *time.Time
	Window7dUtilization   float64
	Window7dResetAt       *time.Time
	Window7dOIUtilization float64
	Window7dOIResetAt     *time.Time
	HasData               bool
}

// SampleAnthropicPassiveUsage samples passive usage from Anthropic headers.
func SampleAnthropicPassiveUsage(headers http.Header, now time.Time) AnthropicPassiveUsage {
	if headers == nil {
		return AnthropicPassiveUsage{}
	}

	var usage AnthropicPassiveUsage

	// 5h window
	if v := headers.Get("anthropic-ratelimit-unified-5h-utilization"); v != "" {
		usage.Window5hUtilization = utilizationAsPercent(v)
		usage.HasData = true
	}
	if reset, ok := parseAnthropicWindowReset(headers, "5h", now); ok {
		usage.Window5hResetAt = &reset
		usage.HasData = true
	}

	// 7d window
	if v := headers.Get("anthropic-ratelimit-unified-7d-utilization"); v != "" {
		usage.Window7dUtilization = utilizationAsPercent(v)
		usage.HasData = true
	}
	if reset, ok := parseAnthropicWindowReset(headers, "7d", now); ok {
		usage.Window7dResetAt = &reset
		usage.HasData = true
	}

	// 7d_oi window
	if v := headers.Get("anthropic-ratelimit-unified-7d_oi-utilization"); v != "" {
		usage.Window7dOIUtilization = utilizationAsPercent(v)
		usage.HasData = true
	}
	if reset, ok := parseAnthropicResetTimestamp(headers.Get("anthropic-ratelimit-unified-7d_oi-reset"), now, 8*24*time.Hour); ok {
		usage.Window7dOIResetAt = &reset
		usage.HasData = true
	}

	return usage
}

// CalculateWindowCostSchedulability determines if an account should be schedulable.
func CalculateWindowCostSchedulability(currentWindowCost, windowCostLimit, windowCostStickyReserve float64) WindowCostSchedulability {
	if windowCostLimit <= 0 {
		return WindowCostSchedulable
	}
	if currentWindowCost < windowCostLimit {
		return WindowCostSchedulable
	}
	if currentWindowCost < windowCostLimit+windowCostStickyReserve {
		return WindowCostStickyOnly
	}
	return WindowCostNotSchedulable
}

// PickSooner returns whichever of the two time pointers is earlier.
func PickSooner(a, b *time.Time) *time.Time {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if a.Before(*b) {
		return a
	}
	return b
}

// Anthropic429Result holds the parsed Anthropic 429 rate-limit information.
type Anthropic429Result struct {
	Window  string
	ResetAt time.Time
}

// CalculateAnthropic429ResetTime parses Anthropic's per-window rate-limit headers.
func CalculateAnthropic429ResetTime(headers http.Header, now time.Time) *Anthropic429Result {
	if headers == nil {
		return nil
	}

	var windows []*struct {
		name    string
		resetAt time.Time
		util    float64
	}

	for _, window := range []string{"5h", "7d"} {
		if exceeded := isAnthropicWindowExceeded(headers, window); exceeded {
			if reset, ok := parseAnthropicWindowReset(headers, window, now); ok {
				util := utilizationAsPercent(headers.Get("anthropic-ratelimit-unified-" + window + "-utilization"))
				windows = append(windows, &struct {
					name    string
					resetAt time.Time
					util    float64
				}{window, reset, util})
			}
		}
	}

	if len(windows) == 0 {
		return nil
	}

	var winner *struct {
		name    string
		resetAt time.Time
		util    float64
	}
	for _, w := range windows {
		if winner == nil || w.resetAt.After(winner.resetAt) {
			winner = w
		}
	}

	if winner == nil {
		return nil
	}

	return &Anthropic429Result{
		Window:  winner.name,
		ResetAt: winner.resetAt,
	}
}

// isAnthropicWindowExceeded checks whether a given Anthropic rate-limit window has been exceeded.
func isAnthropicWindowExceeded(headers http.Header, window string) bool {
	utilHeader := headers.Get("anthropic-ratelimit-unified-" + window + "-utilization")
	surpassedHeader := headers.Get("anthropic-ratelimit-unified-" + window + "-surpassed-threshold")
	statusHeader := headers.Get("anthropic-ratelimit-unified-" + window + "-status")

	if strings.EqualFold(strings.TrimSpace(statusHeader), "rejected") {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(surpassedHeader), "true") {
		return true
	}
	util := utilizationAsPercent(utilHeader)
	return util >= 100.0
}

// ParseOpenAIRateLimitResetTime parses an OpenAI-format 429 response body.
func ParseOpenAIRateLimitResetTime(body []byte) *time.Time {
	if len(body) == 0 {
		return nil
	}

	var parsed struct {
		Error struct {
			ResetsAt        any `json:"resets_at"`
			ResetsInSeconds any `json:"resets_in_seconds"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil
	}

	if parsed.Error.ResetsAt != nil {
		switch v := parsed.Error.ResetsAt.(type) {
		case float64:
			if v > 1e12 {
				ts := time.Unix(int64(v/1000), 0)
				return &ts
			}
			ts := time.Unix(int64(v), 0)
			return &ts
		}
	}

	if parsed.Error.ResetsInSeconds != nil {
		switch v := parsed.Error.ResetsInSeconds.(type) {
		case float64:
			if v > 0 {
				ts := time.Now().Add(time.Duration(v) * time.Second)
				return &ts
			}
		}
	}

	return nil
}

// CalculateOpenAI429ResetTime extracts the reset time from OpenAI 429 response headers.
func CalculateOpenAI429ResetTime(headers http.Header, now time.Time) *time.Time {
	if headers == nil {
		return nil
	}

	for _, key := range []string{
		"x-ratelimit-reset-usage",
		"x-ratelimit-reset-requests",
		"x-ratelimit-reset-tokens",
	} {
		if v := headers.Get(key); v != "" {
			if d, err := parseDuration(v); err == nil && d > 0 {
				ts := now.Add(d)
				return &ts
			}
		}
	}

	if v := headers.Get("retry-after"); v != "" {
		if seconds, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && seconds > 0 {
			ts := now.Add(time.Duration(seconds) * time.Second)
			return &ts
		}
	}

	return nil
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if seconds, err := strconv.Atoi(s); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}
	return time.ParseDuration(s)
}

var openAIImageTryAgainPattern = regexp.MustCompile(`(?i)try again in\s+([0-9]+(?:\.[0-9]+)?)\s*(ms|s|sec|secs|second|seconds|m|min|mins|minute|minutes)`)

// ParseOpenAIImageTryAgainCooldown extracts the cooldown duration from an OpenAI error.
func ParseOpenAIImageTryAgainCooldown(body []byte) time.Duration {
	if len(body) == 0 {
		return 0
	}
	matches := openAIImageTryAgainPattern.FindSubmatch(body)
	if len(matches) < 3 {
		return 0
	}
	value, err := strconv.ParseFloat(string(matches[1]), 64)
	if err != nil {
		return 0
	}
	unit := strings.ToLower(strings.TrimSpace(string(matches[2])))
	switch unit {
	case "ms":
		return time.Duration(value * float64(time.Millisecond))
	case "s", "sec", "secs", "second", "seconds":
		return time.Duration(value * float64(time.Second))
	case "m", "min", "mins", "minute", "minutes":
		return time.Duration(value * float64(time.Minute))
	}
	return 0
}

// IsOpenAIImageRateLimitError detects OpenAI image generation rate-limit errors.
func IsOpenAIImageRateLimitError(statusCode int, body []byte) bool {
	if statusCode != 429 {
		return false
	}
	bodyStr := strings.ToLower(string(body))
	return strings.Contains(bodyStr, "image") &&
		(strings.Contains(bodyStr, "try again") || strings.Contains(bodyStr, "rate limit"))
}

// Clamp01 clamps a float64 to the [0, 1] range.
func Clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

// ParseRetryAfterResetTime parses the Retry-After header as a reset time.
func ParseRetryAfterResetTime(headers http.Header, now time.Time) *time.Time {
	if headers == nil {
		return nil
	}
	v := headers.Get("retry-after")
	if v == "" {
		return nil
	}
	seconds, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return nil
	}
	if seconds <= 0 {
		return nil
	}
	ts := now.Add(time.Duration(seconds) * time.Second)
	return &ts
}
