package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// EventReceiverProvider is an interface for event receiver operations.
// This allows handlers to work with the event receiver without importing
// the server package (avoiding circular dependencies).
type EventReceiverProvider interface {
	IsRunning() bool
	GetRegisteredWorkflows() []*core.Workflow
	GetRegisteredTriggersInfo() []*EventReceiverTriggerInfo
	EmitEvent(event *core.Event) error
}

// EventReceiverTriggerInfo holds workflow and trigger info for API responses.
type EventReceiverTriggerInfo struct {
	WorkflowName string
	TriggerName  string
	Type         string // "event", "cron", "watch", "manual"
	Topic        string
	Enabled      bool
}

// EventReceiverStatusResponse represents the event receiver status
type EventReceiverStatusResponse struct {
	Enabled        bool `json:"enabled"`
	Running        bool `json:"running"`
	WorkflowCount  int  `json:"workflow_count"`
	TriggerCount   int  `json:"trigger_count"`
}

// EventReceiverWorkflowResponse represents a registered event-triggered workflow
type EventReceiverWorkflowResponse struct {
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`
	Description string   `json:"description"`
	Triggers    []string `json:"triggers"`
}

// EventReceiverWorkflowsResponse represents the list of registered workflows
type EventReceiverWorkflowsResponse struct {
	Workflows []EventReceiverWorkflowResponse `json:"workflows"`
	Triggers  []EventReceiverTriggerResponse  `json:"triggers"`
}

// EventReceiverTriggerResponse represents a registered event trigger
type EventReceiverTriggerResponse struct {
	WorkflowName string `json:"workflow_name"`
	TriggerName  string `json:"trigger_name"`
	Topic        string `json:"topic"`
	Enabled      bool   `json:"enabled"`
}

// EmitEventRequest represents the request body for emitting events
type EmitEventRequest struct {
	Topic    string                 `json:"topic"`
	Name     string                 `json:"name,omitempty"`
	Source   string                 `json:"source,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
	DataType string                 `json:"data_type,omitempty"`
}

// EmitEventResponse represents the response for emitting events
type EmitEventResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	EventID string `json:"event_id,omitempty"`
}

// GetEventReceiverStatus returns the status of the event receiver
// @Summary Get event receiver status
// @Description Returns the status of the event receiver including enabled state and counts
// @Tags event-receiver
// @Accept json
// @Produce json
// @Success 200 {object} EventReceiverStatusResponse
// @Router /osm/api/event-receiver/status [get]
func GetEventReceiverStatus(provider EventReceiverProvider) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if provider == nil {
			return c.JSON(EventReceiverStatusResponse{
				Enabled:       false,
				Running:       false,
				WorkflowCount: 0,
				TriggerCount:  0,
			})
		}

		workflows := provider.GetRegisteredWorkflows()
		triggers := provider.GetRegisteredTriggersInfo()

		return c.JSON(EventReceiverStatusResponse{
			Enabled:       true,
			Running:       provider.IsRunning(),
			WorkflowCount: len(workflows),
			TriggerCount:  len(triggers),
		})
	}
}

// ListEventReceiverWorkflows lists all registered event-triggered workflows
// @Summary List event receiver workflows
// @Description Returns all workflows registered with the event receiver
// @Tags event-receiver
// @Accept json
// @Produce json
// @Success 200 {object} EventReceiverWorkflowsResponse
// @Router /osm/api/event-receiver/workflows [get]
func ListEventReceiverWorkflows(provider EventReceiverProvider) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if provider == nil {
			return c.JSON(EventReceiverWorkflowsResponse{
				Workflows: []EventReceiverWorkflowResponse{},
				Triggers:  []EventReceiverTriggerResponse{},
			})
		}

		workflows := provider.GetRegisteredWorkflows()
		triggers := provider.GetRegisteredTriggersInfo()

		// Build workflow responses
		workflowResponses := make([]EventReceiverWorkflowResponse, 0, len(workflows))
		for _, wf := range workflows {
			triggerNames := make([]string, 0)
			for _, t := range wf.GetEventTriggers() {
				triggerNames = append(triggerNames, t.Name)
			}

			workflowResponses = append(workflowResponses, EventReceiverWorkflowResponse{
				Name:        wf.Name,
				Kind:        string(wf.Kind),
				Description: wf.Description,
				Triggers:    triggerNames,
			})
		}

		// Build trigger responses
		triggerResponses := make([]EventReceiverTriggerResponse, 0, len(triggers))
		for _, t := range triggers {
			triggerResponses = append(triggerResponses, EventReceiverTriggerResponse{
				WorkflowName: t.WorkflowName,
				TriggerName:  t.TriggerName,
				Topic:        t.Topic,
				Enabled:      t.Enabled,
			})
		}

		return c.JSON(EventReceiverWorkflowsResponse{
			Workflows: workflowResponses,
			Triggers:  triggerResponses,
		})
	}
}

// EmitEvent emits an event to trigger matching workflows
// @Summary Emit an event
// @Description Emits an event that can trigger event-triggered workflows
// @Tags event-receiver
// @Accept json
// @Produce json
// @Param request body EmitEventRequest true "Event to emit"
// @Success 200 {object} EmitEventResponse
// @Failure 400 {object} map[string]interface{}
// @Router /osm/api/events/emit [post]
func EmitEvent(provider EventReceiverProvider) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if provider == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":   true,
				"message": "Event receiver is not enabled",
			})
		}

		if !provider.IsRunning() {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":   true,
				"message": "Event receiver is not running",
			})
		}

		var req EmitEventRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid request body: " + err.Error(),
			})
		}

		if req.Topic == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Topic is required",
			})
		}

		// Create event
		event := &core.Event{
			Topic:      req.Topic,
			Name:       req.Name,
			Source:     req.Source,
			DataType:   req.DataType,
			ParsedData: req.Data,
			Timestamp:  time.Now(),
		}

		// Emit the event
		if err := provider.EmitEvent(event); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to emit event: " + err.Error(),
			})
		}

		return c.JSON(EmitEventResponse{
			Success: true,
			Message: "Event emitted successfully",
			EventID: event.ID,
		})
	}
}
