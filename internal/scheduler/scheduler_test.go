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

func TestScheduler_EventInputExtraction(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	err = scheduler.Start()
	require.NoError(t, err)
	defer func() { _ = scheduler.Stop() }()

	receivedInput := make(chan string, 1)
	scheduler.SetHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string) error {
		receivedInput <- input
		return nil
	})

	workflow := &core.Workflow{Name: "test", Kind: core.KindModule}

	trigger := &core.Trigger{
		Name:    "with-input",
		On:      core.TriggerEvent,
		Enabled: true,
		Event: &core.EventConfig{
			Topic: "assets.new",
		},
		Input: core.TriggerInput{
			Type:  "event_data",
			Field: "url",
			Name:  "target",
		},
	}

	err = scheduler.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	event := &core.Event{
		Topic:  "assets.new",
		Source: "httpx",
		Data:   `{"url": "https://api.example.com", "status": 200}`,
	}

	err = scheduler.EmitEvent(event)
	require.NoError(t, err)

	select {
	case input := <-receivedInput:
		assert.Equal(t, "https://api.example.com", input)
	case <-time.After(2 * time.Second):
		t.Fatal("event trigger did not fire")
	}
}

func TestScheduler_MultipleEventTriggers(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	err = scheduler.Start()
	require.NoError(t, err)
	defer func() { _ = scheduler.Stop() }()

	triggered := make(chan string, 10)
	scheduler.SetHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string) error {
		triggered <- tr.Name
		return nil
	})

	workflow := &core.Workflow{Name: "test", Kind: core.KindModule}

	// Register multiple triggers for different topics
	triggers := []*core.Trigger{
		{
			Name:    "assets-trigger",
			On:      core.TriggerEvent,
			Enabled: true,
			Event:   &core.EventConfig{Topic: "assets.new"},
		},
		{
			Name:    "vuln-trigger",
			On:      core.TriggerEvent,
			Enabled: true,
			Event:   &core.EventConfig{Topic: "vulnerabilities.new"},
		},
		{
			Name:    "run-trigger",
			On:      core.TriggerEvent,
			Enabled: true,
			Event:   &core.EventConfig{Topic: "run.completed"},
		},
	}

	for _, tr := range triggers {
		err = scheduler.RegisterTrigger(workflow, tr)
		require.NoError(t, err)
	}

	// Emit events for each topic
	_ = scheduler.EmitEvent(&core.Event{Topic: "assets.new", Name: "asset-event"})
	_ = scheduler.EmitEvent(&core.Event{Topic: "vulnerabilities.new", Name: "vuln-event"})
	_ = scheduler.EmitEvent(&core.Event{Topic: "run.completed", Name: "run-event"})

	// Collect all triggered names
	triggeredNames := make(map[string]bool)
	timeout := time.After(3 * time.Second)

	for i := 0; i < 3; i++ {
		select {
		case name := <-triggered:
			triggeredNames[name] = true
		case <-timeout:
			t.Fatalf("only received %d triggers, expected 3", i)
		}
	}

	assert.True(t, triggeredNames["assets-trigger"])
	assert.True(t, triggeredNames["vuln-trigger"])
	assert.True(t, triggeredNames["run-trigger"])
}

func TestScheduler_EventFilterMultipleConditions(t *testing.T) {
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

	workflow := &core.Workflow{Name: "test", Kind: core.KindModule}

	// Trigger with multiple filter conditions (AND logic)
	trigger := &core.Trigger{
		Name:    "multi-filter",
		On:      core.TriggerEvent,
		Enabled: true,
		Event: &core.EventConfig{
			Topic: "vulnerabilities.new",
			Filters: []string{
				"event.source == 'nuclei'",
				"event.data_type == 'finding'",
			},
		},
	}

	err = scheduler.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	// Event that matches only one condition - should NOT trigger
	_ = scheduler.EmitEvent(&core.Event{
		Topic:    "vulnerabilities.new",
		Source:   "nuclei",
		DataType: "other",
	})

	time.Sleep(100 * time.Millisecond)

	// Event that matches both conditions - should trigger
	_ = scheduler.EmitEvent(&core.Event{
		Topic:    "vulnerabilities.new",
		Source:   "nuclei",
		DataType: "finding",
	})

	select {
	case <-triggered:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("multi-filter event trigger did not fire")
	}
}

func TestScheduler_EventFilterWithParsedData(t *testing.T) {
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

	workflow := &core.Workflow{Name: "test", Kind: core.KindModule}

	trigger := &core.Trigger{
		Name:    "severity-filter",
		On:      core.TriggerEvent,
		Enabled: true,
		Event: &core.EventConfig{
			Topic:   "vulnerabilities.new",
			Filters: []string{"event.data.severity == 'critical'"},
		},
	}

	err = scheduler.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	// Create event with parsed data
	event := &core.Event{
		Topic:  "vulnerabilities.new",
		Source: "nuclei",
		Data:   `{"severity": "critical", "template": "CVE-2024-1234"}`,
	}
	_ = event.ParseData()

	_ = scheduler.EmitEvent(event)

	select {
	case <-triggered:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("severity filter event trigger did not fire")
	}
}

func TestScheduler_EventNotTriggeredForWrongTopic(t *testing.T) {
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

	workflow := &core.Workflow{Name: "test", Kind: core.KindModule}

	trigger := &core.Trigger{
		Name:    "assets-only",
		On:      core.TriggerEvent,
		Enabled: true,
		Event:   &core.EventConfig{Topic: "assets.new"},
	}

	err = scheduler.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	// Emit event with wrong topic
	_ = scheduler.EmitEvent(&core.Event{
		Topic:  "vulnerabilities.new",
		Source: "nuclei",
	})

	select {
	case <-triggered:
		t.Fatal("trigger should not fire for wrong topic")
	case <-time.After(500 * time.Millisecond):
		// Success - trigger didn't fire
	}
}

func TestScheduler_EventQueueMetrics(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	err = scheduler.Start()
	require.NoError(t, err)
	defer func() { _ = scheduler.Stop() }()

	// Initial metrics should be zero
	enqueued, dropped := scheduler.GetEventMetrics()
	assert.Equal(t, int64(0), enqueued)
	assert.Equal(t, int64(0), dropped)

	// Emit some events
	for i := 0; i < 5; i++ {
		_ = scheduler.EmitEvent(&core.Event{
			Topic:  "test.topic",
			Name:   "test-event",
			Source: "test",
		})
	}

	// Allow events to be processed
	time.Sleep(100 * time.Millisecond)

	enqueued, dropped = scheduler.GetEventMetrics()
	assert.Equal(t, int64(5), enqueued)
	assert.Equal(t, int64(0), dropped)
}

func TestScheduler_EventQueueStats(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	queueSize, currentLen, backpressureTimeout := scheduler.GetQueueStats()
	assert.Equal(t, 1000, queueSize)
	assert.Equal(t, 0, currentLen)
	assert.Equal(t, 5*time.Second, backpressureTimeout)
}

func TestScheduler_CustomQueueConfig(t *testing.T) {
	scheduler, err := NewSchedulerWithConfig(100, 1*time.Second)
	require.NoError(t, err)

	queueSize, _, backpressureTimeout := scheduler.GetQueueStats()
	assert.Equal(t, 100, queueSize)
	assert.Equal(t, 1*time.Second, backpressureTimeout)
}

func TestScheduler_EmitEventWhenStopped(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	// Don't start scheduler - try to emit event
	event := &core.Event{
		Topic:  "test.topic",
		Source: "test",
	}

	err = scheduler.EmitEvent(event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scheduler not running")
}

func TestScheduler_EventTimestampAutoSet(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	err = scheduler.Start()
	require.NoError(t, err)
	defer func() { _ = scheduler.Stop() }()

	receivedEvent := make(chan *core.Event, 1)
	scheduler.SetHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string) error {
		return nil
	})

	workflow := &core.Workflow{Name: "test", Kind: core.KindModule}
	trigger := &core.Trigger{
		Name:    "catch-all",
		On:      core.TriggerEvent,
		Enabled: true,
		Event:   &core.EventConfig{Topic: "test.topic"},
	}

	_ = scheduler.RegisterTrigger(workflow, trigger)

	// Event without timestamp
	event := &core.Event{
		Topic:  "test.topic",
		Source: "test",
	}
	assert.True(t, event.Timestamp.IsZero())

	err = scheduler.EmitEvent(event)
	require.NoError(t, err)

	// After emitting, timestamp should be set
	assert.False(t, event.Timestamp.IsZero())

	close(receivedEvent)
}

func TestScheduler_RegisterEventTrigger(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	workflow := &core.Workflow{
		Name: "event-workflow",
		Kind: core.KindModule,
	}

	trigger := &core.Trigger{
		Name:    "on-discovery",
		On:      core.TriggerEvent,
		Enabled: true,
		Event: &core.EventConfig{
			Topic: "assets.new",
			Filters: []string{
				"event.source == 'subfinder'",
			},
		},
		Input: core.TriggerInput{
			Type:  "event_data",
			Field: "subdomain",
			Name:  "target",
		},
	}

	err = scheduler.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	triggers := scheduler.ListTriggers()
	assert.Len(t, triggers, 1)
	assert.Equal(t, "on-discovery", triggers[0].Trigger.Name)
}

func TestScheduler_UnregisterTrigger(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	workflow := &core.Workflow{
		Name: "test-workflow",
		Kind: core.KindModule,
	}

	trigger := &core.Trigger{
		Name:    "to-remove",
		On:      core.TriggerEvent,
		Enabled: true,
		Event:   &core.EventConfig{Topic: "test.topic"},
	}

	err = scheduler.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	triggers := scheduler.ListTriggers()
	assert.Len(t, triggers, 1)

	// Unregister - key is "workflowName:triggerName"
	_ = scheduler.UnregisterTrigger("test-workflow:to-remove")

	triggers = scheduler.ListTriggers()
	assert.Len(t, triggers, 0)
}

func TestScheduler_EventChainWorkflows(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	err = scheduler.Start()
	require.NoError(t, err)
	defer func() { _ = scheduler.Stop() }()

	// Track execution order
	executionOrder := make(chan string, 10)

	scheduler.SetHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string) error {
		executionOrder <- w.Name
		return nil
	})

	// Workflow 1: triggers on assets.new
	workflow1 := &core.Workflow{Name: "recon", Kind: core.KindModule}
	trigger1 := &core.Trigger{
		Name:    "on-new-asset",
		On:      core.TriggerEvent,
		Enabled: true,
		Event:   &core.EventConfig{Topic: "assets.new"},
	}

	// Workflow 2: triggers on probing.complete
	workflow2 := &core.Workflow{Name: "scanner", Kind: core.KindModule}
	trigger2 := &core.Trigger{
		Name:    "on-probe-complete",
		On:      core.TriggerEvent,
		Enabled: true,
		Event:   &core.EventConfig{Topic: "probing.complete"},
	}

	_ = scheduler.RegisterTrigger(workflow1, trigger1)
	_ = scheduler.RegisterTrigger(workflow2, trigger2)

	// Simulate chain: assets.new -> probing.complete
	_ = scheduler.EmitEvent(&core.Event{Topic: "assets.new"})

	// Wait for first trigger
	select {
	case name := <-executionOrder:
		assert.Equal(t, "recon", name)
	case <-time.After(2 * time.Second):
		t.Fatal("first workflow didn't trigger")
	}

	// Emit second event (simulating first workflow completion)
	_ = scheduler.EmitEvent(&core.Event{Topic: "probing.complete"})

	// Wait for second trigger
	select {
	case name := <-executionOrder:
		assert.Equal(t, "scanner", name)
	case <-time.After(2 * time.Second):
		t.Fatal("second workflow didn't trigger")
	}
}

func TestScheduler_EventDeduplication(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	err = scheduler.Start()
	require.NoError(t, err)
	defer func() { _ = scheduler.Stop() }()

	triggerCount := make(chan int, 10)
	count := 0
	scheduler.SetHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string) error {
		count++
		triggerCount <- count
		return nil
	})

	workflow := &core.Workflow{Name: "test", Kind: core.KindModule}

	// Trigger with deduplication - 1 second window
	trigger := &core.Trigger{
		Name:    "dedupe-trigger",
		On:      core.TriggerEvent,
		Enabled: true,
		Event: &core.EventConfig{
			Topic:        "assets.new",
			DedupeKey:    "{{event.source}}-{{event.data.url}}",
			DedupeWindow: "1s",
		},
	}

	err = scheduler.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	// Emit the same event 3 times rapidly - only first should trigger
	for i := 0; i < 3; i++ {
		event := &core.Event{
			Topic:  "assets.new",
			Source: "httpx",
			Data:   `{"url": "https://example.com"}`,
		}
		_ = event.ParseData()
		_ = scheduler.EmitEvent(event)
	}

	// Wait for first event to be processed
	select {
	case c := <-triggerCount:
		assert.Equal(t, 1, c)
	case <-time.After(2 * time.Second):
		t.Fatal("trigger did not fire for first event")
	}

	// Give time for duplicate events to be processed (they should be skipped)
	time.Sleep(200 * time.Millisecond)

	// No more triggers should have fired
	select {
	case c := <-triggerCount:
		t.Fatalf("unexpected trigger fired, count: %d", c)
	default:
		// Good - no more triggers
	}
}

func TestScheduler_EventDeduplicationDifferentKeys(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	err = scheduler.Start()
	require.NoError(t, err)
	defer func() { _ = scheduler.Stop() }()

	triggerCount := 0
	triggered := make(chan bool, 10)
	scheduler.SetHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string) error {
		triggerCount++
		triggered <- true
		return nil
	})

	workflow := &core.Workflow{Name: "test", Kind: core.KindModule}

	trigger := &core.Trigger{
		Name:    "dedupe-trigger",
		On:      core.TriggerEvent,
		Enabled: true,
		Event: &core.EventConfig{
			Topic:        "assets.new",
			DedupeKey:    "{{event.data.url}}",
			DedupeWindow: "1s",
		},
	}

	err = scheduler.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	// Emit events with different URLs - all should trigger
	urls := []string{
		"https://example1.com",
		"https://example2.com",
		"https://example3.com",
	}

	for _, url := range urls {
		event := &core.Event{
			Topic:  "assets.new",
			Source: "httpx",
			Data:   `{"url": "` + url + `"}`,
		}
		_ = event.ParseData()
		_ = scheduler.EmitEvent(event)
	}

	// Wait for all 3 triggers
	for i := 0; i < 3; i++ {
		select {
		case <-triggered:
			// Good
		case <-time.After(2 * time.Second):
			t.Fatalf("expected trigger %d did not fire", i+1)
		}
	}

	assert.Equal(t, 3, triggerCount, "all 3 unique events should trigger")
}

func TestScheduler_DedupeExpiry(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	err = scheduler.Start()
	require.NoError(t, err)
	defer func() { _ = scheduler.Stop() }()

	triggerCount := 0
	triggered := make(chan bool, 10)
	scheduler.SetHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string) error {
		triggerCount++
		triggered <- true
		return nil
	})

	workflow := &core.Workflow{Name: "test", Kind: core.KindModule}

	// Very short dedupe window for testing
	trigger := &core.Trigger{
		Name:    "short-dedupe",
		On:      core.TriggerEvent,
		Enabled: true,
		Event: &core.EventConfig{
			Topic:        "assets.new",
			DedupeKey:    "{{event.data.url}}",
			DedupeWindow: "200ms",
		},
	}

	err = scheduler.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	// Emit first event
	event := &core.Event{
		Topic:  "assets.new",
		Source: "httpx",
		Data:   `{"url": "https://example.com"}`,
	}
	_ = event.ParseData()
	_ = scheduler.EmitEvent(event)

	// Wait for first trigger
	select {
	case <-triggered:
		// Good
	case <-time.After(2 * time.Second):
		t.Fatal("first trigger did not fire")
	}

	// Wait for dedupe window to expire
	time.Sleep(300 * time.Millisecond)

	// Emit same event again - should trigger because window expired
	event2 := &core.Event{
		Topic:  "assets.new",
		Source: "httpx",
		Data:   `{"url": "https://example.com"}`,
	}
	_ = event2.ParseData()
	_ = scheduler.EmitEvent(event2)

	select {
	case <-triggered:
		// Good - triggered after window expired
	case <-time.After(2 * time.Second):
		t.Fatal("second trigger did not fire after window expired")
	}

	assert.Equal(t, 2, triggerCount, "event should trigger twice after window expiry")
}

func TestScheduler_DedupeKeyTemplate(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	// Test computeDedupeKey directly
	event := &core.Event{
		Topic:    "assets.new",
		Name:     "test-event",
		Source:   "httpx",
		ID:       "evt-123",
		DataType: "url",
		Data:     `{"url": "https://example.com", "status": 200}`,
	}
	_ = event.ParseData()

	tests := []struct {
		template string
		expected string
	}{
		{
			template: "{{event.source}}",
			expected: "httpx",
		},
		{
			template: "{{event.topic}}-{{event.source}}",
			expected: "assets.new-httpx",
		},
		{
			template: "{{event.source}}-{{event.data.url}}",
			expected: "httpx-https://example.com",
		},
		{
			template: "{{event.id}}",
			expected: "evt-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			key := scheduler.computeDedupeKey(tt.template, event)
			assert.Equal(t, tt.expected, key)
		})
	}
}

func TestDedupeCache_IsDuplicate(t *testing.T) {
	cache := newDedupeCache()

	// First check - not a duplicate
	assert.False(t, cache.IsDuplicate("key1", 1*time.Second))

	// Mark the key
	cache.Mark("key1", 1*time.Second)

	// Second check - should be duplicate
	assert.True(t, cache.IsDuplicate("key1", 1*time.Second))

	// Different key - not a duplicate
	assert.False(t, cache.IsDuplicate("key2", 1*time.Second))
}

func TestDedupeCache_Cleanup(t *testing.T) {
	cache := newDedupeCache()

	// Mark with very short window
	cache.Mark("short", 10*time.Millisecond)
	cache.Mark("long", 5*time.Second)

	// Initially both should be duplicates
	assert.True(t, cache.IsDuplicate("short", 10*time.Millisecond))
	assert.True(t, cache.IsDuplicate("long", 5*time.Second))

	// Wait for short to expire
	time.Sleep(50 * time.Millisecond)

	// Run cleanup
	cache.cleanup()

	// Short should no longer be duplicate (expired and cleaned)
	assert.False(t, cache.IsDuplicate("short", 10*time.Millisecond))

	// Long should still be duplicate
	assert.True(t, cache.IsDuplicate("long", 5*time.Second))
}

func TestTrigger_GetDebounceDuration(t *testing.T) {
	tests := []struct {
		name     string
		debounce string
		expected time.Duration
	}{
		{"empty", "", 0},
		{"500ms", "500ms", 500 * time.Millisecond},
		{"1s", "1s", 1 * time.Second},
		{"invalid", "invalid", 0},
		{"100ms", "100ms", 100 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger := &core.Trigger{Debounce: tt.debounce}
			assert.Equal(t, tt.expected, trigger.GetDebounceDuration())
		})
	}
}

func TestTrigger_HasDebounce(t *testing.T) {
	trigger := &core.Trigger{Debounce: "500ms"}
	assert.True(t, trigger.HasDebounce())

	trigger2 := &core.Trigger{}
	assert.False(t, trigger2.HasDebounce())

	trigger3 := &core.Trigger{Debounce: "invalid"}
	assert.False(t, trigger3.HasDebounce())
}

func TestEventConfig_GetDedupeWindow(t *testing.T) {
	tests := []struct {
		name     string
		config   *core.EventConfig
		expected time.Duration
	}{
		{"nil config", nil, 0},
		{"empty window", &core.EventConfig{DedupeWindow: ""}, 0},
		{"1s", &core.EventConfig{DedupeWindow: "1s"}, 1 * time.Second},
		{"10s", &core.EventConfig{DedupeWindow: "10s"}, 10 * time.Second},
		{"invalid", &core.EventConfig{DedupeWindow: "invalid"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.GetDedupeWindow())
		})
	}
}

func TestEventConfig_HasDeduplication(t *testing.T) {
	tests := []struct {
		name     string
		config   *core.EventConfig
		expected bool
	}{
		{"nil config", nil, false},
		{"no key or window", &core.EventConfig{}, false},
		{"key but no window", &core.EventConfig{DedupeKey: "{{event.id}}"}, false},
		{"window but no key", &core.EventConfig{DedupeWindow: "1s"}, false},
		{"both set", &core.EventConfig{DedupeKey: "{{event.id}}", DedupeWindow: "1s"}, true},
		{"invalid window", &core.EventConfig{DedupeKey: "{{event.id}}", DedupeWindow: "invalid"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.HasDeduplication())
		})
	}
}

func TestScheduler_ResolveDotNotation(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	event := &core.Event{
		Topic:        "assets.new",
		Name:         "test-event",
		Source:       "httpx",
		ID:           "evt-123",
		DataType:     "url",
		Workspace:    "/tmp/workspace",
		RunUUID:      "run-456",
		WorkflowName: "test-workflow",
		ParsedData: map[string]interface{}{
			"url":    "https://example.com",
			"status": 200,
			"metadata": map[string]interface{}{
				"port":     443,
				"protocol": "https",
			},
		},
	}

	tests := []struct {
		name     string
		expr     string
		expected string
	}{
		{
			name:     "event_data simple field",
			expr:     "event_data.url",
			expected: "https://example.com",
		},
		{
			name:     "event_data nested field",
			expr:     "event_data.metadata.port",
			expected: "443",
		},
		{
			name:     "event_data nested string field",
			expr:     "event_data.metadata.protocol",
			expected: "https",
		},
		{
			name:     "event.topic",
			expr:     "event.topic",
			expected: "assets.new",
		},
		{
			name:     "event.source",
			expr:     "event.source",
			expected: "httpx",
		},
		{
			name:     "event.name",
			expr:     "event.name",
			expected: "test-event",
		},
		{
			name:     "event.id",
			expr:     "event.id",
			expected: "evt-123",
		},
		{
			name:     "event.workspace",
			expr:     "event.workspace",
			expected: "/tmp/workspace",
		},
		{
			name:     "event.run_uuid",
			expr:     "event.run_uuid",
			expected: "run-456",
		},
		{
			name:     "event.workflow_name",
			expr:     "event.workflow_name",
			expected: "test-workflow",
		},
		{
			name:     "non-existent field",
			expr:     "event_data.nonexistent",
			expected: "",
		},
		{
			name:     "invalid prefix",
			expr:     "invalid.field",
			expected: "",
		},
		{
			name:     "no dot",
			expr:     "nodot",
			expected: "nodot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scheduler.resolveDotNotation(tt.expr, event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestScheduler_ResolveEventVars(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	event := &core.Event{
		Topic:  "assets.new",
		Source: "httpx",
		ParsedData: map[string]interface{}{
			"url":  "https://example.com",
			"type": "subdomain",
			"desc": "  test description  ",
		},
	}

	trigger := &core.Trigger{
		Name:    "vars-trigger",
		On:      core.TriggerEvent,
		Enabled: true,
		Event:   &core.EventConfig{Topic: "assets.new"},
		Input: core.TriggerInput{
			Vars: map[string]string{
				"target":     "event_data.url",
				"asset_type": "event_data.type",
				"source":     "event.source",
			},
		},
	}

	vars := scheduler.resolveEventVars(trigger, event)
	require.NotNil(t, vars)

	assert.Equal(t, "https://example.com", vars["target"])
	assert.Equal(t, "subdomain", vars["asset_type"])
	assert.Equal(t, "httpx", vars["source"])
}

func TestScheduler_ResolveEventVars_EmptyWhenNoVars(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	event := &core.Event{
		Topic:  "assets.new",
		Source: "httpx",
	}

	// Legacy syntax trigger (no Vars)
	trigger := &core.Trigger{
		Name:    "legacy-trigger",
		On:      core.TriggerEvent,
		Enabled: true,
		Event:   &core.EventConfig{Topic: "assets.new"},
		Input: core.TriggerInput{
			Type:  "event_data",
			Field: "url",
			Name:  "target",
		},
	}

	vars := scheduler.resolveEventVars(trigger, event)
	assert.Nil(t, vars)
}

func TestScheduler_ResolveVarExpression(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	event := &core.Event{
		Topic:  "assets.new",
		Source: "httpx",
		ParsedData: map[string]interface{}{
			"url":  "https://example.com",
			"desc": "  test description  ",
		},
	}

	tests := []struct {
		name     string
		expr     string
		expected string
	}{
		{
			name:     "simple dot notation",
			expr:     "event_data.url",
			expected: "https://example.com",
		},
		{
			name:     "event metadata",
			expr:     "event.source",
			expected: "httpx",
		},
		{
			name:     "function call - trim",
			expr:     "trim(event_data.desc)",
			expected: "test description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scheduler.resolveVarExpression(tt.expr, event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetNestedField(t *testing.T) {
	data := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"value": "deep-value",
			},
			"simple": "simple-value",
		},
		"top": "top-value",
	}

	tests := []struct {
		name     string
		path     []string
		expected interface{}
	}{
		{
			name:     "top level",
			path:     []string{"top"},
			expected: "top-value",
		},
		{
			name:     "nested one level",
			path:     []string{"level1", "simple"},
			expected: "simple-value",
		},
		{
			name:     "nested two levels",
			path:     []string{"level1", "level2", "value"},
			expected: "deep-value",
		},
		{
			name:     "non-existent",
			path:     []string{"nonexistent"},
			expected: nil,
		},
		{
			name:     "partial path non-existent",
			path:     []string{"level1", "nonexistent"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getNestedField(data, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestScheduler_EventInputVarsSyntax(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	err = scheduler.Start()
	require.NoError(t, err)
	defer func() { _ = scheduler.Stop() }()

	// Track received vars
	receivedVars := make(chan map[string]string, 1)
	scheduler.SetEventHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string, envelope string, resolvedVars map[string]string) error {
		receivedVars <- resolvedVars
		return nil
	})

	workflow := &core.Workflow{Name: "test", Kind: core.KindModule}

	// New Vars syntax trigger
	trigger := &core.Trigger{
		Name:    "vars-trigger",
		On:      core.TriggerEvent,
		Enabled: true,
		Event:   &core.EventConfig{Topic: "test.vars"},
		Input: core.TriggerInput{
			Vars: map[string]string{
				"target":     "event_data.url",
				"asset_type": "event_data.type",
				"source":     "event.source",
			},
		},
	}

	err = scheduler.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	// Emit event
	event := &core.Event{
		Topic:  "test.vars",
		Source: "test-source",
		ParsedData: map[string]interface{}{
			"url":  "https://example.com/api",
			"type": "endpoint",
		},
	}

	err = scheduler.EmitEvent(event)
	require.NoError(t, err)

	select {
	case vars := <-receivedVars:
		require.NotNil(t, vars)
		assert.Equal(t, "https://example.com/api", vars["target"])
		assert.Equal(t, "endpoint", vars["asset_type"])
		assert.Equal(t, "test-source", vars["source"])
	case <-time.After(2 * time.Second):
		t.Fatal("event trigger did not fire")
	}
}

func TestScheduler_EventInputLegacySyntax(t *testing.T) {
	scheduler, err := NewScheduler()
	require.NoError(t, err)

	err = scheduler.Start()
	require.NoError(t, err)
	defer func() { _ = scheduler.Stop() }()

	// Track received input (legacy style)
	receivedInput := make(chan string, 1)
	receivedVars := make(chan map[string]string, 1)
	scheduler.SetEventHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string, envelope string, resolvedVars map[string]string) error {
		receivedInput <- input
		receivedVars <- resolvedVars
		return nil
	})

	workflow := &core.Workflow{Name: "test", Kind: core.KindModule}

	// Legacy syntax trigger
	trigger := &core.Trigger{
		Name:    "legacy-trigger",
		On:      core.TriggerEvent,
		Enabled: true,
		Event:   &core.EventConfig{Topic: "test.legacy"},
		Input: core.TriggerInput{
			Type:  "event_data",
			Field: "url",
			Name:  "target",
		},
	}

	err = scheduler.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	// Emit event
	event := &core.Event{
		Topic:  "test.legacy",
		Source: "test-source",
		ParsedData: map[string]interface{}{
			"url": "https://legacy.example.com",
		},
	}

	err = scheduler.EmitEvent(event)
	require.NoError(t, err)

	select {
	case input := <-receivedInput:
		assert.Equal(t, "https://legacy.example.com", input)
	case <-time.After(2 * time.Second):
		t.Fatal("event trigger did not fire")
	}

	// Vars should be nil for legacy syntax
	select {
	case vars := <-receivedVars:
		assert.Nil(t, vars)
	case <-time.After(100 * time.Millisecond):
		// OK, might have already drained
	}
}
