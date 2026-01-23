package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/database/repository"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
	oslogger "github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/j3ssie/osmedeus/v5/internal/scheduler"
	"github.com/j3ssie/osmedeus/v5/pkg/server/handlers"
	"go.uber.org/zap"
)

// EventReceiver manages event-triggered workflows in server mode.
// It loads all workflows with event triggers at startup and executes
// them when matching events are received through the scheduler.
type EventReceiver struct {
	scheduler *scheduler.Scheduler
	loader    *parser.Loader
	config    *config.Config
	logger    *zap.Logger
	workflows []*core.Workflow
	triggers  []*RegisteredEventTrigger
	mu        sync.RWMutex
	running   bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// RegisteredEventTrigger holds workflow and trigger info for event-triggered workflows
type RegisteredEventTrigger struct {
	Workflow *core.Workflow
	Trigger  *core.Trigger
}

// NewEventReceiver creates a new event receiver for the server.
// It initializes the scheduler and workflow loader but does not start
// listening for events until Start() is called.
func NewEventReceiver(cfg *config.Config) (*EventReceiver, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	sched, err := scheduler.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	loader := parser.NewLoader(cfg.WorkflowsPath)

	return &EventReceiver{
		scheduler: sched,
		loader:    loader,
		config:    cfg,
		logger:    oslogger.Get(),
		workflows: make([]*core.Workflow, 0),
		triggers:  make([]*RegisteredEventTrigger, 0),
	}, nil
}

// Start loads all event-triggered workflows and begins listening for events.
// It scans the workflows directory, filters to only workflows with enabled
// event triggers, and registers them with the scheduler.
func (er *EventReceiver) Start(ctx context.Context) error {
	er.mu.Lock()
	defer er.mu.Unlock()

	if er.running {
		return nil
	}

	// Create cancellable context for workflow executions
	er.ctx, er.cancel = context.WithCancel(ctx)

	// Load all workflows
	allWorkflows, err := er.loader.LoadAllWorkflows()
	if err != nil {
		er.logger.Warn("Failed to load some workflows", zap.Error(err))
		// Continue with partial results
	}

	// Filter to only workflows with event triggers
	var workflowCount, triggerCount int
	for _, wf := range allWorkflows {
		eventTriggers := wf.GetEventTriggers()
		if len(eventTriggers) == 0 {
			continue
		}

		er.workflows = append(er.workflows, wf)
		workflowCount++

		// Register each event trigger
		for i := range eventTriggers {
			trigger := &eventTriggers[i]
			if err := er.scheduler.RegisterTrigger(wf, trigger); err != nil {
				er.logger.Warn("Failed to register event trigger",
					zap.String("workflow", wf.Name),
					zap.String("trigger", trigger.Name),
					zap.Error(err),
				)
				continue
			}

			er.triggers = append(er.triggers, &RegisteredEventTrigger{
				Workflow: wf,
				Trigger:  trigger,
			})
			triggerCount++

			er.logger.Debug("Registered event trigger",
				zap.String("workflow", wf.Name),
				zap.String("trigger", trigger.Name),
				zap.String("topic", trigger.Event.Topic),
			)
		}
	}

	// Set handler for event triggers (use SetEventHandler to receive full event envelope)
	er.scheduler.SetEventHandler(string(core.TriggerEvent), er.handleEventTriggerWithEnvelope)

	// Set handler for cron triggers
	er.scheduler.SetHandler(string(core.TriggerCron), er.handleCronTrigger)

	// Set handler for watch triggers
	er.scheduler.SetHandler(string(core.TriggerWatch), er.handleWatchTrigger)

	// Start the scheduler
	if err := er.scheduler.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	er.running = true

	er.logger.Debug("Event receiver started",
		zap.Int("workflows", workflowCount),
		zap.Int("triggers", triggerCount),
	)

	// Process any queued events from the database asynchronously
	go er.processQueuedEvents()

	return nil
}

// Stop gracefully stops the event receiver and its scheduler.
func (er *EventReceiver) Stop() error {
	er.mu.Lock()
	defer er.mu.Unlock()

	if !er.running {
		return nil
	}

	// Cancel any running workflow executions
	if er.cancel != nil {
		er.cancel()
	}

	// Stop the scheduler
	if err := er.scheduler.Stop(); err != nil {
		er.logger.Warn("Error stopping scheduler", zap.Error(err))
	}

	er.running = false
	er.logger.Info("Event receiver stopped")

	return nil
}

// handleEventTriggerWithEnvelope is called when an event matches a trigger.
// It receives the full event envelope JSON and resolved variables (from the new Vars syntax).
func (er *EventReceiver) handleEventTriggerWithEnvelope(workflow *core.Workflow, trigger *core.Trigger, input string, eventEnvelope string, resolvedVars map[string]string) error {
	er.logger.Info("Event trigger fired",
		zap.String("workflow", workflow.Name),
		zap.String("trigger", trigger.Name),
		zap.String("input", input),
	)
	er.logger.Debug("Event envelope received",
		zap.String("event_envelope", eventEnvelope),
	)
	if resolvedVars != nil {
		er.logger.Debug("Resolved vars from Vars syntax",
			zap.Any("resolved_vars", resolvedVars),
		)
	}

	// Execute workflow in goroutine to not block the scheduler
	go func() {
		// Build params from trigger input
		params := make(map[string]string)

		// If this trigger is linked to a Schedule record, look it up to get Target/Params first
		if trigger.ScheduleID != "" {
			schedule, err := database.GetScheduleByID(er.ctx, trigger.ScheduleID)
			if err != nil {
				er.logger.Warn("Failed to load schedule for event trigger",
					zap.String("schedule_id", trigger.ScheduleID),
					zap.Error(err),
				)
			} else {
				// Skip execution if schedule is disabled
				if !schedule.IsEnabled {
					er.logger.Info("Skipping disabled schedule",
						zap.String("schedule_id", trigger.ScheduleID),
						zap.String("workflow", workflow.Name),
					)
					return
				}
				// Use target from schedule (can be overridden by event input)
				if schedule.Target != "" {
					params["target"] = schedule.Target
				}
				// Use params from schedule
				if schedule.Params != nil {
					for k, v := range schedule.Params {
						if s, ok := v.(string); ok {
							params[k] = s
						}
					}
				}
			}
		}

		// New Vars syntax: set all resolved variables (overrides schedule params)
		if resolvedVars != nil {
			for name, value := range resolvedVars {
				params[name] = value
			}
		} else {
			// Legacy syntax: set the input parameter using the trigger's input.name field
			if trigger.Input.Name != "" {
				params[trigger.Input.Name] = input
			}

			// Default to "target" if no input name specified
			if input != "" && trigger.Input.Name == "" {
				params["target"] = input
			}
		}

		// Pass the event envelope as a special parameter
		if eventEnvelope != "" {
			params[scheduler.EventEnvelopeKey] = eventEnvelope
		}

		// Create executor with server mode
		exec := executor.NewExecutor()
		exec.SetServerMode(true)
		exec.SetSchedulerInvocation(true) // Mark as scheduler invocation
		exec.SetLoader(er.loader)

		var err error
		var result *core.WorkflowResult

		if workflow.IsFlow() {
			result, err = exec.ExecuteFlow(er.ctx, workflow, params, er.config)
		} else {
			result, err = exec.ExecuteModule(er.ctx, workflow, params, er.config)
		}

		if err != nil {
			er.logger.Error("Event-triggered workflow failed",
				zap.String("workflow", workflow.Name),
				zap.String("trigger", trigger.Name),
				zap.Error(err),
			)
			return
		}

		er.logger.Info("Event-triggered workflow completed",
			zap.String("workflow", workflow.Name),
			zap.String("status", string(result.Status)),
			zap.Duration("duration", result.EndTime.Sub(result.StartTime)),
		)
	}()

	return nil
}

// handleCronTrigger is called when a cron trigger fires.
func (er *EventReceiver) handleCronTrigger(workflow *core.Workflow, trigger *core.Trigger, input string) error {
	er.logger.Info("Cron trigger fired",
		zap.String("workflow", workflow.Name),
		zap.String("trigger", trigger.Name),
	)

	go func() {
		params := make(map[string]string)

		// If this trigger is linked to a Schedule record, look it up to get Target/Params
		if trigger.ScheduleID != "" {
			schedule, err := database.GetScheduleByID(er.ctx, trigger.ScheduleID)
			if err != nil {
				er.logger.Warn("Failed to load schedule for cron trigger",
					zap.String("schedule_id", trigger.ScheduleID),
					zap.Error(err),
				)
			} else {
				// Skip execution if schedule is disabled
				if !schedule.IsEnabled {
					er.logger.Info("Skipping disabled schedule",
						zap.String("schedule_id", trigger.ScheduleID),
						zap.String("workflow", workflow.Name),
					)
					return
				}
				// Use target from schedule
				if schedule.Target != "" {
					params["target"] = schedule.Target
				}
				// Use params from schedule
				if schedule.Params != nil {
					for k, v := range schedule.Params {
						if s, ok := v.(string); ok {
							params[k] = s
						}
					}
				}
				er.logger.Debug("Loaded schedule params for cron trigger",
					zap.String("schedule_id", trigger.ScheduleID),
					zap.String("target", schedule.Target),
					zap.Int("param_count", len(schedule.Params)),
				)
			}
		}

		// Set input parameter if specified (can override schedule params)
		if trigger.Input.Name != "" && input != "" {
			params[trigger.Input.Name] = input
		}

		exec := executor.NewExecutor()
		exec.SetServerMode(true)
		exec.SetSchedulerInvocation(true)
		exec.SetLoader(er.loader)

		var err error
		var result *core.WorkflowResult

		if workflow.IsFlow() {
			result, err = exec.ExecuteFlow(er.ctx, workflow, params, er.config)
		} else {
			result, err = exec.ExecuteModule(er.ctx, workflow, params, er.config)
		}

		if err != nil {
			er.logger.Error("Cron-triggered workflow failed",
				zap.String("workflow", workflow.Name),
				zap.String("trigger", trigger.Name),
				zap.Error(err),
			)
			return
		}

		er.logger.Info("Cron-triggered workflow completed",
			zap.String("workflow", workflow.Name),
			zap.String("status", string(result.Status)),
			zap.Duration("duration", result.EndTime.Sub(result.StartTime)),
		)
	}()

	return nil
}

// handleWatchTrigger is called when a file watch trigger fires.
func (er *EventReceiver) handleWatchTrigger(workflow *core.Workflow, trigger *core.Trigger, input string) error {
	er.logger.Info("Watch trigger fired",
		zap.String("workflow", workflow.Name),
		zap.String("trigger", trigger.Name),
		zap.String("path", trigger.Path),
	)

	go func() {
		params := make(map[string]string)

		// If this trigger is linked to a Schedule record, look it up to get Target/Params
		if trigger.ScheduleID != "" {
			schedule, err := database.GetScheduleByID(er.ctx, trigger.ScheduleID)
			if err != nil {
				er.logger.Warn("Failed to load schedule for watch trigger",
					zap.String("schedule_id", trigger.ScheduleID),
					zap.Error(err),
				)
			} else {
				// Skip execution if schedule is disabled
				if !schedule.IsEnabled {
					er.logger.Info("Skipping disabled schedule",
						zap.String("schedule_id", trigger.ScheduleID),
						zap.String("workflow", workflow.Name),
					)
					return
				}
				// Use target from schedule
				if schedule.Target != "" {
					params["target"] = schedule.Target
				}
				// Use params from schedule
				if schedule.Params != nil {
					for k, v := range schedule.Params {
						if s, ok := v.(string); ok {
							params[k] = s
						}
					}
				}
			}
		}

		// Set input parameter if specified (can override schedule params)
		if trigger.Input.Name != "" && input != "" {
			params[trigger.Input.Name] = input
		}

		exec := executor.NewExecutor()
		exec.SetServerMode(true)
		exec.SetSchedulerInvocation(true)
		exec.SetLoader(er.loader)

		var err error
		var result *core.WorkflowResult

		if workflow.IsFlow() {
			result, err = exec.ExecuteFlow(er.ctx, workflow, params, er.config)
		} else {
			result, err = exec.ExecuteModule(er.ctx, workflow, params, er.config)
		}

		if err != nil {
			er.logger.Error("Watch-triggered workflow failed",
				zap.String("workflow", workflow.Name),
				zap.String("trigger", trigger.Name),
				zap.Error(err),
			)
			return
		}

		er.logger.Info("Watch-triggered workflow completed",
			zap.String("workflow", workflow.Name),
			zap.String("status", string(result.Status)),
			zap.Duration("duration", result.EndTime.Sub(result.StartTime)),
		)
	}()

	return nil
}

// GetScheduler returns the underlying scheduler instance.
func (er *EventReceiver) GetScheduler() *scheduler.Scheduler {
	return er.scheduler
}

// GetRegisteredWorkflows returns the list of registered event-triggered workflows.
func (er *EventReceiver) GetRegisteredWorkflows() []*core.Workflow {
	er.mu.RLock()
	defer er.mu.RUnlock()
	return er.workflows
}

// GetRegisteredTriggers returns the list of registered event triggers.
func (er *EventReceiver) GetRegisteredTriggers() []*RegisteredEventTrigger {
	er.mu.RLock()
	defer er.mu.RUnlock()
	return er.triggers
}

// IsRunning returns true if the event receiver is currently running.
func (er *EventReceiver) IsRunning() bool {
	er.mu.RLock()
	defer er.mu.RUnlock()
	return er.running
}

// EmitEvent emits an event to trigger matching workflows.
// This is used by the API endpoint to trigger workflows via events.
func (er *EventReceiver) EmitEvent(event *core.Event) error {
	if !er.running {
		return fmt.Errorf("event receiver not running")
	}
	return er.scheduler.EmitEvent(event)
}

// GetRegisteredTriggersInfo returns trigger information for the API.
// This implements the EventReceiverProvider interface method.
func (er *EventReceiver) GetRegisteredTriggersInfo() []*handlers.EventReceiverTriggerInfo {
	er.mu.RLock()
	defer er.mu.RUnlock()

	result := make([]*handlers.EventReceiverTriggerInfo, 0, len(er.triggers))
	for _, t := range er.triggers {
		topic := ""
		if t.Trigger.Event != nil {
			topic = t.Trigger.Event.Topic
		}
		result = append(result, &handlers.EventReceiverTriggerInfo{
			WorkflowName: t.Workflow.Name,
			TriggerName:  t.Trigger.Name,
			Type:         string(t.Trigger.On),
			Topic:        topic,
			Enabled:      t.Trigger.Enabled,
		})
	}
	return result
}

// GetWorkflowLoader returns the workflow loader for dynamic schedule registration.
// This implements the EventReceiverProvider interface method.
func (er *EventReceiver) GetWorkflowLoader() handlers.WorkflowLoader {
	return er.loader
}

// RegisterSchedule dynamically registers a trigger with the running scheduler.
// This implements the EventReceiverProvider interface method.
func (er *EventReceiver) RegisterSchedule(workflow *core.Workflow, trigger *core.Trigger) error {
	if !er.running {
		return fmt.Errorf("event receiver not running")
	}

	err := er.scheduler.RegisterTrigger(workflow, trigger)
	if err != nil {
		return err
	}

	// Track the trigger
	er.mu.Lock()
	er.triggers = append(er.triggers, &RegisteredEventTrigger{
		Workflow: workflow,
		Trigger:  trigger,
	})
	er.mu.Unlock()

	er.logger.Info("Dynamically registered trigger",
		zap.String("workflow", workflow.Name),
		zap.String("trigger", trigger.Name),
		zap.String("type", string(trigger.On)),
	)

	return nil
}

// processQueuedEvents processes any unprocessed events from the database.
// This is called asynchronously when the event receiver starts to handle
// events that were queued while the server was unavailable.
func (er *EventReceiver) processQueuedEvents() {
	db := database.GetDB()
	if db == nil {
		er.logger.Debug("Database not available, skipping queued event processing")
		return
	}

	repo := repository.NewEventLogRepository(db)
	ctx := er.ctx

	// Fetch unprocessed events
	events, err := repo.ListUnprocessed(ctx, 100)
	if err != nil {
		er.logger.Warn("Failed to fetch unprocessed events", zap.Error(err))
		return
	}

	if len(events) == 0 {
		er.logger.Debug("No queued events to process")
		return
	}

	er.logger.Info("Processing queued events", zap.Int("count", len(events)))

	processedCount := 0
	for _, eventLog := range events {
		// Convert EventLog to core.Event
		event := &core.Event{
			Topic:     eventLog.Topic,
			ID:        eventLog.EventID,
			Name:      eventLog.Name,
			Source:    eventLog.Source,
			DataType:  eventLog.DataType,
			Data:      eventLog.Data,
			Timestamp: eventLog.CreatedAt,
		}

		// Parse the data into ParsedData
		if eventLog.Data != "" {
			var parsedData map[string]interface{}
			if err := json.Unmarshal([]byte(eventLog.Data), &parsedData); err == nil {
				event.ParsedData = parsedData
			} else {
				// If not valid JSON, wrap the value
				event.ParsedData = map[string]interface{}{"value": eventLog.Data}
			}
		}

		// Emit the event to trigger matching workflows
		emitErr := er.scheduler.EmitEvent(event)

		// Mark as processed (even if emit failed, to avoid retrying forever)
		errMsg := ""
		if emitErr != nil {
			errMsg = emitErr.Error()
			er.logger.Warn("Failed to emit queued event",
				zap.Int64("id", eventLog.ID),
				zap.String("topic", eventLog.Topic),
				zap.Error(emitErr),
			)
		} else {
			processedCount++
			er.logger.Debug("Emitted queued event",
				zap.Int64("id", eventLog.ID),
				zap.String("topic", eventLog.Topic),
			)
		}

		if markErr := repo.MarkProcessed(ctx, eventLog.ID, errMsg); markErr != nil {
			er.logger.Warn("Failed to mark event as processed",
				zap.Int64("id", eventLog.ID),
				zap.Error(markErr),
			)
		}

		// Small delay to avoid overwhelming the scheduler
		time.Sleep(50 * time.Millisecond)
	}

	er.logger.Info("Finished processing queued events",
		zap.Int("processed", processedCount),
		zap.Int("total", len(events)),
	)
}
