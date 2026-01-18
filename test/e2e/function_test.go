package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFunction_List(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing function list command")

	stdout, _, err := runCLIWithLog(t, log, "function", "list")
	require.NoError(t, err)

	log.Info("Asserting stdout contains function categories")
	assert.Contains(t, stdout, "| File")
	assert.Contains(t, stdout, "| String")

	log.Success("function list displays all function categories")
}

func TestFunction_Eval_Simple(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing function eval with simple expression")

	log.Info("Evaluating expression: 1+1")
	stdout, _, err := runCLIWithLog(t, log, "function", "eval", "-e", "1+1")
	require.NoError(t, err)

	log.Info("Asserting stdout contains result: 2")
	assert.Contains(t, stdout, "2")

	log.Success("function eval evaluates simple expressions")
}

func TestFunction_Eval_WithTarget(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing function eval with target variable")

	log.Info("Evaluating expression: target with -t example.com")
	stdout, _, err := runCLIWithLog(t, log, "function", "eval", "-e", "target", "-t", "example.com")
	require.NoError(t, err)

	log.Info("Asserting stdout contains target value")
	assert.Contains(t, stdout, "example.com")

	log.Success("function eval resolves target variable")
}

func TestFunction_Eval_StringFunc(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing function eval with string function")

	log.Info("Evaluating expression: trim('  hello  ')")
	stdout, _, err := runCLIWithLog(t, log, "function", "eval", "-e", "trim('  hello  ')")
	require.NoError(t, err)

	log.Info("Asserting stdout contains trimmed result")
	assert.Contains(t, stdout, "hello")

	log.Success("function eval executes string functions")
}
