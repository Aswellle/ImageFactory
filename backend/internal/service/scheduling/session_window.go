package scheduling

import (
	"encoding/json"
	"math"
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
	// WindowStart is when the current 5h window began.
	WindowStart *time.Time
	// WindowEnd is when the current 5h window expires.
	WindowEnd *time.Time
	// Utilization is the current usage as a percentage (0-100).
	Utilization float64
	// RequestCount is the number of requests in the current window.
	RequestCount int
	// Cost is the accumulated cost in the current window (USD).
	Cost float64
}

// WindowCostSchedulability indicates whether an account can be scheduled
// based on its current window cost.
type WindowCostSchedulability int

const (
	// WindowCostSchedulable means the account is fully schedulable.
	WindowCostSchedulable WindowCostSchedulability = iota
	// WindowCostStickyOnly means only sticky sessions can use this account.
	WindowCostStickyOnly
	// WindowCostNotSchedulable means the account is not schedulable.
	WindowCostNotSchedulable
)

// AnthropicRateLimitHeaders holds parsed Anthropic rate-limit information
// from response headers.
//
// Ported from Sub2API: ratelimit_service.go calculateAnthropic429ResetTime
type AnthropicRateLimitHeaders struct {
	// Window is which quota window triggered (5h, 7d, 7d_oi).
	Window string
	// ResetAt is when the window resets.
	ResetAt time.Time
	// Utilization is the current window utilization percentage.
	Utilization float64
	// Status is "rejected" when the window is exhausted.
	Status string
}

// ParseAnthropicRateLimitHeaders parses Anthropic rate-limit headers from an HTTP response.
// This is the header-reading portion of Sub2API's UpdateSessionWindow and
// calculateAnthropic429ResetTime, extracted as a pure function.
//
// Headers parsed:
//   - anthropic-ratelimit-unified-5h-utilization
//   - anthropic-ratelimit-unified-5h-reset
//   - anthropic-ratelimit-unified-7d-utilization
//   - anthropic-ratelimit-unified-7d-reset
//   - anthropic-ratelimit-unified-reset (aggregated)
func ParseAnthropicRateLimitHeaders(headers http.Header, now time.Time) *AnthropicRateLimitHeaders {
	if headers == nil {
		return nil
	}

	result := &AnthropicRateLimitHeaders{}

	// Determine which window is rejected
	for _, window := range []string{"5h", "7d", "7d_oi"} {
		if isAnthropicWindowRejected(headers, window) {
			result.Window = window
			break
		}
	}

	// Parse utilization from any available window
	for _, window := range []string{"5h", "7d"} {
		key := "anthropic-ratelimit-unified-" + window + "-utilization"
		if v := headers.Get(key); v != "" {
			result.Utilization = schedulingPercentValue(v)
			break
		}
	}

	// Parse reset time from per-window or aggregated header
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

// isAnthropicWindowRejected checks whether a given Anthropic rate-limit window
// (e.g. "5h" or "7d") is in "rejected" status.
//
// Ported from Sub2API: ratelimit_service.go isAnthropicWindowRejected
func isAnthropicWindowRejected(headers http.Header, window string) bool {
	return strings.EqualFold(strings.TrimSpace(headers.Get("anthropic-ratelimit-unified-"+window+"-status")), "rejected")
}

// isAnthropic5hRejected is a convenience check for the 5h window.
func isAnthropic5hRejected(headers http.Header) bool {
	return isAnthropicWindowRejected(headers, "5h")
}

// parseAnthropicWindowReset parses a per-window Anthropic reset header.
// It tries both Unix timestamp (seconds/milliseconds) and RFC3339 formats.
//
// Ported from Sub2API: ratelimit_service.go parseAnthropicWindowReset
func parseAnthropicWindowReset(headers http.Header, window string, now time.Time) (time.Time, bool) {
	raw := headers.Get("anthropic-ratelimit-unified-" + window + "-reset")
	return parseAnthropicResetTimestamp(raw, now, 8*24*time.Hour)
}

// parseAnthropicAggregateReset parses the aggregated reset header.
func parseAnthropicAggregateReset(headers http.Header, now time.Time) (time.Time, bool) {
	return parseAnthropicResetTimestamp(headers.Get("anthropic-ratelimit-unified-reset"), now, 8*24*time.Hour)
}

// parseAnthropicResetTimestamp parses an Anthropic reset header Unix timestamp
// (auto-detecting milliseconds), and validates it falls within (now, now+maxAge].
//
// Ported from Sub2API: ratelimit_service.go parseAnthropicResetTimestamp
func parseAnthropicResetTimestamp(raw string, now time.Time, maxAge time.Duration) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}

	var unixSec int64
	if strings.ContainsAny(raw, "T-:") {
		// Try RFC3339
		ts, err := parseTime(raw)
		if err != nil {
			return time.Time{}, false
		}
		unixSec = ts.Unix()
	} else {
		// Parse as Unix timestamp (seconds or milliseconds)
		val, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return time.Time{}, false
		}
		if val > 1e12 {
			// Milliseconds
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
//
// Ported from Sub2API: ratelimit_service.go samplePassiveUsageFromHeaders
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
// Returns a struct with all available window data. The HasData field indicates
// whether any usage information was found.
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

// CalculateWindowCostSchedulability determines if an account should be schedulable
// based on its current window cost vs threshold and sticky reserve.
//
//   - cost < limit: schedulable
//   - limit <= cost < limit+reserve: sticky-only
//   - cost >= limit+reserve: not schedulable
//
// Ported from Sub2API: account.go CheckWindowCostSchedulability
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
// If only one is non-nil, it is returned. If both are nil, returns nil.
//
// Ported from Sub2API: ratelimit_service.go pickSooner
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
	Window        string
	ResetAt       time.Time
	IsFableWindow bool
}

// CalculateAnthropic429ResetTime parses Anthropic's per-window rate-limit headers
// to determine which window (5h or 7d) actually triggered the 429.
//
// Ported from Sub2API: ratelimit_service.go calculateAnthropic429ResetTime
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

	// Pick the window with the latest reset (most restrictive)
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

// isAnthropicWindowExceeded checks whether a given Anthropic rate-limit window
// has been exceeded, using utilization and surpassed-threshold headers.
//
// Ported from Sub2API: ratelimit_service.go isAnthropicWindowExceeded
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
	// utilization >= 100 means window is at capacity
	util := utilizationAsPercent(utilHeader)
	return util >= 100.0
}

// ParseOpenAIRateLimitResetTime parses an OpenAI-format 429 response body,
// returning the reset time.
//
// Ported from Sub2API: ratelimit_service.go parseOpenAIRateLimitResetTime
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

	// Prefer resets_at
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

	// Fall back to resets_in_seconds from now
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
//
// Ported from Sub2API: ratelimit_service.go calculateOpenAI429ResetTime
func CalculateOpenAI429ResetTime(headers http.Header, now time.Time) *time.Time {
	if headers == nil {
		return nil
	}

	// x-ratelimit-reset-usage: "1h2m3s" format or seconds
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

	// retry-after: seconds
	if v := headers.Get("retry-after"); v != "" {
		if seconds, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && seconds > 0 {
			ts := now.Add(time.Duration(seconds) * time.Second)
			return &ts
		}
	}

	return nil
}

// parseDuration parses a duration string, supporting both "1h2m3s" and plain seconds.
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if seconds, err := strconv.Atoi(s); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}
	return time.ParseDuration(s)
}

// OpenAI image rate-limit pattern: "try again in 5s", "try again in 1.5 minutes", etc.
var openAIImageTryAgainPattern = regexp.MustCompile(`(?i)try again in\s+([0-9]+(?:\.[0-9]+)?)\s*(ms|s|sec|secs|second|seconds|m|min|mins|minute|minutes)`)

// ParseOpenAIImageTryAgainCooldown extracts the cooldown duration from an OpenAI
// image generation "try again in X" error message.
//
// Ported from Sub2API: ratelimit_service.go parseOpenAIImageTryAgainCooldown
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
//
// Ported from Sub2API: ratelimit_service.go isOpenAIImageRateLimitError
func IsOpenAIImageRateLimitError(statusCode int, body []byte) bool {
	if statusCode != 429 {
		return false
	}
	bodyStr := strings.ToLower(string(body))
	return strings.Contains(bodyStr, "image") &&
		(strings.Contains(bodyStr, "try again") || strings.Contains(bodyStr, "rate limit"))
}

// Clamp01 clamps a float64 to the [0, 1] range.
//
// Ported from Sub2API: openai_account_scheduler.go clamp01
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
//
// Ported from Sub2API: ratelimit_service.go parseRetryAfterResetTime
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

// Ensure math is imported
var _ = math.Pi
