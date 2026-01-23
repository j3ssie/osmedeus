package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-co-op/gocron/v2"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/functions"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// TriggerHandler is called when a trigger fires
type TriggerHandler func(workflow *core.Workflow, trigger *core.Trigger, input string) error

// EventTriggerHandler is called when an event trigger fires (includes event envelope and resolved vars)
// Parameters:
//   - workflow: the workflow to execute
//   - trigger: the trigger configuration
//   - input: the resolved input value (legacy syntax)
//   - eventEnvelope: JSON-encoded event data
//   - resolvedVars: map of variable names to resolved values (new Vars syntax), nil if legacy syntax
type EventTriggerHandler func(workflow *core.Workflow, trigger *core.Trigger, input string, eventEnvelope string, resolvedVars map[string]string) error

// Default configuration values
const (
	defaultEventQueueSize      = 1000
	defaultBackpressureTimeout = 5 * time.Second
)

// EventEnvelopeKey is the reserved param key for the full event JSON
const EventEnvelopeKey = "_event_envelope"

// Scheduler manages workflow triggers and scheduling
type Scheduler struct {
	scheduler     gocron.Scheduler
	triggers      map[string]*RegisteredTrigger
	handlers      map[string]TriggerHandler
	eventHandlers map[string]EventTriggerHandler // handlers that receive full event envelope
	events        chan *core.Event
	mu            sync.RWMutex
	logger        *zap.Logger
	running       bool

	// Event queue configuration
	queueSize           int           // configurable, default 1000
	backpressureTimeout time.Duration // default 5s

	// Event metrics (atomic counters)
	eventsEnqueued int64 // total events successfully enqueued
	eventsDropped  int64 // total events dropped due to full queue

	// File watcher fields
	watcher    *fsnotify.Watcher
	watchPaths map[string][]*RegisteredTrigger // path → triggers mapping

	// VM pool for JavaScript filter evaluation (no utility functions)
	vmPool *functions.VMPool
	// VM pool with utility functions for filter_functions evaluation
	vmPoolWithFunctions *functions.VMPool

	// Debounce state for watch triggers
	debounceTimers map[string]*debounceState
	debounceMu     sync.Mutex

	// Deduplication cache for event triggers
	dedupeCache *dedupeCache
	stopCleanup chan struct{} // signal to stop cleanup goroutine
}

// debounceState holds the timer state for debounced triggers
type debounceState struct {
	timer *time.Timer
	mu    sync.Mutex
}

// dedupeCache provides time-based deduplication for events
type dedupeCache struct {
	entries sync.Map // key -> expiresAt (time.Time)
}

// newDedupeCache creates a new deduplication cache
func newDedupeCache() *dedupeCache {
	return &dedupeCache{}
}

// IsDuplicate checks if a key already exists and hasn't expired
func (c *dedupeCache) IsDuplicate(key string, window time.Duration) bool {
	if v, ok := c.entries.Load(key); ok {
		expiresAt := v.(time.Time)
		if time.Now().Before(expiresAt) {
			return true
		}
		// Entry expired, will be cleaned up or replaced
	}
	return false
}

// Mark records a key with the given window duration
func (c *dedupeCache) Mark(key string, window time.Duration) {
	c.entries.Store(key, time.Now().Add(window))
}

// cleanup removes expired entries from the cache
func (c *dedupeCache) cleanup() {
	now := time.Now()
	c.entries.Range(func(key, value interface{}) bool {
		expiresAt := value.(time.Time)
		if now.After(expiresAt) {
			c.entries.Delete(key)
		}
		return true
	})
}

// RegisteredTrigger holds trigger information
type RegisteredTrigger struct {
	Workflow *core.Workflow
	Trigger  *core.Trigger
	JobID    string
	Enabled  bool
}

// NewScheduler creates a new scheduler with default configuration
func NewScheduler() (*Scheduler, error) {
	return NewSchedulerWithConfig(defaultEventQueueSize, defaultBackpressureTimeout)
}

// NewSchedulerWithConfig creates a new scheduler with custom configuration
func NewSchedulerWithConfig(queueSize int, backpressureTimeout time.Duration) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Use defaults if invalid values provided
	if queueSize <= 0 {
		queueSize = defaultEventQueueSize
	}
	if backpressureTimeout <= 0 {
		backpressureTimeout = defaultBackpressureTimeout
	}

	// Create GojaRuntime to get a VMPool with all utility functions registered
	gojaRuntime := functions.NewGojaRuntime()

	return &Scheduler{
		scheduler:           s,
		triggers:            make(map[string]*RegisteredTrigger),
		handlers:            make(map[string]TriggerHandler),
		eventHandlers:       make(map[string]EventTriggerHandler),
		events:              make(chan *core.Event, queueSize),
		queueSize:           queueSize,
		backpressureTimeout: backpressureTimeout,
		logger:              logger.Get(),
		watcher:             watcher,
		watchPaths:          make(map[string][]*RegisteredTrigger),
		vmPool:              functions.NewVMPool(nil), // No custom functions needed for filters
		vmPoolWithFunctions: gojaRuntime.GetPool(),    // Pool with utility functions for filter_functions
		debounceTimers:      make(map[string]*debounceState),
		dedupeCache:         newDedupeCache(),
		stopCleanup:         make(chan struct{}),
	}, nil
}

// SetHandler sets the handler for a specific trigger type
func (s *Scheduler) SetHandler(triggerType string, handler TriggerHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[triggerType] = handler
}

// SetEventHandler sets a handler that receives the full event envelope
// This is preferred for event triggers as it provides access to all event metadata
func (s *Scheduler) SetEventHandler(triggerType string, handler EventTriggerHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.eventHandlers[triggerType] = handler
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
	s.logger.Debug("Registered trigger",
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
			s.logger.Debug("File change detected",
				zap.String("path", event.Name),
				zap.String("trigger", reg.Trigger.Name),
				zap.String("op", event.Op.String()),
			)

			// Check if debounce is configured
			if reg.Trigger.HasDebounce() {
				s.handleDebouncedTrigger(reg)
			} else {
				s.logger.Info("Triggering workflow (no debounce)",
					zap.String("path", event.Name),
					zap.String("trigger", reg.Trigger.Name),
				)
				go s.handleTrigger(reg)
			}
		}
	}
}

// handleDebouncedTrigger handles a trigger with debounce
func (s *Scheduler) handleDebouncedTrigger(reg *RegisteredTrigger) {
	triggerKey := fmt.Sprintf("%s:%s", reg.Workflow.Name, reg.Trigger.Name)
	debounceDuration := reg.Trigger.GetDebounceDuration()

	s.debounceMu.Lock()
	state, exists := s.debounceTimers[triggerKey]
	if !exists {
		state = &debounceState{}
		s.debounceTimers[triggerKey] = state
	}
	s.debounceMu.Unlock()

	state.mu.Lock()
	defer state.mu.Unlock()

	// Stop existing timer if any
	if state.timer != nil {
		state.timer.Stop()
	}

	// Create new timer
	state.timer = time.AfterFunc(debounceDuration, func() {
		s.logger.Info("Debounce timer fired, triggering workflow",
			zap.String("trigger", reg.Trigger.Name),
			zap.String("workflow", reg.Workflow.Name),
			zap.Duration("debounce", debounceDuration),
		)
		s.handleTrigger(reg)
	})

	s.logger.Debug("Debounce timer reset",
		zap.String("trigger", triggerKey),
		zap.Duration("debounce", debounceDuration),
	)
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

	// Start dedupe cache cleanup goroutine
	go s.dedupeCleanupLoop()

	s.logger.Debug("Scheduler started")
	return nil
}

// dedupeCleanupLoop periodically cleans up expired dedupe entries
func (s *Scheduler) dedupeCleanupLoop() {
	ticker := time.NewTicker(30 * time.Second) // Cleanup every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.dedupeCache.cleanup()
		case <-s.stopCleanup:
			return
		}
	}
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

	// Stop cleanup goroutine
	close(s.stopCleanup)

	// Stop all debounce timers
	s.debounceMu.Lock()
	for _, state := range s.debounceTimers {
		state.mu.Lock()
		if state.timer != nil {
			state.timer.Stop()
		}
		state.mu.Unlock()
	}
	s.debounceMu.Unlock()

	s.running = false
	close(s.events)

	s.logger.Info("Scheduler stopped")
	return nil
}

// EmitEvent emits a core.Event to trigger workflows with backpressure support.
// Fast path: non-blocking send if queue has space.
// Slow path: waits up to backpressureTimeout before dropping the event.
func (s *Scheduler) EmitEvent(event *core.Event) error {
	if !s.running {
		return fmt.Errorf("scheduler not running")
	}

	// Set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Fast path: non-blocking send
	select {
	case s.events <- event:
		atomic.AddInt64(&s.eventsEnqueued, 1)
		s.logger.Debug("Event emitted",
			zap.String("topic", event.Topic),
			zap.String("id", event.ID),
			zap.String("name", event.Name),
			zap.String("source", event.Source),
			zap.String("data_type", event.DataType),
			zap.String("data", event.Data),
			zap.Any("parsed_data", event.ParsedData),
			zap.Time("timestamp", event.Timestamp),
		)
		return nil
	default:
		// Queue full - apply backpressure
	}

	// Backpressure: wait with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.backpressureTimeout)
	defer cancel()

	select {
	case s.events <- event:
		atomic.AddInt64(&s.eventsEnqueued, 1)
		s.logger.Debug("Event queued after backpressure",
			zap.String("topic", event.Topic),
			zap.String("name", event.Name),
		)
		return nil
	case <-ctx.Done():
		atomic.AddInt64(&s.eventsDropped, 1)
		s.logger.Warn("Event dropped - queue full",
			zap.String("topic", event.Topic),
			zap.String("name", event.Name),
			zap.Int("queue_size", s.queueSize),
			zap.Duration("timeout", s.backpressureTimeout),
		)
		return fmt.Errorf("event queue full after %v", s.backpressureTimeout)
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

// GetEventMetrics returns event queue metrics (enqueued and dropped counts).
// These are atomic counters that track total events processed since scheduler start.
func (s *Scheduler) GetEventMetrics() (enqueued, dropped int64) {
	return atomic.LoadInt64(&s.eventsEnqueued), atomic.LoadInt64(&s.eventsDropped)
}

// GetQueueStats returns current queue configuration and status
func (s *Scheduler) GetQueueStats() (queueSize int, currentLen int, backpressureTimeout time.Duration) {
	return s.queueSize, len(s.events), s.backpressureTimeout
}

// eventListener listens for events and triggers workflows
func (s *Scheduler) eventListener() {
	for event := range s.events {
		// Parse event data if needed
		if event.ParsedData == nil && event.Data != "" {
			_ = event.ParseData()
		}

		s.logger.Debug("Processing event from queue",
			zap.String("topic", event.Topic),
			zap.String("id", event.ID),
			zap.String("source", event.Source),
			zap.Any("parsed_data", event.ParsedData),
		)

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

	// Evaluate filters if defined (simple JS expressions without utility functions)
	if trigger.HasFilters() {
		if !s.evaluateFilters(trigger.GetFilters(), event) {
			return false
		}
	}

	// Evaluate filter_functions if defined (JS expressions with utility functions)
	if trigger.HasFilterFunctions() {
		if !s.evaluateFilterFunctions(trigger.GetFilterFunctions(), event) {
			return false
		}
	}

	// Check deduplication
	if trigger.Event != nil && trigger.Event.HasDeduplication() {
		dedupeKey := s.computeDedupeKey(trigger.Event.DedupeKey, event)
		window := trigger.Event.GetDedupeWindow()

		if s.dedupeCache.IsDuplicate(dedupeKey, window) {
			s.logger.Debug("Event deduplicated",
				zap.String("key", dedupeKey),
				zap.String("topic", event.Topic),
				zap.Duration("window", window),
			)
			return false
		}

		// Mark this key as seen
		s.dedupeCache.Mark(dedupeKey, window)
	}

	return true
}

// computeDedupeKey computes the deduplication key from a template and event
func (s *Scheduler) computeDedupeKey(template string, event *core.Event) string {
	key := template

	// Replace event fields
	key = replaceTemplateVar(key, "event.topic", event.Topic)
	key = replaceTemplateVar(key, "event.name", event.Name)
	key = replaceTemplateVar(key, "event.source", event.Source)
	key = replaceTemplateVar(key, "event.id", event.ID)
	key = replaceTemplateVar(key, "event.data_type", event.DataType)

	// Replace event.data fields if parsed data is available
	if event.ParsedData != nil {
		key = s.replaceDataFields(key, event.ParsedData, "event.data")
	}

	return key
}

// replaceDataFields replaces template variables with values from a nested map
func (s *Scheduler) replaceDataFields(template string, data map[string]interface{}, prefix string) string {
	result := template
	for k, v := range data {
		placeholder := fmt.Sprintf("{{%s.%s}}", prefix, k)
		switch val := v.(type) {
		case string:
			result = replaceTemplateVar(result, fmt.Sprintf("%s.%s", prefix, k), val)
		case map[string]interface{}:
			// Handle nested maps
			result = s.replaceDataFields(result, val, fmt.Sprintf("%s.%s", prefix, k))
		default:
			result = replaceTemplateVar(result, fmt.Sprintf("%s.%s", prefix, k), fmt.Sprintf("%v", v))
		}
		_ = placeholder // silence unused variable warning
	}
	return result
}

// evaluateFilters evaluates JavaScript filter expressions using pooled Goja VMs
func (s *Scheduler) evaluateFilters(filters []string, event *core.Event) bool {
	vmCtx := s.vmPool.Get()
	defer s.vmPool.Put(vmCtx)

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

	if err := vmCtx.SetVariables(map[string]interface{}{"event": eventObj}); err != nil {
		s.logger.Warn("Failed to set event in JS context", zap.Error(err))
		return false
	}

	// All filters must pass
	for _, filter := range filters {
		// Render template variables in the expression (for consistency with filter_functions)
		rendered := s.renderFilterFunctionTemplates(filter, event)

		result, err := vmCtx.Run(rendered)
		if err != nil {
			s.logger.Warn("Filter evaluation failed",
				zap.String("filter", filter),
				zap.String("rendered", rendered),
				zap.Error(err),
			)
			return false
		}

		if !result.ToBoolean() {
			return false
		}
	}

	return true
}

// evaluateFilterFunctions evaluates filter expressions with utility functions using pooled Goja VMs
func (s *Scheduler) evaluateFilterFunctions(filterFuncs []string, event *core.Event) bool {
	vmCtx := s.vmPoolWithFunctions.Get()
	defer s.vmPoolWithFunctions.Put(vmCtx)

	// Build event object for JS context
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

	if err := vmCtx.SetVariables(map[string]interface{}{"event": eventObj}); err != nil {
		s.logger.Warn("Failed to set event in JS context for filter_functions", zap.Error(err))
		return false
	}

	// All filter functions must pass
	for _, filterFunc := range filterFuncs {
		// Render template variables in the expression (e.g., {{event.data.url}})
		rendered := s.renderFilterFunctionTemplates(filterFunc, event)

		result, err := vmCtx.Run(rendered)
		if err != nil {
			s.logger.Warn("Filter function evaluation failed",
				zap.String("filter_function", filterFunc),
				zap.String("rendered", rendered),
				zap.Error(err),
			)
			return false
		}

		if !result.ToBoolean() {
			return false
		}
	}

	return true
}

// renderFilterFunctionTemplates replaces template variables like {{event.data.field}} with actual values
func (s *Scheduler) renderFilterFunctionTemplates(expr string, event *core.Event) string {
	result := expr

	// Replace event fields
	result = replaceTemplateVar(result, "event.topic", event.Topic)
	result = replaceTemplateVar(result, "event.name", event.Name)
	result = replaceTemplateVar(result, "event.source", event.Source)
	result = replaceTemplateVar(result, "event.id", event.ID)
	result = replaceTemplateVar(result, "event.data_type", event.DataType)

	// Replace event.data fields if parsed data is available
	if event.ParsedData != nil {
		result = s.replaceDataFields(result, event.ParsedData, "event.data")
	}

	return result
}

// handleEventTrigger handles an event-based trigger
func (s *Scheduler) handleEventTrigger(reg *RegisteredTrigger, event *core.Event) {
	s.mu.RLock()
	eventHandler, hasEventHandler := s.eventHandlers[string(core.TriggerEvent)]
	legacyHandler, hasLegacyHandler := s.handlers[string(core.TriggerEvent)]
	s.mu.RUnlock()

	if !hasEventHandler && !hasLegacyHandler {
		return
	}

	// Try new Vars syntax first
	var resolvedVars map[string]string
	if reg.Trigger.Input.HasVars() {
		resolvedVars = s.resolveEventVars(reg.Trigger, event)
	}

	// Resolve legacy input (will be empty string if Vars syntax is used)
	input := ""
	if resolvedVars == nil {
		input = s.resolveEventInput(reg.Trigger, event)
	}

	// Prefer EventTriggerHandler (with envelope) over legacy TriggerHandler
	if hasEventHandler {
		envelope := s.serializeEventEnvelope(event)
		if err := eventHandler(reg.Workflow, reg.Trigger, input, envelope, resolvedVars); err != nil {
			s.logger.Error("Event trigger handler failed",
				zap.String("workflow", reg.Workflow.Name),
				zap.String("event", event.Name),
				zap.String("topic", event.Topic),
				zap.Error(err),
			)
		}
		return
	}

	// Fall back to legacy handler (doesn't support Vars syntax)
	if err := legacyHandler(reg.Workflow, reg.Trigger, input); err != nil {
		s.logger.Error("Event trigger handler failed",
			zap.String("workflow", reg.Workflow.Name),
			zap.String("event", event.Name),
			zap.String("topic", event.Topic),
			zap.Error(err),
		)
	}
}

// serializeEventEnvelope serializes an event to JSON for the EventEnvelope template variable
func (s *Scheduler) serializeEventEnvelope(event *core.Event) string {
	// Create envelope structure with all event fields
	envelope := map[string]interface{}{
		"topic":         event.Topic,
		"id":            event.ID,
		"name":          event.Name,
		"source":        event.Source,
		"data":          event.ParsedData,
		"data_raw":      event.Data,
		"data_type":     event.DataType,
		"workspace":     event.Workspace,
		"run_uuid":      event.RunUUID,
		"workflow_name": event.WorkflowName,
		"timestamp":     event.Timestamp.Format(time.RFC3339),
	}

	jsonBytes, err := json.Marshal(envelope)
	if err != nil {
		s.logger.Warn("Failed to serialize event envelope", zap.Error(err))
		return "{}"
	}
	return string(jsonBytes)
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

// functionCallPattern matches function calls like trim(...), jq(...)
var functionCallPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*\s*\(`)

// resolveEventVars resolves all variables from the new Vars syntax
func (s *Scheduler) resolveEventVars(trigger *core.Trigger, event *core.Event) map[string]string {
	if !trigger.Input.HasVars() {
		return nil
	}

	result := make(map[string]string, len(trigger.Input.Vars))
	for varName, expr := range trigger.Input.Vars {
		result[varName] = s.resolveVarExpression(expr, event)
	}
	return result
}

// resolveVarExpression resolves a single expression from the Vars syntax.
// Supports:
//   - event_data.field - access parsed event data fields
//   - event.topic, event.source, etc. - access event metadata
//   - trim(event_data.desc) - function calls with dot notation arguments
func (s *Scheduler) resolveVarExpression(expr string, event *core.Event) string {
	// Check for function call pattern (e.g., trim(...), jq(...))
	if functionCallPattern.MatchString(expr) {
		return s.resolveVarFunction(expr, event)
	}

	// Handle dot notation: event_data.field or event.topic
	return s.resolveDotNotation(expr, event)
}

// resolveVarFunction resolves a function call expression in Vars syntax.
// It first replaces dot notation references with their values, then executes the function.
func (s *Scheduler) resolveVarFunction(expr string, event *core.Event) string {
	// Replace event_data.* references with actual values
	rendered := s.replaceDotNotationInExpr(expr, event)

	// Execute the function expression using the existing function execution infrastructure
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
		s.logger.Warn("Failed to execute var function",
			zap.String("expr", expr),
			zap.String("rendered", rendered),
			zap.Error(err),
		)
		return ""
	}

	return fmt.Sprintf("%v", result)
}

// replaceDotNotationInExpr replaces dot notation references like event_data.url with quoted string values
func (s *Scheduler) replaceDotNotationInExpr(expr string, event *core.Event) string {
	// Pattern to match event_data.field or event.field references
	dotNotationPattern := regexp.MustCompile(`(event_data|event)\.([a-zA-Z_][a-zA-Z0-9_.]*)`)

	return dotNotationPattern.ReplaceAllStringFunc(expr, func(match string) string {
		value := s.resolveDotNotation(match, event)
		// Return as a quoted string for function arguments
		return fmt.Sprintf("%q", value)
	})
}

// resolveDotNotation handles "prefix.field" field access for event data
func (s *Scheduler) resolveDotNotation(expr string, event *core.Event) string {
	parts := strings.Split(expr, ".")
	if len(parts) < 2 {
		return expr
	}

	switch parts[0] {
	case "event_data":
		// Access parsed event data fields
		if event.ParsedData != nil {
			if val := getNestedField(event.ParsedData, parts[1:]); val != nil {
				return fmt.Sprintf("%v", val)
			}
		}
	case "event":
		// Access event metadata
		switch parts[1] {
		case "topic":
			return event.Topic
		case "source":
			return event.Source
		case "name":
			return event.Name
		case "id":
			return event.ID
		case "data_type":
			return event.DataType
		case "workspace":
			return event.Workspace
		case "run_uuid":
			return event.RunUUID
		case "workflow_name":
			return event.WorkflowName
		}
	}
	return ""
}

// getNestedField retrieves a nested value from a map using a path like ["foo", "bar"]
func getNestedField(data map[string]interface{}, path []string) interface{} {
	current := interface{}(data)
	for _, key := range path {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[key]
		} else {
			return nil
		}
	}
	return current
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
