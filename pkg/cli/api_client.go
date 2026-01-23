package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// ScheduleClient handles schedule registration with the osmedeus server
type ScheduleClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewScheduleClient creates a client from config
func NewScheduleClient(cfg *config.Config) *ScheduleClient {
	return &ScheduleClient{
		baseURL: cfg.Server.GetServerURL(),
		apiKey:  cfg.Server.AuthAPIKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetBaseURL overrides the base URL (for --server-url flag)
func (c *ScheduleClient) SetBaseURL(url string) {
	c.baseURL = url
}

// IsServerAvailable checks if the server is reachable via GET /osm/server-info
func (c *ScheduleClient) IsServerAvailable() bool {
	if c.baseURL == "" {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/osm/server-info", nil)
	if err != nil {
		return false
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}

// scheduleRequest represents the JSON body for creating a schedule
type scheduleRequest struct {
	Name         string            `json:"name"`
	WorkflowName string            `json:"workflow_name"`
	WorkflowKind string            `json:"workflow_kind"`
	Target       string            `json:"target"`
	Schedule     string            `json:"schedule"`
	Params       map[string]string `json:"params,omitempty"`
	Enabled      bool              `json:"enabled"`
}

// RegisterCronTrigger POSTs to /osm/api/schedules to register a cron trigger
// Returns nil on 201 Created or 409 Conflict (schedule already exists)
func (c *ScheduleClient) RegisterCronTrigger(ctx context.Context, workflow *core.Workflow, trigger *core.Trigger, target string, params map[string]string) error {
	if c.baseURL == "" {
		return fmt.Errorf("server URL not configured")
	}

	// Build schedule name: workflow-name-trigger-name
	scheduleName := fmt.Sprintf("%s-%s", workflow.Name, trigger.Name)

	reqBody := scheduleRequest{
		Name:         scheduleName,
		WorkflowName: workflow.Name,
		WorkflowKind: string(workflow.Kind),
		Target:       target,
		Schedule:     trigger.Schedule,
		Params:       params,
		Enabled:      trigger.Enabled,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/osm/api/schedules", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("x-osm-api-key", c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// 201 Created = success
	// 409 Conflict = schedule already exists (treat as success)
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		return nil
	}

	// Read error body for debugging
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
}
