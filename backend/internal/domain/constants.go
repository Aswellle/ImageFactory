// package domain holds ImageForge domain-wide constants. Provider-neutral:
// no Sub2API-specific identifiers leak here.
package domain

// Role constants.
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User status.
const (
	UserStatusActive    = "active"
	UserStatusSuspended = "suspended"
	UserStatusDeleted   = "deleted"
)

// Job types.
const (
	JobTypeGeneration = "generation"
	JobTypeEdit       = "edit"
)

// Job statuses.
const (
	JobStatusPending    = "pending"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
	JobStatusCancelled  = "cancelled"
)

// Asset statuses.
const (
	AssetStatusActive  = "active"
	AssetStatusDeleted = "deleted"
)

// Asset sources.
const (
	AssetSourceGenerated = "generated"
	AssetSourceUploaded  = "uploaded"
	AssetSourceEdited    = "edited"
)

// API key statuses.
const (
	APIKeyStatusActive  = "active"
	APIKeyStatusRevoked = "revoked"
)

// Stable, provider-independent error codes exposed to clients.
// (Full definitions live in pkg/errors.)
const (
	ErrorCodeInvalidRequest      = "INVALID_REQUEST"
	ErrorCodeUnauthorized        = "UNAUTHORIZED"
	ErrorCodeForbidden           = "FORBIDDEN"
	ErrorCodeNotFound            = "NOT_FOUND"
	ErrorCodeConflict            = "CONFLICT"
	ErrorCodeImageGeneration     = "IMAGE_GENERATION_FAILED"
	ErrorCodeImageEdit           = "IMAGE_EDIT_FAILED"
	ErrorCodeUpstreamTimeout     = "UPSTREAM_TIMEOUT"
	ErrorCodeUpstreamRateLimited = "UPSTREAM_RATE_LIMITED"
	ErrorCodeModelUnavailable    = "MODEL_UNAVAILABLE"
	ErrorCodeAuthExpired         = "AUTHENTICATION_EXPIRED"
	ErrorCodeStorageFailed       = "STORAGE_FAILED"
	ErrorCodeQuotaExceeded       = "QUOTA_EXCEEDED"
	ErrorCodeInternal            = "INTERNAL_ERROR"
)
