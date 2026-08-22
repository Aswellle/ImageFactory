package batchimage

import (
	"context"
	"fmt"
	"io"
)

// DownloadService handles image retrieval from object storage.
type DownloadService struct {
	// storage would be storage.Storage in the full implementation
}

// NewDownloadService builds a DownloadService.
func NewDownloadService() *DownloadService {
	return &DownloadService{}
}

// OpenImage opens a single image for streaming.
// Returns a ReadCloser that the caller must close.
func (s *DownloadService) OpenImage(ctx context.Context, storageKey string) (io.ReadCloser, error) {
	if storageKey == "" {
		return nil, fmt.Errorf("empty storage key")
	}
	// In the full implementation: fetch from object storage
	return nil, fmt.Errorf("not implemented: storage integration pending")
}

// OpenImageURL returns a presigned URL for direct image access.
func (s *DownloadService) OpenImageURL(ctx context.Context, storageKey string, ttlSeconds int) (string, error) {
	if storageKey == "" {
		return "", fmt.Errorf("empty storage key")
	}
	// In the full implementation: generate presigned URL from object storage
	return "", fmt.Errorf("not implemented: storage integration pending")
}

// StreamZip builds a ZIP archive of multiple images.
// The callback is called with a writer for each file in the archive.
func (s *DownloadService) StreamZip(ctx context.Context, storageKeys []string, w io.Writer) error {
	if len(storageKeys) == 0 {
		return fmt.Errorf("no images to zip")
	}
	// In the full implementation: stream ZIP with manifest.json + errors.json
	return fmt.Errorf("not implemented: storage integration pending")
}
