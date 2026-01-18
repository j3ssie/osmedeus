package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCdnUpload_EmptyLocalPath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdnUpload("", "remote/path")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestCdnUpload_EmptyRemotePath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdnUpload("/local/path", "")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestCdnUpload_UndefinedArguments(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdnUpload()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestCdnDownload_EmptyRemotePath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdnDownload("", "/local/path")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestCdnDownload_EmptyLocalPath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdnDownload("remote/path", "")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestCdnExists_EmptyPath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdnExists("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestCdnDelete_EmptyPath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdnDelete("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

// Note: Actual CDN upload/download/delete tests require a configured
// S3-compatible storage and are not included here. The functions will
// return false when storage is not configured, which is expected.
