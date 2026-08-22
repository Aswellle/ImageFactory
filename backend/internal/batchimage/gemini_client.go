package batchimage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"
)

// GeminiBatchClient defines the interface for Gemini API operations.
type GeminiBatchClient interface {
	UploadJSONL(ctx context.Context, apiKey, displayName string, r io.Reader) (*GeminiUploadedFile, error)
	CreateBatch(ctx context.Context, apiKey, model, fileName, displayName string) (*GeminiBatchJob, error)
	GetBatch(ctx context.Context, apiKey, batchName string) (*GeminiBatchJob, error)
	CancelBatch(ctx context.Context, apiKey, batchName string) error
	DownloadFile(ctx context.Context, apiKey, fileName string) (io.ReadCloser, string, error)
	DeleteFile(ctx context.Context, apiKey, fileName string) error
}

// GeminiUploadedFile represents a file uploaded to Gemini.
type GeminiUploadedFile struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	URI         string `json:"uri"`
	MimeType    string `json:"mimeType"`
}

// GeminiBatchJob represents a Gemini batch job status.
type GeminiBatchJob struct {
	Name     string               `json:"name"`
	State    string               `json:"state"`
	Dest     *GeminiBatchDest     `json:"dest"`
	Response *GeminiBatchResponse `json:"response"`
	Error    *GeminiBatchError    `json:"error"`
	Raw      map[string]any       `json:"-"`
}

// GeminiBatchDest represents the destination of a Gemini batch job.
type GeminiBatchDest struct {
	FileName      string `json:"fileName"`
	FileNameSnake string `json:"file_name"`
}

// GeminiBatchResponse represents the response of a Gemini batch job.
type GeminiBatchResponse struct {
	ResponsesFile      string `json:"responsesFile"`
	ResponsesFileSnake string `json:"responses_file"`
	InlinedResponses   []any  `json:"inlinedResponses"`
	InlinedResponsesAlt []any `json:"inlined_responses"`
}

// GeminiBatchError represents an error from Gemini.
type GeminiBatchError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// GeminiBatchHTTPClient implements GeminiBatchClient with direct HTTP calls.
type GeminiBatchHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewGeminiBatchHTTPClient creates a new Gemini HTTP client.
func NewGeminiBatchHTTPClient(baseURL, apiKey string) *GeminiBatchHTTPClient {
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	return &GeminiBatchHTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// UploadJSONL uploads a JSONL file to Gemini's file API.
func (c *GeminiBatchHTTPClient) UploadJSONL(ctx context.Context, apiKey, displayName string, r io.Reader) (*GeminiUploadedFile, error) {
	url := fmt.Sprintf("%s/upload/v1beta/files?uploadType=multipart&key=%s", c.baseURL, apiKey)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	metadata := map[string]string{"file": displayName}
	metaJSON, _ := json.Marshal(metadata)
	part, _ := writer.CreatePart(textproto.MIMEHeader{"Content-Type": []string{"application/json"}})
	part.Write(metaJSON)

	dataPart, _ := writer.CreateFormFile("data", "batch.jsonl")
	if _, err := io.Copy(dataPart, r); err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini upload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini upload failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result GeminiUploadedFile
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode upload response: %w", err)
	}
	return &result, nil
}

// CreateBatch creates a new batch generation job.
func (c *GeminiBatchHTTPClient) CreateBatch(ctx context.Context, apiKey, model, fileName, displayName string) (*GeminiBatchJob, error) {
	url := fmt.Sprintf("%s/models/%s:batchGenerateContent?key=%s", c.baseURL, model, apiKey)
	payload := map[string]any{
		"display_name": displayName,
		"input_file":   fileName,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini create batch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini create batch failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result GeminiBatchJob
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode batch response: %w", err)
	}
	return &result, nil
}

// GetBatch polls the status of a batch job.
func (c *GeminiBatchHTTPClient) GetBatch(ctx context.Context, apiKey, batchName string) (*GeminiBatchJob, error) {
	url := fmt.Sprintf("%s/%s?key=%s", c.baseURL, batchName, apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini get batch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini get batch failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result GeminiBatchJob
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode batch status: %w", err)
	}
	return &result, nil
}

// CancelBatch cancels a batch job.
func (c *GeminiBatchHTTPClient) CancelBatch(ctx context.Context, apiKey, batchName string) error {
	url := fmt.Sprintf("%s/%s:cancel?key=%s", c.baseURL, batchName, apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("gemini cancel batch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gemini cancel failed (status %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// DownloadFile downloads a file from Gemini.
func (c *GeminiBatchHTTPClient) DownloadFile(ctx context.Context, apiKey, fileName string) (io.ReadCloser, string, error) {
	url := fmt.Sprintf("%s/%s?key=%s", c.baseURL, fileName, apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("gemini download: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, "", fmt.Errorf("gemini download failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	return resp.Body, ct, nil
}

// DeleteFile deletes a file from Gemini.
func (c *GeminiBatchHTTPClient) DeleteFile(ctx context.Context, apiKey, fileName string) error {
	url := fmt.Sprintf("%s/%s?key=%s", c.baseURL, fileName, apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("gemini delete file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gemini delete failed (status %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}
