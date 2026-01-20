package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventEmission tests that events can be emitted and basic event structure
func TestEventEmission(t *testing.T) {
	sched, err := scheduler.NewScheduler()
	require.NoError(t, err)

	err = sched.Start()
	require.NoError(t, err)
	defer func() { _ = sched.Stop() }()

	// Test emitting a simple event
	event := &core.Event{
		Topic:    "test.emission",
		Name:     "test-event",
		Source:   "e2e-test",
		DataType: "test-data",
		Data:     `{"key": "value"}`,
	}

	err = sched.EmitEvent(event)
	require.NoError(t, err)

	// Verify timestamp was set
	assert.False(t, event.Timestamp.IsZero())

	// Verify metrics
	enqueued, dropped := sched.GetEventMetrics()
	assert.Equal(t, int64(1), enqueued)
	assert.Equal(t, int64(0), dropped)
}

// TestEventTriggerWorkflow tests that events trigger registered workflows
func TestEventTriggerWorkflow(t *testing.T) {
	sched, err := scheduler.NewScheduler()
	require.NoError(t, err)

	err = sched.Start()
	require.NoError(t, err)
	defer func() { _ = sched.Stop() }()

	// Track triggered workflows
	triggered := make(chan *triggerResult, 10)

	sched.SetHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string) error {
		triggered <- &triggerResult{
			workflowName: w.Name,
			triggerName:  tr.Name,
			input:        input,
		}
		return nil
	})

	// Register a workflow with event trigger
	workflow := &core.Workflow{
		Name: "event-triggered-workflow",
		Kind: core.KindModule,
	}

	trigger := &core.Trigger{
		Name:    "on-test-event",
		On:      core.TriggerEvent,
		Enabled: true,
		Event: &core.EventConfig{
			Topic: "test.trigger",
		},
		Input: core.TriggerInput{
			Type:  "event_data",
			Field: "target",
			Name:  "target",
		},
	}

	err = sched.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	// Emit an event
	event := &core.Event{
		Topic:  "test.trigger",
		Source: "e2e-test",
		Data:   `{"target": "example.com"}`,
	}

	err = sched.EmitEvent(event)
	require.NoError(t, err)

	// Wait for trigger
	select {
	case result := <-triggered:
		assert.Equal(t, "event-triggered-workflow", result.workflowName)
		assert.Equal(t, "on-test-event", result.triggerName)
		assert.Equal(t, "example.com", result.input)
	case <-time.After(3 * time.Second):
		t.Fatal("workflow was not triggered by event")
	}
}

type triggerResult struct {
	workflowName string
	triggerName  string
	input        string
}

// TestEventFiltering tests that event filters work correctly
func TestEventFiltering(t *testing.T) {
	sched, err := scheduler.NewScheduler()
	require.NoError(t, err)

	err = sched.Start()
	require.NoError(t, err)
	defer func() { _ = sched.Stop() }()

	triggered := make(chan string, 10)

	sched.SetHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string) error {
		triggered <- tr.Name
		return nil
	})

	workflow := &core.Workflow{Name: "filter-test", Kind: core.KindModule}

	// Register trigger that only accepts events from 'nuclei' source
	trigger := &core.Trigger{
		Name:    "nuclei-only",
		On:      core.TriggerEvent,
		Enabled: true,
		Event: &core.EventConfig{
			Topic:   "vulnerabilities.new",
			Filters: []string{"event.source == 'nuclei'"},
		},
	}

	err = sched.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	// Emit event from different source - should NOT trigger
	_ = sched.EmitEvent(&core.Event{
		Topic:  "vulnerabilities.new",
		Source: "other-scanner",
	})

	// Short wait to ensure non-matching event doesn't trigger
	select {
	case <-triggered:
		t.Fatal("should not trigger for non-matching source")
	case <-time.After(200 * time.Millisecond):
		// Expected - no trigger
	}

	// Emit event from nuclei - should trigger
	_ = sched.EmitEvent(&core.Event{
		Topic:  "vulnerabilities.new",
		Source: "nuclei",
	})

	select {
	case name := <-triggered:
		assert.Equal(t, "nuclei-only", name)
	case <-time.After(2 * time.Second):
		t.Fatal("should trigger for matching source")
	}
}

// TestEventDataFiltering tests filtering on parsed event data
func TestEventDataFiltering(t *testing.T) {
	sched, err := scheduler.NewScheduler()
	require.NoError(t, err)

	err = sched.Start()
	require.NoError(t, err)
	defer func() { _ = sched.Stop() }()

	triggered := make(chan bool, 1)

	sched.SetHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string) error {
		triggered <- true
		return nil
	})

	workflow := &core.Workflow{Name: "data-filter-test", Kind: core.KindModule}

	// Register trigger that filters on parsed data fields
	trigger := &core.Trigger{
		Name:    "critical-only",
		On:      core.TriggerEvent,
		Enabled: true,
		Event: &core.EventConfig{
			Topic:   "vulnerabilities.new",
			Filters: []string{"event.data.severity == 'critical'"},
		},
	}

	err = sched.RegisterTrigger(workflow, trigger)
	require.NoError(t, err)

	// Emit low severity - should NOT trigger
	lowSeverityEvent := &core.Event{
		Topic:  "vulnerabilities.new",
		Source: "nuclei",
		Data:   `{"severity": "low", "template": "info-disclosure"}`,
	}
	_ = lowSeverityEvent.ParseData()
	_ = sched.EmitEvent(lowSeverityEvent)

	select {
	case <-triggered:
		t.Fatal("should not trigger for low severity")
	case <-time.After(200 * time.Millisecond):
		// Expected
	}

	// Emit critical severity - should trigger
	criticalEvent := &core.Event{
		Topic:  "vulnerabilities.new",
		Source: "nuclei",
		Data:   `{"severity": "critical", "template": "CVE-2024-1234"}`,
	}
	_ = criticalEvent.ParseData()
	_ = sched.EmitEvent(criticalEvent)

	select {
	case <-triggered:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("should trigger for critical severity")
	}
}

// TestEventChaining tests that workflows can chain via events
func TestEventChaining(t *testing.T) {
	sched, err := scheduler.NewScheduler()
	require.NoError(t, err)

	err = sched.Start()
	require.NoError(t, err)
	defer func() { _ = sched.Stop() }()

	executionOrder := make(chan string, 10)

	sched.SetHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string) error {
		executionOrder <- w.Name

		// Simulate workflow 1 emitting event for workflow 2
		if w.Name == "stage1" {
			// In real scenario, this would happen during workflow execution
			go func() {
				time.Sleep(50 * time.Millisecond)
				_ = sched.EmitEvent(&core.Event{
					Topic:  "stage1.complete",
					Source: "stage1",
					Data:   `{"next_target": "processed.example.com"}`,
				})
			}()
		}
		return nil
	})

	// Stage 1: triggered by initial event
	workflow1 := &core.Workflow{Name: "stage1", Kind: core.KindModule}
	trigger1 := &core.Trigger{
		Name:    "start-stage1",
		On:      core.TriggerEvent,
		Enabled: true,
		Event:   &core.EventConfig{Topic: "pipeline.start"},
	}

	// Stage 2: triggered by stage1 completion
	workflow2 := &core.Workflow{Name: "stage2", Kind: core.KindModule}
	trigger2 := &core.Trigger{
		Name:    "on-stage1-complete",
		On:      core.TriggerEvent,
		Enabled: true,
		Event:   &core.EventConfig{Topic: "stage1.complete"},
	}

	_ = sched.RegisterTrigger(workflow1, trigger1)
	_ = sched.RegisterTrigger(workflow2, trigger2)

	// Start the chain
	_ = sched.EmitEvent(&core.Event{Topic: "pipeline.start"})

	// Verify stage1 triggers first
	select {
	case name := <-executionOrder:
		assert.Equal(t, "stage1", name)
	case <-time.After(2 * time.Second):
		t.Fatal("stage1 didn't trigger")
	}

	// Verify stage2 triggers after stage1 emits completion event
	select {
	case name := <-executionOrder:
		assert.Equal(t, "stage2", name)
	case <-time.After(2 * time.Second):
		t.Fatal("stage2 didn't trigger after stage1")
	}
}

// TestMultipleTriggersPerWorkflow tests that a workflow can have multiple event triggers
func TestMultipleTriggersPerWorkflow(t *testing.T) {
	sched, err := scheduler.NewScheduler()
	require.NoError(t, err)

	err = sched.Start()
	require.NoError(t, err)
	defer func() { _ = sched.Stop() }()

	triggered := make(chan string, 10)

	sched.SetHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string) error {
		triggered <- tr.Name
		return nil
	})

	workflow := &core.Workflow{Name: "multi-trigger", Kind: core.KindModule}

	// Multiple triggers for different events
	triggers := []*core.Trigger{
		{
			Name:    "on-assets",
			On:      core.TriggerEvent,
			Enabled: true,
			Event:   &core.EventConfig{Topic: "assets.new"},
		},
		{
			Name:    "on-vulns",
			On:      core.TriggerEvent,
			Enabled: true,
			Event:   &core.EventConfig{Topic: "vulnerabilities.new"},
		},
	}

	for _, tr := range triggers {
		err = sched.RegisterTrigger(workflow, tr)
		require.NoError(t, err)
	}

	// Emit both types of events
	_ = sched.EmitEvent(&core.Event{Topic: "assets.new"})
	_ = sched.EmitEvent(&core.Event{Topic: "vulnerabilities.new"})

	triggeredNames := make(map[string]bool)
	for i := 0; i < 2; i++ {
		select {
		case name := <-triggered:
			triggeredNames[name] = true
		case <-time.After(2 * time.Second):
			t.Fatalf("expected 2 triggers, got %d", i)
		}
	}

	assert.True(t, triggeredNames["on-assets"])
	assert.True(t, triggeredNames["on-vulns"])
}

// TestEventParsedDataAccess tests accessing parsed JSON data in events
func TestEventParsedDataAccess(t *testing.T) {
	event := &core.Event{
		Topic:    "test.parsed",
		Source:   "test",
		Data:     `{"url": "https://example.com", "status_code": 200, "nested": {"key": "value"}}`,
		DataType: "http-response",
	}

	// Parse the data
	err := event.ParseData()
	require.NoError(t, err)

	// Access fields
	assert.Equal(t, "https://example.com", event.GetDataField("url"))
	assert.Equal(t, float64(200), event.GetDataField("status_code")) // JSON numbers are float64

	// Nested access
	nested := event.GetDataField("nested")
	require.NotNil(t, nested)
	nestedMap, ok := nested.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "value", nestedMap["key"])

	// Non-existent field
	assert.Nil(t, event.GetDataField("nonexistent"))
}

// TestEventQueueBackpressure tests queue behavior under load
func TestEventQueueBackpressure(t *testing.T) {
	// Create scheduler with small queue for testing
	sched, err := scheduler.NewSchedulerWithConfig(10, 100*time.Millisecond)
	require.NoError(t, err)

	err = sched.Start()
	require.NoError(t, err)
	defer func() { _ = sched.Stop() }()

	// Register a slow handler to cause backpressure
	sched.SetHandler(string(core.TriggerEvent), func(w *core.Workflow, tr *core.Trigger, input string) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	workflow := &core.Workflow{Name: "slow-handler", Kind: core.KindModule}
	trigger := &core.Trigger{
		Name:    "catch-all",
		On:      core.TriggerEvent,
		Enabled: true,
		Event:   &core.EventConfig{Topic: "test.backpressure"},
	}
	_ = sched.RegisterTrigger(workflow, trigger)

	// Emit many events quickly
	for i := 0; i < 20; i++ {
		_ = sched.EmitEvent(&core.Event{
			Topic: "test.backpressure",
			Name:  "flood-event",
		})
	}

	// Check metrics
	time.Sleep(500 * time.Millisecond)
	enqueued, dropped := sched.GetEventMetrics()

	// Some events should have been enqueued
	assert.Greater(t, enqueued, int64(0))
	// With small queue and fast emission, some may be dropped
	t.Logf("Events enqueued: %d, dropped: %d", enqueued, dropped)
}

// TestEventTriggeredWorkflowFile tests loading and validating event trigger workflow files
func TestEventTriggeredWorkflowFile(t *testing.T) {
	testdataPath := getTestdataPath(t)

	// Test event receiver workflow YAML structure
	receiverPath := filepath.Join(testdataPath, "test-event-receiver.yaml")
	data, err := os.ReadFile(receiverPath)
	require.NoError(t, err)

	// Verify YAML is valid (basic check)
	assert.Contains(t, string(data), "trigger:")
	assert.Contains(t, string(data), "on: event")
	assert.Contains(t, string(data), "topic:")
	assert.Contains(t, string(data), "filters:")
}

// TestEventStructureSerialization tests event JSON serialization
func TestEventStructureSerialization(t *testing.T) {
	event := &core.Event{
		Topic:     "test.serialization",
		ID:        "test-id-123",
		Name:      "test-event",
		Source:    "test-source",
		Data:      `{"key": "value"}`,
		DataType:  "json",
		Timestamp: time.Now(),
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(event)
	require.NoError(t, err)

	// Deserialize
	var decoded core.Event
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.Topic, decoded.Topic)
	assert.Equal(t, event.ID, decoded.ID)
	assert.Equal(t, event.Name, decoded.Name)
	assert.Equal(t, event.Source, decoded.Source)
	assert.Equal(t, event.Data, decoded.Data)
	assert.Equal(t, event.DataType, decoded.DataType)
}
