package batchimage

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PublicService orchestrates batch-image generation. It is a simplified,
// self-contained version of Sub2API's batch_image_public.go, adapted for
// ImageForge. It uses the BatchImageProvider interface to delegate upstream
type PublicService struct {
	registry *BatchImageProviderRegistry
	mu       sync.RWMutex
	jobs     map[string]*BatchImageJob
	geminiAPIKey string
}

// NewPublicService builds a PublicService with the default provider registry.
func NewPublicService(geminiAPIKey string) *PublicService {
	s := &PublicService{
		registry: NewRegistry(
			NewGeminiAPIBatchImageProvider(nil),
		),
		jobs:         make(map[string]*BatchImageJob),
		geminiAPIKey: geminiAPIKey,
	}
	return s
}

// NewPublicServiceWithRegistry builds a PublicService with a custom registry.
func NewPublicServiceWithRegistry(registry *BatchImageProviderRegistry, geminiAPIKey string) *PublicService {
	return &PublicService{
		registry:     registry,
		jobs:         make(map[string]*BatchImageJob),
		geminiAPIKey: geminiAPIKey,
	}
}

// SubmitInput is the provider-independent submission request.
type SubmitInput struct {
	UserID    int64
	AccountID *int64
	Provider  string
	Model     string
	TaskName  string
	Prompt    string
	// AspectRatio e.g. "1:1", "16:9"
	AspectRatio string
	// ImageSize e.g. "1024x1024"
	ImageSize string
}

// SubmitResult returns the created job's tracking ID.
type SubmitResult struct {
	BatchID   string    `json:"batch_id"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Submit creates a batch-image job and dispatches it to the selected provider.
func (s *PublicService) Submit(ctx context.Context, input SubmitInput, account *Account) (*SubmitResult, error) {
	if input.Prompt == "" {
		return nil, ErrBatchImageProviderInvalidInput
	}

	provider, ok := s.registry.Get(input.Provider)
	if !ok {
		// Fall back to the first available provider.
		provider, ok = s.registry.Get(BatchImageProviderGeminiAPI)
		if !ok {
			return nil, ErrBatchImageInvalidProvider
		}
	}

	if !provider.SupportsAccount(account) {
		return nil, ErrBatchImageProviderUnsupportedAccount
	}

	batchID := "imgbatch_" + uuid.New().String()
	now := time.Now()

	job := &BatchImageJob{
		BatchID:          batchID,
		UserID:           input.UserID,
		AccountID:        input.AccountID,
		Provider:         input.Provider,
		Model:            input.Model,
		TaskName:         input.TaskName,
		Status:           BatchImageJobStatusCreated,
		AspectRatio:      input.AspectRatio,
		ImageSize:        input.ImageSize,
		ItemCount:        1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	batchInput := BatchImageInput{
		BatchID:     batchID,
		Model:       input.Model,
		DisplayName: input.TaskName,
		AspectRatio: input.AspectRatio,
		ImageSize:   input.ImageSize,
		Items: []BatchImageInputItem{
			{CustomID: "item_0", Prompt: input.Prompt},
		},
	}

	providerJob, err := provider.Submit(ctx, job, account, batchInput)
	if err != nil {
		job.Status = BatchImageJobStatusFailed
		job.LastErrorMessage = ptr(err.Error())
		s.mu.Lock()
		s.jobs[batchID] = job
		s.mu.Unlock()
		return nil, err
	}

	job.Status = BatchImageJobStatusSubmitted
	job.ProviderJobName = &providerJob.ProviderJobName
	job.ProviderInputRef = &providerJob.ProviderInputRef
	job.StartedAt = &now
	s.mu.Lock()
	s.jobs[batchID] = job
	s.mu.Unlock()


	return &SubmitResult{
		BatchID:   batchID,
		Provider:  input.Provider,
		Model:     input.Model,
		Status:    job.Status,
		CreatedAt: job.CreatedAt,
	}, nil
}

// Get polls the current status of a batch-image job.
func (s *PublicService) Get(ctx context.Context, batchID string, account *Account) (*BatchImageJob, error) {
	s.mu.RLock()
	job, ok := s.jobs[batchID]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrBatchImageJobNotFound
	}

	provider, ok := s.registry.Get(job.Provider)
	if !ok {
		return job, nil
	}

	status, err := provider.Get(ctx, job, account)
	if err != nil {
		return job, nil // Return last known status on poll error
	}

	// Update job status based on provider status.
	switch status.InternalState {
	case BatchProviderStateSucceeded:
		job.Status = BatchImageJobStatusCompleted
		now := time.Now()
		job.FinishedAt = &now
		if status.ProviderOutputRef != "" {
			job.ProviderOutputRef = &status.ProviderOutputRef
		}
	case BatchProviderStateFailed:
		job.Status = BatchImageJobStatusFailed
		now := time.Now()
		job.FinishedAt = &now
		if status.ErrorCode != "" {
			job.LastErrorCode = &status.ErrorCode
		}
		if status.ErrorMessage != "" {
			job.LastErrorMessage = &status.ErrorMessage
		}
	case BatchProviderStateCancelled:
		job.Status = BatchImageJobStatusCancelled
	case BatchProviderStateRunning:
		job.Status = BatchImageJobStatusRunning
	}
	job.UpdatedAt = time.Now()
	s.mu.Lock()
	s.jobs[batchID] = job
	s.mu.Unlock()

	return job, nil
}

// List returns all tracked jobs for a user.
// TODO: add pagination once job count grows.
func (s *PublicService) List(userID int64) []*BatchImageJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*BatchImageJob
	for _, job := range s.jobs {
		if job.UserID == userID {
			result = append(result, job)
		}
	}
	return result
}

// Cancel cancels a batch-image job.
func (s *PublicService) Cancel(ctx context.Context, batchID string, account *Account) error {
	s.mu.RLock()
	job, ok := s.jobs[batchID]
	if !ok {
		s.mu.RUnlock()
		return ErrBatchImageJobNotFound
	}
	provider, ok := s.registry.Get(job.Provider)
	s.mu.RUnlock()
	if !ok {
		return ErrBatchImageInvalidProvider
	}
	return provider.Cancel(ctx, job, account)
}



// ptr returns a pointer to the given string value.
// Note: new(T) zero-initializes; for non-zero values we need an explicit helper.
func ptr[T any](v T) *T {
	return &v
}
