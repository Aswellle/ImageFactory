// Copyright 2024 ImageForge
// PublicService 编排批处理图片生成。
// 状态持久化到 GenerationJob.batch_image_state，重启后可恢复。

package batchimage

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/imageforge/imageforge/ent"
)

// SubmitInput is the provider-independent submission request.
type SubmitInput struct {
	UserID    int64
	AccountID *int64
	Provider  string
	Model     string
	TaskName  string
	Prompt    string
	AspectRatio string
	ImageSize   string
}

// SubmitResult returns the created job's tracking ID.
type SubmitResult struct {
	BatchID   string    `json:"batch_id"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// PublicService orchestrates batch-image generation.
type PublicService struct {
	registry     *BatchImageProviderRegistry
	mu           sync.RWMutex
	jobs         map[string]*BatchImageJob
	items        map[string][]BatchImageItem
	geminiAPIKey string
	db           *ent.Client // 可选：持久化存储
}

// NewPublicService creates a PublicService without persistence.
func NewPublicService(geminiAPIKey string) *PublicService {
	return &PublicService{
		registry: NewRegistry(
			NewGeminiAPIBatchImageProvider(nil),
		),
		jobs:         make(map[string]*BatchImageJob),
		items:        make(map[string][]BatchImageItem),
		geminiAPIKey: geminiAPIKey,
	}
}

// NewPublicServiceWithDB creates a PublicService with database persistence.
func NewPublicServiceWithDB(geminiAPIKey string, db *ent.Client) *PublicService {
	s := NewPublicService(geminiAPIKey)
	s.db = db
	s.restoreFromDB(context.Background())
	return s
}

// NewPublicServiceWithRegistry creates a PublicService with a custom registry.
func NewPublicServiceWithRegistry(registry *BatchImageProviderRegistry, geminiAPIKey string) *PublicService {
	return &PublicService{
		registry:     registry,
		jobs:         make(map[string]*BatchImageJob),
		items:        make(map[string][]BatchImageItem),
		geminiAPIKey: geminiAPIKey,
	}
}

// Submit creates a batch-image job and dispatches it to the selected provider.
func (s *PublicService) Submit(ctx context.Context, input SubmitInput, account *Account) (*SubmitResult, error) {
	if input.Prompt == "" {
		return nil, ErrBatchImageProviderInvalidInput
	}

	provider, ok := s.registry.Get(input.Provider)
	if !ok {
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
		BatchID:     batchID,
		UserID:      input.UserID,
		AccountID:   input.AccountID,
		Provider:    input.Provider,
		Model:       input.Model,
		TaskName:    input.TaskName,
		Status:      BatchImageJobStatusCreated,
		AspectRatio: input.AspectRatio,
		ImageSize:   input.ImageSize,
		ItemCount:   1,
		CreatedAt:   now,
		UpdatedAt:   now,
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
		job.LastErrorMessage = new(string)
		*job.LastErrorMessage = err.Error()
		s.mu.Lock()
		s.jobs[batchID] = job
		s.items[batchID] = []BatchImageItem{{JobID: batchID, CustomID: "item_0", Status: BatchImageItemStatusFailed, CreatedAt: now}}
		s.mu.Unlock()
		s.persistState(batchID)
		return nil, err
	}

	job.Status = BatchImageJobStatusSubmitted
	job.ProviderJobName = new(string)
	*job.ProviderJobName = providerJob.ProviderJobName
	job.ProviderInputRef = new(string)
	*job.ProviderInputRef = providerJob.ProviderInputRef
	job.StartedAt = &now
	s.mu.Lock()
	s.jobs[batchID] = job
	s.items[batchID] = []BatchImageItem{{JobID: batchID, CustomID: "item_0", Status: BatchImageItemStatusPending, CreatedAt: now}}
	s.mu.Unlock()
	s.persistState(batchID)

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
		// Try to restore from DB
		job = s.getFromDB(batchID)
		if job == nil {
			return nil, ErrBatchImageJobNotFound
		}
		s.mu.Lock()
		s.jobs[batchID] = job
		s.mu.Unlock()
	}

	provider, ok := s.registry.Get(job.Provider)
	if !ok {
		return job, nil
	}

	status, err := provider.Get(ctx, job, account)
	if err != nil {
		return job, nil
	}

	switch status.InternalState {
	case BatchProviderStateSucceeded:
		job.Status = BatchImageJobStatusCompleted
		now := time.Now()
		job.FinishedAt = &now
		if status.ProviderOutputRef != "" {
			job.ProviderOutputRef = &status.ProviderOutputRef
		}
		s.markItemStatus(batchID, BatchImageItemStatusSuccess)
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
		s.markItemStatus(batchID, BatchImageItemStatusFailed)
	case BatchProviderStateCancelled:
		job.Status = BatchImageJobStatusCancelled
		s.markItemStatus(batchID, BatchImageItemStatusCancelled)
	case BatchProviderStateRunning:
		job.Status = BatchImageJobStatusRunning
	}
	job.UpdatedAt = time.Now()
	s.mu.Lock()
	s.jobs[batchID] = job
	s.mu.Unlock()
	s.persistState(batchID)

	return job, nil
}

// List returns all tracked jobs for a user.
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
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[batchID]
	if !ok {
		return ErrBatchImageJobNotFound
	}

	provider, ok := s.registry.Get(job.Provider)
	if !ok {
		return ErrBatchImageInvalidProvider
	}

	if err := provider.Cancel(ctx, job, account); err != nil {
		return err
	}

	job.Status = BatchImageJobStatusCancelled
	job.UpdatedAt = time.Now()
	s.jobs[batchID] = job
	s.persistState(batchID)
	return nil
}

// ListModels returns the catalog of models available for batch generation.
func (s *PublicService) ListModels() *BatchImagePublicModelsResponse {
	models := []BatchImagePublicModel{}
	knownProviders := []string{BatchImageProviderGeminiAPI, BatchImageProviderVertex}
	for _, name := range knownProviders {
		provider, ok := s.registry.Get(name)
		if !ok {
			continue
		}
		_ = provider
		models = append(models, BatchImagePublicModel{
			ID:       name,
			Object:   "model",
			Provider: name,
		})
	}
	return &BatchImagePublicModelsResponse{Object: "list", Data: models}
}

// ListItems returns the items tracked for a batch.
func (s *PublicService) ListItems(batchID string, owner BatchImageOwner) (*BatchImagePublicItemsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items, ok := s.items[batchID]
	if !ok {
		return nil, ErrBatchImageJobNotFound
	}
	// Convert to public view
	publicItems := make([]BatchImagePublicItem, 0, len(items))
	for _, item := range items {
		publicItems = append(publicItems, BatchImagePublicItem{
			Object:   "batch_item",
			CustomID: item.CustomID,
			Status:   item.Status,
		})
	}
	return &BatchImagePublicItemsResponse{Object: "list", Data: publicItems}, nil
}

// Delete removes a batch and its items from the tracking store.
func (s *PublicService) Delete(batchID string, owner BatchImageOwner) error {
	s.mu.Lock()
	delete(s.jobs, batchID)
	delete(s.items, batchID)
	s.mu.Unlock()

	if s.db != nil {
		_, _ = s.db.ExecContext(context.Background(),
			"UPDATE generation_jobs SET batch_image_state = NULL WHERE sub2api_task_id = $1",
			batchID)
	}
	return nil
}

// markItemStatus updates the status of every item for a batch.
func (s *PublicService) markItemStatus(batchID, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := s.items[batchID]
	for i := range items {
		items[i].Status = status
	}
	s.items[batchID] = items
	s.persistState(batchID)
}

// --- Persistence helpers ---

// persistState 将作业和条目状态持久化到数据库。
func (s *PublicService) persistState(batchID string) {
	if s.db == nil {
		return
	}
	s.mu.RLock()
	job := s.jobs[batchID]
	items := s.items[batchID]
	s.mu.RUnlock()

	if job == nil {
		return
	}

	state := map[string]any{
		"job":   job,
		"items": items,
	}
	data, err := json.Marshal(state)
	if err != nil {
		return
	}

	go func() {
		_, _ = s.db.ExecContext(context.Background(),
			`UPDATE generation_jobs SET batch_image_state = $1, updated_at = NOW()
			 WHERE sub2api_task_id = $2`,
			string(data), batchID)
	}()
}

// getFromDB 从数据库恢复作业状态。
func (s *PublicService) getFromDB(batchID string) *BatchImageJob {
	if s.db == nil {
		return nil
	}
	var data string
	rows, err := s.db.QueryContext(context.Background(),
		"SELECT batch_image_state FROM generation_jobs WHERE sub2api_task_id = $1", batchID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	if !rows.Next() {
		return nil
	}
	if err := rows.Scan(&data); err != nil || data == "" {
		return nil
	}
	var state struct {
		Job   *BatchImageJob  `json:"job"`
		Items []BatchImageItem `json:"items"`
	}
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil
	}
	if state.Job != nil {
		s.items[batchID] = state.Items
	}
	return state.Job
}

// restoreFromDB 启动时从数据库恢复所有活跃的批处理作业。
func (s *PublicService) restoreFromDB(ctx context.Context) {
	if s.db == nil {
		return
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT sub2api_task_id, batch_image_state FROM generation_jobs
		 WHERE batch_image_state IS NOT NULL AND batch_image_state <> '{}'
		   AND status IN ('pending', 'processing')`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var batchID, data string
		if err := rows.Scan(&batchID, &data); err != nil {
			continue
		}
		var state struct {
			Job   *BatchImageJob  `json:"job"`
			Items []BatchImageItem `json:"items"`
		}
		if err := json.Unmarshal([]byte(data), &state); err != nil {
			continue
		}
		s.jobs[batchID] = state.Job
		s.items[batchID] = state.Items
	}
}

// UpdateJobStatus 更新作业状态并持久化（供 Worker 调用）。
func (s *PublicService) UpdateJobStatus(batchID, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[batchID]
	if !ok {
		return
	}
	now := time.Now()
	job.Status = status
	job.UpdatedAt = now
	if status == BatchImageJobStatusCompleted || status == BatchImageJobStatusFailed {
		job.FinishedAt = &now
	}
	s.jobs[batchID] = job
	s.persistState(batchID)
}
