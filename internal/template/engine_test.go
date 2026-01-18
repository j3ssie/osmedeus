package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixIncompleteExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "greater than at end of string",
			input:    "fileLength('file.txt') > ",
			expected: "fileLength('file.txt') > 0",
		},
		{
			name:     "less than at end of string",
			input:    "fileLength('file.txt') < ",
			expected: "fileLength('file.txt') < 0",
		},
		{
			name:     "greater than or equal at end of string",
			input:    "fileLength('file.txt') >= ",
			expected: "fileLength('file.txt') >= 0",
		},
		{
			name:     "less than or equal at end of string",
			input:    "fileLength('file.txt') <= ",
			expected: "fileLength('file.txt') <= 0",
		},
		{
			name:     "equal at end of string",
			input:    "someVar == ",
			expected: "someVar == ''",
		},
		{
			name:     "not equal at end of string",
			input:    "someVar != ",
			expected: "someVar != ''",
		},
		{
			name:     "greater than before closing paren",
			input:    "(fileLength('file.txt') > )",
			expected: "(fileLength('file.txt') > 0)",
		},
		{
			name:     "less than before closing paren",
			input:    "(fileLength('file.txt') < )",
			expected: "(fileLength('file.txt') < 0)",
		},
		{
			name:     "greater than or equal before closing paren",
			input:    "(fileLength('file.txt') >= )",
			expected: "(fileLength('file.txt') >= 0)",
		},
		{
			name:     "less than or equal before closing paren",
			input:    "(fileLength('file.txt') <= )",
			expected: "(fileLength('file.txt') <= 0)",
		},
		{
			name:     "equal before closing paren",
			input:    "(someVar == )",
			expected: "(someVar == 0)",
		},
		{
			name:     "not equal before closing paren",
			input:    "(someVar != )",
			expected: "(someVar != 0)",
		},
		{
			name:     "complete expression unchanged",
			input:    "fileLength('file.txt') > 100",
			expected: "fileLength('file.txt') > 100",
		},
		{
			name:     "no comparison operator unchanged",
			input:    "log_info('hello')",
			expected: "log_info('hello')",
		},
		{
			name:     "complex expression with multiple comparisons",
			input:    "fileLength('a.txt') > 0 && fileLength('b.txt') > ",
			expected: "fileLength('a.txt') > 0 && fileLength('b.txt') > 0",
		},
		{
			name:     "empty string unchanged",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fixIncompleteExpressions(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRender_UndefinedVariables(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		template string
		ctx      map[string]interface{}
		expected string
	}{
		{
			name:     "undefined variable in comparison becomes 0",
			template: "fileLength('file.txt') > {{limit}}",
			ctx:      map[string]interface{}{},
			expected: "fileLength('file.txt') > 0",
		},
		{
			name:     "defined variable renders normally",
			template: "fileLength('file.txt') > {{limit}}",
			ctx:      map[string]interface{}{"limit": 100},
			expected: "fileLength('file.txt') > 100",
		},
		{
			name:     "multiple undefined variables",
			template: "fileLength('{{file}}') > {{limit}}",
			ctx:      map[string]interface{}{},
			expected: "fileLength('') > 0",
		},
		{
			name:     "mixed defined and undefined",
			template: "fileLength('{{file}}') > {{limit}}",
			ctx:      map[string]interface{}{"file": "test.txt"},
			expected: "fileLength('test.txt') > 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, tt.ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
