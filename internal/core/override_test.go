package core

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParamOverrideUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantDefault any
		wantType    *string
		wantReq     *bool
		wantGen     *string
		wantErr     bool
	}{
		{
			name:        "shorthand string",
			input:       `"hello"`,
			wantDefault: "hello",
		},
		{
			name:        "shorthand bool true",
			input:       `true`,
			wantDefault: true,
		},
		{
			name:        "shorthand bool false",
			input:       `false`,
			wantDefault: false,
		},
		{
			name:        "shorthand int",
			input:       `42`,
			wantDefault: 42,
		},
		{
			name:        "shorthand float",
			input:       `3.14`,
			wantDefault: 3.14,
		},
		{
			name:        "shorthand nil",
			input:       `null`,
			wantDefault: nil,
		},
		{
			name:        "verbose default only",
			input:       `default: "verbose-value"`,
			wantDefault: "verbose-value",
		},
		{
			name:        "verbose with type",
			input:       "default: \"value\"\ntype: \"string\"",
			wantDefault: "value",
			wantType:    ptr("string"),
		},
		{
			name:        "verbose with required",
			input:       "default: \"value\"\nrequired: true",
			wantDefault: "value",
			wantReq:     ptr(true),
		},
		{
			name:        "verbose with generator",
			input:       "default: \"value\"\ngenerator: \"uuid()\"",
			wantDefault: "value",
			wantGen:     ptr("uuid()"),
		},
		{
			name:        "verbose full struct",
			input:       "default: \"full\"\ntype: \"string\"\nrequired: true\ngenerator: \"gen()\"",
			wantDefault: "full",
			wantType:    ptr("string"),
			wantReq:     ptr(true),
			wantGen:     ptr("gen()"),
		},
		{
			name:        "verbose bool default",
			input:       `default: false`,
			wantDefault: false,
		},
		{
			name:        "verbose int default",
			input:       `default: 100`,
			wantDefault: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p ParamOverride
			err := yaml.Unmarshal([]byte(tt.input), &p)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantDefault, p.Default)
			assert.Equal(t, tt.wantType, p.Type)
			assert.Equal(t, tt.wantReq, p.Required)
			assert.Equal(t, tt.wantGen, p.Generator)
		})
	}
}

func TestParamOverrideMapUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]*ParamOverride
		wantErr bool
	}{
		{
			name: "mixed shorthand and verbose",
			input: `
param-a: "shorthand-value"
param-b:
  default: "verbose-value"
  type: "string"
param-c: false
param-d: 42
`,
			want: map[string]*ParamOverride{
				"param-a": {Default: "shorthand-value"},
				"param-b": {Default: "verbose-value", Type: ptr("string")},
				"param-c": {Default: false},
				"param-d": {Default: 42},
			},
		},
		{
			name: "all shorthand",
			input: `
name: "test"
enabled: true
count: 5
`,
			want: map[string]*ParamOverride{
				"name":    {Default: "test"},
				"enabled": {Default: true},
				"count":   {Default: 5},
			},
		},
		{
			name: "all verbose",
			input: `
param-a:
  default: "value-a"
  required: true
param-b:
  default: "value-b"
  generator: "gen()"
`,
			want: map[string]*ParamOverride{
				"param-a": {Default: "value-a", Required: ptr(true)},
				"param-b": {Default: "value-b", Generator: ptr("gen()")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]*ParamOverride
			err := yaml.Unmarshal([]byte(tt.input), &result)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, len(tt.want), len(result))

			for k, want := range tt.want {
				got, ok := result[k]
				require.True(t, ok, "missing key %s", k)
				assert.Equal(t, want.Default, got.Default, "key %s: Default mismatch", k)
				assert.Equal(t, want.Type, got.Type, "key %s: Type mismatch", k)
				assert.Equal(t, want.Required, got.Required, "key %s: Required mismatch", k)
				assert.Equal(t, want.Generator, got.Generator, "key %s: Generator mismatch", k)
			}
		})
	}
}

func TestWorkflowOverrideParamsUnmarshalYAML(t *testing.T) {
	input := `
params:
  param-a: "shorthand"
  param-b:
    default: "verbose"
    type: "string"
steps:
  mode: append
`
	var override WorkflowOverride
	err := yaml.Unmarshal([]byte(input), &override)
	require.NoError(t, err)

	require.NotNil(t, override.Params)
	require.Len(t, override.Params, 2)

	assert.Equal(t, "shorthand", override.Params["param-a"].Default)

	assert.Equal(t, "verbose", override.Params["param-b"].Default)
	require.NotNil(t, override.Params["param-b"].Type)
	assert.Equal(t, "string", *override.Params["param-b"].Type)

	require.NotNil(t, override.Steps)
	assert.Equal(t, OverrideModeAppend, override.Steps.Mode)
}

// ptr is a helper to create a pointer to a value
func ptr[T any](v T) *T {
	return &v
}
