package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/generationjob"
	"github.com/imageforge/imageforge/ent/usagerecord"

	"github.com/imageforge/imageforge/internal/batchimage"
	"github.com/imageforge/imageforge/internal/job"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/storage"
	"go.uber.org/zap"
)

// GenerationService orchestrates image generation. It creates a job record,
// delegates to the batch-image pipeline (ported from Sub2API), and tracks
// completion. The actual upstream calls go through batchimage providers.
type GenerationService struct {
	db      *ent.Client
	batch   *batchimage.PublicService
	store   storage.Storage
	queue   job.Queue
	assets  *AssetService
	usage   *UsageService
	log     *zap.Logger
}

// NewGenerationService builds a GenerationService.
func NewGenerationService(db *ent.Client, batch *batchimage.PublicService, store storage.Storage, queue job.Queue, assets *AssetService, usage *UsageService) *GenerationService {
	return &GenerationService{db: db, batch: batch, store: store, queue: queue, assets: assets, usage: usage, log: zap.NewNop()}
}

// SubmitRequest is the provider-independent generation request from a user.
type SubmitRequest struct {
	UserID         int64
	ProjectID      *int64
	Type           string
	Prompt         string
	NegativePrompt string
	Model          string
	Size           string
	Quality        string
	Style          string
	ImageCount     int
	OutputFormat   string
}

// SubmitResult returns the created job for the caller to track.
type SubmitResult struct {
	JobID       string `json:"job_id"`
	Status      string `json:"status"`
	Provider    string `json:"provider"`
	Model       string `json:"model"`
	CreatedAt   string `json:"created_at"`
}

// Submit creates a job, sends it to the batch-image pipeline, and enqueues a poll.
func (s *GenerationService) Submit(ctx context.Context, req SubmitRequest) (*SubmitResult, error) {
	if req.Prompt == "" {
		return nil, errors.New(errors.ErrInvalidRequest, "prompt is required")
	}
	if req.ImageCount <= 0 {
		req.ImageCount = 1
	}
	if req.ImageCount > 4 {
		req.ImageCount = 4
	}
	if req.OutputFormat == "" {
		req.OutputFormat = "url"
	}

	aspect := parseAspect(req.Size)
	externalID := "job_" + uuid.New().String()

	// 1. Create the job record (pending).
	created, err := s.db.GenerationJob.Create().
		SetExternalID(externalID).
		SetUserID(req.UserID).
		SetNillableProjectID(req.ProjectID).
		SetType(generationjob.Type(req.Type)).
		SetStatus(generationjob.StatusPending).
		SetModel(req.Model).
		SetPrompt(req.Prompt).
		SetNegativePrompt(req.NegativePrompt).
		SetAspectRatio(aspect).
		SetImageCount(req.ImageCount).
		SetOutputFormat(req.OutputFormat).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to create job", err)
	}
	_ = created

	// 2. Submit to the batch-image pipeline.
	account := &batchimage.Account{
		ID:       1, // Placeholder; full port links to ent.Account
		Platform: "gemini",
		Credentials: map[string]string{"api_key": ""},
	}

	result, err := s.batch.Submit(ctx, batchimage.SubmitInput{
		UserID:      req.UserID,
		Provider:    "gemini_api",
		Model:       req.Model,
		TaskName:    externalID,
		Prompt:      req.Prompt,
		AspectRatio: aspect,
		ImageSize:   req.Size,
	}, account)
	if err != nil {
		_, _ = s.db.GenerationJob.Update().
			Where(generationjob.ExternalID(externalID)).
			SetStatus(generationjob.StatusFailed).
			SetErrorCode("SUBMIT_FAILED").
			SetErrorMessage(err.Error()).
			Save(ctx)
		return nil, errors.Wrap(errors.ErrImageGeneration, "failed to submit generation", err)
	}

	// 3. Update job with processing status.
	_, err = s.db.GenerationJob.Update().
		Where(generationjob.ExternalID(externalID)).
		SetStatus(generationjob.StatusProcessing).
		SetProvider(result.Provider).
		SetStartedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to update job", err)
	}

	// 4. Enqueue a poll task for the worker.
	if s.queue != nil {
		id := result.BatchID
		queueTask := &job.Task{
			Job: &job.Job{ID: id, UserID: req.UserID},
			Run: func(pollCtx context.Context) error {
				return s.pollUntilDone(pollCtx, externalID, id, account)
			},
		}
		if err := s.queue.Submit(ctx, queueTask); err != nil {
			s.log.Error("failed to enqueue poll task", zap.String("external_id", externalID), zap.Error(err))
			_, _ = s.db.GenerationJob.Update().
				Where(generationjob.ExternalID(externalID)).
				SetStatus(generationjob.StatusFailed).
				SetErrorCode("QUEUE_FULL").
				SetErrorMessage("internal error: failed to enqueue generation task").
			Save(ctx)
		}
	}

	return &SubmitResult{
		JobID:     externalID,
		Status:    string(generationjob.StatusProcessing),
		Provider:  result.Provider,
		Model:     result.Model,
		CreatedAt: result.CreatedAt.Format(time.RFC3339),
	}, nil
}

// pollUntilDone polls the batch-image pipeline until the task completes or fails.
func (s *GenerationService) pollUntilDone(ctx context.Context, externalID, batchID string, account *batchimage.Account) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeout := time.After(30 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			_, _ = s.db.GenerationJob.Update().
				Where(generationjob.ExternalID(externalID)).
				SetStatus(generationjob.StatusFailed).
				SetErrorCode("UPSTREAM_TIMEOUT").
				SetErrorMessage("image generation timed out").
				Save(ctx)
			return ctx.Err()
		case <-ticker.C:
			job, err := s.batch.Get(ctx, batchID, account)
			if err != nil {
				continue
			}
			switch job.Status {
			case batchimage.BatchImageJobStatusCompleted:
				return s.handleCompleted(ctx, externalID, job)
			case batchimage.BatchImageJobStatusFailed:
				_, _ = s.db.GenerationJob.Update().
					Where(generationjob.ExternalID(externalID)).
					SetStatus(generationjob.StatusFailed).
					SetErrorCode(toString(job.LastErrorCode)).
					SetErrorMessage(toString(job.LastErrorMessage)).
					SetCompletedAt(time.Now()).
					Save(ctx)
				return nil
			}
		}
	}
}

// handleCompleted marks the job completed and creates an asset record from the
// batch output. Storage keys are derived from the batch job's ProviderOutputRef
// (the stub equivalent of an object-storage key).
func (s *GenerationService) handleCompleted(ctx context.Context, externalID string, bj *batchimage.BatchImageJob) error {
	job, err := s.db.GenerationJob.Query().
		Where(generationjob.ExternalID(externalID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New(errors.ErrNotFound, "job not found")
		}
		return errors.Wrap(errors.ErrInternal, "failed to fetch job", err)
	}

	keys := storageKeysFromBatch(bj)

	if keys.Original != "" {
		var projectID *int64
		if job.ProjectID != 0 {
			projectID = &job.ProjectID
		}
		if _, err := s.assets.CreateFromGeneration(ctx, externalID, job.UserID, projectID, job.Prompt, job.Model, keys); err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to create asset from generation", err)
		}
	}

	usageType := string(job.Type)
	if usageType == "" {
		usageType = string(usagerecord.TypeGeneration)
	}
	if s.usage != nil {
		// Best-effort accounting; a failed record must not fail the request.
		_ = s.usage.RecordUsage(ctx, job.UserID, usageType, job.ImageCount)
	}

	_, err = s.db.GenerationJob.Update().
		Where(generationjob.ExternalID(externalID)).
		SetStatus(generationjob.StatusCompleted).
		SetCompletedAt(time.Now()).
		Save(ctx)
	return err
}


// storageKeysFromBatch derives storage keys from a completed batch-image job.
// In the full port the download pipeline uploads bytes to object storage and
// returns real keys; here the ProviderOutputRef stands in for the original.
func storageKeysFromBatch(bj *batchimage.BatchImageJob) StorageKeys {
	if bj == nil {
		return StorageKeys{}
	}
	var keys StorageKeys
	if bj.ProviderOutputRef != nil {
		keys.Original = *bj.ProviderOutputRef
	}
	return keys
}


// Get returns a job by external ID, enforcing user ownership.
func (s *GenerationService) Get(ctx context.Context, externalID string, userID int64) (*ent.GenerationJob, error) {
	j, err := s.db.GenerationJob.Query().
		Where(generationjob.ExternalID(externalID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New(errors.ErrNotFound, "job not found")
		}
		return nil, errors.Wrap(errors.ErrInternal, "failed to fetch job", err)
	}
	if j.UserID != userID {
		return nil, errors.New(errors.ErrForbidden, "access denied")
	}
	return j, nil
}

// List returns a user's jobs, newest first.
func (s *GenerationService) List(ctx context.Context, userID int64, limit, offset int) ([]*ent.GenerationJob, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.db.GenerationJob.Query().
		Where(generationjob.UserID(userID)).
		Order(ent.Desc(generationjob.FieldCreatedAt)).
		Limit(limit).
		Offset(offset).
		All(ctx)
}

// parseAspect converts a size like "1024x1024" to an aspect ratio like "1:1".
func parseAspect(size string) string {
	switch size {
	case "256x256", "512x512", "1024x1024", "1536x1536":
		return "1:1"
	case "1792x1024", "1536x1024":
		return "16:9"
	case "1024x1792", "1024x1536":
		return "9:16"
	}
	return "1:1"
}

func toString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

