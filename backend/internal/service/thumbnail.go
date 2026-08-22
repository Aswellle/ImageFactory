// package service holds the business-logic services for ImageForge.
package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/imageforge/imageforge/internal/pkg/image"
	"github.com/imageforge/imageforge/internal/storage"
)

// Thumbnail sizes. Thumbnails are used in list views; mediums in detail views.
const (
	ThumbnailWidth  = 256
	ThumbnailHeight = 256
	MediumWidth     = 512
	MediumHeight    = 512
)

// ThumbnailResult holds the storage keys produced for one source image.
type ThumbnailResult struct {
	StorageKey   string
	ThumbnailKey string
	MediumKey    string
	MimeType     string
	// Err captures a non-fatal generation error so callers can log/continue.
	Err error
}

// ThumbnailService generates and stores resized versions of a source image.
// Implementations must be safe for concurrent use; generation is idempotent —
// calling Generate twice for the same key overwrites the prior thumbnails.
type ThumbnailService interface {
	// Generate downloads the original at storageKey, produces thumbnail and
	// medium variants, stores them, and returns their keys. A non-nil error
	// means no thumbnail was produced; callers should fall back to the original.
	Generate(ctx context.Context, storageKey, contentType string) (*ThumbnailResult, error)

	// GenerateBatch generates thumbnails for many keys. Each result is returned
	// in the same order as the input; individual failures are surfaced via
	// ThumbnailResult.Err rather than failing the whole batch.
	GenerateBatch(ctx context.Context, storageKeys []string) []ThumbnailResult
}

// thumbnailService is the default ThumbnailService. It uses the storage
// backend for originals and the image package for resizing.
type thumbnailService struct {
	store storage.Storage
}

// NewThumbnailService builds a ThumbnailService backed by store.
func NewThumbnailService(store storage.Storage) ThumbnailService {
	return &thumbnailService{store: store}
}

// Generate implements ThumbnailService.
func (s *thumbnailService) Generate(ctx context.Context, storageKey, contentType string) (*ThumbnailResult, error) {
	result := &ThumbnailResult{StorageKey: storageKey}

	// 1. Download the original.
	rc, err := s.store.Get(ctx, storageKey)
	if err != nil {
		result.Err = fmt.Errorf("get original %s: %w", storageKey, err)
		return result, result.Err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		result.Err = fmt.Errorf("read original %s: %w", storageKey, err)
		return result, result.Err
	}

	mime := contentType
	if mime == "" {
		mime = image.MimeType(data)
	}
	result.MimeType = mime

	// 2. Resize. On failure (unsupported format / corruption) fall back to
	// storing the original bytes so the asset still has a usable thumbnail.
	medium, err := image.Resize(data, MediumWidth, MediumHeight)
	if err != nil {
		medium = data
	}
	thumb, err := image.Resize(data, ThumbnailWidth, ThumbnailHeight)
	if err != nil {
		thumb = data
	}

	// 3. Derive destination keys by swapping the storage kind segment.
	mediumKey := deriveKey(storageKey, storage.KindMedium)
	thumbKey := deriveKey(storageKey, storage.KindThumbnail)

	// 4. Upload the variants.
	if _, err := s.store.Put(ctx, storage.PutInput{
		Key:         mediumKey,
		Body:        bytes.NewReader(medium),
		Size:        int64(len(medium)),
		ContentType: mime,
	}); err != nil {
		result.Err = fmt.Errorf("put medium %s: %w", mediumKey, err)
		return result, result.Err
	}
	result.MediumKey = mediumKey

	if _, err := s.store.Put(ctx, storage.PutInput{
		Key:         thumbKey,
		Body:        bytes.NewReader(thumb),
		Size:        int64(len(thumb)),
		ContentType: mime,
	}); err != nil {
		result.Err = fmt.Errorf("put thumbnail %s: %w", thumbKey, err)
		return result, result.Err
	}
	result.ThumbnailKey = thumbKey

	return result, nil
}

// GenerateBatch implements ThumbnailService.
func (s *thumbnailService) GenerateBatch(ctx context.Context, storageKeys []string) []ThumbnailResult {
	results := make([]ThumbnailResult, 0, len(storageKeys))
	for _, key := range storageKeys {
		res, err := s.Generate(ctx, key, "")
		if err != nil {
			results = append(results, ThumbnailResult{StorageKey: key, Err: err})
			continue
		}
		results = append(results, *res)
	}
	return results
}

// deriveKey maps an original storage key to the key for a derived variant by
// replacing the kind segment. The layout is:
//
//	users/{uid}/projects/{pid}/{kind}/{name}
//
// where kind ∈ {originals, thumbnails, mediums, versions, uploads}.
// Replacing the kind keeps derived objects colocated with the original.
func deriveKey(original string, target storage.Kind) string {
	known := []storage.Kind{
		storage.KindOriginal, storage.KindThumbnail, storage.KindMedium,
		storage.KindVersion, storage.KindUpload,
	}
	parts := strings.Split(original, "/")
	for i, p := range parts {
		for _, k := range known {
			if storage.Kind(p) == k {
				parts[i] = string(target)
				return strings.Join(parts, "/")
			}
		}
	}
	// No recognized kind segment: append the target kind before the filename.
	dir := filepath.Dir(original)
	base := filepath.Base(original)
	return filepath.Join(dir, string(target), base)
}
