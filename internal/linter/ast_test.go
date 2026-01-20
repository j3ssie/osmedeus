package linter

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseWorkflowAST(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "test", "testdata", "workflows", "linter")

	t.Run("valid workflow", func(t *testing.T) {
		path := filepath.Join(testDataDir, "valid-workflow.yaml")
		ast, err := ParseWorkflowAST(path)
		require.NoError(t, err)
		assert.NotNil(t, ast)
		assert.NotNil(t, ast.Workflow)
		assert.Equal(t, "valid-workflow", ast.Workflow.Name)
		assert.NotNil(t, ast.Root)
		assert.NotEmpty(t, ast.NodeMap)
		assert.NotEmpty(t, ast.Source)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := ParseWorkflowAST("nonexistent.yaml")
		require.Error(t, err)
	})
}

func TestParseWorkflowASTFromContent(t *testing.T) {
	content := []byte(`
name: test-workflow
kind: module
steps:
  - name: step-one
    type: bash
    command: echo "hello"
`)

	ast, err := ParseWorkflowASTFromContent(content, "test.yaml")
	require.NoError(t, err)
	assert.NotNil(t, ast)
	assert.Equal(t, "test-workflow", ast.Workflow.Name)
	assert.Equal(t, "test.yaml", ast.FilePath)
}

func TestGetNodePosition(t *testing.T) {
	content := []byte(`name: test-workflow
kind: module
steps:
  - name: step-one
    type: bash
    command: echo "hello"
`)

	ast, err := ParseWorkflowASTFromContent(content, "test.yaml")
	require.NoError(t, err)

	// Test getting position for known paths
	line, _ := ast.GetNodePosition("name")
	assert.Greater(t, line, 0, "should have a valid line number")

	line, _ = ast.GetNodePosition("steps")
	assert.Greater(t, line, 0, "should have a valid line number for steps")

	// Test unknown path returns 0, 0
	line, col := ast.GetNodePosition("unknown.path")
	assert.Equal(t, 0, line)
	assert.Equal(t, 0, col)
}

func TestGetLine(t *testing.T) {
	content := []byte(`line one
line two
line three`)

	ast, err := ParseWorkflowASTFromContent(content, "test.yaml")
	// This may fail parsing but we're testing GetLine
	if err != nil {
		ast = &WorkflowAST{Source: content}
	}

	assert.Equal(t, "line one", ast.GetLine(1))
	assert.Equal(t, "line two", ast.GetLine(2))
	assert.Equal(t, "line three", ast.GetLine(3))
	assert.Equal(t, "", ast.GetLine(0))  // Out of bounds
	assert.Equal(t, "", ast.GetLine(10)) // Out of bounds
}

func TestFindStepPosition(t *testing.T) {
	content := []byte(`name: test-workflow
kind: module
steps:
  - name: first-step
    type: bash
    command: echo "one"
  - name: second-step
    type: bash
    command: echo "two"
`)

	ast, err := ParseWorkflowASTFromContent(content, "test.yaml")
	require.NoError(t, err)

	line, _ := ast.FindStepPosition("first-step")
	assert.Greater(t, line, 0, "should find first-step position")

	line, _ = ast.FindStepPosition("second-step")
	assert.Greater(t, line, 0, "should find second-step position")

	line, _ = ast.FindStepPosition("nonexistent-step")
	assert.Equal(t, 0, line, "should return 0 for nonexistent step")
}
