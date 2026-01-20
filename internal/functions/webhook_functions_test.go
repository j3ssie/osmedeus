package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifyWebhook_EmptyMessage(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`notify_webhook("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestNotifyWebhook_UndefinedMessage(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`notify_webhook()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestNotifyWebhook_WhitespaceMessage(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`notify_webhook("   ")`,
		map[string]interface{}{},
	)

	// Whitespace-only message is not empty string, so it tries to send
	// but will fail because webhook is not configured
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSendWebhookEvent_EmptyEventType(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`send_webhook_event("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSendWebhookEvent_UndefinedEventType(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`send_webhook_event()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSendWebhookEvent_WithData(t *testing.T) {
	registry := NewRegistry()
	// Even with valid event type and data, it will fail because webhook is not configured
	result, err := registry.Execute(
		`send_webhook_event("test_event", {key: "value", count: 42})`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Returns false because webhook is not configured in global config
	assert.Equal(t, false, result)
}

func TestSendWebhookEvent_WithEmptyData(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`send_webhook_event("test_event", {})`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Returns false because webhook is not configured
	assert.Equal(t, false, result)
}

func TestSendWebhookEvent_WithoutData(t *testing.T) {
	registry := NewRegistry()
	// Event type provided but no data object
	result, err := registry.Execute(
		`send_webhook_event("test_event")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Returns false because webhook is not configured
	assert.Equal(t, false, result)
}

// Note: Actual webhook message/event sending tests require configured webhooks
// and are not included here. The functions will return false when webhooks
// are not configured, which is the expected behavior.
