package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForeachExecutor_VariablePreProcess(t *testing.T) {
	// Create temp directory for test files
	tmpDir := t.TempDir()

	// Create test input file with URLs
	inputFile := filepath.Join(tmpDir, "urls.txt")
	err := os.WriteFile(inputFile, []byte(`https://example.com/path/file.php?id=1
https://example.com/api/v1/users
https://example.com/deep/nested/path/resource
`), 0644)
	require.NoError(t, err)

	// Create output file path
	outputFile := filepath.Join(tmpDir, "output.txt")

	// Create dispatcher
	dispatcher := NewStepDispatcher()

	// Create execution context
	execCtx := core.NewExecutionContext("test-workflow", core.KindModule, "test-run-uuid", "test-target")
	execCtx.SetVariable("Output", tmpDir)

	// Test case 1: Basic pre-processing with get_parent_url
	t.Run("get_parent_url pre-processing", func(t *testing.T) {
		step := &core.Step{
			Name:               "test-preprocess",
			Type:               core.StepTypeForeach,
			Input:              inputFile,
			Variable:           "url",
			VariablePreProcess: "get_parent_url([[url]])",
			Threads:            "1",
			Step: &core.Step{
				Type:    core.StepTypeBash,
				Command: "echo '[[url]]' >> " + outputFile,
			},
		}

		// Clean output file
		_ = os.Remove(outputFile)

		ctx := context.Background()
		result, err := dispatcher.Dispatch(ctx, step, execCtx)
		require.NoError(t, err)
		assert.Equal(t, core.StepStatusSuccess, result.Status)

		// Verify output contains parent URLs
		output, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		outputStr := string(output)

		assert.Contains(t, outputStr, "https://example.com/path/")
		assert.Contains(t, outputStr, "https://example.com/api/v1/")
		assert.Contains(t, outputStr, "https://example.com/deep/nested/path/")
	})

	// Test case 2: No pre-processing (existing behavior unchanged)
	t.Run("no pre-processing", func(t *testing.T) {
		step := &core.Step{
			Name:     "test-no-preprocess",
			Type:     core.StepTypeForeach,
			Input:    inputFile,
			Variable: "url",
			Threads:  "1",
			Step: &core.Step{
				Type:    core.StepTypeBash,
				Command: "echo '[[url]]' >> " + outputFile,
			},
		}

		// Clean output file
		_ = os.Remove(outputFile)

		ctx := context.Background()
		result, err := dispatcher.Dispatch(ctx, step, execCtx)
		require.NoError(t, err)
		assert.Equal(t, core.StepStatusSuccess, result.Status)

		// Verify output contains original URLs
		output, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		outputStr := string(output)

		assert.Contains(t, outputStr, "https://example.com/path/file.php?id=1")
		assert.Contains(t, outputStr, "https://example.com/api/v1/users")
	})

	// Test case 3: Pre-processing with trim
	t.Run("trim pre-processing", func(t *testing.T) {
		// Create input with whitespace
		trimInputFile := filepath.Join(tmpDir, "trim-input.txt")
		err := os.WriteFile(trimInputFile, []byte(`  hello
  world
`), 0644)
		require.NoError(t, err)

		step := &core.Step{
			Name:               "test-trim-preprocess",
			Type:               core.StepTypeForeach,
			Input:              trimInputFile,
			Variable:           "line",
			VariablePreProcess: "trim([[line]])",
			Threads:            "1",
			Step: &core.Step{
				Type:    core.StepTypeBash,
				Command: "echo '[[line]]' >> " + outputFile,
			},
		}

		// Clean output file
		_ = os.Remove(outputFile)

		ctx := context.Background()
		result, err := dispatcher.Dispatch(ctx, step, execCtx)
		require.NoError(t, err)
		assert.Equal(t, core.StepStatusSuccess, result.Status)

		// Verify output contains trimmed values
		output, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		outputStr := string(output)

		// The file iterator already trims lines, but the pre-process should work anyway
		assert.Contains(t, outputStr, "hello")
		assert.Contains(t, outputStr, "world")
	})

	// Test case 4: Pre-processing with parse_url
	t.Run("parse_url pre-processing", func(t *testing.T) {
		step := &core.Step{
			Name:               "test-parse-url-preprocess",
			Type:               core.StepTypeForeach,
			Input:              inputFile,
			Variable:           "url",
			VariablePreProcess: "parse_url([[url]], '%d')", // Extract domain only
			Threads:            "1",
			Step: &core.Step{
				Type:    core.StepTypeBash,
				Command: "echo '[[url]]' >> " + outputFile,
			},
		}

		// Clean output file
		_ = os.Remove(outputFile)

		ctx := context.Background()
		result, err := dispatcher.Dispatch(ctx, step, execCtx)
		require.NoError(t, err)
		assert.Equal(t, core.StepStatusSuccess, result.Status)

		// Verify output contains just domains
		output, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		outputStr := string(output)

		assert.Contains(t, outputStr, "example.com")
	})

	// Test case 5: Pre-processing with chained functions
	t.Run("chained functions pre-processing", func(t *testing.T) {
		step := &core.Step{
			Name:               "test-chained-preprocess",
			Type:               core.StepTypeForeach,
			Input:              inputFile,
			Variable:           "url",
			VariablePreProcess: "to_lower_case(parse_url([[url]], '%d'))",
			Threads:            "1",
			Step: &core.Step{
				Type:    core.StepTypeBash,
				Command: "echo '[[url]]' >> " + outputFile,
			},
		}

		// Clean output file
		_ = os.Remove(outputFile)

		ctx := context.Background()
		result, err := dispatcher.Dispatch(ctx, step, execCtx)
		require.NoError(t, err)
		assert.Equal(t, core.StepStatusSuccess, result.Status)

		// Verify output
		output, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		outputStr := string(output)

		assert.Contains(t, outputStr, "example.com")
	})
}

func TestForeachExecutor_VariablePreProcess_ErrorHandling(t *testing.T) {
	// Create temp directory for test files
	tmpDir := t.TempDir()

	// Create test input file
	inputFile := filepath.Join(tmpDir, "input.txt")
	err := os.WriteFile(inputFile, []byte(`line1
line2
`), 0644)
	require.NoError(t, err)

	outputFile := filepath.Join(tmpDir, "output.txt")

	// Create dispatcher
	dispatcher := NewStepDispatcher()

	// Create execution context
	execCtx := core.NewExecutionContext("test-workflow", core.KindModule, "test-run-uuid", "test-target")

	// Test case: Invalid function should fall back to original value
	t.Run("invalid function fallback", func(t *testing.T) {
		step := &core.Step{
			Name:               "test-invalid-func",
			Type:               core.StepTypeForeach,
			Input:              inputFile,
			Variable:           "line",
			VariablePreProcess: "invalid_nonexistent_function([[line]])",
			Threads:            "1",
			Step: &core.Step{
				Type:    core.StepTypeBash,
				Command: "echo '[[line]]' >> " + outputFile,
			},
		}

		// Clean output file
		_ = os.Remove(outputFile)

		ctx := context.Background()
		result, err := dispatcher.Dispatch(ctx, step, execCtx)
		require.NoError(t, err)
		assert.Equal(t, core.StepStatusSuccess, result.Status)

		// Should still produce output with original values since invalid function falls back
		output, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		outputStr := string(output)

		assert.Contains(t, outputStr, "line1")
		assert.Contains(t, outputStr, "line2")
	})
}

func TestForeachExecutor_VariablePreProcess_SpecialChars(t *testing.T) {
	// Create temp directory for test files
	tmpDir := t.TempDir()

	// Create test input file with URLs containing special characters
	inputFile := filepath.Join(tmpDir, "special-urls.txt")
	err := os.WriteFile(inputFile, []byte(`https://example.com/path?foo='bar'&baz=1
https://example.com/it's/a/test
`), 0644)
	require.NoError(t, err)

	outputFile := filepath.Join(tmpDir, "output.txt")

	// Create dispatcher
	dispatcher := NewStepDispatcher()

	// Create execution context
	execCtx := core.NewExecutionContext("test-workflow", core.KindModule, "test-run-uuid", "test-target")

	t.Run("special characters in URL", func(t *testing.T) {
		step := &core.Step{
			Name:               "test-special-chars",
			Type:               core.StepTypeForeach,
			Input:              inputFile,
			Variable:           "url",
			VariablePreProcess: "get_parent_url([[url]])",
			Threads:            "1",
			Step: &core.Step{
				Type:    core.StepTypeBash,
				Command: "echo '[[url]]' >> " + outputFile,
			},
		}

		// Clean output file
		_ = os.Remove(outputFile)

		ctx := context.Background()
		result, err := dispatcher.Dispatch(ctx, step, execCtx)
		require.NoError(t, err)
		assert.Equal(t, core.StepStatusSuccess, result.Status)

		// Verify output contains parent URLs
		output, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		outputStr := string(output)

		assert.Contains(t, outputStr, "https://example.com/")
	})
}

func TestAutoQuoteForJS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "'hello'",
		},
		{
			name:     "string with single quote",
			input:    "it's a test",
			expected: "'it\\'s a test'",
		},
		{
			name:     "URL",
			input:    "https://example.com/path?foo=bar",
			expected: "'https://example.com/path?foo=bar'",
		},
		{
			name:     "URL with quotes",
			input:    "https://example.com/path?foo='bar'",
			expected: "'https://example.com/path?foo=\\'bar\\''",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := autoQuoteForJS(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
