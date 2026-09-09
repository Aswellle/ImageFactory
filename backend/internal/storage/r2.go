package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/imageforge/imageforge/internal/config"
)

// R2 implements storage.Storage backed by an S3-compatible endpoint
// (Cloudflare R2, MinIO, etc.).
type R2 struct {
	client     *s3.Client
	bucket     string
	presign    *s3.PresignClient
	publicHost string
}

// Compile-time interface check.
var _ Storage = (*R2)(nil)

// NewR2 builds an R2/MinIO storage client from config.
func NewR2(cfg config.StorageConfig) (*R2, error) {
	staticCreds := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")

	awscfg, err := awsConfig.LoadDefaultConfig(context.Background(),
		awsConfig.WithRegion(cfg.Region),
		awsConfig.WithCredentialsProvider(staticCreds),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	opts := func(o *s3.Options) {
		// Path-style addressing is required for MinIO and most non-AWS endpoints.
		o.UsePathStyle = true
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	}

	client := s3.NewFromConfig(awscfg, opts)
	r2 := &R2{
		client:     client,
		bucket:     cfg.Bucket,
		presign:    s3.NewPresignClient(client),
		publicHost: cfg.PublicHost,
	}
	// Ensure the bucket exists (idempotent). MinIO and R2 both return
	// BucketAlreadyOwnedByYou when the bucket is already present.
	if err := r2.ensureBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("ensure bucket %s: %w", cfg.Bucket, err)
	}
	return r2, nil
}

// ensureBucket creates the bucket if it does not already exist. It is safe to
// call on every startup.
func (r *R2) ensureBucket(ctx context.Context) error {
	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(r.bucket)})
	if err == nil {
		return nil // already exists
	}
	_, err = r.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(r.bucket)})
	if err != nil {
		return err
	}
	return nil
}

func (r *R2) Put(ctx context.Context, in PutInput) (string, error) {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(in.Key),
		Body:        in.Body,
		ContentType: aws.String(in.ContentType),
	})
	if err != nil {
		return "", fmt.Errorf("r2 put %s: %w", in.Key, err)
	}
	return in.Key, nil
}

func (r *R2) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("r2 get %s: %w", key, err)
	}
	return out.Body, nil
}

func (r *R2) Delete(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("r2 delete %s: %w", key, err)
	}
	return nil
}

func (r *R2) Stat(ctx context.Context, key string) (*ObjectMeta, error) {
	out, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("r2 stat %s: %w", key, err)
	}
	return &ObjectMeta{
		Key:          key,
		Size:         aws.ToInt64(out.ContentLength),
		ContentType:  aws.ToString(out.ContentType),
		ETag:         aws.ToString(out.ETag),
		LastModified: aws.ToTime(out.LastModified),
	}, nil
}

func (r *R2) PresignedGetURL(ctx context.Context, key string, expire time.Duration) (string, error) {
	req, err := r.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expire))
	if err != nil {
		return "", fmt.Errorf("r2 presign get %s: %w", key, err)
	}
	return req.URL, nil
}

func (r *R2) PresignedPutURL(ctx context.Context, key string, expire time.Duration, contentType string) (string, error) {
	req, err := r.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expire))
	if err != nil {
		return "", fmt.Errorf("r2 presign put %s: %w", key, err)
	}
	return req.URL, nil
}
