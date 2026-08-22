package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Filesystem implements storage.Storage against the local filesystem. It is
// intended for local development only — production uses R2/MinIO.
type Filesystem struct {
	root string
}

// Compile-time interface check.
var _ Storage = (*Filesystem)(nil)

// NewFilesystem creates a filesystem-backed store rooted at root.
func NewFilesystem(root string) (*Filesystem, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create storage root %s: %w", root, err)
	}
	return &Filesystem{root: root}, nil
}

func (f *Filesystem) resolve(key string) string {
	return filepath.Join(f.root, filepath.FromSlash(key))
}

func (f *Filesystem) Put(ctx context.Context, in PutInput) (string, error) {
	path := f.resolve(in.Key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create %s: %w", path, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, in.Body); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}
	return in.Key, nil
}

func (f *Filesystem) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	file, err := os.Open(f.resolve(key))
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", key, err)
	}
	return file, nil
}

func (f *Filesystem) Delete(ctx context.Context, key string) error {
	err := os.Remove(f.resolve(key))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete %s: %w", key, err)
	}
	return nil
}

func (f *Filesystem) Stat(ctx context.Context, key string) (*ObjectMeta, error) {
	info, err := os.Stat(f.resolve(key))
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", key, err)
	}
	return &ObjectMeta{
		Key:          key,
		Size:         info.Size(),
		LastModified: info.ModTime(),
	}, nil
}

// PresignedGetURL / PresignedPutURL return local file:// placeholders for
// filesystem-backed dev. Production R2/MinIO returns real signed URLs.
func (f *Filesystem) PresignedGetURL(ctx context.Context, key string, expire time.Duration) (string, error) {
	return fmt.Sprintf("file://%s", f.resolve(key)), nil
}

func (f *Filesystem) PresignedPutURL(ctx context.Context, key string, expire time.Duration, contentType string) (string, error) {
	return fmt.Sprintf("file://%s", f.resolve(key)), nil
}
