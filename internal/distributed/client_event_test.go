package distributed

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

func TestDataEnvelopeSerialization(t *testing.T) {
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	dataBytes, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	envelope := DataEnvelope{
		Type:      "test-type",
		Data:      dataBytes,
		Timestamp: time.Now().Truncate(time.Second),
		WorkerID:  "worker-123",
	}

	// Serialize
	serialized, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal envelope: %v", err)
	}

	// Deserialize
	var decoded DataEnvelope
	if err := json.Unmarshal(serialized, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}

	// Verify fields
	if decoded.Type != envelope.Type {
		t.Errorf("Type mismatch: %s vs %s", decoded.Type, envelope.Type)
	}
	if decoded.WorkerID != envelope.WorkerID {
		t.Errorf("WorkerID mismatch: %s vs %s", decoded.WorkerID, envelope.WorkerID)
	}
	if !decoded.Timestamp.Equal(envelope.Timestamp) {
		t.Errorf("Timestamp mismatch: %v vs %v", decoded.Timestamp, envelope.Timestamp)
	}

	// Verify nested data can be decoded
	var decodedData map[string]string
	if err := json.Unmarshal(decoded.Data, &decodedData); err != nil {
		t.Fatalf("Failed to unmarshal nested data: %v", err)
	}
	if decodedData["key1"] != "value1" {
		t.Errorf("Nested data mismatch: expected value1, got %s", decodedData["key1"])
	}
}

func TestEventSerialization(t *testing.T) {
	event := &core.Event{
		Topic:     "test.event",
		ID:        "event-id-123",
		Name:      "test.name",
		Source:    "test-source",
		DataType:  "test-data-type",
		Data:      `{"nested": "value"}`,
		Timestamp: time.Now().Truncate(time.Second),
	}

	// Serialize
	serialized, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	// Deserialize
	var decoded core.Event
	if err := json.Unmarshal(serialized, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	// Verify all fields
	if decoded.Topic != event.Topic {
		t.Errorf("Topic mismatch: %s vs %s", decoded.Topic, event.Topic)
	}
	if decoded.ID != event.ID {
		t.Errorf("ID mismatch: %s vs %s", decoded.ID, event.ID)
	}
	if decoded.Name != event.Name {
		t.Errorf("Name mismatch: %s vs %s", decoded.Name, event.Name)
	}
	if decoded.Source != event.Source {
		t.Errorf("Source mismatch: %s vs %s", decoded.Source, event.Source)
	}
	if decoded.DataType != event.DataType {
		t.Errorf("DataType mismatch: %s vs %s", decoded.DataType, event.DataType)
	}
	if decoded.Data != event.Data {
		t.Errorf("Data mismatch: %s vs %s", decoded.Data, event.Data)
	}
}

func TestKeyConstants(t *testing.T) {
	// Verify key prefixes are properly formatted
	expectedPrefix := "osm:"

	keys := map[string]string{
		"KeyPrefix":        KeyPrefix,
		"KeyEventsPrefix":  KeyEventsPrefix,
		"KeyDataRuns":      KeyDataRuns,
		"KeyDataSteps":     KeyDataSteps,
		"KeyDataEvents":    KeyDataEvents,
		"KeyDataArtifacts": KeyDataArtifacts,
	}

	for name, key := range keys {
		if len(key) < len(expectedPrefix) {
			t.Errorf("%s is too short: %s", name, key)
			continue
		}
		if key[:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("%s doesn't start with %s: %s", name, expectedPrefix, key)
		}
	}

	// Verify specific keys
	if KeyEventsPrefix != "osm:events:" {
		t.Errorf("KeyEventsPrefix mismatch: expected osm:events:, got %s", KeyEventsPrefix)
	}
	if KeyDataRuns != "osm:data:runs" {
		t.Errorf("KeyDataRuns mismatch: expected osm:data:runs, got %s", KeyDataRuns)
	}
	if KeyDataSteps != "osm:data:steps" {
		t.Errorf("KeyDataSteps mismatch: expected osm:data:steps, got %s", KeyDataSteps)
	}
	if KeyDataEvents != "osm:data:events" {
		t.Errorf("KeyDataEvents mismatch: expected osm:data:events, got %s", KeyDataEvents)
	}
	if KeyDataArtifacts != "osm:data:artifacts" {
		t.Errorf("KeyDataArtifacts mismatch: expected osm:data:artifacts, got %s", KeyDataArtifacts)
	}
}

func TestTaskStatusConstants(t *testing.T) {
	// Verify task status constants
	if TaskStatusPending != "pending" {
		t.Errorf("TaskStatusPending mismatch: expected pending, got %s", TaskStatusPending)
	}
	if TaskStatusRunning != "running" {
		t.Errorf("TaskStatusRunning mismatch: expected running, got %s", TaskStatusRunning)
	}
	if TaskStatusCompleted != "completed" {
		t.Errorf("TaskStatusCompleted mismatch: expected completed, got %s", TaskStatusCompleted)
	}
	if TaskStatusFailed != "failed" {
		t.Errorf("TaskStatusFailed mismatch: expected failed, got %s", TaskStatusFailed)
	}
}

func TestNewTask(t *testing.T) {
	task := NewTask("task-123", "test-workflow", "module", "example.com", map[string]interface{}{
		"param1": "value1",
	})

	if task.ID != "task-123" {
		t.Errorf("ID mismatch: expected task-123, got %s", task.ID)
	}
	if task.WorkflowName != "test-workflow" {
		t.Errorf("WorkflowName mismatch: expected test-workflow, got %s", task.WorkflowName)
	}
	if task.WorkflowKind != "module" {
		t.Errorf("WorkflowKind mismatch: expected module, got %s", task.WorkflowKind)
	}
	if task.Target != "example.com" {
		t.Errorf("Target mismatch: expected example.com, got %s", task.Target)
	}
	if task.Status != TaskStatusPending {
		t.Errorf("Status mismatch: expected pending, got %s", task.Status)
	}
	if task.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestTaskMarkRunning(t *testing.T) {
	task := NewTask("task-123", "test-workflow", "module", "example.com", nil)

	task.MarkRunning("worker-456")

	if task.Status != TaskStatusRunning {
		t.Errorf("Status mismatch: expected running, got %s", task.Status)
	}
	if task.WorkerID != "worker-456" {
		t.Errorf("WorkerID mismatch: expected worker-456, got %s", task.WorkerID)
	}
	if task.StartedAt == nil {
		t.Error("StartedAt should not be nil")
	}
}

func TestTaskMarkCompleted(t *testing.T) {
	task := NewTask("task-123", "test-workflow", "module", "example.com", nil)

	task.MarkCompleted()

	if task.Status != TaskStatusCompleted {
		t.Errorf("Status mismatch: expected completed, got %s", task.Status)
	}
	if task.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}
}

func TestTaskMarkFailed(t *testing.T) {
	task := NewTask("task-123", "test-workflow", "module", "example.com", nil)

	task.MarkFailed("something went wrong")

	if task.Status != TaskStatusFailed {
		t.Errorf("Status mismatch: expected failed, got %s", task.Status)
	}
	if task.Error != "something went wrong" {
		t.Errorf("Error mismatch: expected 'something went wrong', got %s", task.Error)
	}
	if task.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}
}

func TestTaskResultSerialization(t *testing.T) {
	result := &TaskResult{
		TaskID:      "task-123",
		Status:      TaskStatusCompleted,
		Output:      "task output",
		Exports:     map[string]interface{}{"key": "value"},
		CompletedAt: time.Now().Truncate(time.Second),
	}

	// Serialize
	data, err := result.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	// Deserialize
	decoded, err := UnmarshalTaskResult(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if decoded.TaskID != result.TaskID {
		t.Errorf("TaskID mismatch: %s vs %s", decoded.TaskID, result.TaskID)
	}
	if decoded.Status != result.Status {
		t.Errorf("Status mismatch: %s vs %s", decoded.Status, result.Status)
	}
	if decoded.Output != result.Output {
		t.Errorf("Output mismatch: %s vs %s", decoded.Output, result.Output)
	}
}

func TestWorkerInfoSerialization(t *testing.T) {
	info := &WorkerInfo{
		ID:            "worker-123",
		Hostname:      "test-host",
		Status:        "idle",
		CurrentTaskID: "",
		JoinedAt:      time.Now().Truncate(time.Second),
		LastHeartbeat: time.Now().Truncate(time.Second),
		TasksComplete: 5,
		TasksFailed:   1,
	}

	// Serialize
	data, err := info.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal worker info: %v", err)
	}

	// Deserialize
	decoded, err := UnmarshalWorkerInfo(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal worker info: %v", err)
	}

	if decoded.ID != info.ID {
		t.Errorf("ID mismatch: %s vs %s", decoded.ID, info.ID)
	}
	if decoded.Hostname != info.Hostname {
		t.Errorf("Hostname mismatch: %s vs %s", decoded.Hostname, info.Hostname)
	}
	if decoded.Status != info.Status {
		t.Errorf("Status mismatch: %s vs %s", decoded.Status, info.Status)
	}
	if decoded.TasksComplete != info.TasksComplete {
		t.Errorf("TasksComplete mismatch: %d vs %d", decoded.TasksComplete, info.TasksComplete)
	}
	if decoded.TasksFailed != info.TasksFailed {
		t.Errorf("TasksFailed mismatch: %d vs %d", decoded.TasksFailed, info.TasksFailed)
	}
}
