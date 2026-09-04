package scheduling

import (
	"time"
)

// Account is the scheduling package's view of an AI provider account.
//
// It mirrors the subset of Sub2API's service.Account that the scheduling
// algorithms (threshold evaluation, rate-limit handling, session windows)
// actually touch. Callers convert from ent.Account via NewAccount.
type Account struct {
	ID          int64
	Name        string
	Platform    string
	Type        string
	Credentials map[string]any
	Extra       map[string]any

	Priority int
	Status   string

	RateLimitedAt    *time.Time
	RateLimitResetAt *time.Time
	OverloadUntil    *time.Time

	TempUnschedulableUntil  *time.Time
	TempUnschedulableReason string

	SessionWindowStart *time.Time
	SessionWindowEnd   *time.Time
}

// NewAccount creates a scheduling.Account from an ent.Account-like source.
// The Extra map may be nil; all map reads are nil-safe.
func NewAccount(id int64, name, platform, typ string, credentials, extra map[string]any, priority int, status string, rateLimitedAt, rateLimitResetAt, overloadUntil, tempUnschedUntil, sessionWindowStart, sessionWindowEnd *time.Time, tempUnschedReason string) *Account {
	return &Account{
		ID:                      id,
		Name:                    name,
		Platform:                platform,
		Type:                    typ,
		Credentials:             credentials,
		Extra:                   extra,
		Priority:                priority,
		Status:                  status,
		RateLimitedAt:           rateLimitedAt,
		RateLimitResetAt:        rateLimitResetAt,
		OverloadUntil:           overloadUntil,
		TempUnschedulableUntil:  tempUnschedUntil,
		TempUnschedulableReason: tempUnschedReason,
		SessionWindowStart:      sessionWindowStart,
		SessionWindowEnd:        sessionWindowEnd,
	}
}

// IsActive reports whether the account is in active status.
func (a *Account) IsActive() bool {
	return a != nil && a.Status == StatusActive
}

// IsRateLimited reports whether the account is currently rate-limited.
func (a *Account) IsRateLimited() bool {
	if a == nil || a.RateLimitResetAt == nil {
		return false
	}
	return a.RateLimitResetAt.After(time.Now())
}

// IsOverloaded reports whether the account is currently overloaded.
func (a *Account) IsOverloaded() bool {
	if a == nil || a.OverloadUntil == nil {
		return false
	}
	return a.OverloadUntil.After(time.Now())
}

// IsTempUnschedulable reports whether the account is temporarily unschedulable.
func (a *Account) IsTempUnschedulable() bool {
	if a == nil || a.TempUnschedulableUntil == nil {
		return false
	}
	return a.TempUnschedulableUntil.After(time.Now())
}

// IsOAuth reports whether the account uses OAuth authentication.
func (a *Account) IsOAuth() bool {
	return a != nil && (a.Type == AccountTypeOAuth || a.Type == AccountTypeSetupToken)
}

// IsOpenAIOAuth reports whether this is an OpenAI OAuth account.
func (a *Account) IsOpenAIOAuth() bool {
	return a != nil && a.Platform == PlatformOpenAI && a.Type == AccountTypeOAuth
}

// IsOpenAI reports whether this is an OpenAI-platform account.
func (a *Account) IsOpenAI() bool {
	return a != nil && a.Platform == PlatformOpenAI
}

// IsAnthropic reports whether this is an Anthropic-platform account.
func (a *Account) IsAnthropic() bool {
	return a != nil && a.Platform == PlatformAnthropic
}

// IsGrok reports whether this is a Grok-platform account.
func (a *Account) IsGrok() bool {
	return a != nil && a.Platform == PlatformGrok
}

// GetCredential returns the named credential value, or "" when absent.
func (a *Account) GetCredential(key string) string {
	if a == nil || len(a.Credentials) == 0 {
		return ""
	}
	if v, ok := a.Credentials[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetExtraString returns the named extra value as a string, or "" when absent.
func (a *Account) GetExtraString(key string) string {
	if a == nil || len(a.Extra) == 0 {
		return ""
	}
	if v, ok := a.Extra[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetExtraInt returns the named extra value as an int, or 0 when absent/unparseable.
func (a *Account) GetExtraInt(key string) int {
	if a == nil || len(a.Extra) == 0 {
		return 0
	}
	return parseExtraInt(a.Extra[key])
}

// GetExtraFloat64 returns the named extra value as a float64, or 0 when absent/unparseable.
func (a *Account) GetExtraFloat64(key string) float64 {
	if a == nil || len(a.Extra) == 0 {
		return 0
	}
	return parseExtraFloat64(a.Extra[key])
}
