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

// execCmd executes a bash command and returns the stdout output
// Usage: exec_cmd(command) -> string
func (vf *vmFunc) execCmd(call goja.FunctionCall) goja.Value {
	command := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen("execCmd"), zap.String("command", command))

	if command == "undefined" || command == "" {
		logger.Get().Warn("execCmd: empty command provided")
		return vf.vm.ToValue("")
	}

	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		logger.Get().Warn("execCmd: command failed", zap.String("command", command), zap.Error(err))
		return vf.vm.ToValue("")
	}

	logger.Get().Debug(terminal.HiGreen("execCmd")+" result", zap.String("command", command), zap.Int("outputLength", len(output)))
	return vf.vm.ToValue(strings.TrimSpace(string(output)))
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
