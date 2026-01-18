package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifyTelegram_EmptyMessage(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`notifyTelegram("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestNotifyTelegram_UndefinedMessage(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`notifyTelegram()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSendTelegramFile_EmptyPath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`sendTelegramFile("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSendTelegramFile_UndefinedPath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`sendTelegramFile()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

// Note: Actual Telegram message/file sending tests require a configured
// Telegram bot and are not included here. The functions will return false
// when Telegram is not configured, which is the expected behavior.
