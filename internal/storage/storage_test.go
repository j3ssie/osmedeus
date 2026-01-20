package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgressReader(t *testing.T) {
	t.Run("tracks progress correctly", func(t *testing.T) {
		data := []byte("hello world test data for progress tracking")
		reader := bytes.NewReader(data)

		var lastTransferred, lastTotal int64
		callback := func(transferred, total int64) {
			lastTransferred = transferred
			lastTotal = total
		}

		pr := NewProgressReader(reader, int64(len(data)), callback)

		// Read in chunks
		buf := make([]byte, 10)
		totalRead := 0
		for {
			n, err := pr.Read(buf)
			totalRead += n
			if err != nil {
				break
			}
		}

		assert.Equal(t, len(data), totalRead)
		assert.Equal(t, int64(len(data)), lastTransferred)
		assert.Equal(t, int64(len(data)), lastTotal)
	})

	t.Run("works without callback", func(t *testing.T) {
		data := []byte("test data")
		reader := bytes.NewReader(data)

		pr := NewProgressReader(reader, int64(len(data)), nil)

		buf := make([]byte, 100)
		n, _ := pr.Read(buf)

		assert.Equal(t, len(data), n)
	})
}

func TestFileInfo(t *testing.T) {
	t.Run("struct initialization", func(t *testing.T) {
		now := time.Now()
		info := FileInfo{
			Key:          "test/file.txt",
			Size:         1024,
			LastModified: now,
			ETag:         "abc123",
			ContentType:  "text/plain",
		}

		assert.Equal(t, "test/file.txt", info.Key)
		assert.Equal(t, int64(1024), info.Size)
		assert.Equal(t, now, info.LastModified)
		assert.Equal(t, "abc123", info.ETag)
		assert.Equal(t, "text/plain", info.ContentType)
	})
}

func TestSyncResult(t *testing.T) {
	t.Run("struct initialization", func(t *testing.T) {
		result := SyncResult{
			Uploaded:   []string{"file1.txt", "file2.txt"},
			Downloaded: []string{},
			Skipped:    []string{"file3.txt"},
			Deleted:    []string{},
			Errors:     []error{},
		}

		assert.Equal(t, 2, len(result.Uploaded))
		assert.Equal(t, 0, len(result.Downloaded))
		assert.Equal(t, 1, len(result.Skipped))
		assert.Equal(t, 0, len(result.Deleted))
		assert.Equal(t, 0, len(result.Errors))
	})
}

func TestSyncOptions(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		opts := SyncOptions{}

		assert.False(t, opts.Delete)
		assert.False(t, opts.DryRun)
		assert.Nil(t, opts.Progress)
		assert.Equal(t, 0, opts.Concurrency)
	})

	t.Run("with all options set", func(t *testing.T) {
		callback := func(transferred, total int64) {}
		opts := SyncOptions{
			Delete:      true,
			DryRun:      true,
			Progress:    callback,
			Concurrency: 8,
		}

		assert.True(t, opts.Delete)
		assert.True(t, opts.DryRun)
		assert.NotNil(t, opts.Progress)
		assert.Equal(t, 8, opts.Concurrency)
	})
}

func TestNewClient(t *testing.T) {
	t.Run("requires endpoint", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Bucket: "test-bucket",
		}

		_, err := NewClient(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint is required")
	})

	t.Run("requires bucket", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Endpoint: "localhost:9000",
		}

		_, err := NewClient(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bucket is required")
	})

	t.Run("creates client with valid config", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Provider:        "minio",
			Endpoint:        "localhost:9000",
			AccessKeyID:     "minioadmin",
			SecretAccessKey: "minioadmin",
			Bucket:          "test-bucket",
			Region:          "us-east-1",
			UseSSL:          false,
		}

		client, err := NewClient(cfg)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "test-bucket", client.bucket)
	})

	t.Run("defaults to S3 provider", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Provider:        "", // empty provider
			Endpoint:        "localhost:9000",
			AccessKeyID:     "test",
			SecretAccessKey: "test",
			Bucket:          "test-bucket",
		}

		client, err := NewClient(cfg)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("handles unknown provider as S3", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Provider:        "unknown-provider",
			Endpoint:        "localhost:9000",
			AccessKeyID:     "test",
			SecretAccessKey: "test",
			Bucket:          "test-bucket",
		}

		client, err := NewClient(cfg)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestNewClientFromGlobal(t *testing.T) {
	t.Run("fails when global config not loaded", func(t *testing.T) {
		// Ensure no global config
		config.Set(nil)

		_, err := NewClientFromGlobal()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "global config not loaded")
	})

	t.Run("fails when storage not configured", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Storage.Enabled = false
		config.Set(cfg)

		_, err := NewClientFromGlobal()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "storage not configured")
	})
}

func TestGetClient(t *testing.T) {
	t.Run("returns error when no global config", func(t *testing.T) {
		ResetClient()
		config.Set(nil)

		_, err := GetClient()
		assert.Error(t, err)
	})

	t.Run("returns same client on multiple calls", func(t *testing.T) {
		ResetClient()
		cfg := config.DefaultConfig()
		cfg.Storage = config.StorageConfig{
			Provider:        "minio",
			Endpoint:        "localhost:9000",
			AccessKeyID:     "minioadmin",
			SecretAccessKey: "minioadmin",
			Bucket:          "test-bucket",
			Enabled:         true,
		}
		config.Set(cfg)

		client1, err1 := GetClient()
		client2, err2 := GetClient()

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Same(t, client1, client2, "should return same client instance")

		ResetClient()
	})
}

func TestResetClient(t *testing.T) {
	t.Run("clears singleton client", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Storage = config.StorageConfig{
			Provider:        "minio",
			Endpoint:        "localhost:9000",
			AccessKeyID:     "minioadmin",
			SecretAccessKey: "minioadmin",
			Bucket:          "test-bucket",
			Enabled:         true,
		}
		config.Set(cfg)

		client1, _ := GetClient()
		ResetClient()
		client2, _ := GetClient()

		assert.NotSame(t, client1, client2, "should create new client after reset")

		ResetClient()
	})
}

func TestS3ClientGetURL(t *testing.T) {
	t.Run("generates HTTP URL when SSL disabled", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Endpoint: "localhost:9000",
			Bucket:   "test-bucket",
			UseSSL:   false,
		}
		client := &S3Client{
			bucket: cfg.Bucket,
			cfg:    cfg,
		}

		url := client.GetURL("path/to/file.txt")
		assert.Equal(t, "http://localhost:9000/test-bucket/path/to/file.txt", url)
	})

	t.Run("generates HTTPS URL when SSL enabled", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Endpoint: "s3.amazonaws.com",
			Bucket:   "test-bucket",
			UseSSL:   true,
		}
		client := &S3Client{
			bucket: cfg.Bucket,
			cfg:    cfg,
		}

		url := client.GetURL("path/to/file.txt")
		assert.Equal(t, "https://s3.amazonaws.com/test-bucket/path/to/file.txt", url)
	})
}

func TestStorageConfigHelpers(t *testing.T) {
	t.Run("ResolveEndpoint for R2", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Provider:  "r2",
			AccountID: "abc123",
		}

		endpoint := cfg.ResolveEndpoint()
		assert.Equal(t, "abc123.r2.cloudflarestorage.com", endpoint)
	})

	t.Run("ResolveEndpoint for GCS", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Provider: "gcs",
		}

		endpoint := cfg.ResolveEndpoint()
		assert.Equal(t, "storage.googleapis.com", endpoint)
	})

	t.Run("ResolveEndpoint for Spaces", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Provider: "spaces",
			Region:   "nyc3",
		}

		endpoint := cfg.ResolveEndpoint()
		assert.Equal(t, "nyc3.digitaloceanspaces.com", endpoint)
	})

	t.Run("ResolveEndpoint for OCI", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Provider:  "oci",
			AccountID: "myns",
			Region:    "us-ashburn-1",
		}

		endpoint := cfg.ResolveEndpoint()
		assert.Equal(t, "myns.compat.objectstorage.us-ashburn-1.oraclecloud.com", endpoint)
	})

	t.Run("ResolveEndpoint for S3", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Provider: "s3",
			Region:   "us-west-2",
		}

		endpoint := cfg.ResolveEndpoint()
		assert.Equal(t, "s3.us-west-2.amazonaws.com", endpoint)
	})

	t.Run("ResolveEndpoint uses explicit endpoint", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Provider: "s3",
			Endpoint: "custom.endpoint.com",
			Region:   "us-west-2",
		}

		endpoint := cfg.ResolveEndpoint()
		assert.Equal(t, "custom.endpoint.com", endpoint)
	})

	t.Run("GetPresignExpiry default", func(t *testing.T) {
		cfg := &config.StorageConfig{}

		expiry := cfg.GetPresignExpiry()
		assert.Equal(t, time.Hour, expiry)
	})

	t.Run("GetPresignExpiry custom", func(t *testing.T) {
		cfg := &config.StorageConfig{
			PresignExpiry: "30m",
		}

		expiry := cfg.GetPresignExpiry()
		assert.Equal(t, 30*time.Minute, expiry)
	})

	t.Run("GetPresignExpiry invalid falls back to default", func(t *testing.T) {
		cfg := &config.StorageConfig{
			PresignExpiry: "invalid",
		}

		expiry := cfg.GetPresignExpiry()
		assert.Equal(t, time.Hour, expiry)
	})

	t.Run("ShouldUseSSL explicit true", func(t *testing.T) {
		cfg := &config.StorageConfig{
			UseSSL: true,
		}

		assert.True(t, cfg.ShouldUseSSL())
	})

	t.Run("ShouldUseSSL uses provider default", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Provider: "gcs",
			UseSSL:   false,
		}

		// GCS provider default is true
		assert.True(t, cfg.ShouldUseSSL())
	})

	t.Run("ShouldUsePathStyle explicit true", func(t *testing.T) {
		cfg := &config.StorageConfig{
			PathStyle: true,
		}

		assert.True(t, cfg.ShouldUsePathStyle())
	})

	t.Run("ShouldUsePathStyle uses provider default", func(t *testing.T) {
		cfg := &config.StorageConfig{
			Provider:  "r2",
			PathStyle: false,
		}

		// R2 provider default is true
		assert.True(t, cfg.ShouldUsePathStyle())
	})
}

// Integration tests that require actual storage (skipped by default)
func TestIntegration(t *testing.T) {
	// Skip if no MINIO_ENDPOINT environment variable
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		t.Skip("Skipping integration tests: MINIO_ENDPOINT not set")
	}

	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucket := os.Getenv("MINIO_BUCKET")

	if accessKey == "" || secretKey == "" || bucket == "" {
		t.Skip("Skipping integration tests: MINIO credentials not set")
	}

	cfg := &config.StorageConfig{
		Provider:        "minio",
		Endpoint:        endpoint,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		Bucket:          bucket,
		UseSSL:          false,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	t.Run("Upload and Download", func(t *testing.T) {
		ctx := t.Context()

		// Create temp file
		tmpDir := t.TempDir()
		localPath := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(localPath, []byte("test content"), 0644)
		require.NoError(t, err)

		remotePath := "test/integration/test.txt"

		// Upload
		err = client.Upload(ctx, localPath, remotePath)
		require.NoError(t, err)

		// Check exists
		exists, err := client.Exists(ctx, remotePath)
		require.NoError(t, err)
		assert.True(t, exists)

		// Download
		downloadPath := filepath.Join(tmpDir, "downloaded.txt")
		err = client.Download(ctx, remotePath, downloadPath)
		require.NoError(t, err)

		// Verify content
		content, err := os.ReadFile(downloadPath)
		require.NoError(t, err)
		assert.Equal(t, "test content", string(content))

		// Stat
		info, err := client.Stat(ctx, remotePath)
		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, remotePath, info.Key)
		assert.Equal(t, int64(12), info.Size)

		// Delete
		err = client.Delete(ctx, remotePath)
		require.NoError(t, err)

		// Verify deleted
		exists, err = client.Exists(ctx, remotePath)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("UploadWithProgress", func(t *testing.T) {
		ctx := t.Context()

		// Create temp file with larger content
		tmpDir := t.TempDir()
		localPath := filepath.Join(tmpDir, "progress-test.txt")
		content := bytes.Repeat([]byte("x"), 10000)
		err := os.WriteFile(localPath, content, 0644)
		require.NoError(t, err)

		remotePath := "test/integration/progress-test.txt"

		var progressCalls int
		var lastTransferred int64
		progress := func(transferred, total int64) {
			progressCalls++
			lastTransferred = transferred
		}

		err = client.UploadWithProgress(ctx, localPath, remotePath, progress)
		require.NoError(t, err)

		assert.Greater(t, progressCalls, 0)
		assert.Equal(t, int64(len(content)), lastTransferred)

		// Cleanup
		_ = client.Delete(ctx, remotePath)
	})

	t.Run("ListWithInfo", func(t *testing.T) {
		ctx := t.Context()

		// Upload test files
		tmpDir := t.TempDir()
		for i := 0; i < 3; i++ {
			localPath := filepath.Join(tmpDir, "list-test.txt")
			err := os.WriteFile(localPath, []byte("content"), 0644)
			require.NoError(t, err)

			remotePath := "test/integration/list/file" + string(rune('0'+i)) + ".txt"
			err = client.Upload(ctx, localPath, remotePath)
			require.NoError(t, err)
		}

		// List files
		files, err := client.ListWithInfo(ctx, "test/integration/list/")
		require.NoError(t, err)
		assert.Equal(t, 3, len(files))

		for _, f := range files {
			assert.NotEmpty(t, f.Key)
			assert.Greater(t, f.Size, int64(0))
			assert.False(t, f.LastModified.IsZero())
		}

		// Cleanup
		for _, f := range files {
			_ = client.Delete(ctx, f.Key)
		}
	})

	t.Run("PresignedGetURL", func(t *testing.T) {
		ctx := t.Context()

		// Upload test file
		tmpDir := t.TempDir()
		localPath := filepath.Join(tmpDir, "presign-test.txt")
		err := os.WriteFile(localPath, []byte("presigned content"), 0644)
		require.NoError(t, err)

		remotePath := "test/integration/presign-test.txt"
		err = client.Upload(ctx, localPath, remotePath)
		require.NoError(t, err)

		// Generate presigned URL
		url, err := client.PresignedGetURL(ctx, remotePath, time.Hour)
		require.NoError(t, err)
		assert.NotEmpty(t, url)
		assert.Contains(t, url, remotePath)

		// Cleanup
		_ = client.Delete(ctx, remotePath)
	})

	t.Run("SyncUpload", func(t *testing.T) {
		ctx := t.Context()

		// Create local directory structure
		tmpDir := t.TempDir()
		localDir := filepath.Join(tmpDir, "sync-test")
		require.NoError(t, os.MkdirAll(filepath.Join(localDir, "subdir"), 0755))

		require.NoError(t, os.WriteFile(filepath.Join(localDir, "file1.txt"), []byte("file1"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(localDir, "file2.txt"), []byte("file2"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(localDir, "subdir", "file3.txt"), []byte("file3"), 0644))

		remotePrefix := "test/integration/sync/"

		// Sync upload
		result, err := client.SyncUpload(ctx, localDir, remotePrefix, nil)
		require.NoError(t, err)
		assert.Equal(t, 3, len(result.Uploaded))
		assert.Equal(t, 0, len(result.Skipped))
		assert.Equal(t, 0, len(result.Errors))

		// Sync again - should skip all
		result, err = client.SyncUpload(ctx, localDir, remotePrefix, nil)
		require.NoError(t, err)
		assert.Equal(t, 0, len(result.Uploaded))
		assert.Equal(t, 3, len(result.Skipped))

		// Cleanup
		files, _ := client.List(ctx, remotePrefix)
		for _, f := range files {
			_ = client.Delete(ctx, f)
		}
	})

	t.Run("SyncDownload", func(t *testing.T) {
		ctx := t.Context()

		// Upload test files first
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "src")
		require.NoError(t, os.MkdirAll(srcDir, 0755))

		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "dl1.txt"), []byte("download1"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "dl2.txt"), []byte("download2"), 0644))

		remotePrefix := "test/integration/download/"
		_, err := client.SyncUpload(ctx, srcDir, remotePrefix, nil)
		require.NoError(t, err)

		// Sync download to new directory
		destDir := filepath.Join(tmpDir, "dest")
		result, err := client.SyncDownload(ctx, remotePrefix, destDir, nil)
		require.NoError(t, err)
		assert.Equal(t, 2, len(result.Downloaded))
		assert.Equal(t, 0, len(result.Errors))

		// Verify files exist
		content1, err := os.ReadFile(filepath.Join(destDir, "dl1.txt"))
		require.NoError(t, err)
		assert.Equal(t, "download1", string(content1))

		// Cleanup
		files, _ := client.List(ctx, remotePrefix)
		for _, f := range files {
			_ = client.Delete(ctx, f)
		}
	})
}
