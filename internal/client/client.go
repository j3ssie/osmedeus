// Package client provides HTTP client utilities for interacting with a remote osmedeus server.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	// EnvRemoteURL is the environment variable for the remote server URL
	EnvRemoteURL = "OSM_REMOTE_URL"
	// EnvAuthKey is the environment variable for the API authentication key
	EnvAuthKey = "OSM_REMOTE_AUTH_KEY"
	// APIBasePath is the base path for API endpoints
	APIBasePath = "/osm/api"
	// DefaultTimeout is the default HTTP request timeout
	DefaultTimeout = 30 * time.Second
)

// Client is an HTTP client for interacting with a remote osmedeus server
type Client struct {
	baseURL    string
	authKey    string
	httpClient *http.Client
}

// NewClient creates a new client with the given base URL and auth key.
// If baseURL or authKey are empty, they will be read from environment variables.
func NewClient(baseURL, authKey string) (*Client, error) {
	// Resolve base URL
	if baseURL == "" {
		baseURL = os.Getenv(EnvRemoteURL)
	}
	if baseURL == "" {
		return nil, fmt.Errorf("remote URL is required (set %s or use --remote-url)", EnvRemoteURL)
	}

	// Validate URL format
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid remote URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("invalid URL scheme: %s (must be http or https)", parsedURL.Scheme)
	}

	// Remove trailing slash from base URL
	baseURL = strings.TrimRight(baseURL, "/")

	// Resolve auth key (optional - server may allow unauthenticated access)
	if authKey == "" {
		authKey = os.Getenv(EnvAuthKey)
	}

	return &Client{
		baseURL: baseURL,
		authKey: authKey,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}, nil
}

// buildURL constructs the full URL for an API endpoint
func (c *Client) buildURL(path string, query url.Values) string {
	fullPath := c.baseURL + APIBasePath + path
	if len(query) > 0 {
		fullPath += "?" + query.Encode()
	}
	return fullPath
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(ctx context.Context, method, fullURL string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.authKey != "" {
		req.Header.Set("x-osm-api-key", c.authKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// handleResponse reads and parses the response, returning an error for non-2xx status codes
func handleResponse(resp *http.Response, result interface{}) error {
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse
		if json.Unmarshal(bodyBytes, &errResp) == nil && errResp.Message != "" {
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    errResp.Message,
			}
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(bodyBytes)),
		}
	}

	// Parse successful response
	if result != nil && len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// Get performs a GET request to the given path with optional query parameters
func (c *Client) Get(ctx context.Context, path string, query url.Values, result interface{}) error {
	fullURL := c.buildURL(path, query)
	resp, err := c.doRequest(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}
	return handleResponse(resp, result)
}

// Post performs a POST request to the given path with a JSON body
func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	fullURL := c.buildURL(path, nil)
	resp, err := c.doRequest(ctx, http.MethodPost, fullURL, bodyReader)
	if err != nil {
		return err
	}
	return handleResponse(resp, result)
}

// Delete performs a DELETE request to the given path
func (c *Client) Delete(ctx context.Context, path string, result interface{}) error {
	fullURL := c.buildURL(path, nil)
	resp, err := c.doRequest(ctx, http.MethodDelete, fullURL, nil)
	if err != nil {
		return err
	}
	return handleResponse(resp, result)
}

// GetRaw performs a GET request and returns the raw response body
func (c *Client) GetRaw(ctx context.Context, path string, query url.Values) ([]byte, error) {
	fullURL := c.buildURL(path, query)
	resp, err := c.doRequest(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse
		if json.Unmarshal(bodyBytes, &errResp) == nil && errResp.Message != "" {
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Message:    errResp.Message,
			}
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(bodyBytes)),
		}
	}

	return bodyBytes, nil
}
