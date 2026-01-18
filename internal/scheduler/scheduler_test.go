package scheduler

import (
	"testing"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduler_New(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)
	assert.NotNil(t, scheduler)
}

func TestScheduler_RegisterCronTrigger(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	workflow := &core.Workflow{
		Name: "test-workflow",
		Kind: core.KindModule,
	}

	trigger := &core.Trigger{
		Name:     "test-cron",
		On:       core.TriggerCron,
		Schedule: "* * * * *",
		Enabled:  true,
	}

	err = scheduler.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	triggers := scheduler.ListTriggers()
	assert.Len(t, triggers, 1)
}

func TestScheduler_DisabledTrigger(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	workflow := &core.Workflow{
		Name: "test-workflow",
		Kind: core.KindModule,
	}

	trigger := &core.Trigger{
		Name:     "test-cron",
		On:       core.TriggerCron,
		Schedule: "* * * * *",
		Enabled:  false,
	}

	err = scheduler.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	triggers := scheduler.ListTriggers()
	assert.Len(t, triggers, 0)
}

func TestScheduler_EmitEvent(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	err = scheduler.Start()
	require.NoError(t, err)
	defer func() { _ = scheduler.Stop() }()

	triggered := make(chan bool, 1)
	scheduler.SetHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string) error {
		triggered <- true
		return nil
	})

	workflow := &core.Workflow{
		Name: "test-workflow",
		Kind: core.KindModule,
	}

	trigger := &core.Trigger{
		Name:    "on-asset",
		On:      core.TriggerEvent,
		Enabled: true,
		Event: &core.EventConfig{
			Topic: "assets.new",
		},
	}

	err = scheduler.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	event := &core.Event{
		Topic:  "assets.new",
		Name:   "test-event",
		Source: "test",
	}

	err = scheduler.EmitEvent(event)
	require.NoError(t, err)

	select {
	case <-triggered:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("event trigger did not fire")
	}
}

func TestScheduler_EventFilters(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	err = scheduler.Start()
	require.NoError(t, err)
	defer func() { _ = scheduler.Stop() }()

	triggered := make(chan string, 1)
	scheduler.SetHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string) error {
		triggered <- tr.Name
		return nil
	})

	workflow := &core.Workflow{Name: "test", Kind: core.KindModule}

	trigger := &core.Trigger{
		Name:    "filtered",
		On:      core.TriggerEvent,
		Enabled: true,
		Event: &core.EventConfig{
			Topic:   "assets.new",
			Filters: []string{"event.source == 'httpx'"},
		},
	}

	err = scheduler.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	// Emit non-matching event
	_ = scheduler.EmitEvent(&core.Event{
		Topic:  "assets.new",
		Source: "nuclei",
	})

	// Give it a moment
	time.Sleep(100 * time.Millisecond)

	// Emit matching event
	_ = scheduler.EmitEvent(&core.Event{
		Topic:  "assets.new",
		Source: "httpx",
	})

	select {
	case name := <-triggered:
		assert.Equal(t, "filtered", name)
	case <-time.After(2 * time.Second):
		t.Fatal("filtered event trigger did not fire")
	}
}

func TestScheduler_StartStop(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	err = scheduler.Start()
	require.NoError(t, err)

	err = scheduler.Stop()
	require.NoError(t, err)
}

func TestScheduler_EmitEventSimple(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	err = scheduler.Start()
	require.NoError(t, err)
	defer func() { _ = scheduler.Stop() }()

	err = scheduler.EmitEventSimple("test.topic", "test-name", map[string]interface{}{
		"url": "https://example.com",
	})
	require.NoError(t, err)
}
