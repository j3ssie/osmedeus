package core

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchesTopic(t *testing.T) {
	tests := []struct {
		name         string
		triggerTopic string
		eventTopic   string
		triggerType  TriggerType
		want         bool
	}{
		// Non-event triggers should not match
		{
			name:        "cron trigger should not match",
			triggerType: TriggerCron,
			eventTopic:  "test.topic",
			want:        false,
		},
		// Empty topic matches all
		{
			name:         "empty topic matches all events",
			triggerType:  TriggerEvent,
			triggerTopic: "",
			eventTopic:   "any.topic",
			want:         true,
		},
		// Exact matching
		{
			name:         "exact match succeeds",
			triggerType:  TriggerEvent,
			triggerTopic: "assets.new",
			eventTopic:   "assets.new",
			want:         true,
		},
		{
			name:         "exact match fails on different topic",
			triggerType:  TriggerEvent,
			triggerTopic: "assets.new",
			eventTopic:   "assets.updated",
			want:         false,
		},
		// Wildcard patterns
		{
			name:         "star matches everything",
			triggerType:  TriggerEvent,
			triggerTopic: "*",
			eventTopic:   "any.topic.here",
			want:         true,
		},
		{
			name:         "prefix wildcard - matches",
			triggerType:  TriggerEvent,
			triggerTopic: "test*",
			eventTopic:   "test.asset.new",
			want:         true,
		},
		{
			name:         "prefix wildcard - does not match",
			triggerType:  TriggerEvent,
			triggerTopic: "test*",
			eventTopic:   "other.topic",
			want:         false,
		},
		{
			name:         "suffix wildcard - matches",
			triggerType:  TriggerEvent,
			triggerTopic: "*.new",
			eventTopic:   "assets.new",
			want:         true,
		},
		{
			name:         "suffix wildcard - does not match",
			triggerType:  TriggerEvent,
			triggerTopic: "*.new",
			eventTopic:   "assets.updated",
			want:         false,
		},
		{
			name:         "middle wildcard - matches",
			triggerType:  TriggerEvent,
			triggerTopic: "assets.*.created",
			eventTopic:   "assets.subdomain.created",
			want:         true,
		},
		{
			name:         "middle wildcard - does not match different suffix",
			triggerType:  TriggerEvent,
			triggerTopic: "assets.*.created",
			eventTopic:   "assets.subdomain.updated",
			want:         false,
		},
		{
			name:         "question mark wildcard - matches single char",
			triggerType:  TriggerEvent,
			triggerTopic: "test?",
			eventTopic:   "test1",
			want:         true,
		},
		{
			name:         "question mark wildcard - does not match multiple chars",
			triggerType:  TriggerEvent,
			triggerTopic: "test?",
			eventTopic:   "test123",
			want:         false,
		},
		{
			name:         "character class - matches",
			triggerType:  TriggerEvent,
			triggerTopic: "test[abc]",
			eventTopic:   "testa",
			want:         true,
		},
		{
			name:         "character class - does not match",
			triggerType:  TriggerEvent,
			triggerTopic: "test[abc]",
			eventTopic:   "testd",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger := &Trigger{
				On: tt.triggerType,
			}
			if tt.triggerType == TriggerEvent {
				trigger.Event = &EventConfig{
					Topic: tt.triggerTopic,
				}
			}

			got := trigger.MatchesTopic(tt.eventTopic)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMatchesTopic_NilEvent(t *testing.T) {
	trigger := &Trigger{
		On:    TriggerEvent,
		Event: nil,
	}

	got := trigger.MatchesTopic("any.topic")
	assert.False(t, got, "nil event should not match")
}

func TestContainsWildcard(t *testing.T) {
	tests := []struct {
		pattern string
		want    bool
	}{
		{"simple.topic", false},
		{"*", true},
		{"test*", true},
		{"*.new", true},
		{"test.*.new", true},
		{"test?", true},
		{"test[abc]", true},
		{"test[a-z]", true},
		{"no-wildcards-here", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got := containsWildcard(tt.pattern)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTriggerInputUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name         string
		yamlInput    string
		wantHasVars  bool
		wantVars     map[string]string
		wantType     string
		wantField    string
		wantName     string
		wantFunction string
	}{
		{
			name: "legacy syntax with type event_data",
			yamlInput: `
type: event_data
field: url
name: target
`,
			wantHasVars: false,
			wantType:    "event_data",
			wantField:   "url",
			wantName:    "target",
		},
		{
			name: "legacy syntax with function",
			yamlInput: `
type: function
function: 'trim({{event.data}})'
name: result
`,
			wantHasVars:  false,
			wantType:     "function",
			wantFunction: "trim({{event.data}})",
			wantName:     "result",
		},
		{
			name: "new vars syntax - simple field access",
			yamlInput: `
target: event_data.url
source: event.source
`,
			wantHasVars: true,
			wantVars: map[string]string{
				"target": "event_data.url",
				"source": "event.source",
			},
		},
		{
			name: "new vars syntax - with function calls",
			yamlInput: `
target: event_data.url
description: trim(event_data.desc)
asset_type: event_data.type
`,
			wantHasVars: true,
			wantVars: map[string]string{
				"target":      "event_data.url",
				"description": "trim(event_data.desc)",
				"asset_type":  "event_data.type",
			},
		},
		{
			name: "new vars syntax - nested field access",
			yamlInput: `
target: event_data.metadata.url
port: event_data.metadata.port
`,
			wantHasVars: true,
			wantVars: map[string]string{
				"target": "event_data.metadata.url",
				"port":   "event_data.metadata.port",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input TriggerInput
			err := yaml.Unmarshal([]byte(tt.yamlInput), &input)
			require.NoError(t, err)

			assert.Equal(t, tt.wantHasVars, input.HasVars())

			if tt.wantHasVars {
				assert.Equal(t, tt.wantVars, input.Vars)
				// Legacy fields should be empty
				assert.Empty(t, input.Type)
				assert.Empty(t, input.Field)
				assert.Empty(t, input.Name)
			} else {
				assert.Equal(t, tt.wantType, input.Type)
				assert.Equal(t, tt.wantField, input.Field)
				assert.Equal(t, tt.wantName, input.Name)
				assert.Equal(t, tt.wantFunction, input.Function)
				assert.Nil(t, input.Vars)
			}
		})
	}
}

func TestTriggerInputHasVars(t *testing.T) {
	tests := []struct {
		name string
		ti   TriggerInput
		want bool
	}{
		{
			name: "empty TriggerInput",
			ti:   TriggerInput{},
			want: false,
		},
		{
			name: "legacy syntax",
			ti: TriggerInput{
				Type:  "event_data",
				Field: "url",
				Name:  "target",
			},
			want: false,
		},
		{
			name: "new vars syntax",
			ti: TriggerInput{
				Vars: map[string]string{
					"target": "event_data.url",
				},
			},
			want: true,
		},
		{
			name: "empty vars map",
			ti: TriggerInput{
				Vars: map[string]string{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ti.HasVars())
		})
	}
}
