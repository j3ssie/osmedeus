package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestWorkflowHooksYAMLParsing(t *testing.T) {
	t.Run("both pre and post hooks", func(t *testing.T) {
		data := `
kind: module
name: test-hooks
hooks:
  pre_scan_steps:
    - name: pre-1
      type: bash
      command: echo pre
    - name: pre-2
      type: bash
      command: echo pre2
  post_scan_steps:
    - name: post-1
      type: bash
      command: echo post
steps:
  - name: s1
    type: bash
    command: echo main
`
		var wf Workflow
		err := yaml.Unmarshal([]byte(data), &wf)
		require.NoError(t, err)
		require.NotNil(t, wf.Hooks)
		assert.Len(t, wf.Hooks.PreScanSteps, 2)
		assert.Len(t, wf.Hooks.PostScanSteps, 1)
		assert.Equal(t, "pre-1", wf.Hooks.PreScanSteps[0].Name)
		assert.Equal(t, "post-1", wf.Hooks.PostScanSteps[0].Name)
	})

	t.Run("pre only", func(t *testing.T) {
		data := `
kind: module
name: test
hooks:
  pre_scan_steps:
    - name: pre-1
      type: bash
      command: echo pre
steps:
  - name: s1
    type: bash
    command: echo main
`
		var wf Workflow
		err := yaml.Unmarshal([]byte(data), &wf)
		require.NoError(t, err)
		require.NotNil(t, wf.Hooks)
		assert.Len(t, wf.Hooks.PreScanSteps, 1)
		assert.Empty(t, wf.Hooks.PostScanSteps)
	})

	t.Run("post only", func(t *testing.T) {
		data := `
kind: module
name: test
hooks:
  post_scan_steps:
    - name: post-1
      type: bash
      command: echo post
steps:
  - name: s1
    type: bash
    command: echo main
`
		var wf Workflow
		err := yaml.Unmarshal([]byte(data), &wf)
		require.NoError(t, err)
		require.NotNil(t, wf.Hooks)
		assert.Empty(t, wf.Hooks.PreScanSteps)
		assert.Len(t, wf.Hooks.PostScanSteps, 1)
	})

	t.Run("no hooks", func(t *testing.T) {
		data := `
kind: module
name: test
steps:
  - name: s1
    type: bash
    command: echo main
`
		var wf Workflow
		err := yaml.Unmarshal([]byte(data), &wf)
		require.NoError(t, err)
		assert.Nil(t, wf.Hooks)
	})

	t.Run("empty hooks block", func(t *testing.T) {
		data := `
kind: module
name: test
hooks: {}
steps:
  - name: s1
    type: bash
    command: echo main
`
		var wf Workflow
		err := yaml.Unmarshal([]byte(data), &wf)
		require.NoError(t, err)
		// Empty hooks block is parsed but contains no steps
		if wf.Hooks != nil {
			assert.Empty(t, wf.Hooks.PreScanSteps)
			assert.Empty(t, wf.Hooks.PostScanSteps)
		}
	})

	t.Run("flow with hooks", func(t *testing.T) {
		data := `
kind: flow
name: test-flow
hooks:
  pre_scan_steps:
    - name: flow-pre
      type: bash
      command: echo flow-pre
  post_scan_steps:
    - name: flow-post
      type: bash
      command: echo flow-post
modules:
  - name: mod1
    steps:
      - name: inline-step
        type: bash
        command: echo inline
`
		var wf Workflow
		err := yaml.Unmarshal([]byte(data), &wf)
		require.NoError(t, err)
		require.NotNil(t, wf.Hooks)
		assert.Len(t, wf.Hooks.PreScanSteps, 1)
		assert.Len(t, wf.Hooks.PostScanSteps, 1)
	})
}

func TestWorkflowHookCount(t *testing.T) {
	t.Run("nil hooks", func(t *testing.T) {
		wf := &Workflow{}
		assert.Equal(t, 0, wf.HookCount())
	})

	t.Run("empty hooks", func(t *testing.T) {
		wf := &Workflow{Hooks: &WorkflowHooks{}}
		assert.Equal(t, 0, wf.HookCount())
	})

	t.Run("pre only", func(t *testing.T) {
		wf := &Workflow{
			Hooks: &WorkflowHooks{
				PreScanSteps: []Step{{Name: "pre-1"}, {Name: "pre-2"}},
			},
		}
		assert.Equal(t, 2, wf.HookCount())
	})

	t.Run("post only", func(t *testing.T) {
		wf := &Workflow{
			Hooks: &WorkflowHooks{
				PostScanSteps: []Step{{Name: "post-1"}},
			},
		}
		assert.Equal(t, 1, wf.HookCount())
	})

	t.Run("both pre and post", func(t *testing.T) {
		wf := &Workflow{
			Hooks: &WorkflowHooks{
				PreScanSteps:  []Step{{Name: "pre-1"}, {Name: "pre-2"}},
				PostScanSteps: []Step{{Name: "post-1"}, {Name: "post-2"}, {Name: "post-3"}},
			},
		}
		assert.Equal(t, 5, wf.HookCount())
	})
}
