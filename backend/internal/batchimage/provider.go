// Package batchimage provides the batch-image generation pipeline for
// ImageForge. It is ported from Sub2API's internal/service/batch_image*.go
// and adapted to run as a top-level package within ImageForge.
//
// The BatchImageProvider interface is preserved exactly so that provider
// implementations (Gemini, Vertex, OpenAI) can be ported with minimal changes.
package batchimage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// --- Errors (self-contained; originally in Sub2API's internal/pkg/errors) ---

// Error is the application error type.
type Error struct {
	Code       string
	Message    string
	HTTPStatus int
	cause      error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.cause }

func newError(httpStatus int, code, message string) *Error {
	return &Error{Code: code, Message: message, HTTPStatus: httpStatus}
}

func (e *Error) withCause(err error) *Error {
	e.cause = err
	return e
}

var (
	ErrBatchImageJobNotFound                     = newError(http.StatusNotFound, "BATCH_IMAGE_JOB_NOT_FOUND", "batch image job not found")
	ErrBatchImageJobExists                       = newError(http.StatusConflict, "BATCH_IMAGE_JOB_EXISTS", "batch image job already exists")
	ErrBatchImageInvalidProvider                 = newError(http.StatusBadRequest, "BATCH_IMAGE_INVALID_PROVIDER", "invalid batch image provider")
	ErrBatchImageProviderUnsupportedAccount      = newError(http.StatusBadRequest, "BATCH_IMAGE_PROVIDER_UNSUPPORTED_ACCOUNT", "batch image provider does not support this account")
	ErrBatchImageProviderMissingAPIKey           = newError(http.StatusBadRequest, "BATCH_IMAGE_PROVIDER_MISSING_API_KEY", "batch image provider account is missing api key")
	ErrBatchImageProviderMissingServiceAccount   = newError(http.StatusBadRequest, "BATCH_IMAGE_PROVIDER_MISSING_SERVICE_ACCOUNT", "batch image provider account is missing service account credentials")
	ErrBatchImageProviderMissingJobName          = newError(http.StatusBadRequest, "BATCH_IMAGE_PROVIDER_MISSING_JOB_NAME", "batch image provider job name is missing")
	ErrBatchImageProviderMissingResultRef        = newError(http.StatusBadRequest, "BATCH_IMAGE_PROVIDER_MISSING_RESULT_REF", "batch image provider result reference is missing")
	ErrBatchImageProviderInlineResultUnsupported = newError(http.StatusBadRequest, "GEMINI_INLINE_BATCH_RESULT_UNSUPPORTED", "Gemini inline batch result is not supported")
	ErrBatchImageProviderInvalidInput            = newError(http.StatusBadRequest, "BATCH_IMAGE_PROVIDER_INVALID_INPUT", "invalid batch image provider input")
	ErrBatchImageProviderUnsafeCleanupPath       = newError(http.StatusBadRequest, "VERTEX_UNSAFE_CLEANUP_PATH", "unsafe batch image cleanup path")
	ErrUnsupportedCleanupTarget                  = newError(http.StatusBadRequest, "BATCH_IMAGE_PROVIDER_UNSUPPORTED_CLEANUP_TARGET", "unsupported batch image cleanup target")
)

// --- Provider interface (preserved from Sub2API) ---

// BatchImageProvider is the contract for an upstream image-generation provider.
type BatchImageProvider interface {
	Name() string
	SupportsAccount(account *Account) bool
	Submit(ctx context.Context, job *BatchImageJob, account *Account, input BatchImageInput) (*BatchProviderJob, error)
	Get(ctx context.Context, job *BatchImageJob, account *Account) (*BatchProviderStatus, error)
	Cancel(ctx context.Context, job *BatchImageJob, account *Account) error
	OpenResult(ctx context.Context, job *BatchImageJob, account *Account) (io.ReadCloser, string, error)
	Cleanup(ctx context.Context, job *BatchImageJob, account *Account, target CleanupTarget) error
}

// BatchImageProviderRegistry holds named providers.
type BatchImageProviderRegistry struct {
	providers map[string]BatchImageProvider
}

// NewBatchImageProviderRegistry builds a registry from the given providers.
func NewBatchImageProviderRegistry(providers ...BatchImageProvider) *BatchImageProviderRegistry {
	r := &BatchImageProviderRegistry{providers: make(map[string]BatchImageProvider, len(providers))}
	for _, provider := range providers {
		if provider == nil || strings.TrimSpace(provider.Name()) == "" {
			continue
		}
		r.providers[provider.Name()] = provider
	}
	return r
}

// NewRegistry is a convenience alias.
func NewRegistry(providers ...BatchImageProvider) *BatchImageProviderRegistry {
	return NewBatchImageProviderRegistry(providers...)
}

// Get returns a provider by name.
func (r *BatchImageProviderRegistry) Get(provider string) (BatchImageProvider, bool) {
	if r == nil {
		return nil, false
	}
	p, ok := r.providers[provider]
	return p, ok
}

// MustGet returns a provider by name or an error.
func (r *BatchImageProviderRegistry) MustGet(provider string) (BatchImageProvider, error) {
	p, ok := r.Get(provider)
	if !ok {
		return nil, ErrBatchImageInvalidProvider
	}
	return p, nil
}

// --- Input / output types ---

type BatchImageInput struct {
	BatchID          string
	Model            string
	DisplayName      string
	Items            []BatchImageInputItem
	ResponseMimeType string
	AspectRatio      string
	ImageSize        string
	Metadata         map[string]string
}

type BatchImageInputItem struct {
	CustomID        string
	Prompt          string
	ReferenceImages []BatchImageReference
}

type BatchImageReference struct {
	ID       string
	Type     string
	MimeType string
	Data     []byte
	FileURI  string
}

type BatchProviderJob struct {
	ProviderJobName  string
	ProviderInputRef string
	ProviderOutputRef string
	RawState         string
}

// BatchProviderInternalState is the normalized provider state.
type BatchProviderInternalState string

const (
	BatchProviderStateQueued    BatchProviderInternalState = "queued"
	BatchProviderStateRunning   BatchProviderInternalState = "running"
	BatchProviderStateSucceeded BatchProviderInternalState = "succeeded"
	BatchProviderStateFailed    BatchProviderInternalState = "failed"
	BatchProviderStateCancelled BatchProviderInternalState = "cancelled"
	BatchProviderStateExpired   BatchProviderInternalState = "expired"
)

// BatchProviderStatus is the normalized poll result.
type BatchProviderStatus struct {
	RawState             string
	InternalState        BatchProviderInternalState
	Done                 bool
	ProviderOutputRef    string
	ErrorCode            string
	ErrorMessage         string
	SuggestedRequeueAfter time.Duration
}

// CleanupTarget identifies what to clean up.
type CleanupTarget string

const (
	CleanupTargetInput  CleanupTarget = "input"
	CleanupTargetOutput CleanupTarget = "output"
	CleanupTargetAll    CleanupTarget = "all"
)

// --- Provider name constants ---

const (
	BatchImageProviderGeminiAPI = "gemini_api"
	BatchImageProviderVertex    = "vertex"
)

// --- Job status constants ---

const (
	BatchImageJobStatusCreated       = "created"
	BatchImageJobStatusUploading     = "uploading"
	BatchImageJobStatusSubmitted     = "submitted"
	BatchImageJobStatusRunning       = "running"
	BatchImageJobStatusIndexing      = "indexing"
	BatchImageJobStatusSettling      = "settling"
	BatchImageJobStatusCompleted     = "completed"
	BatchImageJobStatusFailed        = "failed"
	BatchImageJobStatusCancelled     = "cancelled"
	BatchImageJobStatusOutputDeleted = "output_deleted"
)

const (
	BatchImageItemStatusPending   = "pending"
	BatchImageItemStatusSuccess   = "success"
	BatchImageItemStatusFailed    = "failed"
	BatchImageItemStatusCancelled = "cancelled"
)

// --- Helper functions ---

func batchImageProviderJobName(job *BatchImageJob) string {
	if job == nil || job.ProviderJobName == nil {
		return ""
	}
	return strings.TrimSpace(*job.ProviderJobName)
}

func batchImageProviderInputRef(job *BatchImageJob) string {
	if job == nil || job.ProviderInputRef == nil {
		return ""
	}
	return strings.TrimSpace(*job.ProviderInputRef)
}

func batchImageProviderOutputRef(job *BatchImageJob) string {
	if job == nil || job.ProviderOutputRef == nil {
		return ""
	}
	return strings.TrimSpace(*job.ProviderOutputRef)
}

func batchImageProviderAPIKey(account *Account) string {
	if account == nil {
		return ""
	}
	return strings.TrimSpace(account.GetCredential("api_key"))
}

func batchImageProviderInputError(format string, args ...any) error {
	return ErrBatchImageProviderInvalidInput.withCause(fmt.Errorf(format, args...))
}
