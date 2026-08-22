// package storage defines the object-storage abstraction. Implementations
// (R2, MinIO, local filesystem) must never expose internal credentials.
//
// Paths use the layout: users/{user_id}/projects/{project_id}/{kind}/{name}
// where kind ∈ {originals, thumbnails, mediums, versions, uploads}.
package storage

import (
	"context"
	"io"
	"time"

	"github.com/imageforge/imageforge/internal/config"
)

// Kind identifies the storage tier of an object.
type Kind string

const (
	KindOriginal  Kind = "originals"
	KindThumbnail Kind = "thumbnails"
	KindMedium    Kind = "mediums"
	KindVersion   Kind = "versions"
	KindUpload    Kind = "uploads"
)

// PutInput describes an object to upload.
type PutInput struct {
	Key         string
	Body        io.Reader
	Size        int64
	ContentType string
}

// ObjectMeta describes a stored object's metadata.
type ObjectMeta struct {
	Key          string
	Size         int64
	ContentType  string
	ETag         string
	LastModified time.Time
}

// Storage is the object-storage contract. All methods take a context for
// cancellation/timeout. No implementation may log secrets.
type Storage interface {
	// Put stores an object and returns its canonical key.
	Put(ctx context.Context, in PutInput) (string, error)

	// Get returns a reader for an object. Caller must close it.
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes an object. Idempotent: deleting a missing key is not an error.
	Delete(ctx context.Context, key string) error

	// Stat returns metadata for an object without fetching its body.
	Stat(ctx context.Context, key string) (*ObjectMeta, error)

	// PresignedGetURL returns a time-limited URL for direct client access.
	// Used for private assets instead of exposing permanent storage URLs.
	PresignedGetURL(ctx context.Context, key string, expire time.Duration) (string, error)

	// PresignedPutURL returns a time-limited URL for direct client upload.
	PresignedPutURL(ctx context.Context, key string, expire time.Duration, contentType string) (string, error)
}

// New returns a Storage implementation for the given config.
func New(cfg config.StorageConfig) (Storage, error) {
	switch cfg.Provider {
	case "r2", "minio":
		return NewR2(cfg)
	default:
		return NewFilesystem("./data/storage")
	}
}
