package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestReportOptionalField(t *testing.T) {
	tests := []struct {
		name     string
		yamlStr  string
		expected bool
	}{
		{
			name: "optional true",
			yamlStr: `
name: test-report
path: /tmp/test.txt
type: text
optional: true
`,
			expected: true,
		},
		{
			name: "optional false",
			yamlStr: `
name: test-report
path: /tmp/test.txt
type: text
optional: false
`,
			expected: false,
		},
		{
			name: "optional omitted defaults to false",
			yamlStr: `
name: test-report
path: /tmp/test.txt
type: text
`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var report Report
			err := yaml.Unmarshal([]byte(tt.yamlStr), &report)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, report.Optional)
			assert.Equal(t, "test-report", report.Name)
			assert.Equal(t, "text", report.Type)
		})
	}
}

func TestReportOptionalInWorkflow(t *testing.T) {
	workflowYAML := `
name: test-workflow
kind: module
description: Test workflow with mixed optional reports

reports:
  - name: required-report
    path: "{{Output}}/required.txt"
    type: text
    description: A required report

  - name: optional-report
    path: "{{Output}}/optional.json"
    type: json
    description: An optional report
    optional: true

  - name: another-required
    path: "{{Output}}/another.csv"
    type: csv

steps:
  - name: test-step
    type: bash
    commands:
      - echo "hello"
`
	var workflow Workflow
	err := yaml.Unmarshal([]byte(workflowYAML), &workflow)
	require.NoError(t, err)

	require.Len(t, workflow.Reports, 3)

	// First report: required (default)
	assert.Equal(t, "required-report", workflow.Reports[0].Name)
	assert.False(t, workflow.Reports[0].Optional)

	// Second report: explicitly optional
	assert.Equal(t, "optional-report", workflow.Reports[1].Name)
	assert.True(t, workflow.Reports[1].Optional)

	// Third report: required (omitted)
	assert.Equal(t, "another-required", workflow.Reports[2].Name)
	assert.False(t, workflow.Reports[2].Optional)
}
