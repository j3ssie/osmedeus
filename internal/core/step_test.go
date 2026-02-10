package core

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStepSuppressDetails(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected bool
	}{
		{
			name: "suppress_details true",
			yaml: `
name: test-step
type: function
suppress_details: true
functions:
  - exec_cmd("echo hello")
`,
			expected: true,
		},
		{
			name: "suppress_details false",
			yaml: `
name: test-step
type: function
suppress_details: false
functions:
  - exec_cmd("echo hello")
`,
			expected: false,
		},
		{
			name: "suppress_details omitted defaults to false",
			yaml: `
name: test-step
type: function
functions:
  - exec_cmd("echo hello")
`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var step Step
			err := yaml.Unmarshal([]byte(tt.yaml), &step)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, step.SuppressDetails)
			assert.Equal(t, "test-step", step.Name)
		})
	}
}
