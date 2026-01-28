package core

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreferences_GetEmptyTarget(t *testing.T) {
	tests := []struct {
		name       string
		prefs      *Preferences
		defaultVal bool
		want       bool
	}{
		{
			name:       "nil preferences returns default true",
			prefs:      nil,
			defaultVal: true,
			want:       true,
		},
		{
			name:       "nil preferences returns default false",
			prefs:      nil,
			defaultVal: false,
			want:       false,
		},
		{
			name:       "nil EmptyTarget returns default true",
			prefs:      &Preferences{},
			defaultVal: true,
			want:       true,
		},
		{
			name:       "nil EmptyTarget returns default false",
			prefs:      &Preferences{},
			defaultVal: false,
			want:       false,
		},
		{
			name:       "EmptyTarget true returns true regardless of default",
			prefs:      &Preferences{EmptyTarget: boolPtr(true)},
			defaultVal: false,
			want:       true,
		},
		{
			name:       "EmptyTarget false returns false regardless of default",
			prefs:      &Preferences{EmptyTarget: boolPtr(false)},
			defaultVal: true,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.prefs.GetEmptyTarget(tt.defaultVal)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPreferences_UnmarshalYAML_EmptyTarget(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantNil  bool
		wantVal  bool
		wantErr  bool
	}{
		{
			name:    "empty_target true",
			input:   "empty_target: true",
			wantNil: false,
			wantVal: true,
		},
		{
			name:    "empty_target false",
			input:   "empty_target: false",
			wantNil: false,
			wantVal: false,
		},
		{
			name:    "empty_target not set",
			input:   "silent: true",
			wantNil: true,
		},
		{
			name:    "empty preferences",
			input:   "{}",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var prefs Preferences
			err := yaml.Unmarshal([]byte(tt.input), &prefs)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.wantNil {
				assert.Nil(t, prefs.EmptyTarget)
			} else {
				require.NotNil(t, prefs.EmptyTarget)
				assert.Equal(t, tt.wantVal, *prefs.EmptyTarget)
			}
		})
	}
}

// boolPtr is a helper to create *bool for tests
func boolPtr(b bool) *bool {
	return &b
}
