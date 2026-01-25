package functions

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCdnUpload_EmptyLocalPath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_upload("", "remote/path")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestCdnUpload_EmptyRemotePath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_upload("/local/path", "")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestCdnUpload_UndefinedArguments(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_upload()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestCdnDownload_EmptyRemotePath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_download("", "/local/path")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestCdnDownload_EmptyLocalPath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_download("remote/path", "")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestCdnExists_EmptyPath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_exists("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestCdnDelete_EmptyPath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_delete("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

// Tests for new CDN functions

func TestCdnSyncUpload_EmptyLocalDir(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_sync_upload("", "remote/prefix/", "json")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Should return JSON string with success: false
	resultStr, ok := result.(string)
	require.True(t, ok)
	var resultMap map[string]interface{}
	err = json.Unmarshal([]byte(resultStr), &resultMap)
	require.NoError(t, err)
	assert.Equal(t, false, resultMap["success"])
}

func TestCdnSyncUpload_UndefinedArguments(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_sync_upload(undefined, undefined, "json")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Should return JSON string with success: false
	resultStr, ok := result.(string)
	require.True(t, ok)
	var resultMap map[string]interface{}
	err = json.Unmarshal([]byte(resultStr), &resultMap)
	require.NoError(t, err)
	assert.Equal(t, false, resultMap["success"])
}

func TestCdnSyncDownload_EmptyLocalDir(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_sync_download("remote/prefix/", "", "json")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Should return JSON string with success: false
	resultStr, ok := result.(string)
	require.True(t, ok)
	var resultMap map[string]interface{}
	err = json.Unmarshal([]byte(resultStr), &resultMap)
	require.NoError(t, err)
	assert.Equal(t, false, resultMap["success"])
}

func TestCdnSyncDownload_UndefinedArguments(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_sync_download(undefined, undefined, "json")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Should return JSON string with success: false
	resultStr, ok := result.(string)
	require.True(t, ok)
	var resultMap map[string]interface{}
	err = json.Unmarshal([]byte(resultStr), &resultMap)
	require.NoError(t, err)
	assert.Equal(t, false, resultMap["success"])
}

func TestCdnGetPresignedURL_EmptyPath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_get_presigned_url("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestCdnGetPresignedURL_WithExpiry(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_get_presigned_url("test/file.txt", 60)`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Returns empty string when storage not configured
	assert.Equal(t, "", result)
}

func TestCdnGetPresignedURL_NoExpiry(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_get_presigned_url("test/file.txt")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Returns empty string when storage not configured
	assert.Equal(t, "", result)
}

func TestCdnList_EmptyPrefix(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_list("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Should return empty JSON array when storage not configured
	resultStr, ok := result.(string)
	require.True(t, ok)
	var resultSlice []interface{}
	err = json.Unmarshal([]byte(resultStr), &resultSlice)
	require.NoError(t, err)
	assert.Equal(t, 0, len(resultSlice))
}

func TestCdnList_WithPrefix(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_list("scans/")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Should return empty JSON array when storage not configured
	resultStr, ok := result.(string)
	require.True(t, ok)
	var resultSlice []interface{}
	err = json.Unmarshal([]byte(resultStr), &resultSlice)
	require.NoError(t, err)
	assert.Equal(t, 0, len(resultSlice))
}

func TestCdnList_NoArguments(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_list()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Should return empty JSON array when storage not configured
	resultStr, ok := result.(string)
	require.True(t, ok)
	var resultSlice []interface{}
	err = json.Unmarshal([]byte(resultStr), &resultSlice)
	require.NoError(t, err)
	assert.Equal(t, 0, len(resultSlice))
}

func TestCdnStat_EmptyPath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_stat("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Should return "null" string for empty path
	resultStr, ok := result.(string)
	require.True(t, ok)
	assert.Equal(t, "null", resultStr)
}

func TestCdnStat_NonExistentFile(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_stat("nonexistent/file.txt")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Should return "null" string when storage not configured
	resultStr, ok := result.(string)
	require.True(t, ok)
	assert.Equal(t, "null", resultStr)
}

func TestCdnStat_UndefinedArgument(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_stat()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Should return "null" string for undefined argument
	resultStr, ok := result.(string)
	require.True(t, ok)
	assert.Equal(t, "null", resultStr)
}

// Test return types for sync operations
func TestCdnSyncUpload_ReturnStructure(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_sync_upload("/nonexistent", "prefix/", "json")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	resultStr, ok := result.(string)
	require.True(t, ok)
	var resultMap map[string]interface{}
	err = json.Unmarshal([]byte(resultStr), &resultMap)
	require.NoError(t, err)

	// Check all expected fields exist
	_, hasSuccess := resultMap["success"]
	_, hasUploaded := resultMap["uploaded"]
	_, hasSkipped := resultMap["skipped"]
	_, hasDeleted := resultMap["deleted"]
	_, hasErrorCount := resultMap["errorCount"]

	assert.True(t, hasSuccess, "should have 'success' field")
	assert.True(t, hasUploaded, "should have 'uploaded' field")
	assert.True(t, hasSkipped, "should have 'skipped' field")
	assert.True(t, hasDeleted, "should have 'deleted' field")
	assert.True(t, hasErrorCount, "should have 'errorCount' field")
}

func TestCdnSyncDownload_ReturnStructure(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`cdn_sync_download("prefix/", "/nonexistent", "json")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	resultStr, ok := result.(string)
	require.True(t, ok)
	var resultMap map[string]interface{}
	err = json.Unmarshal([]byte(resultStr), &resultMap)
	require.NoError(t, err)

	// Check all expected fields exist
	_, hasSuccess := resultMap["success"]
	_, hasDownloaded := resultMap["downloaded"]
	_, hasSkipped := resultMap["skipped"]
	_, hasDeleted := resultMap["deleted"]
	_, hasErrorCount := resultMap["errorCount"]

	assert.True(t, hasSuccess, "should have 'success' field")
	assert.True(t, hasDownloaded, "should have 'downloaded' field")
	assert.True(t, hasSkipped, "should have 'skipped' field")
	assert.True(t, hasDeleted, "should have 'deleted' field")
	assert.True(t, hasErrorCount, "should have 'errorCount' field")
}

// Note: Actual CDN upload/download/delete tests require a configured
// S3-compatible storage and are not included here. The functions will
// return false when storage is not configured, which is expected.
