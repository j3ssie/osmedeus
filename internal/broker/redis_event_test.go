package broker

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// MockRedisEventBroker is a mock implementation for testing
type MockRedisEventBroker struct {
	mu              sync.Mutex
	publishedEvents []*core.Event
	subscribers     []func(*core.Event)
}

// NewMockBroker creates a mock broker for testing
func NewMockBroker() *MockRedisEventBroker {
	return &MockRedisEventBroker{
		publishedEvents: make([]*core.Event, 0),
		subscribers:     make([]func(*core.Event), 0),
	}
}

// PublishEvent stores the event and notifies subscribers
func (m *MockRedisEventBroker) PublishEvent(ctx context.Context, event *core.Event) error {
	m.mu.Lock()
	m.publishedEvents = append(m.publishedEvents, event)
	subs := make([]func(*core.Event), len(m.subscribers))
	copy(subs, m.subscribers)
	m.mu.Unlock()

	// Notify subscribers
	for _, sub := range subs {
		sub(event)
	}
	return nil
}

// Subscribe adds a subscriber
func (m *MockRedisEventBroker) Subscribe(handler func(*core.Event)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribers = append(m.subscribers, handler)
}

// GetPublishedEvents returns all published events
func (m *MockRedisEventBroker) GetPublishedEvents() []*core.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]*core.Event, len(m.publishedEvents))
	copy(result, m.publishedEvents)
	return result
}

// Reset clears all events and subscribers
func (m *MockRedisEventBroker) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedEvents = m.publishedEvents[:0]
	m.subscribers = m.subscribers[:0]
}

func TestMockBrokerPublishAndSubscribe(t *testing.T) {
	broker := NewMockBroker()
	ctx := context.Background()

	// Subscribe before publishing
	var received *core.Event
	var wg sync.WaitGroup
	wg.Add(1)
	broker.Subscribe(func(e *core.Event) {
		received = e
		wg.Done()
	})

	// Publish an event
	event := &core.Event{
		Topic:     "test.topic",
		ID:        "test-id-123",
		Source:    "unit-test",
		DataType:  "test-data",
		Data:      `{"key": "value"}`,
		Timestamp: time.Now(),
	}

	err := broker.PublishEvent(ctx, event)
	if err != nil {
		t.Fatalf("PublishEvent failed: %v", err)
	}

	// Wait for subscriber to receive
	wg.Wait()

	// Verify event was received
	if received == nil {
		t.Fatal("Event not received by subscriber")
	}
	if received.Topic != event.Topic {
		t.Errorf("Expected topic %s, got %s", event.Topic, received.Topic)
	}
	if received.ID != event.ID {
		t.Errorf("Expected ID %s, got %s", event.ID, received.ID)
	}

	// Verify event was stored
	published := broker.GetPublishedEvents()
	if len(published) != 1 {
		t.Errorf("Expected 1 published event, got %d", len(published))
	}
}

func TestEventSerialization(t *testing.T) {
	event := &core.Event{
		Topic:     "scan.completed",
		ID:        "event-456",
		Name:      "vulnerability.found",
		Source:    "nuclei",
		DataType:  "vulnerability",
		Data:      `{"severity": "high", "template": "cve-2024-1234"}`,
		Timestamp: time.Now().Truncate(time.Second), // Truncate for comparison
	}

	// Serialize
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	// Deserialize
	var decoded core.Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	// Verify fields
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

func TestTryPublishEventWithoutRedis(t *testing.T) {
	// Reset shared broker to ensure clean state
	ResetSharedBroker()

	// Try to publish without Redis configured
	published, err := TryPublishEvent("test-workspace", "test.topic", "test", "data", "test-run-id", "test-workflow", map[string]string{"key": "value"})

	// Should return false (not published) since Redis isn't configured
	if published {
		t.Error("Expected TryPublishEvent to return false when Redis not configured")
	}
	if err != nil {
		t.Errorf("Expected no error when Redis not configured, got: %v", err)
	}
}

func TestMultipleSubscribers(t *testing.T) {
	broker := NewMockBroker()
	ctx := context.Background()

	// Create multiple subscribers
	var wg sync.WaitGroup
	received := make([]*core.Event, 3)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		idx := i
		broker.Subscribe(func(e *core.Event) {
			received[idx] = e
			wg.Done()
		})
	}

	// Publish an event
	event := &core.Event{
		Topic:  "multi.test",
		ID:     "multi-id",
		Source: "test",
	}
	_ = broker.PublishEvent(ctx, event)

	// Wait for all subscribers
	wg.Wait()

	// Verify all received
	for i, e := range received {
		if e == nil {
			t.Errorf("Subscriber %d did not receive event", i)
		} else if e.Topic != event.Topic {
			t.Errorf("Subscriber %d received wrong topic: %s", i, e.Topic)
		}
	}
}

func TestBrokerReset(t *testing.T) {
	broker := NewMockBroker()
	ctx := context.Background()

	// Publish some events
	for i := 0; i < 5; i++ {
		_ = broker.PublishEvent(ctx, &core.Event{Topic: "reset.test"})
	}

	// Add a subscriber
	broker.Subscribe(func(e *core.Event) {})

	// Verify we have events
	if len(broker.GetPublishedEvents()) != 5 {
		t.Error("Expected 5 events before reset")
	}

	// Reset
	broker.Reset()

	// Verify cleared
	if len(broker.GetPublishedEvents()) != 0 {
		t.Error("Expected 0 events after reset")
	}
}
