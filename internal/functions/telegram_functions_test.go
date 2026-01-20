package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifyTelegram_EmptyMessage(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`notify_telegram("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestNotifyTelegram_UndefinedMessage(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`notify_telegram()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSendTelegramFile_EmptyPath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`send_telegram_file("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSendTelegramFile_UndefinedPath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`send_telegram_file()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSendTelegramFile_NonExistentFile(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`send_telegram_file("/nonexistent/path/to/file.txt")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Returns false because either Telegram is not configured
	// or the file doesn't exist (checked first in notify package)
	assert.Equal(t, false, result)
}

func TestNotifyTelegram_Whitespace(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`notify_telegram("   ")`,
		map[string]interface{}{},
	)

	// Whitespace-only message is not empty string, so it tries to send
	// but will fail because Telegram is not configured
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSendTelegramFile_WithCaption(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`send_telegram_file("/nonexistent/file.txt", "My caption")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Returns false because Telegram is not configured
	assert.Equal(t, false, result)
}

func TestSendTelegramFile_WithEmptyCaption(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`send_telegram_file("/nonexistent/file.txt", "")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Returns false because Telegram is not configured
	assert.Equal(t, false, result)
}

// Note: Actual Telegram message/file sending tests require a configured
// Telegram bot and are not included here. The functions will return false
// when Telegram is not configured, which is the expected behavior.
