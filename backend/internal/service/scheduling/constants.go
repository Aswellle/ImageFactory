package scheduling

import "time"

// Platform constants — mirrors Sub2API's domain.Platform* values.
// These are the upstream AI provider identifiers used throughout the
// scheduling and rate-limit algorithms.
const (
	PlatformAnthropic   = "anthropic"
	PlatformOpenAI      = "openai"
	PlatformGemini      = "gemini"
	PlatformAntigravity = "antigravity"
	PlatformGrok        = "grok"
	PlatformKimi        = "kimi"
	PlatformZhipu       = "zhipu"
	PlatformDeepseek    = "deepseek"
	PlatformComposite   = "composite"
)

// Account status constants.
const (
	StatusActive   = "active"
	StatusDisabled = "disabled"
	StatusError    = "error"
)

// Account type constants.
const (
	AccountTypeOAuth      = "oauth"
	AccountTypeSetupToken = "setup_token"
	AccountTypeAPIKey     = "api_key"
	AccountTypeBedrock    = "bedrock"
)

// AllowedSchedulingThresholdPlatforms is the list of platforms that support
// per-platform auto-pause scheduling thresholds. Platforms not in this list
// (Gemini, Antigravity, Deepseek, etc.) do not trigger threshold-based pauses;
// they use other mechanisms (balance checks, quota prechecks).
//
// Sub2API reference: internal/service/domain_constants.go
var AllowedSchedulingThresholdPlatforms = []string{
	PlatformOpenAI,
	PlatformAnthropic,
	PlatformGrok,
	PlatformKimi,
	PlatformZhipu,
}

// CN provider Coding Plan extra key suffixes.
const (
	cnExtraSuffix5hUsed       = "5h_used_percent"
	cnExtraSuffix5hReset      = "5h_reset_at"
	cnExtraSuffixWeeklyUsed   = "weekly_used_percent"
	cnExtraSuffixWeeklyReset  = "weekly_reset_at"
	cnExtraSuffixUsageUpdated = "usage_updated_at"
)

// cnExtraKey builds a provider-prefixed extra key (e.g. "kimi_5h_used_percent").
func cnExtraKey(provider, suffix string) string {
	return provider + "_" + suffix
}

// openAICodexAutoPauseStaleAfter is the maximum age of a Codex usage snapshot
// before it is considered stale and no longer keeps an account auto-paused.
const openAICodexAutoPauseStaleAfter = 6 * 30 * 24 * time.Hour // ~6 months

// Cooldown durations for various error conditions.
const (
	upstreamModelNotFoundCooldown  = 30 * time.Minute
	upstreamCodexPlanGatedCooldown = 30 * time.Minute
	tempUnschedBodyMaxBytes        = 64 << 10 // 64 KiB
	tempUnschedMessageMaxBytes     = 2048
)

// Reason strings for temp-unschedulable states.
const (
	upstreamModelNotFoundReason       = "upstream_404_model_not_found"
	upstreamCodexPlanGatedModelReason = "upstream_400_codex_plan_gated_model"
)
