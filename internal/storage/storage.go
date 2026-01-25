package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ProgressCallback for tracking upload/download progress
type ProgressCallback func(bytesTransferred, totalBytes int64)

// ProgressReader wraps io.Reader to track progress
type ProgressReader struct {
	reader   io.Reader
	total    int64
	current  int64
	callback ProgressCallback
}

// NewProgressReader creates a new progress tracking reader
func NewProgressReader(reader io.Reader, total int64, callback ProgressCallback) *ProgressReader {
	return &ProgressReader{
		reader:   reader,
		total:    total,
		callback: callback,
	}
}

// Read implements io.Reader with progress tracking
func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.current += int64(n)
		if pr.callback != nil {
			pr.callback(pr.current, pr.total)
		}
	}
	return n, err
}

// FileInfo represents file metadata from cloud storage
type FileInfo struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	ETag         string    `json:"etag"`
	ContentType  string    `json:"contentType"`
}

// SyncResult holds results of a sync operation
type SyncResult struct {
	Uploaded   []string `json:"uploaded"`
	Downloaded []string `json:"downloaded"`
	Skipped    []string `json:"skipped"`
	Deleted    []string `json:"deleted"`
	Errors     []error  `json:"-"`
}

type SyncEvent struct {
	Action string
	Path   string
}

// SyncOptions configures sync behavior
type SyncOptions struct {
	Delete      bool             // Delete remote files not in local
	DryRun      bool             // Don't actually transfer
	Progress    ProgressCallback // Optional progress callback
	Concurrency int              // Parallel transfers (default: 4)
	Event       func(SyncEvent)
}

// Singleton client pattern
var (
	globalClient   *Client
	globalClientMu sync.RWMutex
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

	// Extended methods for enhanced functionality

	// UploadWithProgress uploads a file with progress tracking
	UploadWithProgress(ctx context.Context, localPath, remotePath string, progress ProgressCallback) error

	// DownloadWithProgress downloads a file with progress tracking
	DownloadWithProgress(ctx context.Context, remotePath, localPath string, progress ProgressCallback) error

	// PresignedGetURL generates a presigned URL for downloading
	PresignedGetURL(ctx context.Context, remotePath string, expiry time.Duration) (string, error)

	// PresignedPutURL generates a presigned URL for uploading
	PresignedPutURL(ctx context.Context, remotePath string, expiry time.Duration) (string, error)

	// ListWithInfo lists files with full metadata
	ListWithInfo(ctx context.Context, prefix string) ([]FileInfo, error)

	// Stat returns metadata for a single file
	Stat(ctx context.Context, remotePath string) (*FileInfo, error)

	// ReadContent reads a remote file and returns its content as a string
	ReadContent(ctx context.Context, remotePath string) (string, error)
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

// GetClient returns the singleton storage client, creating it if needed
func GetClient() (*Client, error) {
	globalClientMu.RLock()
	if globalClient != nil {
		globalClientMu.RUnlock()
		return globalClient, nil
	}
	globalClientMu.RUnlock()

	globalClientMu.Lock()
	defer globalClientMu.Unlock()

	// Double-check after acquiring write lock
	if globalClient != nil {
		return globalClient, nil
	}

	client, err := NewClientFromGlobal()
	if err != nil {
		return nil, err
	}
	globalClient = client
	return globalClient, nil
}

// ResetClient clears the singleton client (for testing)
func ResetClient() {
	globalClientMu.Lock()
	defer globalClientMu.Unlock()
	globalClient = nil
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

// UploadWithProgress uploads a file with progress tracking
func (c *S3Client) UploadWithProgress(ctx context.Context, localPath, remotePath string, progress ProgressCallback) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	reader := NewProgressReader(file, fileInfo.Size(), progress)
	contentType := "application/octet-stream"

	_, err = c.client.PutObject(ctx, c.bucket, remotePath, reader, fileInfo.Size(), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	return nil
}

// DownloadWithProgress downloads a file with progress tracking
func (c *S3Client) DownloadWithProgress(ctx context.Context, remotePath, localPath string, progress ProgressCallback) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Get object info first for progress tracking
	objInfo, err := c.client.StatObject(ctx, c.bucket, remotePath, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to stat remote object: %w", err)
	}

	obj, err := c.client.GetObject(ctx, c.bucket, remotePath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer func() { _ = obj.Close() }()

	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer func() { _ = localFile.Close() }()

	reader := NewProgressReader(obj, objInfo.Size, progress)
	_, err = io.Copy(localFile, reader)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	return nil
}

// PresignedGetURL generates a presigned URL for downloading
func (c *S3Client) PresignedGetURL(ctx context.Context, remotePath string, expiry time.Duration) (string, error) {
	url, err := c.client.PresignedGetObject(ctx, c.bucket, remotePath, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned GET URL: %w", err)
	}
	return url.String(), nil
}

// PresignedPutURL generates a presigned URL for uploading
func (c *S3Client) PresignedPutURL(ctx context.Context, remotePath string, expiry time.Duration) (string, error) {
	url, err := c.client.PresignedPutObject(ctx, c.bucket, remotePath, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned PUT URL: %w", err)
	}
	return url.String(), nil
}

// ListWithInfo lists files with full metadata
func (c *S3Client) ListWithInfo(ctx context.Context, prefix string) ([]FileInfo, error) {
	var files []FileInfo

	objectCh := c.client.ListObjects(ctx, c.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}
		files = append(files, FileInfo{
			Key:          object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
			ETag:         object.ETag,
			ContentType:  object.ContentType,
		})
	}

	return files, nil
}

// Stat returns metadata for a single file
func (c *S3Client) Stat(ctx context.Context, remotePath string) (*FileInfo, error) {
	objInfo, err := c.client.StatObject(ctx, c.bucket, remotePath, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to stat object: %w", err)
	}
	return &FileInfo{
		Key:          objInfo.Key,
		Size:         objInfo.Size,
		LastModified: objInfo.LastModified,
		ETag:         objInfo.ETag,
		ContentType:  objInfo.ContentType,
	}, nil
}

// ReadContent reads a remote file and returns its content as a string
func (c *S3Client) ReadContent(ctx context.Context, remotePath string) (string, error) {
	obj, err := c.client.GetObject(ctx, c.bucket, remotePath, minio.GetObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get object: %w", err)
	}
	defer func() { _ = obj.Close() }()

	content, err := io.ReadAll(obj)
	if err != nil {
		return "", fmt.Errorf("failed to read object content: %w", err)
	}
	return string(content), nil
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

// UploadWithProgress uploads a file with progress tracking
func (c *Client) UploadWithProgress(ctx context.Context, localPath, remotePath string, progress ProgressCallback) error {
	return c.provider.UploadWithProgress(ctx, localPath, remotePath, progress)
}

// DownloadWithProgress downloads a file with progress tracking
func (c *Client) DownloadWithProgress(ctx context.Context, remotePath, localPath string, progress ProgressCallback) error {
	return c.provider.DownloadWithProgress(ctx, remotePath, localPath, progress)
}

// PresignedGetURL generates a presigned URL for downloading
func (c *Client) PresignedGetURL(ctx context.Context, remotePath string, expiry time.Duration) (string, error) {
	return c.provider.PresignedGetURL(ctx, remotePath, expiry)
}

// PresignedPutURL generates a presigned URL for uploading
func (c *Client) PresignedPutURL(ctx context.Context, remotePath string, expiry time.Duration) (string, error) {
	return c.provider.PresignedPutURL(ctx, remotePath, expiry)
}

// ListWithInfo lists files with full metadata
func (c *Client) ListWithInfo(ctx context.Context, prefix string) ([]FileInfo, error) {
	return c.provider.ListWithInfo(ctx, prefix)
}

// Stat returns metadata for a single file
func (c *Client) Stat(ctx context.Context, remotePath string) (*FileInfo, error) {
	return c.provider.Stat(ctx, remotePath)
}

// ReadContent reads a remote file and returns its content as a string
func (c *Client) ReadContent(ctx context.Context, remotePath string) (string, error) {
	return c.provider.ReadContent(ctx, remotePath)
}

// SyncUpload synchronizes a local directory to remote storage
func (c *Client) SyncUpload(ctx context.Context, localDir, remotePrefix string, opts *SyncOptions) (*SyncResult, error) {
	if opts == nil {
		opts = &SyncOptions{}
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = 4
	}

	result := &SyncResult{
		Uploaded:   []string{},
		Skipped:    []string{},
		Deleted:    []string{},
		Downloaded: []string{},
		Errors:     []error{},
	}

	// Get remote files for delta detection
	remoteFiles, err := c.ListWithInfo(ctx, remotePrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list remote files: %w", err)
	}

	remoteMap := make(map[string]FileInfo)
	for _, f := range remoteFiles {
		remoteMap[f.Key] = f
	}

	// Walk local directory
	localFiles := make(map[string]string) // relativePath -> absolutePath
	err = filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}
		localFiles[relPath] = path
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk local directory: %w", err)
	}

	// Determine files to upload
	for relPath, absPath := range localFiles {
		remotePath := filepath.Join(remotePrefix, relPath)
		remotePath = filepath.ToSlash(remotePath) // Convert to forward slashes for S3

		localInfo, err := os.Stat(absPath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("stat %s: %w", absPath, err))
			continue
		}

		// Check if remote file exists and is up-to-date
		if remoteInfo, exists := remoteMap[remotePath]; exists {
			// Skip if remote file is same size and modified time is not older
			if remoteInfo.Size == localInfo.Size() && !remoteInfo.LastModified.Before(localInfo.ModTime()) {
				result.Skipped = append(result.Skipped, remotePath)
				if opts.Event != nil {
					opts.Event(SyncEvent{Action: "skipped", Path: remotePath})
				}
				continue
			}
		}

		// Upload file
		if !opts.DryRun {
			if opts.Event != nil {
				opts.Event(SyncEvent{Action: "uploading", Path: remotePath})
			}
			if opts.Progress != nil {
				err = c.UploadWithProgress(ctx, absPath, remotePath, opts.Progress)
			} else {
				err = c.Upload(ctx, absPath, remotePath)
			}
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("upload %s: %w", absPath, err))
				if opts.Event != nil {
					opts.Event(SyncEvent{Action: "error", Path: remotePath})
				}
				continue
			}
		}
		result.Uploaded = append(result.Uploaded, remotePath)
		if opts.Event != nil {
			opts.Event(SyncEvent{Action: "uploaded", Path: remotePath})
		}
	}

	// Handle deletion of remote files not in local
	if opts.Delete {
		for remotePath := range remoteMap {
			relPath := remotePath
			if len(remotePrefix) > 0 {
				relPath = remotePath[len(remotePrefix):]
				relPath = filepath.FromSlash(relPath)
				relPath = filepath.Clean(relPath)
			}
			if _, exists := localFiles[relPath]; !exists {
				if !opts.DryRun {
					if err := c.Delete(ctx, remotePath); err != nil {
						result.Errors = append(result.Errors, fmt.Errorf("delete %s: %w", remotePath, err))
						if opts.Event != nil {
							opts.Event(SyncEvent{Action: "error", Path: remotePath})
						}
						continue
					}
				}
				result.Deleted = append(result.Deleted, remotePath)
				if opts.Event != nil {
					opts.Event(SyncEvent{Action: "deleted", Path: remotePath})
				}
			}
		}
	}

	return result, nil
}

// SyncDownload synchronizes remote storage to a local directory
func (c *Client) SyncDownload(ctx context.Context, remotePrefix, localDir string, opts *SyncOptions) (*SyncResult, error) {
	if opts == nil {
		opts = &SyncOptions{}
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = 4
	}

	result := &SyncResult{
		Uploaded:   []string{},
		Skipped:    []string{},
		Deleted:    []string{},
		Downloaded: []string{},
		Errors:     []error{},
	}

	// Ensure local directory exists
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local directory: %w", err)
	}

	// Get remote files
	remoteFiles, err := c.ListWithInfo(ctx, remotePrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list remote files: %w", err)
	}

	// Build map of local files for delete detection
	localFiles := make(map[string]os.FileInfo)
	if opts.Delete {
		err = filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			relPath, _ := filepath.Rel(localDir, path)
			localFiles[relPath] = info
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk local directory: %w", err)
		}
	}

	// Download remote files
	downloadedPaths := make(map[string]bool)
	for _, remoteInfo := range remoteFiles {
		relPath := remoteInfo.Key
		if len(remotePrefix) > 0 {
			relPath = remoteInfo.Key[len(remotePrefix):]
		}
		relPath = filepath.FromSlash(relPath)
		relPath = filepath.Clean(relPath)
		if relPath == "" || relPath == "." {
			continue
		}

		localPath := filepath.Join(localDir, relPath)
		downloadedPaths[relPath] = true

		// Check if local file exists and is up-to-date
		if localInfo, err := os.Stat(localPath); err == nil {
			if localInfo.Size() == remoteInfo.Size && !localInfo.ModTime().Before(remoteInfo.LastModified) {
				result.Skipped = append(result.Skipped, remoteInfo.Key)
				if opts.Event != nil {
					opts.Event(SyncEvent{Action: "skipped", Path: remoteInfo.Key})
				}
				continue
			}
		}

		// Download file
		if !opts.DryRun {
			if opts.Event != nil {
				opts.Event(SyncEvent{Action: "downloading", Path: remoteInfo.Key})
			}
			if opts.Progress != nil {
				err = c.DownloadWithProgress(ctx, remoteInfo.Key, localPath, opts.Progress)
			} else {
				err = c.Download(ctx, remoteInfo.Key, localPath)
			}
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("download %s: %w", remoteInfo.Key, err))
				if opts.Event != nil {
					opts.Event(SyncEvent{Action: "error", Path: remoteInfo.Key})
				}
				continue
			}
		}
		result.Downloaded = append(result.Downloaded, remoteInfo.Key)
		if opts.Event != nil {
			opts.Event(SyncEvent{Action: "downloaded", Path: remoteInfo.Key})
		}
	}

	// Handle deletion of local files not in remote
	if opts.Delete {
		for relPath := range localFiles {
			if !downloadedPaths[relPath] {
				localPath := filepath.Join(localDir, relPath)
				if !opts.DryRun {
					if err := os.Remove(localPath); err != nil {
						result.Errors = append(result.Errors, fmt.Errorf("delete local %s: %w", localPath, err))
						if opts.Event != nil {
							opts.Event(SyncEvent{Action: "error", Path: localPath})
						}
						continue
					}
				}
				result.Deleted = append(result.Deleted, relPath)
				if opts.Event != nil {
					opts.Event(SyncEvent{Action: "deleted", Path: localPath})
				}
			}
		}
	}

	return result, nil
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
