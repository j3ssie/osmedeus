package notify

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// WebhookPayload represents the data sent to webhook endpoints
type WebhookPayload struct {
	Event     string                 `json:"event"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// WebhookClient handles sending notifications to webhook endpoints
type WebhookClient struct {
	url        string
	headers    map[string]string
	timeout    time.Duration
	retries    int
	skipTLS    bool
	events     []string
	httpClient *http.Client
}

// NewWebhookClient creates a new webhook client from config
func NewWebhookClient(cfg *config.WebhookConfig) (*WebhookClient, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 // default 30 seconds
	}

	retries := cfg.RetryCount
	if retries <= 0 {
		retries = 3 // default 3 retries
	}

	// Create HTTP client with optional TLS skip
	transport := &http.Transport{}
	if cfg.SkipTLSVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}

	return &WebhookClient{
		url:        cfg.URL,
		headers:    cfg.Headers,
		timeout:    time.Duration(timeout) * time.Second,
		retries:    retries,
		skipTLS:    cfg.SkipTLSVerify,
		events:     cfg.Events,
		httpClient: client,
	}, nil
}

// ShouldTrigger checks if this webhook should trigger for the given event
func (c *WebhookClient) ShouldTrigger(eventType string) bool {
	// If no events specified, trigger for all events
	if len(c.events) == 0 {
		return true
	}
	for _, e := range c.events {
		if e == eventType {
			return true
		}
	}
	return false
}

// Send sends a payload to the webhook endpoint with retries
func (c *WebhookClient) Send(payload WebhookPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < c.retries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s...
			time.Sleep(time.Duration(1<<attempt) * time.Second)
		}

		req, err := http.NewRequest(http.MethodPost, c.url, bytes.NewReader(data))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		// Set default content type
		req.Header.Set("Content-Type", "application/json")

		// Set custom headers
		for k, v := range c.headers {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}
		defer func() { _ = resp.Body.Close() }()

		// Success if 2xx status code
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return fmt.Errorf("webhook failed after %d attempts: %w", c.retries, lastErr)
}

// SendMessage sends a simple text message as a webhook payload
func (c *WebhookClient) SendMessage(message string) error {
	payload := WebhookPayload{
		Event:     "message",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"message": message,
		},
	}
	return c.Send(payload)
}

// SendEvent sends an event with data to the webhook
func (c *WebhookClient) SendEvent(eventType string, data map[string]interface{}) error {
	if !c.ShouldTrigger(eventType) {
		return nil // Skip if event not in filter list
	}

	payload := WebhookPayload{
		Event:     eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
	return c.Send(payload)
}

// TriggerWebhooks sends an event to all configured webhooks
func TriggerWebhooks(cfg *config.Config, eventType string, data map[string]interface{}) {
	if cfg == nil || !cfg.Notification.Enabled {
		return
	}

	log := logger.Get()

	for i, whCfg := range cfg.Notification.Webhooks {
		if !whCfg.Enabled {
			continue
		}

		client, err := NewWebhookClient(&whCfg)
		if err != nil {
			log.Warn("Failed to create webhook client",
				zap.Int("webhook_index", i),
				zap.Error(err))
			continue
		}

		if err := client.SendEvent(eventType, data); err != nil {
			log.Warn("Failed to send webhook notification",
				zap.Int("webhook_index", i),
				zap.String("event", eventType),
				zap.Error(err))
		}
	}
}

// SendWebhookMessage sends a message to all enabled webhooks using global config
func SendWebhookMessage(message string) error {
	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("global config not loaded")
	}
	if !cfg.Notification.Enabled || len(cfg.Notification.Webhooks) == 0 {
		return fmt.Errorf("webhooks not configured")
	}

	var lastErr error
	successCount := 0

	for i, whCfg := range cfg.Notification.Webhooks {
		if !whCfg.Enabled {
			continue
		}

		client, err := NewWebhookClient(&whCfg)
		if err != nil {
			lastErr = fmt.Errorf("webhook %d: %w", i, err)
			continue
		}

		if err := client.SendMessage(message); err != nil {
			lastErr = fmt.Errorf("webhook %d: %w", i, err)
			continue
		}
		successCount++
	}

	if successCount == 0 && lastErr != nil {
		return lastErr
	}
	return nil
}

// SendWebhookEvent sends an event to all enabled webhooks using global config
func SendWebhookEvent(eventType string, data map[string]interface{}) error {
	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("global config not loaded")
	}
	if !cfg.Notification.Enabled || len(cfg.Notification.Webhooks) == 0 {
		return fmt.Errorf("webhooks not configured")
	}

	var lastErr error
	successCount := 0

	for i, whCfg := range cfg.Notification.Webhooks {
		if !whCfg.Enabled {
			continue
		}

		client, err := NewWebhookClient(&whCfg)
		if err != nil {
			lastErr = fmt.Errorf("webhook %d: %w", i, err)
			continue
		}

		if err := client.SendEvent(eventType, data); err != nil {
			lastErr = fmt.Errorf("webhook %d: %w", i, err)
			continue
		}
		successCount++
	}

	if successCount == 0 && lastErr != nil {
		return lastErr
	}
	return nil
}
