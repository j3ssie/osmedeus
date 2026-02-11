package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestWorkflowHelpYAMLParsing(t *testing.T) {
	t.Run("both fields", func(t *testing.T) {
		data := `
kind: module
name: test
help:
  example_targets: ['example.com', 'httpbin.org']
  usage: osmedeus run -m test -t <target>
steps:
  - name: s1
    type: bash
    command: echo hi
`
		var wf Workflow
		err := yaml.Unmarshal([]byte(data), &wf)
		require.NoError(t, err)
		require.NotNil(t, wf.Help)
		assert.Equal(t, "osmedeus run -m test -t <target>", wf.Help.Usage)
		assert.Equal(t, []string{"example.com", "httpbin.org"}, wf.Help.ExampleTargets)
	})

	t.Run("usage only", func(t *testing.T) {
		data := `
kind: module
name: test
help:
  usage: osmedeus run -m test -t <target>
steps:
  - name: s1
    type: bash
    command: echo hi
`
		var wf Workflow
		err := yaml.Unmarshal([]byte(data), &wf)
		require.NoError(t, err)
		require.NotNil(t, wf.Help)
		assert.Equal(t, "osmedeus run -m test -t <target>", wf.Help.Usage)
		assert.Empty(t, wf.Help.ExampleTargets)
	})

	t.Run("omitted", func(t *testing.T) {
		data := `
kind: module
name: test
steps:
  - name: s1
    type: bash
    command: echo hi
`
		var wf Workflow
		err := yaml.Unmarshal([]byte(data), &wf)
		require.NoError(t, err)
		assert.Nil(t, wf.Help)
	})
}

func TestWorkflowGetUsage(t *testing.T) {
	t.Run("nil help", func(t *testing.T) {
		wf := &Workflow{}
		assert.Equal(t, "", wf.GetUsage())
	})

	t.Run("non-nil help", func(t *testing.T) {
		wf := &Workflow{
			Help: &WorkflowHelp{Usage: "osmedeus run -m test -t example.com"},
		}
		assert.Equal(t, "osmedeus run -m test -t example.com", wf.GetUsage())
	})
}

func TestWorkflowGetExampleTargets(t *testing.T) {
	t.Run("nil help", func(t *testing.T) {
		wf := &Workflow{}
		assert.Nil(t, wf.GetExampleTargets())
	})

	t.Run("non-nil help", func(t *testing.T) {
		wf := &Workflow{
			Help: &WorkflowHelp{ExampleTargets: []string{"example.com", "httpbin.org"}},
		}
		assert.Equal(t, []string{"example.com", "httpbin.org"}, wf.GetExampleTargets())
	})
}

func TestWorkflowHelpClone(t *testing.T) {
	t.Run("nil help", func(t *testing.T) {
		var h *WorkflowHelp
		assert.Nil(t, h.Clone())
	})

	t.Run("deep copy and mutation isolation", func(t *testing.T) {
		original := &WorkflowHelp{
			ExampleTargets: []string{"example.com", "httpbin.org"},
			Usage:          "osmedeus run -m test -t <target>",
		}
		cloned := original.Clone()

		require.NotNil(t, cloned)
		assert.Equal(t, original.Usage, cloned.Usage)
		assert.Equal(t, original.ExampleTargets, cloned.ExampleTargets)

		// Mutate clone and verify original is unaffected
		cloned.Usage = "changed"
		cloned.ExampleTargets[0] = "changed.com"
		assert.Equal(t, "osmedeus run -m test -t <target>", original.Usage)
		assert.Equal(t, "example.com", original.ExampleTargets[0])
	})

	t.Run("empty example targets", func(t *testing.T) {
		original := &WorkflowHelp{
			Usage: "some usage",
		}
		cloned := original.Clone()
		require.NotNil(t, cloned)
		assert.Equal(t, "some usage", cloned.Usage)
		assert.Nil(t, cloned.ExampleTargets)
	})
}
