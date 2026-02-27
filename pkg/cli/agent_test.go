package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveAgentMessage_Positional(t *testing.T) {
	msg, err := resolveAgentMessage([]string{"hello", "world"})
	assert.NoError(t, err)
	assert.Equal(t, "hello world", msg)
}

func TestResolveAgentMessage_SingleArg(t *testing.T) {
	msg, err := resolveAgentMessage([]string{"test message"})
	assert.NoError(t, err)
	assert.Equal(t, "test message", msg)
}

func TestResolveAgentMessage_NoArgs(t *testing.T) {
	_, err := resolveAgentMessage(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no message provided")
}

func TestResolveAgentMessage_EmptyArgs(t *testing.T) {
	_, err := resolveAgentMessage([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no message provided")
}
