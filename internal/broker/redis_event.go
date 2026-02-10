// Package broker provides Redis-based event pub/sub for distributed mode.
// This package is designed to be a lightweight dependency that can be imported
// without creating import cycles.
package broker

import (
	"context"
	"fmt"
	"github.com/j3ssie/osmedeus/v5/internal/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/redis/rueidis"
)

// Redis key prefixes for events
const (
	KeyPrefix       = "osm:"
	KeyEventsPrefix = KeyPrefix + "events:" // osm:events:{topic}
)

// RedisEventBroker handles event publishing via Redis pub/sub
type RedisEventBroker struct {
	client rueidis.Client
	cfg    *config.RedisConfig
}

var (
	sharedBroker *RedisEventBroker
	brokerOnce   sync.Once
	brokerErr    error
)

// GetSharedBroker returns a singleton Redis event broker.
// Returns nil if Redis is not configured.
func GetSharedBroker() (*RedisEventBroker, error) {
	cfg := config.Get()
	if cfg == nil || !cfg.IsRedisConfigured() {
		return nil, nil
	}

	brokerOnce.Do(func() {
		sharedBroker, brokerErr = NewRedisEventBroker(&cfg.Redis)
	})

	return sharedBroker, brokerErr
}

// ResetSharedBroker resets the shared broker (useful for testing)
func ResetSharedBroker() {
	if sharedBroker != nil {
		sharedBroker.Close()
	}
	sharedBroker = nil
	brokerOnce = sync.Once{}
	brokerErr = nil
}

// NewRedisEventBroker creates a new Redis event broker
func NewRedisEventBroker(cfg *config.RedisConfig) (*RedisEventBroker, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("redis host not configured")
	}

	port := cfg.Port
	if port == 0 {
		port = 6379
	}

	opts := rueidis.ClientOption{
		InitAddress:  []string{fmt.Sprintf("%s:%d", cfg.Host, port)},
		Username:     cfg.Username,
		Password:     cfg.Password,
		SelectDB:     cfg.DB,
		DisableCache: true,
	}

	client, err := rueidis.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	return &RedisEventBroker{
		client: client,
		cfg:    cfg,
	}, nil
}

// Close closes the Redis client
func (b *RedisEventBroker) Close() {
	if b.client != nil {
		b.client.Close()
	}
}

// PublishEvent publishes an event to a topic channel
func (b *RedisEventBroker) PublishEvent(ctx context.Context, event *core.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	channel := KeyEventsPrefix + event.Topic
	cmd := b.client.B().Publish().Channel(channel).Message(string(data)).Build()
	return b.client.Do(ctx, cmd).Error()
}

// SubscribeEvents subscribes to event channels with pattern and calls handler for each event.
// This method blocks until the context is cancelled or an error occurs.
func (b *RedisEventBroker) SubscribeEvents(ctx context.Context, handler func(*core.Event)) error {
	pattern := KeyEventsPrefix + "*"

	err := b.client.Receive(ctx, b.client.B().Psubscribe().Pattern(pattern).Build(),
		func(msg rueidis.PubSubMessage) {
			if msg.Message != "" {
				var event core.Event
				if err := json.Unmarshal([]byte(msg.Message), &event); err == nil {
					handler(&event)
				}
			}
		})

	if err != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}

// PublishEventData is a convenience function that creates an event and publishes it.
func (b *RedisEventBroker) PublishEventData(ctx context.Context, workspace, topic, source, dataType, runID, workflowName string, data interface{}) error {
	// Convert data to JSON string
	var dataJSON string
	switch v := data.(type) {
	case string:
		dataJSON = v
	default:
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			dataJSON = fmt.Sprintf("%v", data)
		} else {
			dataJSON = string(jsonBytes)
		}
	}

	event := &core.Event{
		Topic:        topic,
		ID:           uuid.New().String(),
		Source:       source,
		DataType:     dataType,
		Workspace:    workspace,
		RunUUID:      runID,
		WorkflowName: workflowName,
		Data:         dataJSON,
		Timestamp:    time.Now(),
	}

	return b.PublishEvent(ctx, event)
}

// TryPublishEvent attempts to publish an event via Redis.
// Returns true if the event was published successfully, false otherwise.
// This is a convenience function that doesn't require pre-checking Redis configuration.
func TryPublishEvent(workspace, topic, source, dataType, runID, workflowName string, data interface{}) (bool, error) {
	broker, err := GetSharedBroker()
	if err != nil {
		return false, err
	}
	if broker == nil {
		return false, nil // Redis not configured
	}

	ctx := context.Background()
	if err := broker.PublishEventData(ctx, workspace, topic, source, dataType, runID, workflowName, data); err != nil {
		return false, err
	}
	return true, nil
}
