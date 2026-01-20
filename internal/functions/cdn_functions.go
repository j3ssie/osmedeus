package functions

import (
	"context"
	"time"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/storage"
	"go.uber.org/zap"
)

// cdnUpload uploads a file to cloud storage
// Usage: cdnUpload(localPath, remotePath) -> bool
func (vf *vmFunc) cdnUpload(call goja.FunctionCall) goja.Value {
	localPath := call.Argument(0).String()
	remotePath := call.Argument(1).String()
	logger.Get().Debug("Calling cdnUpload", zap.String("localPath", localPath), zap.String("remotePath", remotePath))

	if localPath == "undefined" || localPath == "" {
		logger.Get().Warn("cdnUpload: empty local path provided")
		return vf.vm.ToValue(false)
	}
	if remotePath == "undefined" || remotePath == "" {
		logger.Get().Warn("cdnUpload: empty remote path provided")
		return vf.vm.ToValue(false)
	}

	ctx := context.Background()
	err := storage.UploadFile(ctx, localPath, remotePath)
	if err != nil {
		logger.Get().Warn("cdnUpload: upload failed", zap.String("localPath", localPath), zap.Error(err))
	} else {
		logger.Get().Debug("cdnUpload result", zap.String("localPath", localPath), zap.String("remotePath", remotePath), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// cdnDownload downloads a file from cloud storage
// Usage: cdnDownload(remotePath, localPath) -> bool
func (vf *vmFunc) cdnDownload(call goja.FunctionCall) goja.Value {
	remotePath := call.Argument(0).String()
	localPath := call.Argument(1).String()
	logger.Get().Debug("Calling cdnDownload", zap.String("remotePath", remotePath), zap.String("localPath", localPath))

	if remotePath == "undefined" || remotePath == "" {
		logger.Get().Warn("cdnDownload: empty remote path provided")
		return vf.vm.ToValue(false)
	}
	if localPath == "undefined" || localPath == "" {
		logger.Get().Warn("cdnDownload: empty local path provided")
		return vf.vm.ToValue(false)
	}

	ctx := context.Background()
	err := storage.DownloadFile(ctx, remotePath, localPath)
	if err != nil {
		logger.Get().Warn("cdnDownload: download failed", zap.String("remotePath", remotePath), zap.Error(err))
	} else {
		logger.Get().Debug("cdnDownload result", zap.String("remotePath", remotePath), zap.String("localPath", localPath), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// cdnExists checks if a file exists in cloud storage
// Usage: cdnExists(remotePath) -> bool
func (vf *vmFunc) cdnExists(call goja.FunctionCall) goja.Value {
	remotePath := call.Argument(0).String()
	logger.Get().Debug("Calling cdnExists", zap.String("remotePath", remotePath))

	if remotePath == "undefined" || remotePath == "" {
		logger.Get().Warn("cdnExists: empty remote path provided")
		return vf.vm.ToValue(false)
	}

	client, err := storage.NewClientFromGlobal()
	if err != nil {
		logger.Get().Warn("cdnExists: failed to create storage client", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	ctx := context.Background()
	exists, _ := client.Exists(ctx, remotePath)
	logger.Get().Debug("cdnExists result", zap.String("remotePath", remotePath), zap.Bool("exists", exists))
	return vf.vm.ToValue(exists)
}

// cdnDelete deletes a file from cloud storage
// Usage: cdnDelete(remotePath) -> bool
func (vf *vmFunc) cdnDelete(call goja.FunctionCall) goja.Value {
	remotePath := call.Argument(0).String()
	logger.Get().Debug("Calling cdnDelete", zap.String("remotePath", remotePath))

	if remotePath == "undefined" || remotePath == "" {
		logger.Get().Warn("cdnDelete: empty remote path provided")
		return vf.vm.ToValue(false)
	}

	ctx := context.Background()
	err := storage.DeleteFile(ctx, remotePath)
	if err != nil {
		logger.Get().Warn("cdnDelete: delete failed", zap.String("remotePath", remotePath), zap.Error(err))
	} else {
		logger.Get().Debug("cdnDelete result", zap.String("remotePath", remotePath), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// cdnSyncUpload synchronizes a local directory to cloud storage
// Usage: cdnSyncUpload(localDir, remotePrefix) -> {success: bool, uploaded: [], skipped: [], errors: []}
func (vf *vmFunc) cdnSyncUpload(call goja.FunctionCall) goja.Value {
	localDir := call.Argument(0).String()
	remotePrefix := call.Argument(1).String()
	logger.Get().Debug("Calling cdnSyncUpload", zap.String("localDir", localDir), zap.String("remotePrefix", remotePrefix))

	result := map[string]interface{}{
		"success":    false,
		"uploaded":   []string{},
		"skipped":    []string{},
		"deleted":    []string{},
		"errorCount": 0,
	}

	if localDir == "undefined" || localDir == "" {
		logger.Get().Warn("cdnSyncUpload: empty local directory provided")
		return vf.vm.ToValue(result)
	}

	client, err := storage.GetClient()
	if err != nil {
		logger.Get().Warn("cdnSyncUpload: failed to get storage client", zap.Error(err))
		return vf.vm.ToValue(result)
	}

	ctx := context.Background()
	syncResult, err := client.SyncUpload(ctx, localDir, remotePrefix, nil)
	if err != nil {
		logger.Get().Warn("cdnSyncUpload: sync failed", zap.String("localDir", localDir), zap.Error(err))
		return vf.vm.ToValue(result)
	}

	result["success"] = len(syncResult.Errors) == 0
	result["uploaded"] = syncResult.Uploaded
	result["skipped"] = syncResult.Skipped
	result["deleted"] = syncResult.Deleted
	result["errorCount"] = len(syncResult.Errors)

	logger.Get().Debug("cdnSyncUpload result",
		zap.String("localDir", localDir),
		zap.String("remotePrefix", remotePrefix),
		zap.Int("uploaded", len(syncResult.Uploaded)),
		zap.Int("skipped", len(syncResult.Skipped)),
		zap.Int("errors", len(syncResult.Errors)))

	return vf.vm.ToValue(result)
}

// cdnSyncDownload synchronizes cloud storage to a local directory
// Usage: cdnSyncDownload(remotePrefix, localDir) -> {success: bool, downloaded: [], skipped: [], errors: []}
func (vf *vmFunc) cdnSyncDownload(call goja.FunctionCall) goja.Value {
	remotePrefix := call.Argument(0).String()
	localDir := call.Argument(1).String()
	logger.Get().Debug("Calling cdnSyncDownload", zap.String("remotePrefix", remotePrefix), zap.String("localDir", localDir))

	result := map[string]interface{}{
		"success":    false,
		"downloaded": []string{},
		"skipped":    []string{},
		"deleted":    []string{},
		"errorCount": 0,
	}

	if localDir == "undefined" || localDir == "" {
		logger.Get().Warn("cdnSyncDownload: empty local directory provided")
		return vf.vm.ToValue(result)
	}

	client, err := storage.GetClient()
	if err != nil {
		logger.Get().Warn("cdnSyncDownload: failed to get storage client", zap.Error(err))
		return vf.vm.ToValue(result)
	}

	ctx := context.Background()
	syncResult, err := client.SyncDownload(ctx, remotePrefix, localDir, nil)
	if err != nil {
		logger.Get().Warn("cdnSyncDownload: sync failed", zap.String("remotePrefix", remotePrefix), zap.Error(err))
		return vf.vm.ToValue(result)
	}

	result["success"] = len(syncResult.Errors) == 0
	result["downloaded"] = syncResult.Downloaded
	result["skipped"] = syncResult.Skipped
	result["deleted"] = syncResult.Deleted
	result["errorCount"] = len(syncResult.Errors)

	logger.Get().Debug("cdnSyncDownload result",
		zap.String("remotePrefix", remotePrefix),
		zap.String("localDir", localDir),
		zap.Int("downloaded", len(syncResult.Downloaded)),
		zap.Int("skipped", len(syncResult.Skipped)),
		zap.Int("errors", len(syncResult.Errors)))

	return vf.vm.ToValue(result)
}

// cdnGetPresignedURL generates a presigned URL for file access
// Usage: cdnGetPresignedURL(remotePath, expiryMins?) -> string
func (vf *vmFunc) cdnGetPresignedURL(call goja.FunctionCall) goja.Value {
	remotePath := call.Argument(0).String()
	logger.Get().Debug("Calling cdnGetPresignedURL", zap.String("remotePath", remotePath))

	if remotePath == "undefined" || remotePath == "" {
		logger.Get().Warn("cdnGetPresignedURL: empty remote path provided")
		return vf.vm.ToValue("")
	}

	// Get expiry from second argument or use default
	var expiry time.Duration
	if len(call.Arguments) > 1 && !goja.IsUndefined(call.Argument(1)) {
		expiryMins := call.Argument(1).ToInteger()
		if expiryMins > 0 {
			expiry = time.Duration(expiryMins) * time.Minute
		}
	}

	// Use config default if not specified
	if expiry == 0 {
		cfg := config.Get()
		if cfg != nil {
			expiry = cfg.Storage.GetPresignExpiry()
		} else {
			expiry = time.Hour
		}
	}

	client, err := storage.GetClient()
	if err != nil {
		logger.Get().Warn("cdnGetPresignedURL: failed to get storage client", zap.Error(err))
		return vf.vm.ToValue("")
	}

	ctx := context.Background()
	url, err := client.PresignedGetURL(ctx, remotePath, expiry)
	if err != nil {
		logger.Get().Warn("cdnGetPresignedURL: failed to generate URL", zap.String("remotePath", remotePath), zap.Error(err))
		return vf.vm.ToValue("")
	}

	logger.Get().Debug("cdnGetPresignedURL result", zap.String("remotePath", remotePath), zap.String("url", url))
	return vf.vm.ToValue(url)
}

// cdnList lists files with metadata from cloud storage
// Usage: cdnList(prefix?) -> [{key, size, lastModified, etag, contentType}]
func (vf *vmFunc) cdnList(call goja.FunctionCall) goja.Value {
	prefix := ""
	if len(call.Arguments) > 0 && !goja.IsUndefined(call.Argument(0)) {
		prefix = call.Argument(0).String()
		if prefix == "undefined" {
			prefix = ""
		}
	}
	logger.Get().Debug("Calling cdnList", zap.String("prefix", prefix))

	client, err := storage.GetClient()
	if err != nil {
		logger.Get().Warn("cdnList: failed to get storage client", zap.Error(err))
		return vf.vm.ToValue([]interface{}{})
	}

	ctx := context.Background()
	files, err := client.ListWithInfo(ctx, prefix)
	if err != nil {
		logger.Get().Warn("cdnList: list failed", zap.String("prefix", prefix), zap.Error(err))
		return vf.vm.ToValue([]interface{}{})
	}

	// Convert to JavaScript-friendly format
	result := make([]map[string]interface{}, 0, len(files))
	for _, f := range files {
		result = append(result, map[string]interface{}{
			"key":          f.Key,
			"size":         f.Size,
			"lastModified": f.LastModified.Format(time.RFC3339),
			"etag":         f.ETag,
			"contentType":  f.ContentType,
		})
	}

	logger.Get().Debug("cdnList result", zap.String("prefix", prefix), zap.Int("count", len(result)))
	return vf.vm.ToValue(result)
}

// cdnStat returns metadata for a single file from cloud storage
// Usage: cdnStat(remotePath) -> {key, size, lastModified, etag, contentType} | null
func (vf *vmFunc) cdnStat(call goja.FunctionCall) goja.Value {
	remotePath := call.Argument(0).String()
	logger.Get().Debug("Calling cdnStat", zap.String("remotePath", remotePath))

	if remotePath == "undefined" || remotePath == "" {
		logger.Get().Warn("cdnStat: empty remote path provided")
		return goja.Null()
	}

	client, err := storage.GetClient()
	if err != nil {
		logger.Get().Warn("cdnStat: failed to get storage client", zap.Error(err))
		return goja.Null()
	}

	ctx := context.Background()
	info, err := client.Stat(ctx, remotePath)
	if err != nil {
		logger.Get().Warn("cdnStat: stat failed", zap.String("remotePath", remotePath), zap.Error(err))
		return goja.Null()
	}

	if info == nil {
		logger.Get().Debug("cdnStat: file not found", zap.String("remotePath", remotePath))
		return goja.Null()
	}

	result := map[string]interface{}{
		"key":          info.Key,
		"size":         info.Size,
		"lastModified": info.LastModified.Format(time.RFC3339),
		"etag":         info.ETag,
		"contentType":  info.ContentType,
	}

	logger.Get().Debug("cdnStat result", zap.String("remotePath", remotePath), zap.Int64("size", info.Size))
	return vf.vm.ToValue(result)
}
