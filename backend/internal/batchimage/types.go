package batchimage

import (
	"context"
	"time"
)

// Account is a minimal stub of Sub2API's service.Account.
// In the full port, this wraps or aliases ent.Account.
type Account struct {
	ID       int64
	Platform string
	Credentials map[string]string
}

func (a *Account) GetCredential(key string) string {
	if a == nil {
		return ""
	}
	return a.Credentials[key]
}

// BatchImageJob is a minimal stub of Sub2API's service.BatchImageJob.
// In the full port, this mirrors ent.BatchImageJob fields.
type BatchImageJob struct {
	ID              int64
	BatchID         string
	UserID          int64
	APIKeyID        *int64
	AccountID       *int64
	Provider        string
	Model           string
	TaskName        string
	ParentBatchID   *string
	AspectRatio     string
	ImageSize       string
	Status          string
	ProviderJobName *string
	ProviderInputRef *string
	ProviderOutputRef *string
	ItemCount       int
	SuccessCount    int
	FailCount       int
	CancelledCount  int
	EstimatedCost   float64
	HoldAmount      *float64
	ActualCost      *float64
	BillableUnitPrice float64
	HoldUnitPrice   float64
	Currency        string
	HoldID          *string
	IdempotencyKey  *string
	RequestHash     *string
	ManifestHash    *string
	SessionID       *string
	PricingSnapshotVersion int
	RetryCount      int
	Version         int
	LastErrorCode   *string
	LastErrorMessage *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	SubmittedAt     *time.Time
	StartedAt       *time.Time
	FinishedAt      *time.Time
	SettledAt       *time.Time
	OutputExpiresAt *time.Time
	InputDeletedAt  *time.Time
	OutputDeletedAt *time.Time
}
// ---------------------------------------------------------------------------
// Additional fields used by the settlement/worker/download/cleanup pipeline.
// The existing BatchImageJob is extended (not replaced) to stay compatible
// with the simplified PublicService facade.
// ---------------------------------------------------------------------------

// BatchImageItem is a minimal stub of Sub2API's service.BatchImageItem.
type BatchImageItem struct {
	ID                   int64
	JobID                string
	CustomID             string
	Status               string
	RequestHash          *string
	PromptPreview        *string
	ProviderSourceObject *string
	SourceLineNumber     *int
	SourceByteOffset     *int64
	SourceByteLength     *int64
	MimeType             *string
	FileExtension        *string
	ImageCount           int
	ErrorCode            *string
	ErrorMessage         *string
	BilledAmount         *float64
	CreatedAt            time.Time
	IndexedAt            *time.Time
}

// BatchImageOwner identifies the caller for owner-scoped operations.
type BatchImageOwner struct {
	UserID   int64
	APIKeyID int64
}

// ---------------------------------------------------------------------------
// Repository + queue contracts (ported from Sub2API). Implementations wrap
// ent.Client and live outside this package; the pipeline depends only on these
// interfaces so it stays testable and free of raw SQL.
// ---------------------------------------------------------------------------

// CreateBatchImageJobParams captures the fields needed to persist a new job.
type CreateBatchImageJobParams struct {
	BatchID         string
	UserID          int64
	APIKeyID        *int64
	AccountID       *int64
	Provider        string
	Model           string
	TaskName        string
	ParentBatchID   *string
	HoldID          *string
	IdempotencyKey  *string
	RequestHash     *string
	ManifestHash    *string
	SessionID       *string
	ItemCount       int
	EstimatedCost   float64
	HoldAmount      *float64
	Currency        string
	PricingSnapshot int
	HoldUnitPrice   float64
	BillableUnitPrice float64
	RetryCount      int
}

// CreateBatchImageItemParams captures the fields needed to persist a new item.
type CreateBatchImageItemParams struct {
	JobID                string
	CustomID             string
	Status               string
	RequestHash          *string
	PromptPreview        *string
	ProviderSourceObject *string
	SourceLineNumber     *int
	SourceByteOffset     *int64
	SourceByteLength     *int64
	MimeType             *string
	FileExtension        *string
	ImageCount           int
	ErrorCode            *string
	ErrorMessage         *string
	BilledAmount         *float64
	IndexedAt            *time.Time
}

// BatchImageItemFilter selects a subset of items for a job.
type BatchImageItemFilter struct {
	Status string
	Limit  int
	Offset int
}

// BatchImageJobFilter selects a subset of jobs for a user.
type BatchImageJobFilter struct {
	Status       string
	TaskNameLike string
	CreatedAfter *time.Time
	Limit        int
	Offset       int
}

// BatchImageTransitionOptions carries event metadata for a status transition.
type BatchImageTransitionOptions struct {
	EventType    string
	EventPayload any
	ErrorCode    *string
	ErrorMessage *string
	Now          *time.Time
}

// MarkBatchImageJobSettledParams captures the fields written when a job settles.
type MarkBatchImageJobSettledParams struct {
	BatchID         string
	ActualCost      float64
	ManifestHash    string
	EventPayload    any
	Now             *time.Time
	OutputExpiresAt *time.Time
}

// UpdateBatchImageJobProviderSubmitParams captures post-submit provider refs.
type UpdateBatchImageJobProviderSubmitParams struct {
	BatchID           string
	ProviderJobName   string
	ProviderInputRef  string
	ProviderOutputRef string
	EventPayload      any
}

// BatchImageCounts aggregates success/fail tallies for a job.
type BatchImageCounts struct {
	SuccessCount int
	FailCount    int
}

// BatchImageEvent is an append-only audit record for a job.
type BatchImageEvent struct {
	ID        int64
	JobID     string
	EventType string
	Payload   []byte
	EventHash *string
	CreatedAt time.Time
}

// BatchImageRepository is the data-access contract for the batch-image
// pipeline. It mirrors Sub2API's interface but is trimmed to the methods the
// ImageForge pipeline actually uses.
type BatchImageRepository interface {
	CreateBatchImageJob(ctx context.Context, params CreateBatchImageJobParams) (*BatchImageJob, error)
	GetBatchImageJobByBatchID(ctx context.Context, batchID string) (*BatchImageJob, error)
	GetBatchImageJobByID(ctx context.Context, id int64) (*BatchImageJob, error)
	ListBatchImageJobsForOwner(ctx context.Context, userID int64, filter BatchImageJobFilter) ([]*BatchImageJob, error)
	TransitionBatchImageJobStatus(ctx context.Context, batchID, toStatus string, opts BatchImageTransitionOptions) error
	UpdateBatchImageJobProviderOutputRef(ctx context.Context, batchID, providerOutputRef string) error
	UpdateBatchImageJobProviderSubmit(ctx context.Context, params UpdateBatchImageJobProviderSubmitParams) error
	MarkBatchImageJobSettled(ctx context.Context, params MarkBatchImageJobSettledParams) error
	SetBatchImageJobSettlementFailed(ctx context.Context, batchID, code, message string) (int, error)
	CreateBatchImageItem(ctx context.Context, params CreateBatchImageItemParams) (*BatchImageItem, error)
	ReplaceBatchImageItemsForJob(ctx context.Context, batchID string, items []CreateBatchImageItemParams, counts BatchImageCounts) error
	ListBatchImageItemsForOwner(ctx context.Context, userID int64, batchID string, filter BatchImageItemFilter) ([]*BatchImageItem, error)
	ListBatchImageItemsForDownload(ctx context.Context, batchID string, status string, limit int) ([]*BatchImageItem, error)
	GetBatchImageJobForDownload(ctx context.Context, userID int64, batchID string) (*BatchImageJob, error)
	GetBatchImageItemForDownload(ctx context.Context, batchID, customID string) (*BatchImageItem, error)
	ListBatchImageJobsDueForInputCleanup(ctx context.Context, cutoff time.Time, limit int) ([]*BatchImageJob, error)
	ListBatchImageJobsDueForOutputCleanup(ctx context.Context, now time.Time, limit int) ([]*BatchImageJob, error)
	MarkBatchImageInputDeleted(ctx context.Context, batchID string, deletedAt time.Time) error
	MarkBatchImageOutputDeleted(ctx context.Context, batchID string, deletedAt time.Time) error
	MarkBatchImageJobUserDeleted(ctx context.Context, userID int64, batchID string, deletedAt time.Time) error
	SetBatchImageOutputExpiresAt(ctx context.Context, batchID string, expiresAt time.Time) error
	RecordBatchImageCleanupFailure(ctx context.Context, batchID, code, message string) error
	AppendBatchImageEvent(ctx context.Context, batchID, eventType string, payload any) error
}

// BatchImageJobLock guards a job during processing.
type BatchImageJobLock interface {
	Release(ctx context.Context) error
}

// ReservedBatchImageJob is a dequeued job ticket.
type ReservedBatchImageJob struct {
	BatchID string
}

// BatchImageQueue is the job-dispatch abstraction (ported from Sub2API).
// Implementations may be Redis-backed or in-memory; the worker depends only
// on this interface.
type BatchImageQueue interface {
	Enqueue(ctx context.Context, batchID string) error
	Reserve(ctx context.Context, blockTimeout time.Duration) (ReservedBatchImageJob, error)
	RequeueAfter(ctx context.Context, batchID string, delay time.Duration) error
	Ack(ctx context.Context, batchID string) error
	Heartbeat(ctx context.Context, batchID string) error
	MoveDueDelayedToReady(ctx context.Context, limit int) (int, error)
	RecoverStaleActive(ctx context.Context, staleAfter time.Duration, limit int) (int, error)
	TryAcquireJobLock(ctx context.Context, batchID string, ttl time.Duration) (BatchImageJobLock, bool, error)
}

// BatchImageAccountResolver resolves an account ID to an Account.
type BatchImageAccountResolver interface {
	ResolveBatchImageAccount(ctx context.Context, accountID int64) (*Account, error)
}

// BatchImagePricingResolver resolves the unit price for a job's model.
type BatchImagePricingResolver interface {
	BatchImageUnitPrice(ctx context.Context, job *BatchImageJob) (float64, error)
}

// UsageBillingRepository captures/releases balance holds for a job.
type UsageBillingRepository interface {
	CaptureBatchImageBalanceHold(ctx context.Context, job *BatchImageJob, actualCost float64, manifestHash string) error
	ReleaseBatchImageBalanceHold(ctx context.Context, job *BatchImageJob, requestHash string) error
}

// APIKeyAuthCacheInvalidator drops cached auth entries after balance changes.
type APIKeyAuthCacheInvalidator interface {
	InvalidateAuthCacheByUserID(ctx context.Context, userID int64)
}

// BatchImageProcessResult tells the worker what to do with a job it processed.
type BatchImageProcessResult struct {
	RequeueAfter time.Duration
	Terminal     bool
}

// BatchImageProcessor drives a single job through the provider poll loop.
type BatchImageProcessor interface {
	Process(ctx context.Context, batchID string) (BatchImageProcessResult, error)
}

// BatchImagePublicModel describes a model available for batch generation.
type BatchImagePublicModel struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Provider string `json:"provider"`
}

// BatchImagePublicModelsResponse is the ListModels response envelope.
type BatchImagePublicModelsResponse struct {
	Object string                `json:"object"`
	Data   []BatchImagePublicModel `json:"data"`
}

// BatchImagePublicItem is the public view of a batch item.
type BatchImagePublicItem struct {
	Object    string  `json:"object"`
	CustomID  string  `json:"custom_id"`
	Status    string  `json:"status"`
	ImageCount int    `json:"image_count"`
	ErrorCode *string `json:"error_code,omitempty"`
}

// BatchImagePublicItemsResponse is the ListItems response envelope.
type BatchImagePublicItemsResponse struct {
	Object  string                 `json:"object"`
	Data    []BatchImagePublicItem `json:"data"`
	HasMore bool                   `json:"has_more"`
}

// BatchImagePublicBatch is the public view of a batch job.
type BatchImagePublicBatch struct {
	Object        string     `json:"object"`
	BatchID       string     `json:"batch_id"`

	Status        string     `json:"status"`
	Model         string     `json:"model"`
	Provider      string     `json:"provider"`
	TaskName      string     `json:"task_name"`
	ItemCount     int        `json:"item_count"`
	SuccessCount  int        `json:"success_count"`
	FailCount     int        `json:"fail_count"`
	CreatedAt     time.Time  `json:"created_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

// ---------------------------------------------------------------------------
// Shared helpers (ported from Sub2API service helpers).
// ---------------------------------------------------------------------------

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func truncateBatchImageMessage(s string, maxLen int) string {
	if maxLen <= 0 {
		return s
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
