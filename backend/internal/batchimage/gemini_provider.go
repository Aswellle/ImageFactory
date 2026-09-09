package batchimage

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"
)

const defaultGeminiBatchRequeueAfter = 30 * time.Second

// GeminiAPIBatchImageProvider implements BatchImageProvider for Gemini API.
type GeminiAPIBatchImageProvider struct {
	client GeminiBatchClient
}

// NewGeminiAPIBatchImageProvider builds a Gemini provider.
func NewGeminiAPIBatchImageProvider(client GeminiBatchClient) *GeminiAPIBatchImageProvider {
	if client == nil {
		client = NewGeminiBatchHTTPClient("", "")
	}
	return &GeminiAPIBatchImageProvider{client: client}
}

func (p *GeminiAPIBatchImageProvider) Name() string {
	return BatchImageProviderGeminiAPI
}

func (p *GeminiAPIBatchImageProvider) SupportsAccount(account *Account) bool {
	if account == nil {
		return false
	}
	return account.Platform == "gemini"
}

func (p *GeminiAPIBatchImageProvider) Submit(ctx context.Context, job *BatchImageJob, account *Account, input BatchImageInput) (*BatchProviderJob, error) {
	apiKey := batchImageProviderAPIKey(account)
	if apiKey == "" {
		return nil, ErrBatchImageProviderMissingAPIKey
	}

	// Build JSONL content for Gemini batch API.
	jsonl, err := buildGeminiJSONL(job, input)
	if err != nil {
		return nil, batchImageProviderInputError("build jsonl: %w", err)
	}

	// Upload JSONL.
	uploaded, err := p.client.UploadJSONL(ctx, apiKey, job.BatchID, jsonl)
	if err != nil {
		return nil, batchImageProviderInputError("upload jsonl: %w", err)
	}

	// Create batch job.
	batch, err := p.client.CreateBatch(ctx, apiKey, job.Model, uploaded.Name, job.BatchID)
	if err != nil {
		return nil, batchImageProviderInputError("create batch: %w", err)
	}

	return &BatchProviderJob{
		ProviderJobName:  batch.Name,
		ProviderInputRef: uploaded.Name,
		RawState:         batch.State,
	}, nil
}

func (p *GeminiAPIBatchImageProvider) Get(ctx context.Context, job *BatchImageJob, account *Account) (*BatchProviderStatus, error) {
	apiKey := batchImageProviderAPIKey(account)
	name := batchImageProviderJobName(job)
	if name == "" {
		return nil, ErrBatchImageProviderMissingJobName
	}

	batch, err := p.client.GetBatch(ctx, apiKey, name)
	if err != nil {
		return nil, err
	}

	return mapGeminiState(batch), nil
}

func (p *GeminiAPIBatchImageProvider) Cancel(ctx context.Context, job *BatchImageJob, account *Account) error {
	apiKey := batchImageProviderAPIKey(account)
	name := batchImageProviderJobName(job)
	if name == "" {
		return ErrBatchImageProviderMissingJobName
	}
	return p.client.CancelBatch(ctx, apiKey, name)
}

func (p *GeminiAPIBatchImageProvider) OpenResult(ctx context.Context, job *BatchImageJob, account *Account) (io.ReadCloser, string, error) {
	apiKey := batchImageProviderAPIKey(account)
	ref := batchImageProviderOutputRef(job)
	if ref == "" {
		return nil, "", ErrBatchImageProviderMissingResultRef
	}
	return p.client.DownloadFile(ctx, apiKey, ref)
}

func (p *GeminiAPIBatchImageProvider) Cleanup(ctx context.Context, job *BatchImageJob, account *Account, target CleanupTarget) error {
	apiKey := batchImageProviderAPIKey(account)
	if apiKey == "" {
		return ErrBatchImageProviderMissingAPIKey
	}

	if target == CleanupTargetInput || target == CleanupTargetAll {
		if ref := batchImageProviderInputRef(job); ref != "" {
			if err := p.client.DeleteFile(ctx, apiKey, ref); err != nil {
				return err
			}
		}
	}
	if target == CleanupTargetOutput || target == CleanupTargetAll {
		if ref := batchImageProviderOutputRef(job); ref != "" {
			if err := p.client.DeleteFile(ctx, apiKey, ref); err != nil {
				return err
			}
		}
	}
	return nil
}

// --- helpers ---

// buildGeminiJSONL builds the JSONL request body for Gemini batch API.
func buildGeminiJSONL(job *BatchImageJob, input BatchImageInput) (io.Reader, error) {
	var lines []string
	for _, item := range input.Items {
		req := map[string]any{
			"request": map[string]any{
				"contents": []map[string]any{
					{
						"parts": []map[string]any{
							{"text": item.Prompt},
						},
						"role": "user",
					},
				},
			},
		}
		line, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		lines = append(lines, string(line))
	}
	return strings.NewReader(strings.Join(lines, "\n")), nil
}

// mapGeminiState maps Gemini batch state to our internal state.
func mapGeminiState(batch *GeminiBatchJob) *BatchProviderStatus {
	status := &BatchProviderStatus{
		RawState: batch.State,
	}

	switch strings.ToUpper(batch.State) {
	case "ACTIVE", "STATE_ACTIVE":
		status.InternalState = BatchProviderStateRunning
		status.Done = false
	case "SUCCEEDED", "STATE_SUCCEEDED":
		status.InternalState = BatchProviderStateSucceeded
		status.Done = true
		if batch.Response != nil {
			// Gemini API returns responses_file (snake_case); fall back to camelCase for older responses
			ref := batch.Response.ResponsesFileSnake
			if ref == "" {
				ref = batch.Response.ResponsesFile
			}
			status.ProviderOutputRef = ref
		}

	case "FAILED", "STATE_FAILED":
		status.InternalState = BatchProviderStateFailed
		status.Done = true
		if batch.Error != nil {
			status.ErrorCode = batch.Error.Code
			status.ErrorMessage = batch.Error.Message
		}
	case "CANCELLED", "STATE_CANCELLED":
		status.InternalState = BatchProviderStateCancelled
		status.Done = true
	default:
		status.InternalState = BatchProviderStateQueued
		status.SuggestedRequeueAfter = defaultGeminiBatchRequeueAfter
	}

	return status
}
