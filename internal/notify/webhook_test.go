package notify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helpers

// requestRecorder records HTTP requests for verification
type requestRecorder struct {
	mu       sync.Mutex
	requests []*http.Request
	bodies   [][]byte
}

func (rr *requestRecorder) handler(statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rr.mu.Lock()
		defer rr.mu.Unlock()

		// Read body
		body, _ := io.ReadAll(r.Body)
		rr.bodies = append(rr.bodies, body)

		// Clone request (body already read)
		rr.requests = append(rr.requests, r)

		w.WriteHeader(statusCode)
	}
}

func (rr *requestRecorder) getRequests() ([]*http.Request, [][]byte) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	return rr.requests, rr.bodies
}

func setupMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func setupTestConfig(t *testing.T, cfg *config.Config) {
	config.Set(cfg)
	t.Cleanup(func() { config.Set(nil) })
}

// NewWebhookClient tests

func TestNewWebhookClient_Success(t *testing.T) {
	cfg := &config.WebhookConfig{
		URL:     "https://example.com/webhook",
		Enabled: true,
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://example.com/webhook", client.url)
	// Check default timeout (30 seconds)
	assert.Equal(t, 30*time.Second, client.timeout)
	// Check default retries (3)
	assert.Equal(t, 3, client.retries)
}

func TestNewWebhookClient_EmptyURL(t *testing.T) {
	cfg := &config.WebhookConfig{
		URL:     "",
		Enabled: true,
	}

	client, err := NewWebhookClient(cfg)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "webhook URL is required")
}

func TestNewWebhookClient_CustomTimeout(t *testing.T) {
	cfg := &config.WebhookConfig{
		URL:     "https://example.com/webhook",
		Timeout: 60,
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)
	assert.Equal(t, 60*time.Second, client.timeout)
}

func TestNewWebhookClient_CustomRetries(t *testing.T) {
	cfg := &config.WebhookConfig{
		URL:        "https://example.com/webhook",
		RetryCount: 5,
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)
	assert.Equal(t, 5, client.retries)
}

func TestNewWebhookClient_WithHeaders(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Bearer token123",
		"X-Custom":      "custom-value",
	}
	cfg := &config.WebhookConfig{
		URL:     "https://example.com/webhook",
		Headers: headers,
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)
	assert.Equal(t, headers, client.headers)
}

func TestNewWebhookClient_SkipTLSVerify(t *testing.T) {
	cfg := &config.WebhookConfig{
		URL:           "https://example.com/webhook",
		SkipTLSVerify: true,
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)
	assert.True(t, client.skipTLS)
	// The HTTP client should have an insecure transport
	assert.NotNil(t, client.httpClient)
}

// ShouldTrigger tests

func TestShouldTrigger_NoEventsFilter(t *testing.T) {
	cfg := &config.WebhookConfig{
		URL:    "https://example.com/webhook",
		Events: []string{}, // Empty = trigger all
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)

	assert.True(t, client.ShouldTrigger("scan_complete"))
	assert.True(t, client.ShouldTrigger("step_failed"))
	assert.True(t, client.ShouldTrigger("any_event"))
}

func TestShouldTrigger_MatchingEvent(t *testing.T) {
	cfg := &config.WebhookConfig{
		URL:    "https://example.com/webhook",
		Events: []string{"scan_complete", "scan_failed"},
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)

	assert.True(t, client.ShouldTrigger("scan_complete"))
	assert.True(t, client.ShouldTrigger("scan_failed"))
}

func TestShouldTrigger_NonMatchingEvent(t *testing.T) {
	cfg := &config.WebhookConfig{
		URL:    "https://example.com/webhook",
		Events: []string{"scan_complete", "scan_failed"},
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)

	assert.False(t, client.ShouldTrigger("step_started"))
	assert.False(t, client.ShouldTrigger("unknown_event"))
}

// Send tests

func TestSend_Success(t *testing.T) {
	recorder := &requestRecorder{}
	server := setupMockServer(t, recorder.handler(http.StatusOK))

	cfg := &config.WebhookConfig{
		URL:        server.URL,
		RetryCount: 1,
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)

	payload := WebhookPayload{
		Event:     "test_event",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"key": "value"},
	}

	err = client.Send(payload)
	assert.NoError(t, err)

	requests, bodies := recorder.getRequests()
	assert.Len(t, requests, 1)
	assert.Len(t, bodies, 1)
}

func TestSend_WithCustomHeaders(t *testing.T) {
	var receivedHeaders http.Header
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.WebhookConfig{
		URL: server.URL,
		Headers: map[string]string{
			"Authorization": "Bearer secret-token",
			"X-Custom":      "custom-value",
		},
		RetryCount: 1,
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)

	payload := WebhookPayload{
		Event:     "test_event",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{},
	}

	err = client.Send(payload)
	assert.NoError(t, err)

	assert.Equal(t, "Bearer secret-token", receivedHeaders.Get("Authorization"))
	assert.Equal(t, "custom-value", receivedHeaders.Get("X-Custom"))
	assert.Equal(t, "application/json", receivedHeaders.Get("Content-Type"))
}

func TestSend_ServerError_RetrySuccess(t *testing.T) {
	var callCount int32
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		if count == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.WebhookConfig{
		URL:        server.URL,
		RetryCount: 3,
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)

	payload := WebhookPayload{
		Event:     "test_event",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{},
	}

	err = client.Send(payload)
	assert.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&callCount), "expected 2 calls (1 failure + 1 success)")
}

func TestSend_AllRetriesFail(t *testing.T) {
	var callCount int32
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	})

	cfg := &config.WebhookConfig{
		URL:        server.URL,
		RetryCount: 2,
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)

	payload := WebhookPayload{
		Event:     "test_event",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{},
	}

	err = client.Send(payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhook failed after 2 attempts")
	assert.Equal(t, int32(2), atomic.LoadInt32(&callCount), "expected 2 attempts")
}

func TestSend_PayloadFormat(t *testing.T) {
	var receivedBody []byte
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.WebhookConfig{
		URL:        server.URL,
		RetryCount: 1,
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)

	timestamp := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	payload := WebhookPayload{
		Event:     "scan_complete",
		Timestamp: timestamp,
		Data: map[string]interface{}{
			"target":  "example.com",
			"results": 42,
		},
	}

	err = client.Send(payload)
	require.NoError(t, err)

	var received WebhookPayload
	err = json.Unmarshal(receivedBody, &received)
	require.NoError(t, err)

	assert.Equal(t, "scan_complete", received.Event)
	assert.Equal(t, "example.com", received.Data["target"])
	assert.Equal(t, float64(42), received.Data["results"]) // JSON numbers are float64
}

// SendMessage tests

func TestSendMessage_Success(t *testing.T) {
	var receivedBody []byte
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.WebhookConfig{
		URL:        server.URL,
		RetryCount: 1,
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)

	err = client.SendMessage("Hello, webhook!")
	require.NoError(t, err)

	var received WebhookPayload
	err = json.Unmarshal(receivedBody, &received)
	require.NoError(t, err)

	assert.Equal(t, "message", received.Event)
	assert.Equal(t, "Hello, webhook!", received.Data["message"])
}

// SendEvent tests

func TestSendEvent_TriggeredEvent(t *testing.T) {
	var requestMade bool
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestMade = true
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.WebhookConfig{
		URL:        server.URL,
		Events:     []string{"scan_complete"},
		RetryCount: 1,
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)

	err = client.SendEvent("scan_complete", map[string]interface{}{"target": "example.com"})
	assert.NoError(t, err)
	assert.True(t, requestMade, "request should be made for matching event")
}

func TestSendEvent_FilteredEvent(t *testing.T) {
	var requestMade bool
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestMade = true
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.WebhookConfig{
		URL:        server.URL,
		Events:     []string{"scan_complete"},
		RetryCount: 1,
	}

	client, err := NewWebhookClient(cfg)
	require.NoError(t, err)

	err = client.SendEvent("step_started", map[string]interface{}{"step": "recon"})
	assert.NoError(t, err)
	assert.False(t, requestMade, "request should not be made for non-matching event")
}

// Global function tests

func TestTriggerWebhooks_NotificationDisabled(t *testing.T) {
	var requestMade bool
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestMade = true
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.Config{
		Notification: config.NotificationConfig{
			Enabled: false, // Disabled
			Webhooks: []config.WebhookConfig{
				{URL: server.URL, Enabled: true},
			},
		},
	}

	TriggerWebhooks(cfg, "test_event", map[string]interface{}{})
	assert.False(t, requestMade, "no request when notifications disabled")
}

func TestTriggerWebhooks_NilConfig(t *testing.T) {
	// Should not panic with nil config
	assert.NotPanics(t, func() {
		TriggerWebhooks(nil, "test_event", map[string]interface{}{})
	})
}

func TestTriggerWebhooks_MultipleWebhooks(t *testing.T) {
	var callCount int32

	server1 := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	})

	server2 := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.Config{
		Notification: config.NotificationConfig{
			Enabled: true,
			Webhooks: []config.WebhookConfig{
				{URL: server1.URL, Enabled: true, RetryCount: 1},
				{URL: server2.URL, Enabled: true, RetryCount: 1},
				{URL: "https://disabled.example.com", Enabled: false}, // Disabled
			},
		},
	}

	TriggerWebhooks(cfg, "test_event", map[string]interface{}{})
	assert.Equal(t, int32(2), atomic.LoadInt32(&callCount), "only enabled webhooks should receive event")
}

func TestSendWebhookMessage_NoGlobalConfig(t *testing.T) {
	config.Set(nil)

	err := SendWebhookMessage("test message")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "global config not loaded")
}

func TestSendWebhookMessage_NotConfigured(t *testing.T) {
	cfg := &config.Config{
		Notification: config.NotificationConfig{
			Enabled:  false,
			Webhooks: []config.WebhookConfig{},
		},
	}
	setupTestConfig(t, cfg)

	err := SendWebhookMessage("test message")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhooks not configured")
}

func TestSendWebhookMessage_Success(t *testing.T) {
	var receivedBody []byte
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.Config{
		Notification: config.NotificationConfig{
			Enabled: true,
			Webhooks: []config.WebhookConfig{
				{URL: server.URL, Enabled: true, RetryCount: 1},
			},
		},
	}
	setupTestConfig(t, cfg)

	err := SendWebhookMessage("Hello from global config!")
	require.NoError(t, err)

	var received WebhookPayload
	err = json.Unmarshal(receivedBody, &received)
	require.NoError(t, err)

	assert.Equal(t, "message", received.Event)
	assert.Equal(t, "Hello from global config!", received.Data["message"])
}

func TestSendWebhookEvent_NoGlobalConfig(t *testing.T) {
	config.Set(nil)

	err := SendWebhookEvent("test_event", map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "global config not loaded")
}

func TestSendWebhookEvent_NotConfigured(t *testing.T) {
	cfg := &config.Config{
		Notification: config.NotificationConfig{
			Enabled:  false,
			Webhooks: []config.WebhookConfig{},
		},
	}
	setupTestConfig(t, cfg)

	err := SendWebhookEvent("test_event", map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhooks not configured")
}

func TestSendWebhookEvent_Success(t *testing.T) {
	var receivedBody []byte
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.Config{
		Notification: config.NotificationConfig{
			Enabled: true,
			Webhooks: []config.WebhookConfig{
				{URL: server.URL, Enabled: true, RetryCount: 1},
			},
		},
	}
	setupTestConfig(t, cfg)

	err := SendWebhookEvent("scan_complete", map[string]interface{}{
		"target":  "example.com",
		"results": 100,
	})
	require.NoError(t, err)

	var received WebhookPayload
	err = json.Unmarshal(receivedBody, &received)
	require.NoError(t, err)

	assert.Equal(t, "scan_complete", received.Event)
	assert.Equal(t, "example.com", received.Data["target"])
	assert.Equal(t, float64(100), received.Data["results"])
}
