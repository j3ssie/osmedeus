package functions

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	mrand "math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/retry"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"go.uber.org/zap"
)

// httpRequest makes an HTTP request with custom method, headers, and body
// Usage: httpRequest(url, method, headers, body)
// headers is an object like {"Content-Type": "application/json", "Authorization": "Bearer token"}
// Returns: {statusCode: int, body: string, headers: object, error: string}
func (vf *vmFunc) httpRequest(call goja.FunctionCall) goja.Value {
	urlStr := call.Argument(0).String()
	method := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("httpRequest"), zap.String("url", urlStr), zap.String("method", method))

	if urlStr == "undefined" || urlStr == "" {
		logger.Get().Warn("httpRequest: url is required")
		return vf.vm.ToValue(map[string]interface{}{
			"statusCode": 0,
			"body":       "",
			"headers":    map[string]string{},
			"error":      "url is required",
		})
	}

	if method == "undefined" || method == "" {
		method = "GET"
	}

	// Parse headers
	var headers map[string]interface{}
	if !goja.IsUndefined(call.Argument(2)) {
		exported := call.Argument(2).Export()
		if h, ok := exported.(map[string]interface{}); ok {
			headers = h
		}
	}

	// Parse body
	var body io.Reader
	bodyStr := call.Argument(3).String()
	if bodyStr != "undefined" && bodyStr != "" {
		body = bytes.NewBufferString(bodyStr)
	}

	// Create request
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		logger.Get().Warn("httpRequest: failed to create request", zap.String("url", urlStr), zap.Error(err))
		return vf.vm.ToValue(map[string]interface{}{
			"statusCode": 0,
			"body":       "",
			"headers":    map[string]string{},
			"error":      err.Error(),
		})
	}

	// Set default User-Agent if not provided
	if headers == nil {
		headers = make(map[string]interface{})
	}
	if _, hasUA := headers["User-Agent"]; !hasUA {
		headers["User-Agent"] = core.DefaultUA
	}

	// Set headers
	for k, v := range headers {
		if str, ok := v.(string); ok {
			req.Header.Set(k, str)
		}
	}

	// Create client with timeout and TLS skip verify
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Execute request with retry logic (3 attempts, exponential backoff)
	var resp *http.Response
	ctx := context.Background()
	retryCfg := retry.Config{
		MaxAttempts:  3,
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     2 * time.Second,
		Multiplier:   2.0,
	}

	err = retry.Do(ctx, retryCfg, func() error {
		var reqErr error
		// Need to recreate body for retries since it may have been consumed
		if bodyStr != "undefined" && bodyStr != "" {
			req.Body = io.NopCloser(bytes.NewBufferString(bodyStr))
		}
		resp, reqErr = client.Do(req)
		if reqErr != nil {
			return retry.Retryable(reqErr)
		}
		// Retry on server errors (5xx)
		if resp.StatusCode >= 500 {
			_ = resp.Body.Close()
			return retry.Retryable(fmt.Errorf("server error: %d", resp.StatusCode))
		}
		return nil
	})

	if err != nil {
		logger.Get().Warn("httpRequest: request failed after retries", zap.String("url", urlStr), zap.Error(err))
		return vf.vm.ToValue(map[string]interface{}{
			"statusCode": 0,
			"body":       "",
			"headers":    map[string]string{},
			"error":      err.Error(),
		})
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Get().Warn("httpRequest: failed to read response body", zap.String("url", urlStr), zap.Error(err))
		return vf.vm.ToValue(map[string]interface{}{
			"statusCode": resp.StatusCode,
			"body":       "",
			"headers":    map[string]string{},
			"error":      err.Error(),
		})
	}

	// Extract response headers
	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}

	logger.Get().Debug(terminal.HiGreen("httpRequest")+" result", zap.String("url", urlStr), zap.Int("statusCode", resp.StatusCode), zap.Int("bodySize", len(respBody)))
	return vf.vm.ToValue(map[string]interface{}{
		"statusCode": resp.StatusCode,
		"body":       string(respBody),
		"headers":    respHeaders,
		"error":      "",
	})
}

// httpGet makes a GET request to the specified URL
// Usage: http_get(url)
// Returns: {status_code, error, message, response_headers, response_body, content_length, response_time_ms}
func (vf *vmFunc) httpGet(call goja.FunctionCall) goja.Value {
	start := time.Now()
	urlStr := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("http_get"), zap.String("url", urlStr))

	if urlStr == "undefined" || urlStr == "" {
		logger.Get().Warn("http_get: url is required")
		return vf.vm.ToValue(map[string]interface{}{
			"status_code":      0,
			"error":            "url is required",
			"message":          "error",
			"response_headers": map[string]string{},
			"response_body":    "",
			"content_length":   0,
			"response_time_ms": time.Since(start).Milliseconds(),
		})
	}

	// Create request
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		logger.Get().Warn("http_get: failed to create request", zap.String("url", urlStr), zap.Error(err))
		return vf.vm.ToValue(map[string]interface{}{
			"status_code":      0,
			"error":            err.Error(),
			"message":          "error",
			"response_headers": map[string]string{},
			"response_body":    "",
			"content_length":   0,
			"response_time_ms": time.Since(start).Milliseconds(),
		})
	}

	// Set default User-Agent
	req.Header.Set("User-Agent", core.DefaultUA)

	// Create client with timeout and TLS skip verify
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Execute request with retry logic (3 attempts, exponential backoff)
	var resp *http.Response
	ctx := context.Background()
	retryCfg := retry.Config{
		MaxAttempts:  3,
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     2 * time.Second,
		Multiplier:   2.0,
	}

	err = retry.Do(ctx, retryCfg, func() error {
		var reqErr error
		resp, reqErr = client.Do(req)
		if reqErr != nil {
			return retry.Retryable(reqErr)
		}
		// Retry on server errors (5xx)
		if resp.StatusCode >= 500 {
			_ = resp.Body.Close()
			return retry.Retryable(fmt.Errorf("server error: %d", resp.StatusCode))
		}
		return nil
	})

	if err != nil {
		logger.Get().Warn("http_get: request failed after retries", zap.String("url", urlStr), zap.Error(err))
		return vf.vm.ToValue(map[string]interface{}{
			"status_code":      0,
			"error":            err.Error(),
			"message":          "error",
			"response_headers": map[string]string{},
			"response_body":    "",
			"content_length":   0,
			"response_time_ms": time.Since(start).Milliseconds(),
		})
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Get().Warn("http_get: failed to read response body", zap.String("url", urlStr), zap.Error(err))
		return vf.vm.ToValue(map[string]interface{}{
			"status_code":      resp.StatusCode,
			"error":            err.Error(),
			"message":          "error",
			"response_headers": map[string]string{},
			"response_body":    "",
			"content_length":   0,
			"response_time_ms": time.Since(start).Milliseconds(),
		})
	}

	// Extract response headers
	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}

	elapsed := time.Since(start).Milliseconds()
	logger.Get().Debug(terminal.HiGreen("http_get")+" result", zap.String("url", urlStr), zap.Int("status_code", resp.StatusCode), zap.Int("bodySize", len(respBody)), zap.Int64("response_time_ms", elapsed))
	return vf.vm.ToValue(map[string]interface{}{
		"status_code":      resp.StatusCode,
		"error":            nil,
		"message":          "success",
		"response_headers": respHeaders,
		"response_body":    string(respBody),
		"content_length":   len(respBody),
		"response_time_ms": elapsed,
	})
}

// httpPost makes a POST request to the specified URL with the given body
// Usage: http_post(url, body)
// Returns: {status_code, error, message, response_headers, response_body, content_length, response_time_ms}
func (vf *vmFunc) httpPost(call goja.FunctionCall) goja.Value {
	start := time.Now()
	urlStr := call.Argument(0).String()
	bodyStr := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("http_post"), zap.String("url", urlStr), zap.Int("bodyLength", len(bodyStr)))

	if urlStr == "undefined" || urlStr == "" {
		logger.Get().Warn("http_post: url is required")
		return vf.vm.ToValue(map[string]interface{}{
			"status_code":      0,
			"error":            "url is required",
			"message":          "error",
			"response_headers": map[string]string{},
			"response_body":    "",
			"content_length":   0,
			"response_time_ms": time.Since(start).Milliseconds(),
		})
	}

	// Handle undefined body as empty string
	var body io.Reader
	if bodyStr != "undefined" && bodyStr != "" {
		body = bytes.NewBufferString(bodyStr)
	}

	// Create request
	req, err := http.NewRequest("POST", urlStr, body)
	if err != nil {
		logger.Get().Warn("http_post: failed to create request", zap.String("url", urlStr), zap.Error(err))
		return vf.vm.ToValue(map[string]interface{}{
			"status_code":      0,
			"error":            err.Error(),
			"message":          "error",
			"response_headers": map[string]string{},
			"response_body":    "",
			"content_length":   0,
			"response_time_ms": time.Since(start).Milliseconds(),
		})
	}

	// Set default User-Agent and Content-Type
	req.Header.Set("User-Agent", core.DefaultUA)
	if bodyStr != "undefined" && bodyStr != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Create client with timeout and TLS skip verify
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Execute request with retry logic (3 attempts, exponential backoff)
	var resp *http.Response
	ctx := context.Background()
	retryCfg := retry.Config{
		MaxAttempts:  3,
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     2 * time.Second,
		Multiplier:   2.0,
	}

	err = retry.Do(ctx, retryCfg, func() error {
		var reqErr error
		// Need to recreate body for retries since it may have been consumed
		if bodyStr != "undefined" && bodyStr != "" {
			req.Body = io.NopCloser(bytes.NewBufferString(bodyStr))
		}
		resp, reqErr = client.Do(req)
		if reqErr != nil {
			return retry.Retryable(reqErr)
		}
		// Retry on server errors (5xx)
		if resp.StatusCode >= 500 {
			_ = resp.Body.Close()
			return retry.Retryable(fmt.Errorf("server error: %d", resp.StatusCode))
		}
		return nil
	})

	if err != nil {
		logger.Get().Warn("http_post: request failed after retries", zap.String("url", urlStr), zap.Error(err))
		return vf.vm.ToValue(map[string]interface{}{
			"status_code":      0,
			"error":            err.Error(),
			"message":          "error",
			"response_headers": map[string]string{},
			"response_body":    "",
			"content_length":   0,
			"response_time_ms": time.Since(start).Milliseconds(),
		})
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Get().Warn("http_post: failed to read response body", zap.String("url", urlStr), zap.Error(err))
		return vf.vm.ToValue(map[string]interface{}{
			"status_code":      resp.StatusCode,
			"error":            err.Error(),
			"message":          "error",
			"response_headers": map[string]string{},
			"response_body":    "",
			"content_length":   0,
			"response_time_ms": time.Since(start).Milliseconds(),
		})
	}

	// Extract response headers
	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}

	elapsed := time.Since(start).Milliseconds()
	logger.Get().Debug(terminal.HiGreen("http_post")+" result", zap.String("url", urlStr), zap.Int("status_code", resp.StatusCode), zap.Int("bodySize", len(respBody)), zap.Int64("response_time_ms", elapsed))
	return vf.vm.ToValue(map[string]interface{}{
		"status_code":      resp.StatusCode,
		"error":            nil,
		"message":          "success",
		"response_headers": respHeaders,
		"response_body":    string(respBody),
		"content_length":   len(respBody),
		"response_time_ms": elapsed,
	})
}

// randomString generates a random alphanumeric string of given length
// Usage: randomString(length)
// Default length is 16 if not specified or invalid
func (vf *vmFunc) randomString(call goja.FunctionCall) goja.Value {
	length := call.Argument(0).ToInteger()
	logger.Get().Debug("Calling "+terminal.HiGreen("randomString"), zap.Int64("length", length))

	if length <= 0 {
		length = 16
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rng := mrand.New(mrand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rng.Intn(len(charset))]
	}

	logger.Get().Debug(terminal.HiGreen("randomString")+" result", zap.Int64("length", length))
	return vf.vm.ToValue(string(b))
}

// uuidFunc generates a UUID v4
// Usage: uuid()
func (vf *vmFunc) uuidFunc(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("uuid"))
	id := uuid.New()
	logger.Get().Debug(terminal.HiGreen("uuid")+" result", zap.String("uuid", id.String()))
	return vf.vm.ToValue(id.String())
}

// base64Encode encodes a string to base64
// Usage: base64Encode(str)
func (vf *vmFunc) base64Encode(call goja.FunctionCall) goja.Value {
	s := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("base64Encode"), zap.Int("inputLength", len(s)))

	if s == "undefined" {
		logger.Get().Warn("base64Encode: undefined input")
		return vf.vm.ToValue("")
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(s))
	logger.Get().Debug(terminal.HiGreen("base64Encode")+" result", zap.Int("outputLength", len(encoded)))
	return vf.vm.ToValue(encoded)
}

// base64Decode decodes a base64 string
// Usage: base64Decode(str)
func (vf *vmFunc) base64Decode(call goja.FunctionCall) goja.Value {
	s := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("base64Decode"), zap.Int("inputLength", len(s)))

	if s == "undefined" {
		logger.Get().Warn("base64Decode: undefined input")
		return vf.vm.ToValue("")
	}

	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		logger.Get().Warn("base64Decode: failed to decode", zap.Error(err))
		return vf.vm.ToValue("")
	}

	logger.Get().Debug(terminal.HiGreen("base64Decode")+" result", zap.Int("outputLength", len(decoded)))
	return vf.vm.ToValue(string(decoded))
}

// removeFile removes a file
// Usage: removeFile(path) -> returns true on success, false on failure
// Will fail if path is a directory
func (vf *vmFunc) removeFile(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("removeFile"), zap.String("path", path))

	if path == "undefined" || path == "" {
		logger.Get().Warn("removeFile: empty path provided")
		return vf.vm.ToValue(false)
	}

	// Check if path is a file (not directory)
	info, err := os.Stat(path)
	if err != nil {
		logger.Get().Warn("removeFile: file does not exist", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	if info.IsDir() {
		logger.Get().Warn("removeFile: path is a directory, not a file", zap.String("path", path))
		return vf.vm.ToValue(false)
	}

	err = os.Remove(path)
	if err != nil {
		logger.Get().Warn("removeFile: failed to remove file", zap.String("path", path), zap.Error(err))
	} else {
		logger.Get().Debug(terminal.HiGreen("removeFile")+" result", zap.String("path", path), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// removeFolder removes a folder and all its contents
// Usage: removeFolder(path) -> returns true on success, false on failure
// Will fail if path is not a directory
func (vf *vmFunc) removeFolder(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("removeFolder"), zap.String("path", path))

	if path == "undefined" || path == "" {
		logger.Get().Warn("removeFolder: empty path provided")
		return vf.vm.ToValue(false)
	}

	// Check if path is a directory
	info, err := os.Stat(path)
	if err != nil {
		logger.Get().Warn("removeFolder: path does not exist", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	if !info.IsDir() {
		logger.Get().Warn("removeFolder: path is not a directory", zap.String("path", path))
		return vf.vm.ToValue(false)
	}

	err = os.RemoveAll(path)
	if err != nil {
		logger.Get().Warn("removeFolder: failed to remove folder", zap.String("path", path), zap.Error(err))
	} else {
		logger.Get().Debug(terminal.HiGreen("removeFolder")+" result", zap.String("path", path), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

func (vf *vmFunc) rmRF(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("rm_rf"), zap.String("path", path))

	if path == "undefined" || path == "" {
		logger.Get().Warn("rm_rf: empty path provided")
		return vf.vm.ToValue(false)
	}

	err := os.RemoveAll(path)
	if err != nil {
		logger.Get().Warn("rm_rf: failed to remove path", zap.String("path", path), zap.Error(err))
	} else {
		logger.Get().Debug(terminal.HiGreen("rm_rf")+" result", zap.String("path", path), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

func (vf *vmFunc) removeAllExcept(call goja.FunctionCall) goja.Value {
	folder := call.Argument(0).String()
	keep := filepath.FromSlash(call.Argument(1).String())
	logger.Get().Debug("Calling "+terminal.HiGreen("remove_all_except"), zap.String("folder", folder), zap.String("keep", keep))

	if folder == "undefined" || folder == "" || keep == "undefined" || keep == "" {
		logger.Get().Warn("remove_all_except: empty folder or keep provided")
		return vf.vm.ToValue(false)
	}

	folderAbs := filepath.Clean(folder)
	info, err := os.Stat(folderAbs)
	if err != nil {
		logger.Get().Warn("remove_all_except: folder does not exist", zap.String("folder", folderAbs), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	if !info.IsDir() {
		logger.Get().Warn("remove_all_except: folder is not a directory", zap.String("folder", folderAbs))
		return vf.vm.ToValue(false)
	}

	keepAbs := keep
	if !filepath.IsAbs(keepAbs) {
		keepAbs = filepath.Join(folderAbs, keepAbs)
	}
	keepAbs = filepath.Clean(keepAbs)

	rel, err := filepath.Rel(folderAbs, keepAbs)
	if err != nil {
		logger.Get().Warn("remove_all_except: failed to compute relative keep path", zap.String("folder", folderAbs), zap.String("keep", keepAbs), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	if rel == "." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || rel == ".." {
		logger.Get().Warn("remove_all_except: keep path must be inside folder", zap.String("folder", folderAbs), zap.String("keep", keepAbs))
		return vf.vm.ToValue(false)
	}

	keepInfo, err := os.Stat(keepAbs)
	if err != nil {
		logger.Get().Warn("remove_all_except: keep file does not exist", zap.String("keep", keepAbs), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	if keepInfo.IsDir() {
		logger.Get().Warn("remove_all_except: keep path is a directory", zap.String("keep", keepAbs))
		return vf.vm.ToValue(false)
	}

	parts := strings.Split(rel, string(os.PathSeparator))
	if len(parts) == 0 {
		return vf.vm.ToValue(false)
	}

	if err := pruneExcept(folderAbs, parts); err != nil {
		logger.Get().Warn("remove_all_except: prune failed", zap.String("folder", folderAbs), zap.String("keep", keepAbs), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	return vf.vm.ToValue(true)
}

func pruneExcept(dir string, keepRelParts []string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(keepRelParts) == 0 {
		return nil
	}
	keepName := keepRelParts[0]
	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(dir, name)
		if name != keepName {
			if err := os.RemoveAll(fullPath); err != nil {
				return err
			}
			continue
		}
		if len(keepRelParts) > 1 {
			if !entry.IsDir() {
				return fmt.Errorf("keep path parent is not a directory: %s", fullPath)
			}
			if err := pruneExcept(fullPath, keepRelParts[1:]); err != nil {
				return err
			}
		}
	}
	return nil
}

// printf prints a message to stdout and returns the message
// Usage: printf(message) -> string
func (vf *vmFunc) printf(call goja.FunctionCall) goja.Value {
	msg := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("printf"), zap.Int("msgLength", len(msg)))

	if msg != "undefined" {
		fmt.Println(msg)
	}
	return vf.vm.ToValue(msg)
}

// catFile prints file content to stdout and returns the content
// Usage: cat_file(path) -> string
func (vf *vmFunc) catFile(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("catFile"), zap.String("path", path))

	if path == "undefined" || path == "" {
		logger.Get().Warn("catFile: empty path provided")
		return vf.vm.ToValue("")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		logger.Get().Warn("catFile: failed to read file", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue("")
	}

	logger.Get().Debug(terminal.HiGreen("catFile")+" result", zap.String("path", path), zap.Int("bytes", len(content)))
	fmt.Print(string(content))
	return vf.vm.ToValue(string(content))
}

// exit exits the scan with the given code
// Usage: exit(code) - 0 for success, non-zero for error
func (vf *vmFunc) exit(call goja.FunctionCall) goja.Value {
	code := call.Argument(0).ToInteger()
	logger.Get().Debug("Calling "+terminal.HiGreen("exit"), zap.Int64("code", code))
	os.Exit(int(code))
	return goja.Undefined()
}

// skip stops execution of the remaining steps in the current module.
// The flow continues to the next module. Accepts an optional message.
// Usage: skip(message?) -> void
func (vf *vmFunc) skip(call goja.FunctionCall) goja.Value {
	msg := call.Argument(0).String()
	if msg == "undefined" || msg == "" {
		msg = "skip() called"
	}
	logger.Get().Info("Module skip requested: " + msg)
	panic(vf.vm.NewGoError(&SkipModuleError{Message: msg}))
}

// execCmd executes a bash command and returns the stdout output
// Usage: exec_cmd(command) -> string
func (vf *vmFunc) execCmd(call goja.FunctionCall) goja.Value {
	command := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("execCmd"), zap.String("command", command))

	if command == "undefined" || command == "" {
		logger.Get().Warn("execCmd: empty command provided")
		return vf.vm.ToValue("")
	}

	// @NOTE: This is intentional - exec_cmd() is a utility function exposed to workflow
	// definitions for executing shell commands. Input comes from trusted workflow YAML files.
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		logger.Get().Warn("execCmd: command failed", zap.String("command", command), zap.Error(err))
		return vf.vm.ToValue("")
	}

	logger.Get().Debug(terminal.HiGreen("execCmd")+" result", zap.String("command", command), zap.Int("outputLength", len(output)))
	return vf.vm.ToValue(strings.TrimSpace(string(output)))
}

func (vf *vmFunc) bash(call goja.FunctionCall) goja.Value {
	return vf.execCmd(call)
}

// findPythonBin returns "python3" if available, otherwise "python".
func findPythonBin() string {
	if _, err := exec.LookPath("python3"); err == nil {
		return "python3"
	}
	return "python"
}

// execPython runs inline Python code via `python3 -c '<code>'`.
// Usage: exec_python(code) -> string
func (vf *vmFunc) execPython(call goja.FunctionCall) goja.Value {
	code := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("execPython"), zap.Int("codeLength", len(code)))

	if code == "undefined" || code == "" {
		logger.Get().Warn("execPython: empty code provided")
		return vf.vm.ToValue("")
	}

	pythonBin := findPythonBin()
	// @NOTE: This is intentional - exec_python() is a utility function exposed to workflow
	// definitions for executing Python code. Input comes from trusted workflow YAML files.
	cmd := exec.Command(pythonBin, "-c", code)
	output, err := cmd.Output()
	if err != nil {
		logger.Get().Warn("execPython: command failed", zap.String("python", pythonBin), zap.Error(err))
		return vf.vm.ToValue("")
	}

	logger.Get().Debug(terminal.HiGreen("execPython")+" result", zap.Int("outputLength", len(output)))
	return vf.vm.ToValue(strings.TrimSpace(string(output)))
}

// execPythonFile runs a Python file via `python3 <path>`.
// Usage: exec_python_file(path) -> string
func (vf *vmFunc) execPythonFile(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("execPythonFile"), zap.String("path", path))

	if path == "undefined" || path == "" {
		logger.Get().Warn("execPythonFile: empty path provided")
		return vf.vm.ToValue("")
	}

	pythonBin := findPythonBin()
	// @NOTE: This is intentional - exec_python_file() is a utility function exposed to workflow
	// definitions for executing Python files. Input comes from trusted workflow YAML files.
	cmd := exec.Command(pythonBin, path)
	output, err := cmd.Output()
	if err != nil {
		logger.Get().Warn("execPythonFile: command failed", zap.String("python", pythonBin), zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue("")
	}

	logger.Get().Debug(terminal.HiGreen("execPythonFile")+" result", zap.String("path", path), zap.Int("outputLength", len(output)))
	return vf.vm.ToValue(strings.TrimSpace(string(output)))
}

// commandExists checks if a command is available in PATH
// Usage: commandExists(command) -> bool
func (vf *vmFunc) commandExists(call goja.FunctionCall) goja.Value {
	command := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("commandExists"), zap.String("command", command))

	if command == "undefined" || command == "" {
		logger.Get().Warn("commandExists: empty command provided")
		return vf.vm.ToValue(false)
	}

	_, err := exec.LookPath(command)
	exists := err == nil

	logger.Get().Debug(terminal.HiGreen("commandExists")+" result",
		zap.String("command", command),
		zap.Bool("exists", exists))
	return vf.vm.ToValue(exists)
}

// logDebug logs a debug message with [DEBUG] prefix and returns the message
// Usage: log_debug(message) -> string
func (vf *vmFunc) logDebug(call goja.FunctionCall) goja.Value {
	msg := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("logDebug"), zap.Int("msgLength", len(msg)))

	if msg != "undefined" {
		fmt.Printf("[DEBUG] %s\n", msg)
	}
	return vf.vm.ToValue(msg)
}

// logInfo logs an info message with [INFO] prefix and returns the message
// Usage: log_info(message) -> string
func (vf *vmFunc) logInfo(call goja.FunctionCall) goja.Value {
	msg := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("logInfo"), zap.Int("msgLength", len(msg)))

	if msg != "undefined" {
		fmt.Printf("[INFO] %s\n", msg)
	}
	return vf.vm.ToValue(msg)
}

func (vf *vmFunc) logWarn(call goja.FunctionCall) goja.Value {
	msg := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("logWarn"), zap.Int("msgLength", len(msg)))

	if msg != "undefined" {
		fmt.Printf("%s %s\n", terminal.Yellow("[WARN]"), msg)
	}
	return vf.vm.ToValue(msg)
}

func (vf *vmFunc) logError(call goja.FunctionCall) goja.Value {
	msg := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("logError"), zap.Int("msgLength", len(msg)))

	if msg != "undefined" {
		fmt.Printf("%s %s\n", terminal.Red("[ERROR]"), msg)
	}
	return vf.vm.ToValue(msg)
}

// printGreen prints a message in green color and returns the message
// Usage: print_green(message) -> string
func (vf *vmFunc) printGreen(call goja.FunctionCall) goja.Value {
	msg := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("printGreen"), zap.Int("msgLength", len(msg)))

	if msg != "undefined" {
		fmt.Println(terminal.Green(msg))
	}
	return vf.vm.ToValue(msg)
}

// printBlue prints a message in blue color and returns the message
// Usage: print_blue(message) -> string
func (vf *vmFunc) printBlue(call goja.FunctionCall) goja.Value {
	msg := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("printBlue"), zap.Int("msgLength", len(msg)))

	if msg != "undefined" {
		fmt.Println(terminal.Blue(msg))
	}
	return vf.vm.ToValue(msg)
}

// printYellow prints a message in yellow color and returns the message
// Usage: print_yellow(message) -> string
func (vf *vmFunc) printYellow(call goja.FunctionCall) goja.Value {
	msg := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("printYellow"), zap.Int("msgLength", len(msg)))

	if msg != "undefined" {
		fmt.Println(terminal.Yellow(msg))
	}
	return vf.vm.ToValue(msg)
}

// printRed prints a message in red color and returns the message
// Usage: print_red(message) -> string
func (vf *vmFunc) printRed(call goja.FunctionCall) goja.Value {
	msg := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("printRed"), zap.Int("msgLength", len(msg)))

	if msg != "undefined" {
		fmt.Println(terminal.Red(msg))
	}
	return vf.vm.ToValue(msg)
}

// setVar sets a runtime variable that can be retrieved with get_var
// Also sets the variable on the VM for immediate access in the same execution
// Usage: set_var(name, value) -> string (returns the value)
func (vf *vmFunc) setVar(call goja.FunctionCall) goja.Value {
	name := call.Argument(0).String()
	value := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("setVar"), zap.String("name", name), zap.String("value", value))

	if name == "undefined" || name == "" {
		logger.Get().Warn("setVar: name is required")
		return vf.vm.ToValue("")
	}

	// Handle undefined value as empty string
	if value == "undefined" {
		value = ""
	}

	ctx := vf.getContext()
	if ctx != nil {
		if ctx.RuntimeVars == nil {
			ctx.RuntimeVars = make(map[string]string)
		}
		ctx.RuntimeVars[name] = value
		// Also set on VM for immediate access in same execution
		_ = vf.vm.Set(name, value)
	}

	logger.Get().Debug(terminal.HiGreen("setVar")+" result", zap.String("name", name), zap.String("value", value))
	return vf.vm.ToValue(value)
}

// getVar retrieves a runtime variable set with set_var
// Usage: get_var(name) -> string (returns empty string if not found)
func (vf *vmFunc) getVar(call goja.FunctionCall) goja.Value {
	name := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("getVar"), zap.String("name", name))

	if name == "undefined" || name == "" {
		logger.Get().Warn("getVar: name is required")
		return vf.vm.ToValue("")
	}

	ctx := vf.getContext()
	if ctx != nil && ctx.RuntimeVars != nil {
		if val, ok := ctx.RuntimeVars[name]; ok {
			logger.Get().Debug(terminal.HiGreen("getVar")+" result", zap.String("name", name), zap.String("value", val))
			return vf.vm.ToValue(val)
		}
	}

	logger.Get().Debug(terminal.HiGreen("getVar")+" result (not found)", zap.String("name", name))
	return vf.vm.ToValue("")
}

// sleep pauses execution for the given number of seconds
// Usage: sleep(seconds) -> void
func (vf *vmFunc) sleep(call goja.FunctionCall) goja.Value {
	seconds := call.Argument(0).ToInteger()
	logger.Get().Debug("Calling "+terminal.HiGreen("sleep"), zap.Int64("seconds", seconds))

	if seconds <= 0 {
		seconds = 1
	}

	time.Sleep(time.Duration(seconds) * time.Second)
	logger.Get().Debug(terminal.HiGreen("sleep")+" completed", zap.Int64("seconds", seconds))
	return goja.Undefined()
}

// saveContent saves content to a file
// Usage: save_content(content, path) -> bool
func (vf *vmFunc) saveContent(call goja.FunctionCall) goja.Value {
	content := call.Argument(0).String()
	path := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("saveContent"), zap.String("path", path), zap.Int("contentLength", len(content)))

	if path == "undefined" || path == "" {
		logger.Get().Warn("saveContent: empty path provided")
		return vf.vm.ToValue(false)
	}

	// Handle undefined content as empty string
	if content == "undefined" {
		content = ""
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Get().Warn("saveContent: failed to create directory", zap.String("dir", dir), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Write content to file
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		logger.Get().Warn("saveContent: failed to write file", zap.String("path", path), zap.Error(err))
	} else {
		logger.Get().Debug(terminal.HiGreen("saveContent")+" result", zap.String("path", path), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// moveFile moves a file from source to destination
// Uses os.Rename for same-filesystem moves (instant),
// falls back to buffered copy+delete for cross-filesystem moves
// Usage: moveFile(source, dest) -> bool
func (vf *vmFunc) moveFile(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()
	dest := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("moveFile"),
		zap.String("source", source), zap.String("dest", dest))

	// Validate arguments
	if source == "undefined" || source == "" || dest == "undefined" || dest == "" {
		logger.Get().Warn("moveFile: empty source or dest provided")
		return vf.vm.ToValue(false)
	}

	// Check source exists and is a file
	info, err := os.Stat(source)
	if err != nil {
		logger.Get().Warn("moveFile: source does not exist",
			zap.String("source", source), zap.Error(err))
		return vf.vm.ToValue(false)
	}
	if info.IsDir() {
		logger.Get().Warn("moveFile: source is a directory, not a file",
			zap.String("source", source))
		return vf.vm.ToValue(false)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		logger.Get().Warn("moveFile: failed to create destination directory",
			zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Try atomic rename first (fast, same filesystem)
	err = os.Rename(source, dest)
	if err == nil {
		logger.Get().Debug(terminal.HiGreen("moveFile")+" result (renamed)",
			zap.String("source", source), zap.String("dest", dest), zap.Bool("success", true))
		return vf.vm.ToValue(true)
	}

	// Fallback: buffered copy + delete (cross-filesystem)
	if err := copyFileBuffered(source, dest, info.Mode()); err != nil {
		logger.Get().Warn("moveFile: failed to copy file",
			zap.String("source", source), zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Remove source after successful copy
	if err := os.Remove(source); err != nil {
		logger.Get().Warn("moveFile: copied but failed to remove source",
			zap.String("source", source), zap.Error(err))
		// Still return true since the file was successfully moved
	}

	logger.Get().Debug(terminal.HiGreen("moveFile")+" result (copied)",
		zap.String("source", source), zap.String("dest", dest), zap.Bool("success", true))
	return vf.vm.ToValue(true)
}

// snapshotExport exports a workspace as a ZIP snapshot by running osmedeus snapshot export as a subprocess.
// Usage: snapshot_export(workspace, dest?) -> string (zip path on success, empty on failure)
func (vf *vmFunc) snapshotExport(call goja.FunctionCall) goja.Value {
	workspace := call.Argument(0).String()
	dest := call.Argument(1).String()

	if workspace == "undefined" || workspace == "" {
		logger.Get().Warn("snapshot_export: workspace is required")
		return vf.vm.ToValue("")
	}
	if dest == "undefined" {
		dest = ""
	}

	// Find the current executable
	exePath, err := os.Executable()
	if err != nil {
		logger.Get().Warn("snapshot_export: failed to find executable", zap.Error(err))
		return vf.vm.ToValue("")
	}

	// Build command arguments
	args := []string{"snapshot", "export", workspace}
	if dest != "" {
		args = append(args, "-o", dest)
	}

	logger.Get().Debug("Calling "+terminal.HiGreen("snapshot_export"),
		zap.String("workspace", workspace),
		zap.String("dest", dest))

	// @NOTE: This is intentional - snapshot_export() is a utility function exposed to workflow
	// definitions for exporting workspaces. Input comes from trusted workflow YAML files.
	cmd := exec.Command(exePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Get().Warn("snapshot_export: command failed",
			zap.String("workspace", workspace),
			zap.Error(err),
			zap.String("output", string(output)))
		return vf.vm.ToValue("")
	}

	// Parse output to extract the zip file path from "File: <path>" line
	var zipPath string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if idx := strings.Index(line, "File:"); idx >= 0 {
			zipPath = strings.TrimSpace(line[idx+len("File:"):])
			break
		}
	}

	logger.Get().Debug(terminal.HiGreen("snapshot_export")+" result",
		zap.String("workspace", workspace),
		zap.String("zipPath", zipPath))
	return vf.vm.ToValue(zipPath)
}

// snapshotImport imports a workspace from a ZIP snapshot by running osmedeus snapshot import as a subprocess.
// Usage: snapshot_import(source) -> string (workspace name on success, empty on failure)
func (vf *vmFunc) snapshotImport(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()

	if source == "undefined" || source == "" {
		logger.Get().Warn("snapshot_import: source is required")
		return vf.vm.ToValue("")
	}

	// Find the current executable
	exePath, err := os.Executable()
	if err != nil {
		logger.Get().Warn("snapshot_import: failed to find executable", zap.Error(err))
		return vf.vm.ToValue("")
	}

	// Build command arguments with --force to skip interactive confirmation
	args := []string{"snapshot", "import", source, "--force"}

	logger.Get().Debug("Calling "+terminal.HiGreen("snapshot_import"),
		zap.String("source", source))

	// @NOTE: This is intentional - snapshot_import() is a utility function exposed to workflow
	// definitions for importing workspaces. Input comes from trusted workflow YAML files.
	cmd := exec.Command(exePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Get().Warn("snapshot_import: command failed",
			zap.String("source", source),
			zap.Error(err),
			zap.String("output", string(output)))
		return vf.vm.ToValue("")
	}

	// Parse output to extract workspace name from "Workspace: <name>" line
	var workspaceName string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if idx := strings.Index(line, "Workspace:"); idx >= 0 {
			workspaceName = strings.TrimSpace(line[idx+len("Workspace:"):])
			break
		}
	}

	logger.Get().Debug(terminal.HiGreen("snapshot_import")+" result",
		zap.String("source", source),
		zap.String("workspace", workspaceName))
	return vf.vm.ToValue(workspaceName)
}

// parseParamsToFlags parses a comma-separated "key=value" string into -p flags.
// E.g. "threads=10,deep=true" -> ["-p", "threads=10", "-p", "deep=true"]
func parseParamsToFlags(params string) []string {
	params = strings.TrimSpace(params)
	if params == "" || params == "undefined" {
		return nil
	}
	var flags []string
	for _, pair := range strings.Split(params, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" || !strings.Contains(pair, "=") {
			continue
		}
		flags = append(flags, "-p", pair)
	}
	return flags
}

// runOsmedeus executes the current osmedeus binary with the given flag (-m or -f), name, target, and optional params.
// Returns the combined stdout+stderr output as a string.
func (vf *vmFunc) runOsmedeus(flag, name, target, params, funcName string) goja.Value {
	// Find the current executable
	exePath, err := os.Executable()
	if err != nil {
		logger.Get().Warn(funcName+": failed to find executable", zap.Error(err))
		return vf.vm.ToValue("")
	}

	// Build command arguments
	args := []string{"run", flag, name, "-t", target}
	args = append(args, parseParamsToFlags(params)...)

	logger.Get().Debug("Calling "+terminal.HiGreen(funcName),
		zap.String("exe", exePath),
		zap.Strings("args", args))

	// @NOTE: This is intentional - run_module/run_flow are utility functions exposed to workflow
	// definitions for launching sub-scans. Input comes from trusted workflow YAML files.
	cmd := exec.Command(exePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Get().Warn(funcName+": command failed",
			zap.String("name", name),
			zap.String("target", target),
			zap.Error(err),
			zap.String("output", string(output)))
	}

	logger.Get().Debug(terminal.HiGreen(funcName)+" result",
		zap.String("name", name),
		zap.String("target", target),
		zap.Int("outputLength", len(output)))
	return vf.vm.ToValue(strings.TrimSpace(string(output)))
}

// runModule runs an osmedeus module as a subprocess
// Usage: run_module(module, target, params?) -> string
// params is optional comma-separated key=value pairs: "threads=10,deep=true"
func (vf *vmFunc) runModule(call goja.FunctionCall) goja.Value {
	module := call.Argument(0).String()
	target := call.Argument(1).String()
	params := call.Argument(2).String()

	if module == "undefined" || module == "" || target == "undefined" || target == "" {
		logger.Get().Warn("run_module: module and target are required")
		return vf.vm.ToValue("")
	}

	return vf.runOsmedeus("-m", module, target, params, "run_module")
}

// runFlow runs an osmedeus flow as a subprocess
// Usage: run_flow(flow, target, params?) -> string
// params is optional comma-separated key=value pairs: "threads=10,deep=true"
func (vf *vmFunc) runFlow(call goja.FunctionCall) goja.Value {
	flow := call.Argument(0).String()
	target := call.Argument(1).String()
	params := call.Argument(2).String()

	if flow == "undefined" || flow == "" || target == "undefined" || target == "" {
		logger.Get().Warn("run_flow: flow and target are required")
		return vf.vm.ToValue("")
	}

	return vf.runOsmedeus("-f", flow, target, params, "run_flow")
}

// runOnMaster executes an action on the master node in distributed mode.
// In distributed mode, the request is sent via Redis to the master.
// In standalone mode, falls back to local execution.
//
// Usage:
//
//	run_on_master('func', 'db_import_sarif("ws", "/path/file.sarif")') -> bool
//	run_on_master('run', 'subdomain', 'example.com', 'threads=10') -> bool
//	run_on_master('bash', 'nmap -sV target.com') -> bool
func (vf *vmFunc) runOnMaster(call goja.FunctionCall) goja.Value {
	action := call.Argument(0).String()
	if action == "undefined" || action == "" {
		logger.Get().Warn("run_on_master: action is required (func, run, or bash)")
		return vf.vm.ToValue(false)
	}

	log := logger.Get()

	switch action {
	case "func":
		expr := call.Argument(1).String()
		if expr == "undefined" || expr == "" {
			log.Warn("run_on_master: expression is required for 'func' action")
			return vf.vm.ToValue(false)
		}

		log.Debug("Calling "+terminal.HiGreen("run_on_master"),
			zap.String("action", "func"),
			zap.String("expr", expr))

		// Try distributed path first
		if trySendExecuteRequest("func", expr, "", "", "", "master", "") {
			log.Debug("run_on_master: sent func request to master via Redis")
			return vf.vm.ToValue(true)
		}

		// Fallback: execute locally using the current runtime
		log.Warn("run_on_master: not in distributed mode, executing locally",
			zap.String("action", action),
			zap.String("hint", "ensure worker mode is active and Redis is configured"))
		execCtx := make(map[string]interface{})
		_, err := vf.runtime.Execute(expr, execCtx)
		if err != nil {
			log.Warn("run_on_master: local execution failed",
				zap.String("expr", expr),
				zap.Error(err))
			return vf.vm.ToValue(false)
		}
		return vf.vm.ToValue(true)

	case "run":
		workflow := call.Argument(1).String()
		target := call.Argument(2).String()
		params := call.Argument(3).String()

		if workflow == "undefined" || workflow == "" || target == "undefined" || target == "" {
			log.Warn("run_on_master: workflow and target are required for 'run' action")
			return vf.vm.ToValue(false)
		}
		if params == "undefined" {
			params = ""
		}

		log.Debug("Calling "+terminal.HiGreen("run_on_master"),
			zap.String("action", "run"),
			zap.String("workflow", workflow),
			zap.String("target", target),
			zap.String("params", params))

		// Try distributed path first
		if trySendExecuteRequest("run", "", workflow, target, params, "master", "") {
			log.Debug("run_on_master: sent run request to master via Redis")
			return vf.vm.ToValue(true)
		}

		// Fallback: run locally as subprocess
		log.Warn("run_on_master: not in distributed mode, running locally",
			zap.String("action", action),
			zap.String("hint", "ensure worker mode is active and Redis is configured"))
		vf.runOsmedeus("-m", workflow, target, params, "run_on_master")
		return vf.vm.ToValue(true)

	case "bash":
		command := call.Argument(1).String()
		if command == "undefined" || command == "" {
			log.Warn("run_on_master: command is required for 'bash' action")
			return vf.vm.ToValue(false)
		}

		log.Debug("Calling "+terminal.HiGreen("run_on_master"),
			zap.String("action", "bash"),
			zap.String("command", command))

		// Try distributed path first
		if trySendExecuteRequest("bash", command, "", "", "", "master", "") {
			log.Debug("run_on_master: sent bash request to master via Redis")
			return vf.vm.ToValue(true)
		}

		// Fallback: execute locally
		log.Warn("run_on_master: not in distributed mode, executing bash locally")
		// @NOTE: This is intentional - run_on_master('bash', cmd) is a utility function
		// exposed to workflow definitions. Input comes from trusted workflow YAML files.
		cmd := exec.Command("sh", "-c", command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Warn("run_on_master: local bash execution failed",
				zap.String("command", command), zap.Error(err),
				zap.String("output", string(output)))
			return vf.vm.ToValue(false)
		}
		return vf.vm.ToValue(true)

	default:
		log.Warn("run_on_master: unknown action",
			zap.String("action", action),
			zap.String("valid", "func, run, bash"))
		return vf.vm.ToValue(false)
	}
}

// runOnWorker executes an action on worker node(s) in distributed mode.
// In distributed mode, the request is routed via Redis through the master to target workers.
// In standalone mode, falls back to local execution.
//
// Usage:
//
//	run_on_worker('all', 'func', 'log_info("hello")') -> bool
//	run_on_worker('scanner-1', 'run', 'subdomain', 'example.com', 'threads=10') -> bool
//	run_on_worker('all', 'bash', 'apt update && apt install -y nmap') -> bool
func (vf *vmFunc) runOnWorker(call goja.FunctionCall) goja.Value {
	scope := call.Argument(0).String()
	if scope == "undefined" || scope == "" {
		scope = "all"
	}
	action := call.Argument(1).String()
	if action == "undefined" || action == "" {
		logger.Get().Warn("run_on_worker: action is required (func, run, bash)")
		return vf.vm.ToValue(false)
	}

	log := logger.Get()

	switch action {
	case "func":
		expr := call.Argument(2).String()
		if expr == "undefined" || expr == "" {
			log.Warn("run_on_worker: expression required for 'func' action")
			return vf.vm.ToValue(false)
		}
		if trySendExecuteRequest("func", expr, "", "", "", "worker", scope) {
			return vf.vm.ToValue(true)
		}
		// Fallback: execute locally
		log.Warn("run_on_worker: not in distributed mode, executing locally")
		execCtx := make(map[string]interface{})
		_, err := vf.runtime.Execute(expr, execCtx)
		if err != nil {
			log.Warn("run_on_worker: local execution failed", zap.Error(err))
			return vf.vm.ToValue(false)
		}
		return vf.vm.ToValue(true)

	case "run":
		workflow := call.Argument(2).String()
		target := call.Argument(3).String()
		params := call.Argument(4).String()
		if workflow == "undefined" || workflow == "" || target == "undefined" || target == "" {
			log.Warn("run_on_worker: workflow and target required for 'run' action")
			return vf.vm.ToValue(false)
		}
		if params == "undefined" {
			params = ""
		}
		if trySendExecuteRequest("run", "", workflow, target, params, "worker", scope) {
			return vf.vm.ToValue(true)
		}
		// Fallback: run locally
		log.Warn("run_on_worker: not in distributed mode, running locally")
		vf.runOsmedeus("-m", workflow, target, params, "run_on_worker")
		return vf.vm.ToValue(true)

	case "bash":
		command := call.Argument(2).String()
		if command == "undefined" || command == "" {
			log.Warn("run_on_worker: command required for 'bash' action")
			return vf.vm.ToValue(false)
		}
		if trySendExecuteRequest("bash", command, "", "", "", "worker", scope) {
			return vf.vm.ToValue(true)
		}
		// Fallback: execute locally
		log.Warn("run_on_worker: not in distributed mode, executing bash locally")
		// @NOTE: This is intentional - run_on_worker('scope', 'bash', cmd) is a utility function
		// exposed to workflow definitions. Input comes from trusted workflow YAML files.
		cmd := exec.Command("sh", "-c", command)
		_, _ = cmd.CombinedOutput()
		return vf.vm.ToValue(true)

	default:
		log.Warn("run_on_worker: unknown action (expected 'func', 'run', or 'bash'). "+
			"Did you forget the scope argument? Correct usage: run_on_worker(scope, action, ...args)",
			zap.String("action", action))
		return vf.vm.ToValue(false)
	}
}

// copyFileBuffered copies a file using buffered I/O (memory-efficient for large files)
func copyFileBuffered(source, dest string, mode os.FileMode) error {
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
