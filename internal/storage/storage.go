package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Provider represents a cloud storage provider
type Provider interface {
	// Upload uploads a local file to cloud storage
	Upload(ctx context.Context, localPath, remotePath string) error

	// Download downloads a file from cloud storage to local path
	Download(ctx context.Context, remotePath, localPath string) error

	// Delete deletes a file from cloud storage
	Delete(ctx context.Context, remotePath string) error

	// Exists checks if a file exists in cloud storage
	Exists(ctx context.Context, remotePath string) (bool, error)

	// List lists files with the given prefix
	List(ctx context.Context, prefix string) ([]string, error)

	// GetURL returns a URL for accessing the file (if supported)
	GetURL(remotePath string) string
}

// Client wraps the storage provider with common functionality
type Client struct {
	provider Provider
	bucket   string
	cfg      *config.StorageConfig
}

// S3Client implements Provider for S3-compatible storage (S3, MinIO, etc.)
type S3Client struct {
	client *minio.Client
	bucket string
	cfg    *config.StorageConfig
}

// NewClient creates a new storage client from config
func NewClient(cfg *config.StorageConfig) (*Client, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("storage endpoint is required")
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("storage bucket is required")
	}

	var provider Provider
	var err error

	switch cfg.Provider {
	case "s3", "minio", "":
		provider, err = newS3Client(cfg)
	default:
		// Default to S3-compatible
		provider, err = newS3Client(cfg)
	}

	if err != nil {
		return nil, err
	}

	return &Client{
		provider: provider,
		bucket:   cfg.Bucket,
		cfg:      cfg,
	}, nil
}

// NewClientFromGlobal creates a client from global config
func NewClientFromGlobal() (*Client, error) {
	cfg := config.Get()
	if cfg == nil {
		return nil, fmt.Errorf("global config not loaded")
	}
	if !cfg.IsStorageConfigured() {
		return nil, fmt.Errorf("storage not configured")
	}
	return NewClient(&cfg.Storage)
}

// newS3Client creates an S3-compatible client
func newS3Client(cfg *config.StorageConfig) (*S3Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return &S3Client{
		client: client,
		bucket: cfg.Bucket,
		cfg:    cfg,
	}, nil
}

// Upload uploads a local file to cloud storage
func (c *S3Client) Upload(ctx context.Context, localPath, remotePath string) error {
	// Check if local file exists
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("local file not found: %s", localPath)
	}

	// Detect content type (optional, minio will auto-detect)
	contentType := "application/octet-stream"

	_, err := c.client.FPutObject(ctx, c.bucket, remotePath, localPath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// Download downloads a file from cloud storage to local path
func (c *S3Client) Download(ctx context.Context, remotePath, localPath string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	err := c.client.FGetObject(ctx, c.bucket, remotePath, localPath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

// Delete deletes a file from cloud storage
func (c *S3Client) Delete(ctx context.Context, remotePath string) error {
	err := c.client.RemoveObject(ctx, c.bucket, remotePath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// Exists checks if a file exists in cloud storage
func (c *S3Client) Exists(ctx context.Context, remotePath string) (bool, error) {
	_, err := c.client.StatObject(ctx, c.bucket, remotePath, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// List lists files with the given prefix
func (c *S3Client) List(ctx context.Context, prefix string) ([]string, error) {
	var files []string

	objectCh := c.client.ListObjects(ctx, c.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}
		files = append(files, object.Key)
	}

	return files, nil
}

// GetURL returns a URL for accessing the file
func (c *S3Client) GetURL(remotePath string) string {
	scheme := "http"
	if c.cfg.UseSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, c.cfg.Endpoint, c.bucket, remotePath)
}

// Client wrapper methods

// Upload uploads a local file to cloud storage
func (c *Client) Upload(ctx context.Context, localPath, remotePath string) error {
	return c.provider.Upload(ctx, localPath, remotePath)
}

// Download downloads a file from cloud storage to local path
func (c *Client) Download(ctx context.Context, remotePath, localPath string) error {
	return c.provider.Download(ctx, remotePath, localPath)
}

// Delete deletes a file from cloud storage
func (c *Client) Delete(ctx context.Context, remotePath string) error {
	return c.provider.Delete(ctx, remotePath)
}

// Exists checks if a file exists in cloud storage
func (c *Client) Exists(ctx context.Context, remotePath string) (bool, error) {
	return c.provider.Exists(ctx, remotePath)
}

// List lists files with the given prefix
func (c *Client) List(ctx context.Context, prefix string) ([]string, error) {
	return c.provider.List(ctx, prefix)
}

// GetURL returns a URL for accessing the file
func (c *Client) GetURL(remotePath string) string {
	return c.provider.GetURL(remotePath)
}

// UploadReader uploads data from a reader to cloud storage
func (c *Client) UploadReader(ctx context.Context, reader io.Reader, size int64, remotePath string) error {
	if s3c, ok := c.provider.(*S3Client); ok {
		_, err := s3c.client.PutObject(ctx, c.bucket, remotePath, reader, size, minio.PutObjectOptions{})
		return err
	}
	return fmt.Errorf("upload from reader not supported for this provider")
}

// Convenience functions for quick use without creating a client

// UploadFile uploads a file using global config
func UploadFile(ctx context.Context, localPath, remotePath string) error {
	client, err := NewClientFromGlobal()
	if err != nil {
		return err
	}
	return client.Upload(ctx, localPath, remotePath)
}

// DownloadFile downloads a file using global config
func DownloadFile(ctx context.Context, remotePath, localPath string) error {
	client, err := NewClientFromGlobal()
	if err != nil {
		return err
	}
	return client.Download(ctx, remotePath, localPath)
}

// DeleteFile deletes a file using global config
func DeleteFile(ctx context.Context, remotePath string) error {
	client, err := NewClientFromGlobal()
	if err != nil {
		return err
	}
	return client.Delete(ctx, remotePath)
}

// ListFiles lists files with the given prefix using global config
func ListFiles(ctx context.Context, prefix string) ([]string, error) {
	client, err := NewClientFromGlobal()
	if err != nil {
		return nil, err
	}
	return client.List(ctx, prefix)
}
