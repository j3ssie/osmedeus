package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/j3ssie/osmedeus/v5/internal/broker"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/database/repository"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// ServerEventClient sends events to the Osmedeus server's event receiver endpoint.
// It supports authentication via API key or JWT token.
type ServerEventClient struct {
	serverURL  string
	apiKey     string
	jwtSecret  string
	jwtExpiry  time.Duration
	httpClient *http.Client
}

// ServerEventRequest represents the request body for the /osm/api/events/emit endpoint.
type ServerEventRequest struct {
	Topic        string                 `json:"topic"`
	Name         string                 `json:"name,omitempty"`
	SourceType   string                 `json:"source_type,omitempty"` // "run", "eval", "api" - origin of the event
	Source       string                 `json:"source,omitempty"`
	DataType     string                 `json:"data_type,omitempty"`
	Workspace    string                 `json:"workspace,omitempty"`
	RunUUID      string                 `json:"run_uuid,omitempty"`
	WorkflowName string                 `json:"workflow_name,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

// NewServerEventClient creates a new server event client from the global config.
// Returns nil if no server URL is configured.
func NewServerEventClient() *ServerEventClient {
	cfg := config.Get()
	if cfg == nil {
		return nil
	}

	serverURL := cfg.Server.GetEventReceiverURL()
	if serverURL == "" {
		return nil
	}

	return &ServerEventClient{
		serverURL:  serverURL,
		apiKey:     cfg.Server.AuthAPIKey,
		jwtSecret:  cfg.Server.JWT.SecretSigningKey,
		jwtExpiry:  time.Duration(cfg.Server.JWT.ExpirationMinutes) * time.Minute,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// NewServerEventClientFromConfig creates a server event client from a specific config.
func NewServerEventClientFromConfig(cfg *config.Config) *ServerEventClient {
	if cfg == nil {
		return nil
	}

	serverURL := cfg.Server.GetEventReceiverURL()
	if serverURL == "" {
		return nil
	}

	return &ServerEventClient{
		serverURL:  serverURL,
		apiKey:     cfg.Server.AuthAPIKey,
		jwtSecret:  cfg.Server.JWT.SecretSigningKey,
		jwtExpiry:  time.Duration(cfg.Server.JWT.ExpirationMinutes) * time.Minute,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// SendEvent sends an event to the server's /osm/api/events/emit endpoint.
// Returns nil on success or an error if the request fails.
func (c *ServerEventClient) SendEvent(workspace, topic, source, dataType, runID, workflowName, sourceType string, data interface{}) error {
	if c == nil || c.serverURL == "" {
		return fmt.Errorf("server event client not configured")
	}

	// Build the request
	reqData := ServerEventRequest{
		Topic:        topic,
		SourceType:   sourceType,
		Source:       source,
		DataType:     dataType,
		Workspace:    workspace,
		RunUUID:      runID,
		WorkflowName: workflowName,
	}

	// Convert data to map[string]interface{}
	switch v := data.(type) {
	case map[string]interface{}:
		reqData.Data = v
	case string:
		reqData.Data = map[string]interface{}{"value": v}
	default:
		// Try to convert via JSON marshal/unmarshal
		jsonBytes, err := json.Marshal(data)
		if err == nil {
			var dataMap map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &dataMap); err == nil {
				reqData.Data = dataMap
			} else {
				reqData.Data = map[string]interface{}{"value": data}
			}
		} else {
			reqData.Data = map[string]interface{}{"value": fmt.Sprintf("%v", data)}
		}
	}

	body, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("failed to marshal event request: %w", err)
	}

	url := c.serverURL + "/osm/api/events/emit"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication
	if err := c.addAuth(req); err != nil {
		return fmt.Errorf("failed to add auth: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// addAuth adds authentication headers to the request.
// Prefers API key if set, otherwise generates a JWT token.
func (c *ServerEventClient) addAuth(req *http.Request) error {
	// Prefer API key authentication
	if c.apiKey != "" {
		req.Header.Set("x-osm-api-key", c.apiKey)
		return nil
	}

	// Fall back to JWT authentication
	if c.jwtSecret != "" && c.jwtSecret != "change-this-secret-in-production" {
		token, err := c.generateJWT()
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}

	// No auth configured - server might be running with --no-auth
	return nil
}

// generateJWT creates a JWT token for authentication.
func (c *ServerEventClient) generateJWT() (string, error) {
	expiry := c.jwtExpiry
	if expiry == 0 {
		expiry = 60 * time.Minute
	}

	claims := jwt.MapClaims{
		"username": "event-emitter",
		"exp":      time.Now().Add(expiry).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(c.jwtSecret))
}

// QueueEventToDatabase stores an event in the database for later processing.
// This is used as a fallback when the server is unavailable.
func QueueEventToDatabase(ctx context.Context, topic, source, dataType, sourceType string, data interface{}, errMsg string) error {
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

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

	eventLog := &database.EventLog{
		Topic:      topic,
		EventID:    uuid.New().String(),
		SourceType: sourceType,
		Source:     source,
		DataType:   dataType,
		Data:       dataJSON,
		Processed:  false,
		Error:      errMsg,
		CreatedAt:  time.Now(),
	}

	repo := repository.NewEventLogRepository(db)
	return repo.Create(ctx, eventLog)
}

// SendEventViaRedis publishes an event via Redis pub/sub for distributed mode.
func SendEventViaRedis(workspace, topic, source, dataType, runID, workflowName string, data interface{}) error {
	redisBroker, err := broker.GetSharedBroker()
	if err != nil {
		return fmt.Errorf("failed to get redis broker: %w", err)
	}
	if redisBroker == nil {
		return fmt.Errorf("redis broker not available")
	}

	ctx := context.Background()
	return redisBroker.PublishEventData(ctx, workspace, topic, source, dataType, runID, workflowName, data)
}

// SendEventWithFallback attempts to send an event to the server.
// If the server is unavailable, it queues the event to the database.
// It also sends the event to configured webhooks (existing behavior).
// In distributed mode with Redis configured, it uses Redis pub/sub.
// sourceType indicates the origin of the event: "run", "eval", or "api".
func SendEventWithFallback(workspace, topic, source, dataType, runID, workflowName, sourceType string, data interface{}) error {
	log := logger.Get()
	ctx := context.Background()
	cfg := config.Get()

	// In distributed mode, use Redis pub/sub
	if cfg != nil && cfg.IsRedisConfigured() {
		redisErr := SendEventViaRedis(workspace, topic, source, dataType, runID, workflowName, data)
		if redisErr == nil {
			log.Debug("Event published via Redis",
				zap.String("topic", topic),
				zap.String("source", source),
				zap.String("source_type", sourceType),
				zap.String("workspace", workspace),
				zap.String("run_id", runID),
				zap.String("workflow_name", workflowName),
			)
			// Also send to webhooks
			_ = SendStructuredEvent(topic, source, dataType, data)
			return nil
		}
		log.Debug("Failed to publish event via Redis, falling back to HTTP",
			zap.String("topic", topic),
			zap.Error(redisErr),
		)
		// Fall through to HTTP/DB fallback
	}

	// Try to send to server first
	client := NewServerEventClient()
	var serverErr error

	if client != nil {
		serverErr = client.SendEvent(workspace, topic, source, dataType, runID, workflowName, sourceType, data)
		if serverErr == nil {
			log.Debug("Event sent to server successfully",
				zap.String("topic", topic),
				zap.String("source", source),
				zap.String("source_type", sourceType),
			)
		} else {
			log.Debug("Failed to send event to server, queuing to database",
				zap.String("topic", topic),
				zap.Error(serverErr),
			)
			// Queue to database as fallback
			if queueErr := QueueEventToDatabase(ctx, topic, source, dataType, sourceType, data, serverErr.Error()); queueErr != nil {
				log.Warn("Failed to queue event to database",
					zap.String("topic", topic),
					zap.Error(queueErr),
				)
			}
		}
	} else {
		// No server configured, just queue to database
		if queueErr := QueueEventToDatabase(ctx, topic, source, dataType, sourceType, data, "server not configured"); queueErr != nil {
			log.Debug("Failed to queue event to database (no server configured)",
				zap.String("topic", topic),
				zap.Error(queueErr),
			)
		}
	}

	// Also send to webhooks (existing behavior)
	webhookErr := SendStructuredEvent(topic, source, dataType, data)
	if webhookErr != nil {
		// Log but don't fail - webhooks are optional
		log.Debug("Webhook notification skipped or failed",
			zap.String("topic", topic),
			zap.Error(webhookErr),
		)
	}

	// Return server error if any (webhook errors are not critical)
	return serverErr
}
