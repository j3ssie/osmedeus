package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/fsnotify/fsnotify"
	"github.com/go-co-op/gocron/v2"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/functions"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// TriggerHandler is called when a trigger fires
type TriggerHandler func(workflow *core.Workflow, trigger *core.Trigger, input string) error

// Scheduler manages workflow triggers and scheduling
type Scheduler struct {
	scheduler gocron.Scheduler
	triggers  map[string]*RegisteredTrigger
	handlers  map[string]TriggerHandler
	events    chan *core.Event
	mu        sync.RWMutex
	logger    *zap.Logger
	running   bool

	// File watcher fields
	watcher    *fsnotify.Watcher
	watchPaths map[string][]*RegisteredTrigger // path → triggers mapping
}

// RegisteredTrigger holds trigger information
type RegisteredTrigger struct {
	Workflow *core.Workflow
	Trigger  *core.Trigger
	JobID    string
	Enabled  bool
}

// NewScheduler creates a new scheduler
func NewScheduler() (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &Scheduler{
		scheduler:  s,
		triggers:   make(map[string]*RegisteredTrigger),
		handlers:   make(map[string]TriggerHandler),
		events:     make(chan *core.Event, 100),
		logger:     logger.Get(),
		watcher:    watcher,
		watchPaths: make(map[string][]*RegisteredTrigger),
	}, nil
}

// SetHandler sets the handler for a specific trigger type
func (s *Scheduler) SetHandler(triggerType string, handler TriggerHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[triggerType] = handler
}

// RegisterTrigger registers a workflow trigger
func (s *Scheduler) RegisterTrigger(workflow *core.Workflow, trigger *core.Trigger) error {
	if !trigger.IsEnabled() {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	triggerKey := fmt.Sprintf("%s:%s", workflow.Name, trigger.Name)

	// Remove existing trigger if present
	if existing, ok := s.triggers[triggerKey]; ok {
		s.unregisterTriggerLocked(existing)
	}

	reg := &RegisteredTrigger{
		Workflow: workflow,
		Trigger:  trigger,
		Enabled:  true,
	}

	switch trigger.On {
	case core.TriggerCron:
		if err := s.registerCronTrigger(reg); err != nil {
			return err
		}
	case core.TriggerEvent:
		// Event triggers are handled through EmitEvent
	case core.TriggerWatch:
		if err := s.registerWatchTrigger(reg); err != nil {
			return err
		}
	case core.TriggerManual:
		// Manual triggers are handled through API
	default:
		return fmt.Errorf("unknown trigger type: %s", trigger.On)
	}

	s.triggers[triggerKey] = reg
	s.logger.Info("Registered trigger",
		zap.String("workflow", workflow.Name),
		zap.String("trigger", trigger.Name),
		zap.String("type", string(trigger.On)),
	)

	return nil
}

// registerCronTrigger registers a cron-based trigger
func (s *Scheduler) registerCronTrigger(reg *RegisteredTrigger) error {
	if reg.Trigger.Schedule == "" {
		return fmt.Errorf("cron trigger requires schedule")
	}

	job, err := s.scheduler.NewJob(
		gocron.CronJob(reg.Trigger.Schedule, false),
		gocron.NewTask(func() {
			s.handleTrigger(reg)
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create cron job: %w", err)
	}

	reg.JobID = job.ID().String()
	return nil
}

// registerWatchTrigger registers a file watch trigger
func (s *Scheduler) registerWatchTrigger(reg *RegisteredTrigger) error {
	if reg.Trigger.Path == "" {
		return fmt.Errorf("watch trigger requires path")
	}

	return s.registerWatchPath(reg)
}

// registerWatchPath adds a path to the file watcher
func (s *Scheduler) registerWatchPath(reg *RegisteredTrigger) error {
	path := reg.Trigger.Path

	s.mu.Lock()
	defer s.mu.Unlock()

	// Add to path mapping
	s.watchPaths[path] = append(s.watchPaths[path], reg)

	// Add to watcher if not already watching this path
	if len(s.watchPaths[path]) == 1 {
		if err := s.watcher.Add(path); err != nil {
			// Remove from mapping if watcher.Add fails
			delete(s.watchPaths, path)
			return fmt.Errorf("failed to watch path %s: %w", path, err)
		}
		s.logger.Debug("Added file watch", zap.String("path", path))
	}

	return nil
}

// watchFileEvents processes file system events from fsnotify
func (s *Scheduler) watchFileEvents() {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			s.handleFileEvent(event)
		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			s.logger.Error("File watcher error", zap.Error(err))
		}
	}
}

// handleFileEvent processes a file system event
func (s *Scheduler) handleFileEvent(event fsnotify.Event) {
	// Only handle write/create events
	if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
		return
	}

	s.mu.RLock()
	triggers := s.watchPaths[event.Name]
	s.mu.RUnlock()

	for _, reg := range triggers {
		if reg.Enabled {
			s.logger.Info("File change detected, triggering workflow",
				zap.String("path", event.Name),
				zap.String("trigger", reg.Trigger.Name),
				zap.String("op", event.Op.String()),
			)
			go s.handleTrigger(reg)
		}
	}
}

// handleTrigger handles a trigger firing
func (s *Scheduler) handleTrigger(reg *RegisteredTrigger) {
	s.mu.RLock()
	handler, ok := s.handlers[string(reg.Trigger.On)]
	s.mu.RUnlock()

	if !ok {
		s.logger.Warn("No handler for trigger type", zap.String("type", string(reg.Trigger.On)))
		return
	}

	// Get input based on trigger input configuration
	input := s.resolveInput(reg.Trigger)

	if err := handler(reg.Workflow, reg.Trigger, input); err != nil {
		s.logger.Error("Trigger handler failed",
			zap.String("workflow", reg.Workflow.Name),
			zap.String("trigger", reg.Trigger.Name),
			zap.Error(err),
		)
	}
}

// resolveInput resolves the input for a trigger
func (s *Scheduler) resolveInput(trigger *core.Trigger) string {
	switch trigger.Input.Type {
	case "param":
		return trigger.Input.Name
	case "file":
		// Read from file
		return trigger.Input.Path
	default:
		return ""
	}
}

// UnregisterTrigger removes a trigger by name
func (s *Scheduler) UnregisterTrigger(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	reg, ok := s.triggers[name]
	if !ok {
		return fmt.Errorf("trigger not found: %s", name)
	}

	s.unregisterTriggerLocked(reg)
	delete(s.triggers, name)

	return nil
}

// unregisterTriggerLocked removes a trigger (caller must hold lock)
func (s *Scheduler) unregisterTriggerLocked(reg *RegisteredTrigger) {
	// Remove from scheduler if JobID exists
	// Note: gocron v2 requires job reference, simplified here
	reg.Enabled = false
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.scheduler.Start()
	s.running = true

	// Start file watcher event loop
	go s.watchFileEvents()

	// Start event listener
	go s.eventListener()

	s.logger.Info("Scheduler started")
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if err := s.scheduler.Shutdown(); err != nil {
		return fmt.Errorf("failed to stop scheduler: %w", err)
	}

	// Close the file watcher
	if s.watcher != nil {
		if err := s.watcher.Close(); err != nil {
			s.logger.Warn("Failed to close file watcher", zap.Error(err))
		}
	}

	s.running = false
	close(s.events)

	s.logger.Info("Scheduler stopped")
	return nil
}

// EmitEvent emits a core.Event to trigger workflows
func (s *Scheduler) EmitEvent(event *core.Event) error {
	if !s.running {
		return fmt.Errorf("scheduler not running")
	}

	// Set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	select {
	case s.events <- event:
		s.logger.Debug("Event emitted",
			zap.String("topic", event.Topic),
			zap.String("name", event.Name),
		)
		return nil
	default:
		return fmt.Errorf("event queue full")
	}
}

// EmitEventSimple emits a simple event with topic and name
func (s *Scheduler) EmitEventSimple(topic, name string, data map[string]interface{}) error {
	event := &core.Event{
		Topic:      topic,
		Name:       name,
		Timestamp:  time.Now(),
		ParsedData: data,
	}
	return s.EmitEvent(event)
}

// eventListener listens for events and triggers workflows
func (s *Scheduler) eventListener() {
	for event := range s.events {
		// Parse event data if needed
		if event.ParsedData == nil && event.Data != "" {
			_ = event.ParseData()
		}

		s.mu.RLock()
		for _, reg := range s.triggers {
			if reg.Trigger.On == core.TriggerEvent && reg.Enabled {
				if s.matchesEventTrigger(reg.Trigger, event) {
					go s.handleEventTrigger(reg, event)
				}
			}
		}
		s.mu.RUnlock()
	}
}

// matchesEventTrigger checks if an event matches the trigger configuration
func (s *Scheduler) matchesEventTrigger(trigger *core.Trigger, event *core.Event) bool {
	// Check topic match
	if !trigger.MatchesTopic(event.Topic) {
		return false
	}

	// Evaluate filters if defined
	if trigger.HasFilters() {
		return s.evaluateFilters(trigger.GetFilters(), event)
	}

	return true
}

// evaluateFilters evaluates JavaScript filter expressions using Goja
func (s *Scheduler) evaluateFilters(filters []string, event *core.Event) bool {
	vm := goja.New()

	// Set event object in JS context
	eventObj := map[string]interface{}{
		"topic":     event.Topic,
		"id":        event.ID,
		"name":      event.Name,
		"source":    event.Source,
		"data_type": event.DataType,
	}

	// Add parsed data if available
	if event.ParsedData != nil {
		eventObj["data"] = event.ParsedData
	} else if event.Data != "" {
		eventObj["data"] = event.Data
	}

	if err := vm.Set("event", eventObj); err != nil {
		s.logger.Warn("Failed to set event in JS context", zap.Error(err))
		return false
	}

	// All filters must pass
	for _, filter := range filters {
		result, err := vm.RunString(filter)
		if err != nil {
			s.logger.Warn("Filter evaluation failed",
				zap.String("filter", filter),
				zap.Error(err),
			)
			return false
		}

		boolResult := result.ToBoolean()
		if !boolResult {
			return false
		}
	}

	return true
}

// handleEventTrigger handles an event-based trigger
func (s *Scheduler) handleEventTrigger(reg *RegisteredTrigger, event *core.Event) {
	s.mu.RLock()
	handler, ok := s.handlers[string(core.TriggerEvent)]
	s.mu.RUnlock()

	if !ok {
		return
	}

	// Resolve input from event
	input := s.resolveEventInput(reg.Trigger, event)

	if err := handler(reg.Workflow, reg.Trigger, input); err != nil {
		s.logger.Error("Event trigger handler failed",
			zap.String("workflow", reg.Workflow.Name),
			zap.String("event", event.Name),
			zap.String("topic", event.Topic),
			zap.Error(err),
		)
	}
}

// resolveEventInput resolves the input value from an event based on trigger configuration
func (s *Scheduler) resolveEventInput(trigger *core.Trigger, event *core.Event) string {
	switch trigger.Input.Type {
	case "function":
		// Execute function like jq("{{event.data}}", ".url")
		return s.resolveEventFunction(trigger.Input.Function, event)
	case "event_data":
		// Extract field from event data
		if trigger.Input.Field != "" {
			if v := event.GetDataField(trigger.Input.Field); v != nil {
				return fmt.Sprintf("%v", v)
			}
		}
		return ""
	case "param":
		return trigger.Input.Name
	case "file":
		return trigger.Input.Path
	default:
		return ""
	}
}

// resolveEventFunction executes a function expression with event context
func (s *Scheduler) resolveEventFunction(funcExpr string, event *core.Event) string {
	if funcExpr == "" {
		return ""
	}

	// Replace template variables like {{event.data}} with actual values
	rendered := funcExpr

	// Replace {{event.data}} with the actual JSON data
	if event.Data != "" {
		rendered = replaceTemplateVar(rendered, "event.data", event.Data)
	}

	// Replace other event fields
	rendered = replaceTemplateVar(rendered, "event.topic", event.Topic)
	rendered = replaceTemplateVar(rendered, "event.name", event.Name)
	rendered = replaceTemplateVar(rendered, "event.source", event.Source)
	rendered = replaceTemplateVar(rendered, "event.id", event.ID)
	rendered = replaceTemplateVar(rendered, "event.data_type", event.DataType)

	// Execute the function expression
	ctx := map[string]interface{}{
		"event": map[string]interface{}{
			"topic":     event.Topic,
			"id":        event.ID,
			"name":      event.Name,
			"source":    event.Source,
			"data":      event.ParsedData,
			"data_type": event.DataType,
		},
	}

	result, err := functions.Execute(rendered, ctx)
	if err != nil {
		s.logger.Warn("Failed to execute function",
			zap.String("function", funcExpr),
			zap.Error(err),
		)
		return ""
	}

	return fmt.Sprintf("%v", result)
}

// replaceTemplateVar replaces {{varName}} with value
func replaceTemplateVar(s, varName, value string) string {
	placeholder := "{{" + varName + "}}"
	result := s
	for {
		idx := indexOf(result, placeholder)
		if idx < 0 {
			break
		}
		result = result[:idx] + value + result[idx+len(placeholder):]
	}
	return result
}

// indexOf finds the index of substr in s, returns -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ListTriggers returns all registered triggers
func (s *Scheduler) ListTriggers() []*RegisteredTrigger {
	s.mu.RLock()
	defer s.mu.RUnlock()

	triggers := make([]*RegisteredTrigger, 0, len(s.triggers))
	for _, t := range s.triggers {
		triggers = append(triggers, t)
	}
	return triggers
}

// TriggerManually manually fires a trigger
func (s *Scheduler) TriggerManually(ctx context.Context, workflowName, triggerName, input string) error {
	s.mu.RLock()
	key := fmt.Sprintf("%s:%s", workflowName, triggerName)
	reg, ok := s.triggers[key]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("trigger not found: %s", key)
	}

	handler, ok := s.handlers[string(core.TriggerManual)]
	if !ok {
		// Try generic handler
		handler, ok = s.handlers["default"]
		if !ok {
			return fmt.Errorf("no handler for manual triggers")
		}
	}

	return handler(reg.Workflow, reg.Trigger, input)
}

// DefaultScheduler is the global scheduler instance
var (
	DefaultScheduler *Scheduler
	once             sync.Once
)

// GetScheduler returns the global scheduler instance
func GetScheduler() (*Scheduler, error) {
	var err error
	once.Do(func() {
		DefaultScheduler, err = NewScheduler()
	})
	return DefaultScheduler, err
}
