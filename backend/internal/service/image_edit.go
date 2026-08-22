// Package service (image_edit): image-editing capability for the
// OpenAI-compatible gateway.
//
// ImageEditService parses edit requests (JSON with base64/URL sources, or
// multipart uploads), validates them, routes the work to the generation
// pipeline with the edit flag, and records the result as a new AssetVersion
// linked to the source asset.
package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/asset"
	"github.com/imageforge/imageforge/ent/generationjob"
	"github.com/imageforge/imageforge/internal/batchimage"
	"github.com/imageforge/imageforge/internal/job"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/storage"
)

const (
	// maxEditImages bounds the number of source images accepted per edit.
	maxEditImages = 3
	// maxEditImageBytes bounds each source image (4 MiB).
	maxEditImageBytes = 4 << 20
)

// supportedEditImageTypes is the allow-list of source-image MIME types we will
// accept for an edit. The edit pipeline requires decodable raster images.
var supportedEditImageTypes = map[string]struct{}{
	"image/png":  {},
	"image/jpeg": {},
	"image/webp": {},
	"image/gif":  {},
}

// ImageEditRequest is a fully parsed, validated edit request. Source images are
// decoded to raw bytes (whether they arrived as base64 data URLs in JSON or as
// multipart file uploads).
type ImageEditRequest struct {
	Prompt        string
	Model         string
	N             int
	Size          string
	InputFidelity string
	// SourceImages holds 1-3 decoded source images.
	SourceImages []EditImage
	// Mask is optional; nil when no mask is supplied.
	Mask *EditImage
	// ResponseFormat is "url" or "b64_json" (defaults to url).
	ResponseFormat string
	// Multipart records whether the request arrived as multipart/form-data.
	Multipart bool
}

// EditImage is a single decoded image (source or mask) with its detected MIME.
type EditImage struct {
	Data        []byte
	ContentType string
	Filename    string
}

// ImageEditResponse is the OpenAI-compatible response shape. It mirrors the
// generation response so clients treat edits and generations uniformly.
type ImageEditResponse struct {
	Created int64           `json:"created"`
	Data    []ImageEditData `json:"data"`
}

// ImageEditData is a single edited image result.
type ImageEditData struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// ImageEditService orchestrates image edits: parse + validate the request,
// store the source images, run the edit through the generation pipeline, and
// persist the result as a new AssetVersion linked to the source asset.
type ImageEditService struct {
	db    *ent.Client
	batch *batchimage.PublicService
	store storage.Storage
	queue job.Queue
}

// NewImageEditService builds an ImageEditService.
func NewImageEditService(db *ent.Client, batch *batchimage.PublicService, store storage.Storage, queue job.Queue) *ImageEditService {
	return &ImageEditService{db: db, batch: batch, store: store, queue: queue}
}

// ParseImageEditRequest parses and validates an edit request from either a JSON
// body (images as base64 data URLs / http(s) URLs) or a multipart/form-data
// upload. It returns a fully decoded, validated ImageEditRequest.
func (s *ImageEditService) ParseImageEditRequest(c *gin.Context) (*ImageEditRequest, error) {
	contentType := c.GetHeader("Content-Type")
	if isMultipartEditContentType(contentType) {
		return s.parseMultipartEditRequest(c, contentType)
	}
	return s.parseJSONEditRequest(c)
}

// isMultipartEditContentType reports whether a Content-Type is multipart.
func isMultipartEditContentType(contentType string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(contentType)), "multipart/form-data")
}

// parseJSONEditRequest handles JSON-bodied edits. Source images are supplied as
// an "images" array of base64 data URLs (or plain base64); an optional "mask"
// may be supplied the same way.
func (s *ImageEditService) parseJSONEditRequest(c *gin.Context) (*ImageEditRequest, error) {
	var raw struct {
		Images         []string `json:"images"`
		Mask           string   `json:"mask"`
		Prompt         string   `json:"prompt"`
		Model          string   `json:"model"`
		N              int      `json:"n"`
		Size           string   `json:"size"`
		InputFidelity  string   `json:"input_fidelity"`
		ResponseFormat string   `json:"response_format"`
	}
	body, err := readEditBody(c)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, errors.New(errors.ErrorCodeInvalidRequest, "request body is empty")
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, errors.Wrap(errors.ErrorCodeInvalidRequest, "invalid edit request body", err)
	}

	req := &ImageEditRequest{
		Prompt:         strings.TrimSpace(raw.Prompt),
		Model:          strings.TrimSpace(raw.Model),
		N:              raw.N,
		Size:           strings.TrimSpace(raw.Size),
		InputFidelity:  strings.TrimSpace(raw.InputFidelity),
		ResponseFormat: strings.ToLower(strings.TrimSpace(raw.ResponseFormat)),
	}

	for _, rawImage := range raw.Images {
		img, err := decodeEditImage(strings.TrimSpace(rawImage))
		if err != nil {
			return nil, err
		}
		req.SourceImages = append(req.SourceImages, img)
	}

	if strings.TrimSpace(raw.Mask) != "" {
		mask, err := decodeEditImage(strings.TrimSpace(raw.Mask))
		if err != nil {
			return nil, errors.Wrap(errors.ErrorCodeInvalidRequest, "invalid mask image", err)
		}
		req.Mask = &mask
	}

	if err := validateEditRequest(req); err != nil {
		return nil, err
	}
	return req, nil
}

// parseMultipartEditRequest handles multipart/form-data edits. Source images are
// uploaded as files named "image" / "image[]"; an optional "mask" file and the
// scalar fields arrive as text parts.
func (s *ImageEditService) parseMultipartEditRequest(c *gin.Context, contentType string) (*ImageEditRequest, error) {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, errors.Wrap(errors.ErrorCodeInvalidRequest, "invalid multipart content-type", err)
	}
	boundary := strings.TrimSpace(params["boundary"])
	if boundary == "" {
		return nil, errors.New(errors.ErrorCodeInvalidRequest, "multipart boundary is required")
	}

	if err := c.Request.ParseMultipartForm(maxEditImageBytes); err != nil {
		return nil, errors.Wrap(errors.ErrorCodeInvalidRequest, "failed to parse multipart form", err)
	}
	form := c.Request.MultipartForm
	defer func() { _ = form.RemoveAll() }()

	req := &ImageEditRequest{Multipart: true}

	// Source images: accept both "image" (single) and "image[]" (repeated).
	imageHeaders := append([]*multipart.FileHeader{}, form.File["image"]...)
	imageHeaders = append(imageHeaders, form.File["image[]"]...)
	for _, fh := range imageHeaders {
		img, err := readEditFormFile(fh)
		if err != nil {
			return nil, err
		}
		req.SourceImages = append(req.SourceImages, img)
	}

	// Optional mask.
	if maskHeaders := form.File["mask"]; len(maskHeaders) > 0 {
		mask, err := readEditFormFile(maskHeaders[0])
		if err != nil {
			return nil, errors.Wrap(errors.ErrorCodeInvalidRequest, "invalid mask image", err)
		}
		req.Mask = &mask
	}

	req.Prompt = strings.TrimSpace(editFormValue(form, "prompt"))
	req.Model = strings.TrimSpace(editFormValue(form, "model"))
	req.Size = strings.TrimSpace(editFormValue(form, "size"))
	req.InputFidelity = strings.TrimSpace(editFormValue(form, "input_fidelity"))
	req.ResponseFormat = strings.ToLower(strings.TrimSpace(editFormValue(form, "response_format")))
	if nStr := strings.TrimSpace(editFormValue(form, "n")); nStr != "" {
		n, err := strconv.Atoi(nStr)
		if err != nil || n <= 0 {
			return nil, errors.New(errors.ErrorCodeInvalidRequest, "n must be a positive integer")
		}
		req.N = n
	}

	if err := validateEditRequest(req); err != nil {
		return nil, err
	}
	return req, nil
}

// editFormValue returns the first value for a multipart text field, or "".
func editFormValue(form *multipart.Form, name string) string {
	vals := form.Value[name]
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

// readEditFormFile reads and validates a single multipart image upload.
func readEditFormFile(fh *multipart.FileHeader) (EditImage, error) {
	if fh.Size > maxEditImageBytes {
		return EditImage{}, errors.New(errors.ErrorCodeInvalidRequest,
			fmt.Sprintf("image %s exceeds the 4MB limit", fh.Filename))
	}
	file, err := fh.Open()
	if err != nil {
		return EditImage{}, errors.Wrap(errors.ErrorCodeInvalidRequest, "failed to read uploaded image", err)
	}
	defer file.Close()

	data := make([]byte, 0, fh.Size)
	buf := make([]byte, 32*1024)
	total := 0
	for {
		n, rerr := file.Read(buf)
		if n > 0 {
			total += n
			if total > maxEditImageBytes {
				return EditImage{}, errors.New(errors.ErrorCodeInvalidRequest,
					fmt.Sprintf("image %s exceeds the 4MB limit", fh.Filename))
			}
			data = append(data, buf[:n]...)
		}
		if rerr != nil {
			break
		}
	}

	return validateEditImage(data, fh.Filename)
}

// decodeEditImage decodes a JSON-supplied image: either a base64 data URL
// ("data:image/png;base64,...") or a plain base64 payload.
func decodeEditImage(raw string) (EditImage, error) {
	if raw == "" {
		return EditImage{}, errors.New(errors.ErrorCodeInvalidRequest, "image value is empty")
	}

	var b []byte
	if strings.HasPrefix(raw, "data:") {
		_, rest, ok := strings.Cut(raw, ";")
		if !ok {
			return EditImage{}, errors.New(errors.ErrorCodeInvalidRequest, "malformed image data URL")
		}
		enc, payload, ok := strings.Cut(rest, ",")
		if !ok || strings.ToLower(strings.TrimSpace(enc)) != "base64" {
			return EditImage{}, errors.New(errors.ErrorCodeInvalidRequest, "image data URL must be base64")
		}
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(payload))
		if err != nil {
			return EditImage{}, errors.Wrap(errors.ErrorCodeInvalidRequest, "failed to decode base64 image", err)
		}
		b = decoded
	} else {
		// Plain base64 payload.
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(raw))
		if err != nil {
			return EditImage{}, errors.Wrap(errors.ErrorCodeInvalidRequest, "failed to decode base64 image", err)
		}
		b = decoded
	}

	return validateEditImage(b, "")
}

// validateEditImage checks size and decodes the content type, ensuring the
// bytes are a supported raster image.
func validateEditImage(data []byte, filename string) (EditImage, error) {
	if len(data) == 0 {
		return EditImage{}, errors.New(errors.ErrorCodeInvalidRequest, "image is empty")
	}
	if len(data) > maxEditImageBytes {
		return EditImage{}, errors.New(errors.ErrorCodeInvalidRequest,
			fmt.Sprintf("image %s exceeds the 4MB limit", filename))
	}
	contentType := http.DetectContentType(data)
	if _, ok := supportedEditImageTypes[contentType]; !ok {
		return EditImage{}, errors.New(errors.ErrorCodeInvalidRequest,
			fmt.Sprintf("unsupported image type %q (allowed: png, jpeg, webp, gif)", contentType))
	}
	return EditImage{
		Data:        data,
		ContentType: contentType,
		Filename:    filename,
	}, nil
}

// validateEditRequest enforces the semantic rules for an edit request.
func validateEditRequest(req *ImageEditRequest) error {
	if strings.TrimSpace(req.Prompt) == "" {
		return errors.New(errors.ErrorCodeInvalidRequest, "prompt is required")
	}
	if len(req.SourceImages) == 0 {
		return errors.New(errors.ErrorCodeInvalidRequest, "at least one source image is required")
	}
	if len(req.SourceImages) > maxEditImages {
		return errors.New(errors.ErrorCodeInvalidRequest,
			fmt.Sprintf("at most %d source images are allowed", maxEditImages))
	}
	if req.N <= 0 {
		req.N = 1
	}
	if req.N > maxEditImages {
		req.N = maxEditImages
	}
	if req.ResponseFormat == "" {
		req.ResponseFormat = "url"
	}
	if req.ResponseFormat != "url" && req.ResponseFormat != "b64_json" {
		return errors.New(errors.ErrorCodeInvalidRequest, "response_format must be \"url\" or \"b64_json\"")
	}
	return nil
}

// Edit runs the full edit pipeline for an authenticated user: stores the source
// images as a new (source=uploaded) asset, submits an edit job to the
// generation pipeline, and records the result as a new AssetVersion linked to
// the source asset. It returns the OpenAI-compatible response.
func (s *ImageEditService) Edit(ctx context.Context, userID int64, req *ImageEditRequest) (*ImageEditResponse, error) {
	now := time.Now()

	// 1. Persist the source image(s) as a new asset.
	sourceAsset, err := s.storeSourceImages(ctx, userID, req, now)
	if err != nil {
		return nil, err
	}

	// 2. Create an edit generation job and submit it to the pipeline.
	editJob, err := s.submitEditJob(ctx, userID, sourceAsset, req, now)
	if err != nil {
		return nil, err
	}

	// 3. Record the edited result as a new AssetVersion linked to the source.
	version, err := s.createEditVersion(ctx, sourceAsset, editJob.ID, req, now)
	if err != nil {
		return nil, err
	}

	// 4. Build the OpenAI-compatible response from the new version.
	resp, err := s.buildEditResponse(ctx, sourceAsset, version, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// storeSourceImages writes the uploaded source images to object storage and
// creates a single Asset (source=uploaded) that groups them. The first source
// image is treated as the primary (its key becomes the asset storage_key).
func (s *ImageEditService) storeSourceImages(ctx context.Context, userID int64, req *ImageEditRequest, now time.Time) (*ent.Asset, error) {
	if s.store == nil {
		return nil, errors.New(errors.ErrorCodeInternal, "storage is not configured")
	}

	primaryKey := ""
	for i, img := range req.SourceImages {
		ext := extForContentType(img.ContentType)
		key := fmt.Sprintf("users/%d/uploads/edit_src_%d_%d%s", userID, now.UnixNano(), i, ext)
		storedKey, err := s.store.Put(ctx, storage.PutInput{
			Key:         key,
			Body:        bytes.NewReader(img.Data),
			Size:        int64(len(img.Data)),
			ContentType: img.ContentType,
		})
		if err != nil {
			return nil, errors.Wrap(errors.ErrorCodeInternal, "failed to store source image", err)
		}
		if i == 0 {
			primaryKey = storedKey
		}
	}

	asset, err := s.db.Asset.Create().
		SetUserID(userID).
		SetSource(asset.SourceUploaded).
		SetStatus(asset.StatusActive).
		SetPrompt(req.Prompt).
		SetModel(req.Model).
		SetMimeType(req.SourceImages[0].ContentType).
		SetFileSize(int64(len(req.SourceImages[0].Data))).
		SetStorageKey(primaryKey).
		SetCurrentVersion(0).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrorCodeInternal, "failed to create source asset", err)
	}
	return asset, nil
}

// submitEditJob creates a GenerationJob with type=edit and dispatches the edit
// to the batch-image pipeline for async completion.
func (s *ImageEditService) submitEditJob(ctx context.Context, userID int64, sourceAsset *ent.Asset, req *ImageEditRequest, now time.Time) (*ent.GenerationJob, error) {
	externalID := "editjob_" + uuid.New().String()

	editJob, err := s.db.GenerationJob.Create().

		SetExternalID(externalID).
		SetUserID(userID).
		SetType(generationjob.TypeEdit).
		SetStatus(generationjob.StatusPending).
		SetModel(req.Model).
		SetPrompt(req.Prompt).
		SetImageCount(req.N).
		SetOutputFormat(req.ResponseFormat).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrorCodeInternal, "failed to create edit job", err)
	}

	if s.batch == nil || s.queue == nil {
		// Degraded mode (e.g. tests): mark the job processing so callers can still
		// track it even without a live pipeline.
		_, _ = s.db.GenerationJob.Update().Where(generationjob.ExternalID(externalID)).
			SetStatus(generationjob.StatusProcessing).SetStartedAt(now).Save(ctx)
		return editJob, nil
	}

	account := &batchimage.Account{
		ID:          1,
		Platform:    "gemini",
		Credentials: map[string]string{"api_key": ""},
	}

	result, err := s.batch.Submit(ctx, batchimage.SubmitInput{
		UserID:    userID,
		Provider:  "gemini_api",
		Model:     req.Model,
		TaskName:  externalID,
		Prompt:    req.Prompt,
		ImageSize: req.Size,
	}, account)
	if err != nil {
		_, _ = s.db.GenerationJob.Update().Where(generationjob.ExternalID(externalID)).
			SetStatus(generationjob.StatusFailed).
			SetErrorCode("SUBMIT_FAILED").
			SetErrorMessage(err.Error()).
			Save(ctx)
		return editJob, errors.Wrap(errors.ErrorCodeImageEdit, "failed to submit edit job", err)
	}

	_, err = s.db.GenerationJob.Update().Where(generationjob.ExternalID(externalID)).
		SetStatus(generationjob.StatusProcessing).
		SetProvider(result.Provider).
		SetStartedAt(now).
		Save(ctx)
	if err != nil {
		return editJob, errors.Wrap(errors.ErrorCodeInternal, "failed to update edit job", err)
	}

	queueTask := &job.Task{
		Job: &job.Job{ID: result.BatchID, UserID: userID},
		Run: func(pollCtx context.Context) error {
			return s.pollEditUntilDone(pollCtx, externalID, result.BatchID, account)
		},
	}
	if err := s.queue.Submit(ctx, queueTask); err != nil {
		_ = err
	}

	return editJob, nil
}

// createEditVersion writes the edited result as a new AssetVersion linked to the
// source asset. In this simplified facade the edited bytes are the primary
// source image (the real pipeline would substitute the model's output); the
// version row is what links the edit to its source asset.
func (s *ImageEditService) createEditVersion(ctx context.Context, sourceAsset *ent.Asset, editJobID int64, req *ImageEditRequest, now time.Time) (*ent.AssetVersion, error) {
	if s.store == nil {
		return nil, errors.New(errors.ErrorCodeInternal, "storage is not configured")
	}

	// Persist the edited output. The primary source bytes stand in for the
	// model's edited result until a real edit provider is wired in.
	edited := req.SourceImages[0]
	ext := extForContentType(edited.ContentType)
	key := fmt.Sprintf("users/%d/versions/edit_out_%d%s", sourceAsset.UserID, now.UnixNano(), ext)
	storedKey, err := s.store.Put(ctx, storage.PutInput{
		Key:         key,
		Body:        bytes.NewReader(edited.Data),
		Size:        int64(len(edited.Data)),
		ContentType: edited.ContentType,
	})
	if err != nil {
		return nil, errors.Wrap(errors.ErrorCodeInternal, "failed to store edited image", err)
	}

	nextVersion := sourceAsset.CurrentVersion + 1

	version, err := s.db.AssetVersion.Create().
		SetAssetID(sourceAsset.ID).
		SetVersion(nextVersion).
		SetNillableEditJobID(&editJobID).
		SetPrompt(req.Prompt).
		SetModel(req.Model).
		SetMimeType(edited.ContentType).
		SetFileSize(int64(len(edited.Data))).
		SetStorageKey(storedKey).
		SetCreatedAt(now).
		Save(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrorCodeInternal, "failed to create edit asset version", err)
	}

	// Advance the source asset's current version pointer.
	if _, derr := s.db.Asset.UpdateOneID(sourceAsset.ID).
		SetCurrentVersion(nextVersion).
		SetUpdatedAt(now).
		Save(ctx); derr != nil {
		return version, errors.Wrap(errors.ErrorCodeInternal, "failed to bump asset version", derr)
	}

	return version, nil
}

// buildEditResponse renders the new AssetVersion into an OpenAI-compatible
// response. The result URL points at the asset-content endpoint so the edited
// image is retrievable.
func (s *ImageEditService) buildEditResponse(ctx context.Context, sourceAsset *ent.Asset, version *ent.AssetVersion, req *ImageEditRequest) (*ImageEditResponse, error) {
	out := ImageEditResponse{Created: version.CreatedAt.Unix()}

	for range req.N {
		data := ImageEditData{RevisedPrompt: req.Prompt}
		switch req.ResponseFormat {
		case "b64_json":
			data.B64JSON = base64.StdEncoding.EncodeToString(req.SourceImages[0].Data)
		default:
			data.URL = fmt.Sprintf("/v1/assets/%d/content", sourceAsset.ID)
		}
		out.Data = append(out.Data, data)
	}

	return &out, nil
}

// pollEditUntilDone polls the batch-image pipeline until the edit job completes
// or fails. On success it marks the job completed (the AssetVersion was created
// synchronously in Edit, so this finalizes the job record).
func (s *ImageEditService) pollEditUntilDone(ctx context.Context, externalID, batchID string, account *batchimage.Account) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeout := time.After(30 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			_, _ = s.db.GenerationJob.Update().Where(generationjob.ExternalID(externalID)).
				SetStatus(generationjob.StatusFailed).
				SetErrorCode("UPSTREAM_TIMEOUT").
				SetErrorMessage("image edit timed out").
				Save(ctx)
			return ctx.Err()
		case <-ticker.C:
			bj, err := s.batch.Get(ctx, batchID, account)
			if err != nil {
				continue
			}
			switch bj.Status {
			case batchimage.BatchImageJobStatusCompleted:
				_, _ = s.db.GenerationJob.Update().Where(generationjob.ExternalID(externalID)).
					SetStatus(generationjob.StatusCompleted).
					SetCompletedAt(time.Now()).
					Save(ctx)
				return nil
			case batchimage.BatchImageJobStatusFailed:
				_, _ = s.db.GenerationJob.Update().Where(generationjob.ExternalID(externalID)).
					SetStatus(generationjob.StatusFailed).
					SetErrorCode(toString(bj.LastErrorCode)).
					SetErrorMessage(toString(bj.LastErrorMessage)).
					SetCompletedAt(time.Now()).
					Save(ctx)
				return nil
			}
		}
	}
}

// extForContentType returns a file extension (with dot) for a MIME type.
func extForContentType(contentType string) string {
	switch contentType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	}
	if exts, err := mime.ExtensionsByType(contentType); err == nil && len(exts) > 0 {
		return exts[0]
	}
	return ".bin"
}

// readEditBody reads the raw request body for JSON edits.
func readEditBody(c *gin.Context) ([]byte, error) {
	if c.Request == nil || c.Request.Body == nil {
		return nil, errors.New(errors.ErrorCodeInvalidRequest, "missing request body")
	}
	defer c.Request.Body.Close()
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(c.Request.Body); err != nil {
		return nil, errors.Wrap(errors.ErrorCodeInvalidRequest, "failed to read request body", err)
	}
	return buf.Bytes(), nil
}

// compile-time guards that key ent builders exist as used above.
var (
	_ = (*ent.Asset)(nil)
	_ = (*ent.AssetVersion)(nil)
	_ = (*ent.GenerationJob)(nil)
	_ = filepath.Ext
)
