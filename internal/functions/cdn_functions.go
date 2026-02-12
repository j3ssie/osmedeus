package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/storage"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
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

// deleteWorkItem represents a single file to delete
type deleteWorkItem struct {
	index int
	key   string
}

// deleteWorkResult represents the result of a delete operation
type deleteWorkResult struct {
	index int
	key   string
	err   error
}

// cdnDelete deletes a file from cloud storage
// Usage: cdnDelete(remotePath, mode?) -> bool
// mode: "json" to suppress console output
func (vf *vmFunc) cdnDelete(call goja.FunctionCall) goja.Value {
	remotePath := call.Argument(0).String()
	mode := ""
	if len(call.Arguments) > 1 && !goja.IsUndefined(call.Argument(1)) {
		mode = strings.ToLower(call.Argument(1).String())
	}
	jsonOnly := mode == "json" || vf.getContext().suppressDetails
	logger.Get().Debug("Calling cdnDelete", zap.String("remotePath", remotePath))

	if remotePath == "undefined" || remotePath == "" {
		logger.Get().Warn("cdnDelete: empty remote path provided")
		return vf.vm.ToValue(false)
	}

	client, err := storage.GetClient()
	if err != nil {
		logger.Get().Warn("cdnDelete: failed to get storage client", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Define console output prefixes
	var infoPrefix, errPrefix string
	if !jsonOnly {
		infoPrefix = terminal.InfoSymbol() + " " + terminal.HiBlue("cdn_delete")
		errPrefix = terminal.ErrorSymbol() + " " + terminal.HiBlue("cdn_delete")
	}

	ctx := context.Background()
	deletedCount := 0
	errorCount := 0
	folderMode := strings.HasSuffix(remotePath, "/")

	// Print started message
	if !jsonOnly {
		fmt.Printf("%s %s %s\n", infoPrefix, terminal.Cyan("started"), terminal.Gray(remotePath))
	}

	if !folderMode {
		exists, existsErr := client.Exists(ctx, remotePath)
		if existsErr != nil {
			logger.Get().Warn("cdnDelete: failed to check existence", zap.String("remotePath", remotePath), zap.Error(existsErr))
		}
		if exists {
			if !jsonOnly {
				fmt.Printf("%s %s %s\n", infoPrefix, terminal.Blue("deleting"), terminal.Gray(remotePath))
			}
			if err := client.Delete(ctx, remotePath); err != nil {
				logger.Get().Warn("cdnDelete: delete failed", zap.String("remotePath", remotePath), zap.Error(err))
				if !jsonOnly {
					fmt.Printf("%s %s %s\n", errPrefix, terminal.Red("error"), terminal.Gray(remotePath))
				}
				errorCount++
			} else {
				if !jsonOnly {
					fmt.Printf("%s %s %s\n", infoPrefix, terminal.Green("deleted"), terminal.Gray(remotePath))
				}
				deletedCount++
			}
		} else {
			folderMode = true
		}
	}

	if folderMode {
		prefix := remotePath
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		files, listErr := client.List(ctx, prefix)
		if listErr != nil {
			logger.Get().Warn("cdnDelete: list failed", zap.String("prefix", prefix), zap.Error(listErr))
			errorCount++
		} else if len(files) > 0 {
			// Use concurrent deletion with 4 workers
			const concurrency = 4
			workQueue := make(chan deleteWorkItem, concurrency*2)
			results := make(chan deleteWorkResult, concurrency*2)

			var workerWg sync.WaitGroup
			var outputMu sync.Mutex

			// Start fixed worker pool
			for i := 0; i < concurrency; i++ {
				workerWg.Add(1)
				go func() {
					defer workerWg.Done()
					for work := range workQueue {
						// Check context cancellation
						if ctx.Err() != nil {
							results <- deleteWorkResult{index: work.index, key: work.key, err: ctx.Err()}
							continue
						}

						// Print deleting status
						if !jsonOnly {
							outputMu.Lock()
							fmt.Printf("%s %s %s\n", infoPrefix, terminal.Blue("deleting"), terminal.Gray(work.key))
							outputMu.Unlock()
						}

						// Delete file
						deleteErr := client.Delete(ctx, work.key)
						results <- deleteWorkResult{index: work.index, key: work.key, err: deleteErr}
					}
				}()
			}

			// Producer: feed work items into queue
			go func() {
				defer close(workQueue)
				for idx, key := range files {
					select {
					case workQueue <- deleteWorkItem{index: idx, key: key}:
					case <-ctx.Done():
						return
					}
				}
			}()

			// Collector: close results when all workers done
			go func() {
				workerWg.Wait()
				close(results)
			}()

			// Collect results
			for r := range results {
				if r.err != nil {
					logger.Get().Warn("cdnDelete: delete failed", zap.String("remotePath", r.key), zap.Error(r.err))
					if !jsonOnly {
						outputMu.Lock()
						fmt.Printf("%s %s %s\n", errPrefix, terminal.Red("error"), terminal.Gray(r.key))
						outputMu.Unlock()
					}
					errorCount++
				} else {
					if !jsonOnly {
						outputMu.Lock()
						fmt.Printf("%s %s %s\n", infoPrefix, terminal.Green("deleted"), terminal.Gray(r.key))
						outputMu.Unlock()
					}
					deletedCount++
				}
			}
		}
	}

	// Print summary
	if !jsonOnly {
		fmt.Printf("%s %s %s\n",
			infoPrefix,
			terminal.HiGreen("summary"),
			terminal.Gray(fmt.Sprintf("deleted=%d errors=%d", deletedCount, errorCount)))
	}

	if errorCount > 0 {
		logger.Get().Warn("cdnDelete: completed with errors",
			zap.String("remotePath", remotePath),
			zap.Int("deleted", deletedCount),
			zap.Int("errors", errorCount))
		return vf.vm.ToValue(false)
	}

	logger.Get().Debug("cdnDelete result",
		zap.String("remotePath", remotePath),
		zap.Int("deleted", deletedCount),
		zap.Bool("success", deletedCount > 0))
	return vf.vm.ToValue(deletedCount > 0)
}

// cdnSyncUpload synchronizes a local directory to cloud storage
// Usage: cdnSyncUpload(localDir, remotePrefix) -> JSON string {success: bool, uploaded: [], skipped: [], errors: []}
func (vf *vmFunc) cdnSyncUpload(call goja.FunctionCall) goja.Value {
	localDir := call.Argument(0).String()
	remotePrefix := call.Argument(1).String()
	mode := ""
	if len(call.Arguments) > 2 && !goja.IsUndefined(call.Argument(2)) {
		mode = strings.ToLower(call.Argument(2).String())
	}
	jsonOnly := mode == "json" || vf.getContext().suppressDetails
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
		jsonBytes, _ := json.Marshal(result)
		if jsonOnly {
			return vf.vm.ToValue(string(jsonBytes))
		}
		return vf.vm.ToValue("uploaded=0 skipped=0 deleted=0 errors=0")
	}

	client, err := storage.GetClient()
	if err != nil {
		logger.Get().Warn("cdnSyncUpload: failed to get storage client", zap.Error(err))
		jsonBytes, _ := json.Marshal(result)
		if jsonOnly {
			return vf.vm.ToValue(string(jsonBytes))
		}
		return vf.vm.ToValue("uploaded=0 skipped=0 deleted=0 errors=0")
	}

	ctx := context.Background()
	if !jsonOnly {
		prefix := terminal.InfoSymbol() + " " + terminal.HiBlue("cdn_sync_upload")
		fmt.Printf("%s %s %s\n", prefix, terminal.Cyan("started"), terminal.Gray(fmt.Sprintf("%s -> %s", localDir, remotePrefix)))
	}
	opts := &storage.SyncOptions{}
	if !jsonOnly {
		infoPrefix := terminal.InfoSymbol() + " " + terminal.HiBlue("cdn_sync_upload")
		warnPrefix := terminal.WarningSymbol() + " " + terminal.HiBlue("cdn_sync_upload")
		errPrefix := terminal.ErrorSymbol() + " " + terminal.HiBlue("cdn_sync_upload")
		opts.Event = func(event storage.SyncEvent) {
			switch event.Action {
			case "uploading":
				fmt.Printf("%s %s %s\n", infoPrefix, terminal.Blue("uploading"), terminal.Gray(event.Path))
			case "uploaded":
				fmt.Printf("%s %s %s\n", infoPrefix, terminal.Green("uploaded"), terminal.Gray(event.Path))
			case "skipped":
				fmt.Printf("%s %s %s\n", warnPrefix, terminal.Yellow("skipped"), terminal.Gray(event.Path))
			case "deleted":
				fmt.Printf("%s %s %s\n", warnPrefix, terminal.Magenta("deleted"), terminal.Gray(event.Path))
			case "error":
				fmt.Printf("%s %s %s\n", errPrefix, terminal.Red("error"), terminal.Gray(event.Path))
			}
		}
	}
	syncResult, err := client.SyncUpload(ctx, localDir, remotePrefix, opts)
	if err != nil {
		logger.Get().Warn("cdnSyncUpload: sync failed", zap.String("localDir", localDir), zap.Error(err))
		jsonBytes, _ := json.Marshal(result)
		if jsonOnly {
			return vf.vm.ToValue(string(jsonBytes))
		}
		return vf.vm.ToValue("uploaded=0 skipped=0 deleted=0 errors=0")
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

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		logger.Get().Warn("cdnSyncUpload: failed to marshal result", zap.Error(err))
		if jsonOnly {
			return vf.vm.ToValue("{}")
		}
		return vf.vm.ToValue("uploaded=0 skipped=0 deleted=0 errors=0")
	}
	if jsonOnly {
		return vf.vm.ToValue(string(jsonBytes))
	}
	if !jsonOnly {
		prefix := terminal.InfoSymbol() + " " + terminal.HiBlue("cdn_sync_upload")
		fmt.Printf("%s %s %s\n",
			prefix,
			terminal.HiGreen("summary"),
			terminal.Gray(fmt.Sprintf("uploaded=%d skipped=%d deleted=%d errors=%d",
				len(syncResult.Uploaded),
				len(syncResult.Skipped),
				len(syncResult.Deleted),
				len(syncResult.Errors))))
	}
	return vf.vm.ToValue(fmt.Sprintf("uploaded=%d skipped=%d deleted=%d errors=%d",
		len(syncResult.Uploaded),
		len(syncResult.Skipped),
		len(syncResult.Deleted),
		len(syncResult.Errors)))
}

// cdnSyncDownload synchronizes cloud storage to a local directory
// Usage: cdnSyncDownload(remotePrefix, localDir) -> JSON string {success: bool, downloaded: [], skipped: [], errors: []}
func (vf *vmFunc) cdnSyncDownload(call goja.FunctionCall) goja.Value {
	remotePrefix := call.Argument(0).String()
	localDir := call.Argument(1).String()
	mode := ""
	if len(call.Arguments) > 2 && !goja.IsUndefined(call.Argument(2)) {
		mode = strings.ToLower(call.Argument(2).String())
	}
	jsonOnly := mode == "json" || vf.getContext().suppressDetails
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
		jsonBytes, _ := json.Marshal(result)
		if jsonOnly {
			return vf.vm.ToValue(string(jsonBytes))
		}
		return vf.vm.ToValue("downloaded=0 skipped=0 deleted=0 errors=0")
	}

	client, err := storage.GetClient()
	if err != nil {
		logger.Get().Warn("cdnSyncDownload: failed to get storage client", zap.Error(err))
		jsonBytes, _ := json.Marshal(result)
		if jsonOnly {
			return vf.vm.ToValue(string(jsonBytes))
		}
		return vf.vm.ToValue("downloaded=0 skipped=0 deleted=0 errors=0")
	}

	ctx := context.Background()
	if !jsonOnly {
		prefix := terminal.InfoSymbol() + " " + terminal.HiBlue("cdn_sync_download")
		fmt.Printf("%s %s %s\n", prefix, terminal.Cyan("started"), terminal.Gray(fmt.Sprintf("%s -> %s", remotePrefix, localDir)))
	}
	opts := &storage.SyncOptions{}
	if !jsonOnly {
		infoPrefix := terminal.InfoSymbol() + " " + terminal.HiBlue("cdn_sync_download")
		warnPrefix := terminal.WarningSymbol() + " " + terminal.HiBlue("cdn_sync_download")
		errPrefix := terminal.ErrorSymbol() + " " + terminal.HiBlue("cdn_sync_download")
		opts.Event = func(event storage.SyncEvent) {
			switch event.Action {
			case "downloading":
				fmt.Printf("%s %s %s\n", infoPrefix, terminal.Blue("downloading"), terminal.Gray(event.Path))
			case "downloaded":
				fmt.Printf("%s %s %s\n", infoPrefix, terminal.Green("downloaded"), terminal.Gray(event.Path))
			case "skipped":
				fmt.Printf("%s %s %s\n", warnPrefix, terminal.Yellow("skipped"), terminal.Gray(event.Path))
			case "deleted":
				fmt.Printf("%s %s %s\n", warnPrefix, terminal.Magenta("deleted"), terminal.Gray(event.Path))
			case "error":
				fmt.Printf("%s %s %s\n", errPrefix, terminal.Red("error"), terminal.Gray(event.Path))
			}
		}
	}
	syncResult, err := client.SyncDownload(ctx, remotePrefix, localDir, opts)
	if err != nil {
		logger.Get().Warn("cdnSyncDownload: sync failed", zap.String("remotePrefix", remotePrefix), zap.Error(err))
		jsonBytes, _ := json.Marshal(result)
		if jsonOnly {
			return vf.vm.ToValue(string(jsonBytes))
		}
		return vf.vm.ToValue("downloaded=0 skipped=0 deleted=0 errors=0")
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

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		logger.Get().Warn("cdnSyncDownload: failed to marshal result", zap.Error(err))
		if jsonOnly {
			return vf.vm.ToValue("{}")
		}
		return vf.vm.ToValue("downloaded=0 skipped=0 deleted=0 errors=0")
	}
	if jsonOnly {
		return vf.vm.ToValue(string(jsonBytes))
	}
	if !jsonOnly {
		prefix := terminal.InfoSymbol() + " " + terminal.HiBlue("cdn_sync_download")
		fmt.Printf("%s %s %s\n",
			prefix,
			terminal.HiGreen("summary"),
			terminal.Gray(fmt.Sprintf("downloaded=%d skipped=%d deleted=%d errors=%d",
				len(syncResult.Downloaded),
				len(syncResult.Skipped),
				len(syncResult.Deleted),
				len(syncResult.Errors))))
	}
	return vf.vm.ToValue(fmt.Sprintf("downloaded=%d skipped=%d deleted=%d errors=%d",
		len(syncResult.Downloaded),
		len(syncResult.Skipped),
		len(syncResult.Deleted),
		len(syncResult.Errors)))
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

// cdnList lists files with metadata from cloud storage with optional glob pattern
// Usage: cdnList(pattern?) -> JSON string of [{key, size, lastModified, etag, contentType}]
// Supports glob patterns like "*", "scans/*", "de*", "/reports/*.html"
func (vf *vmFunc) cdnList(call goja.FunctionCall) goja.Value {
	pattern := ""
	if len(call.Arguments) > 0 && !goja.IsUndefined(call.Argument(0)) {
		pattern = call.Argument(0).String()
		if pattern == "undefined" {
			pattern = ""
		}
	}
	logger.Get().Debug("Calling cdnList", zap.String("pattern", pattern))

	client, err := storage.GetClient()
	if err != nil {
		logger.Get().Warn("cdnList: failed to get storage client", zap.Error(err))
		return vf.vm.ToValue("[]")
	}

	// Extract prefix for efficient listing (everything before the first wildcard)
	prefix := extractPrefixFromPattern(pattern)

	ctx := context.Background()
	files, err := client.ListWithInfo(ctx, prefix)
	if err != nil {
		logger.Get().Warn("cdnList: list failed", zap.String("pattern", pattern), zap.Error(err))
		return vf.vm.ToValue("[]")
	}

	// Filter by glob pattern if specified
	var filtered []storage.FileInfo
	if pattern != "" && pattern != prefix {
		for _, f := range files {
			if matchGlobPattern(pattern, f.Key) {
				filtered = append(filtered, f)
			}
		}
	} else {
		filtered = files
	}

	// Convert to JSON-friendly format
	result := make([]map[string]interface{}, 0, len(filtered))
	for _, f := range filtered {
		result = append(result, map[string]interface{}{
			"key":          f.Key,
			"size":         f.Size,
			"lastModified": f.LastModified.Format(time.RFC3339),
			"etag":         f.ETag,
			"contentType":  f.ContentType,
		})
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		logger.Get().Warn("cdnList: failed to marshal result", zap.Error(err))
		return vf.vm.ToValue("[]")
	}

	logger.Get().Debug("cdnList result", zap.String("pattern", pattern), zap.Int("count", len(result)))
	return vf.vm.ToValue(string(jsonBytes))
}

// extractPrefixFromPattern extracts the prefix before any wildcard character
func extractPrefixFromPattern(pattern string) string {
	for i, c := range pattern {
		if c == '*' || c == '?' || c == '[' {
			return pattern[:i]
		}
	}
	return pattern
}

// matchGlobPattern matches a key against a glob pattern
func matchGlobPattern(pattern, key string) bool {
	matched, err := filepath.Match(pattern, key)
	if err != nil {
		return false
	}
	if matched {
		return true
	}
	// Also try matching just the filename for patterns like "*.html"
	if !strings.Contains(pattern, "/") {
		matched, _ = filepath.Match(pattern, filepath.Base(key))
		return matched
	}
	return false
}

// cdnRead reads a file from cloud storage and returns its content
// Usage: cdnRead(remotePath) -> string
func (vf *vmFunc) cdnRead(call goja.FunctionCall) goja.Value {
	remotePath := call.Argument(0).String()
	logger.Get().Debug("Calling cdnRead", zap.String("remotePath", remotePath))

	if remotePath == "undefined" || remotePath == "" {
		logger.Get().Warn("cdnRead: empty remote path provided")
		return vf.vm.ToValue("")
	}

	client, err := storage.GetClient()
	if err != nil {
		logger.Get().Warn("cdnRead: failed to get storage client", zap.Error(err))
		return vf.vm.ToValue("")
	}

	ctx := context.Background()
	content, err := client.ReadContent(ctx, remotePath)
	if err != nil {
		logger.Get().Warn("cdnRead: read failed", zap.String("remotePath", remotePath), zap.Error(err))
		return vf.vm.ToValue("")
	}

	logger.Get().Debug("cdnRead result", zap.String("remotePath", remotePath), zap.Int("length", len(content)))
	return vf.vm.ToValue(content)
}

// cdnStat returns metadata for a single file from cloud storage
// Usage: cdnStat(remotePath) -> JSON string {key, size, lastModified, etag, contentType} | "null"
func (vf *vmFunc) cdnStat(call goja.FunctionCall) goja.Value {
	remotePath := call.Argument(0).String()
	logger.Get().Debug("Calling cdnStat", zap.String("remotePath", remotePath))

	if remotePath == "undefined" || remotePath == "" {
		logger.Get().Warn("cdnStat: empty remote path provided")
		return vf.vm.ToValue("null")
	}

	client, err := storage.GetClient()
	if err != nil {
		logger.Get().Warn("cdnStat: failed to get storage client", zap.Error(err))
		return vf.vm.ToValue("null")
	}

	ctx := context.Background()
	info, err := client.Stat(ctx, remotePath)
	if err != nil {
		logger.Get().Warn("cdnStat: stat failed", zap.String("remotePath", remotePath), zap.Error(err))
		return vf.vm.ToValue("null")
	}

	if info == nil {
		logger.Get().Debug("cdnStat: file not found", zap.String("remotePath", remotePath))
		return vf.vm.ToValue("null")
	}

	result := map[string]interface{}{
		"key":          info.Key,
		"size":         info.Size,
		"lastModified": info.LastModified.Format(time.RFC3339),
		"etag":         info.ETag,
		"contentType":  info.ContentType,
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		logger.Get().Warn("cdnStat: failed to marshal result", zap.Error(err))
		return vf.vm.ToValue("null")
	}

	logger.Get().Debug("cdnStat result", zap.String("remotePath", remotePath), zap.Int64("size", info.Size))
	return vf.vm.ToValue(string(jsonBytes))
}

// cdnLsTree lists files from cloud storage in a tree format
// Usage: cdnLsTree(prefix?) -> string (tree format output)
func (vf *vmFunc) cdnLsTree(call goja.FunctionCall) goja.Value {
	prefix := ""
	if len(call.Arguments) > 0 && !goja.IsUndefined(call.Argument(0)) {
		prefix = call.Argument(0).String()
		if prefix == "undefined" {
			prefix = ""
		}
	}
	depth := int64(1)
	if len(call.Arguments) > 1 && !goja.IsUndefined(call.Argument(1)) && !goja.IsNull(call.Argument(1)) {
		depth = call.Argument(1).ToInteger()
		if depth < 1 {
			depth = 1
		}
	}
	logger.Get().Debug("Calling cdnLsTree", zap.String("prefix", prefix))

	client, err := storage.GetClient()
	if err != nil {
		logger.Get().Warn("cdnLsTree: failed to get storage client", zap.Error(err))
		return vf.vm.ToValue("")
	}

	ctx := context.Background()
	files, err := client.ListWithInfo(ctx, prefix)
	if err != nil {
		logger.Get().Warn("cdnLsTree: list failed", zap.String("prefix", prefix), zap.Error(err))
		return vf.vm.ToValue("")
	}

	// Build tree structure
	tree := buildFileTree(files, prefix)
	output := renderTree(tree, prefix, int(depth))

	logger.Get().Debug("cdnLsTree result", zap.String("prefix", prefix), zap.Int("files", len(files)))
	return vf.vm.ToValue(output)
}

// treeNode represents a node in the file tree
type treeNode struct {
	name     string
	isDir    bool
	size     int64
	children map[string]*treeNode
}

// buildFileTree builds a tree structure from flat file list
func buildFileTree(files []storage.FileInfo, prefix string) *treeNode {
	root := &treeNode{
		name:     prefix,
		isDir:    true,
		children: make(map[string]*treeNode),
	}

	for _, f := range files {
		// Remove prefix from key for relative path
		relPath := f.Key
		if prefix != "" {
			relPath = strings.TrimPrefix(f.Key, prefix)
		}
		relPath = strings.TrimPrefix(relPath, "/")

		if relPath == "" {
			continue
		}

		parts := strings.Split(relPath, "/")
		current := root

		for i, part := range parts {
			if part == "" {
				continue
			}

			if current.children[part] == nil {
				isDir := i < len(parts)-1
				current.children[part] = &treeNode{
					name:     part,
					isDir:    isDir,
					children: make(map[string]*treeNode),
				}
			}
			if i == len(parts)-1 {
				current.children[part].size = f.Size
			}
			current = current.children[part]
		}
	}

	return root
}

// renderTree renders the tree structure as a string
func renderTree(root *treeNode, prefix string, depth int) string {
	var sb strings.Builder

	// Write root
	rootName := prefix
	if rootName == "" {
		rootName = "."
	}
	sb.WriteString(terminal.Cyan(rootName))
	sb.WriteString("\n")

	// Get sorted children
	if depth >= 1 {
		children := getSortedChildren(root)
		for i, child := range children {
			isLast := i == len(children)-1
			renderTreeNode(&sb, child, "", isLast, depth, 1)
		}
	}

	return sb.String()
}

// renderTreeNode recursively renders a tree node
func renderTreeNode(sb *strings.Builder, node *treeNode, indent string, isLast bool, depth, level int) {
	// Choose connector
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	sb.WriteString(indent)
	sb.WriteString(connector)
	if node.isDir {
		sb.WriteString(terminal.Cyan(node.name))
	} else {
		sb.WriteString(terminal.Green(node.name))
	}
	if !node.isDir {
		fmt.Fprintf(sb, " (%s)", formatSize(node.size))
	}
	sb.WriteString("\n")

	// Update indent for children
	childIndent := indent
	if isLast {
		childIndent += "    "
	} else {
		childIndent += "│   "
	}

	// Render children
	if node.isDir && level < depth {
		children := getSortedChildren(node)
		for i, child := range children {
			renderTreeNode(sb, child, childIndent, i == len(children)-1, depth, level+1)
		}
	}
}

// getSortedChildren returns children sorted: directories first, then files, alphabetically
func getSortedChildren(node *treeNode) []*treeNode {
	var dirs, files []*treeNode
	for _, child := range node.children {
		if child.isDir {
			dirs = append(dirs, child)
		} else {
			files = append(files, child)
		}
	}

	// Sort each group alphabetically
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].name < dirs[j].name })
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })

	return append(dirs, files...)
}

// formatSize formats bytes into human-readable size
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
