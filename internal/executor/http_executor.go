package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/retry"
	"github.com/j3ssie/osmedeus/v5/internal/template"
)

// HTTPExecutor executes HTTP steps
type HTTPExecutor struct {
	templateEngine template.TemplateEngine
	client         *http.Client
}

// NewHTTPExecutor creates a new HTTP executor with pooled connections
func NewHTTPExecutor(engine template.TemplateEngine) *HTTPExecutor {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	return &HTTPExecutor{
		templateEngine: engine,
		client: &http.Client{
			Transport: transport,
			// No global timeout - we use per-request context timeout
		},
	}
}

// Name returns the executor name for logging/debugging
func (e *HTTPExecutor) Name() string {
	return "http"
}

// StepTypes returns the step types this executor handles
func (e *HTTPExecutor) StepTypes() []core.StepType {
	return []core.StepType{core.StepTypeHTTP}
}

// sanitizeStepName converts step name to a valid variable name
// Converts hyphens to underscores
func sanitizeStepName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

// Execute executes an HTTP step
func (e *HTTPExecutor) Execute(ctx context.Context, step *core.Step, execCtx *core.ExecutionContext) (*core.StepResult, error) {
	result := &core.StepResult{
		StepName:  step.Name,
		Status:    core.StepStatusRunning,
		StartTime: time.Now(),
		Exports:   make(map[string]interface{}),
	}

	// Validate required fields
	if step.URL == "" {
		err := fmt.Errorf("HTTP step '%s' requires 'url' field", step.Name)
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	stepTimeout, err := step.Timeout.Duration()
	if err != nil {
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Default method to GET
	method := step.Method
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)

	// Build HTTP response structure
	httpResp := map[string]interface{}{
		"status_code":      0,
		"error":            nil,
		"message":          "",
		"response_headers": map[string]string{},
		"response_body":    "",
		"content_length":   0,
		"response_time_ms": 0,
	}

	// Create request
	var reqBody io.Reader
	if step.RequestBody != "" {
		reqBody = strings.NewReader(step.RequestBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, step.URL, reqBody)
	if err != nil {
		httpResp["error"] = err.Error()
		httpResp["message"] = "failed to create request"
		e.exportHTTPResponse(result, step.Name, httpResp)
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Add headers
	for key, value := range step.Headers {
		req.Header.Set(key, value)
	}

	// Set timeout via context (allows connection reuse via shared client)
	timeout := 30 * time.Second
	if stepTimeout > 0 {
		timeout = stepTimeout
	}

	// Create request context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Update request with timeout context
	req = req.WithContext(reqCtx)

	// Execute request with retry for transient errors
	startTime := time.Now()
	var resp *http.Response
	err = retry.Do(reqCtx, retry.Config{
		MaxAttempts:  3,
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     2 * time.Second,
		Multiplier:   2.0,
	}, func() error {
		var reqErr error
		resp, reqErr = e.client.Do(req)
		if reqErr != nil {
			// Network errors are retryable
			return retry.Retryable(reqErr)
		}
		// Retry on 5xx server errors
		if resp.StatusCode >= 500 {
			_ = resp.Body.Close()
			return retry.Retryable(fmt.Errorf("server error: %d", resp.StatusCode))
		}
		return nil
	})
	responseTimeMs := time.Since(startTime).Milliseconds()
	httpResp["response_time_ms"] = responseTimeMs

	if err != nil {
		httpResp["error"] = err.Error()
		httpResp["message"] = "request failed"
		e.exportHTTPResponse(result, step.Name, httpResp)
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		httpResp["status_code"] = resp.StatusCode
		httpResp["error"] = err.Error()
		httpResp["message"] = "failed to read response body"
		e.exportHTTPResponse(result, step.Name, httpResp)
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Build response headers map
	respHeaders := make(map[string]string)
	for key, values := range resp.Header {
		respHeaders[strings.ToLower(key)] = strings.Join(values, ", ")
	}

	// Populate successful response
	httpResp["status_code"] = resp.StatusCode
	httpResp["error"] = nil
	httpResp["message"] = "success"
	httpResp["response_headers"] = respHeaders
	httpResp["response_body"] = string(body)
	httpResp["content_length"] = len(body)

	// Export the response
	e.exportHTTPResponse(result, step.Name, httpResp)

	// Set output to response body for logging/display
	result.Output = string(body)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Status = core.StepStatusSuccess

	return result, nil
}

// exportHTTPResponse exports the HTTP response to the result's exports map
// Export key is: <sanitized_step_name>_http_resp
func (e *HTTPExecutor) exportHTTPResponse(result *core.StepResult, stepName string, httpResp map[string]interface{}) {
	exportKey := sanitizeStepName(stepName) + "_http_resp"
	result.Exports[exportKey] = httpResp
}

// CanHandle returns true if this executor can handle the given step type
func (e *HTTPExecutor) CanHandle(stepType core.StepType) bool {
	return stepType == core.StepTypeHTTP
}
