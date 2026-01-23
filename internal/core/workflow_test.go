package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCronTriggers(t *testing.T) {
	tests := []struct {
		name     string
		workflow Workflow
		want     int
	}{
		{
			name: "no triggers",
			workflow: Workflow{
				Name:     "test",
				Triggers: nil,
			},
			want: 0,
		},
		{
			name: "one enabled cron trigger",
			workflow: Workflow{
				Name: "test",
				Triggers: []Trigger{
					{Name: "cron-1", On: TriggerCron, Schedule: "* * * * *", Enabled: true},
				},
			},
			want: 1,
		},
		{
			name: "one disabled cron trigger",
			workflow: Workflow{
				Name: "test",
				Triggers: []Trigger{
					{Name: "cron-1", On: TriggerCron, Schedule: "* * * * *", Enabled: false},
				},
			},
			want: 0,
		},
		{
			name: "mixed triggers - only enabled cron returned",
			workflow: Workflow{
				Name: "test",
				Triggers: []Trigger{
					{Name: "cron-enabled", On: TriggerCron, Schedule: "0 0 * * *", Enabled: true},
					{Name: "cron-disabled", On: TriggerCron, Schedule: "0 1 * * *", Enabled: false},
					{Name: "event-enabled", On: TriggerEvent, Enabled: true},
					{Name: "manual", On: TriggerManual, Enabled: true},
				},
			},
			want: 1,
		},
		{
			name: "multiple enabled cron triggers",
			workflow: Workflow{
				Name: "test",
				Triggers: []Trigger{
					{Name: "hourly", On: TriggerCron, Schedule: "0 * * * *", Enabled: true},
					{Name: "daily", On: TriggerCron, Schedule: "0 0 * * *", Enabled: true},
					{Name: "weekly", On: TriggerCron, Schedule: "0 0 * * 0", Enabled: true},
				},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.workflow.GetCronTriggers()
			assert.Equal(t, tt.want, len(got))

			// Verify all returned triggers are cron and enabled
			for _, trigger := range got {
				assert.Equal(t, TriggerCron, trigger.On)
				assert.True(t, trigger.Enabled)
			}
		})
	}
}

func TestGetEventTriggers(t *testing.T) {
	tests := []struct {
		name     string
		workflow Workflow
		want     int
	}{
		{
			name: "no triggers",
			workflow: Workflow{
				Name:     "test",
				Triggers: nil,
			},
			want: 0,
		},
		{
			name: "one enabled event trigger",
			workflow: Workflow{
				Name: "test",
				Triggers: []Trigger{
					{Name: "event-1", On: TriggerEvent, Enabled: true},
				},
			},
			want: 1,
		},
		{
			name: "mixed triggers - only enabled events returned",
			workflow: Workflow{
				Name: "test",
				Triggers: []Trigger{
					{Name: "cron-1", On: TriggerCron, Schedule: "* * * * *", Enabled: true},
					{Name: "event-enabled", On: TriggerEvent, Enabled: true},
					{Name: "event-disabled", On: TriggerEvent, Enabled: false},
				},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.workflow.GetEventTriggers()
			assert.Equal(t, tt.want, len(got))

			// Verify all returned triggers are event and enabled
			for _, trigger := range got {
				assert.Equal(t, TriggerEvent, trigger.On)
				assert.True(t, trigger.Enabled)
			}
		})
	}
}
