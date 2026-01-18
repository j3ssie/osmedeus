package installer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSubPath(t *testing.T) {
	tests := []struct {
		name     string
		parent   string
		child    string
		expected bool
	}{
		{
			name:     "child inside parent",
			parent:   "/home/user/osmedeus-base",
			child:    "/home/user/osmedeus-base/workflows",
			expected: true,
		},
		{
			name:     "child is parent",
			parent:   "/home/user/osmedeus-base",
			child:    "/home/user/osmedeus-base",
			expected: true,
		},
		{
			name:     "child outside parent",
			parent:   "/home/user/osmedeus-base",
			child:    "/opt/workflows",
			expected: false,
		},
		{
			name:     "child is sibling",
			parent:   "/home/user/osmedeus-base",
			child:    "/home/user/other-folder",
			expected: false,
		},
		{
			name:     "empty parent",
			parent:   "",
			child:    "/home/user/workflows",
			expected: false,
		},
		{
			name:     "empty child",
			parent:   "/home/user/osmedeus-base",
			child:    "",
			expected: false,
		},
		{
			name:     "relative paths - child inside",
			parent:   "osmedeus-base",
			child:    "osmedeus-base/workflows",
			expected: true,
		},
		{
			name:     "relative paths - child outside",
			parent:   "osmedeus-base",
			child:    "other-folder",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSubPath(tt.parent, tt.child)
			assert.Equal(t, tt.expected, result)
		})
	}
}
