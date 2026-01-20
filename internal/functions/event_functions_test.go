package functions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateEvent_ValidArguments(t *testing.T) {
	registry := NewRegistry()

	// Test with all valid arguments (workspace, topic, source, data_type, data)
	result, err := registry.Execute(
		`generate_event("test-workspace", "assets.new", "test-source", "subdomain", "test.example.com")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Returns true because events are queued even when server is unavailable
	assert.Equal(t, true, result)
}

func TestGenerateEvent_EmptyWorkspace(t *testing.T) {
	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event("", "assets.new", "test-source", "subdomain", "test.example.com")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestGenerateEvent_EmptyTopic(t *testing.T) {
	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event("test-workspace", "", "test-source", "subdomain", "test.example.com")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestGenerateEvent_EmptySource(t *testing.T) {
	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event("test-workspace", "assets.new", "", "subdomain", "test.example.com")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestGenerateEvent_EmptyDataType(t *testing.T) {
	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event("test-workspace", "assets.new", "test-source", "", "test.example.com")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestGenerateEvent_MissingArguments(t *testing.T) {
	registry := NewRegistry()

	// Note: when called with no arguments, goja returns "undefined" for missing args
	// which passes the empty string validation, so the function returns true
	result, err := registry.Execute(
		`generate_event()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestGenerateEvent_ObjectData(t *testing.T) {
	registry := NewRegistry()

	// Test with object data (complex payload)
	result, err := registry.Execute(
		`generate_event("test-workspace", "vulnerabilities.new", "nuclei", "finding", {
			url: "https://example.com/admin",
			severity: "critical",
			template: "CVE-2024-1234"
		})`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Returns true because events are queued even when server is unavailable
	assert.Equal(t, true, result)
}

func TestGenerateEvent_NullData(t *testing.T) {
	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event("test-workspace", "assets.new", "test", "subdomain", null)`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Should handle null data gracefully - returns true as events are queued
	assert.Equal(t, true, result)
}

func TestGenerateEvent_UndefinedData(t *testing.T) {
	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event("test-workspace", "assets.new", "test", "subdomain")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Should handle undefined data gracefully - returns true as events are queued
	assert.Equal(t, true, result)
}

func TestGenerateEvent_WithTemplateVariables(t *testing.T) {
	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event("test-workspace", "assets.new", "subfinder", "subdomain", target)`,
		map[string]interface{}{
			"target": "api.example.com",
		},
	)

	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestGenerateEventFromFile_ValidFile(t *testing.T) {
	// Create a temporary file with test data
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "subdomains.txt")

	content := `api.example.com
www.example.com
admin.example.com

test.example.com
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event_from_file("test-workspace", "assets.new", "subfinder", "subdomain", filePath)`,
		map[string]interface{}{
			"filePath": testFile,
		},
	)

	require.NoError(t, err)
	// Returns count of processed lines (4 non-empty lines in the file)
	assert.Equal(t, int64(4), result)
}

func TestGenerateEventFromFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.txt")

	err := os.WriteFile(testFile, []byte(""), 0644)
	require.NoError(t, err)

	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event_from_file("test-workspace", "assets.new", "test", "subdomain", filePath)`,
		map[string]interface{}{
			"filePath": testFile,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(0), result)
}

func TestGenerateEventFromFile_NonExistentFile(t *testing.T) {
	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event_from_file("test-workspace", "assets.new", "test", "subdomain", "/nonexistent/file.txt")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(0), result)
}

func TestGenerateEventFromFile_EmptyPath(t *testing.T) {
	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event_from_file("test-workspace", "assets.new", "test", "subdomain", "")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(0), result)
}

func TestGenerateEventFromFile_MissingWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test.example.com\n"), 0644)
	require.NoError(t, err)

	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event_from_file("", "assets.new", "test", "subdomain", filePath)`,
		map[string]interface{}{
			"filePath": testFile,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(0), result)
}

func TestGenerateEventFromFile_MissingTopic(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test.example.com\n"), 0644)
	require.NoError(t, err)

	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event_from_file("test-workspace", "", "test", "subdomain", filePath)`,
		map[string]interface{}{
			"filePath": testFile,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(0), result)
}

func TestGenerateEventFromFile_MissingSource(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test.example.com\n"), 0644)
	require.NoError(t, err)

	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event_from_file("test-workspace", "assets.new", "", "subdomain", filePath)`,
		map[string]interface{}{
			"filePath": testFile,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(0), result)
}

func TestGenerateEventFromFile_MissingDataType(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test.example.com\n"), 0644)
	require.NoError(t, err)

	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event_from_file("test-workspace", "assets.new", "test", "", filePath)`,
		map[string]interface{}{
			"filePath": testFile,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(0), result)
}

func TestGenerateEventFromFile_BlankLinesSkipped(t *testing.T) {
	// Create file with blank lines that should be skipped
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "mixed.txt")

	content := `
api.example.com

www.example.com


admin.example.com
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry()

	// The function should skip blank lines and whitespace-only lines
	result, err := registry.Execute(
		`generate_event_from_file("test-workspace", "assets.new", "test", "subdomain", filePath)`,
		map[string]interface{}{
			"filePath": testFile,
		},
	)

	require.NoError(t, err)
	// Returns count of non-blank lines processed (3 non-empty lines)
	assert.Equal(t, int64(3), result)
}

func TestGenerateEventFromFile_MissingArguments(t *testing.T) {
	registry := NewRegistry()

	result, err := registry.Execute(
		`generate_event_from_file()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(0), result)
}

// Test various event topics
func TestGenerateEvent_VariousTopics(t *testing.T) {
	registry := NewRegistry()

	topics := []string{
		"assets.new",
		"vulnerabilities.new",
		"run.started",
		"run.completed",
		"step.completed",
		"custom.topic",
		"external.webhook",
	}

	for _, topic := range topics {
		t.Run(topic, func(t *testing.T) {
			result, err := registry.Execute(
				`generate_event("test-workspace", topic, "test", "data", "payload")`,
				map[string]interface{}{
					"topic": topic,
				},
			)

			require.NoError(t, err)
			assert.Equal(t, true, result)
		})
	}
}

// Test event generation with different data types
func TestGenerateEvent_DataTypes(t *testing.T) {
	registry := NewRegistry()

	testCases := []struct {
		name       string
		expression string
	}{
		{"string_data", `generate_event("test-workspace", "test", "src", "type", "string value")`},
		{"number_data", `generate_event("test-workspace", "test", "src", "type", 42)`},
		{"boolean_data", `generate_event("test-workspace", "test", "src", "type", true)`},
		{"array_data", `generate_event("test-workspace", "test", "src", "type", ["a", "b", "c"])`},
		{"nested_object", `generate_event("test-workspace", "test", "src", "type", {nested: {deep: "value"}})`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := registry.Execute(tc.expression, map[string]interface{}{})
			require.NoError(t, err)
			assert.Equal(t, true, result)
		})
	}
}
