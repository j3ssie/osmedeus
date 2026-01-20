package functions

import (
	"bufio"
	"os"
	"strings"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/notify"
	"go.uber.org/zap"
)

// generateEvent sends a structured event with workspace, topic, source, data_type, and data.
// It attempts to send to the server first, falls back to database queuing if unavailable,
// and also sends to configured webhooks.
// Usage: generate_event(workspace, topic, source, data_type, data)
// RunID and WorkflowName are automatically populated from the runtime context.
func (vf *vmFunc) generateEvent(call goja.FunctionCall) goja.Value {
	workspace := call.Argument(0).String()
	topic := call.Argument(1).String()
	source := call.Argument(2).String()
	dataType := call.Argument(3).String()

	// Validate required fields
	if workspace == "" || topic == "" || source == "" || dataType == "" {
		zap.L().Warn("generateEvent: missing required fields",
			zap.String("workspace", workspace),
			zap.String("topic", topic),
			zap.String("source", source),
			zap.String("data_type", dataType))
		return vf.vm.ToValue(false)
	}

	// Get data (can be string or object)
	var data interface{}
	if !goja.IsUndefined(call.Argument(4)) && !goja.IsNull(call.Argument(4)) {
		data = call.Argument(4).Export()
	}

	// Get runtime context for RunID and WorkflowName
	var runID, workflowName string
	if ctx := vf.getContext(); ctx != nil {
		runID = ctx.scanID
		workflowName = ctx.workflowName
	}

	// Use SendEventWithFallback to try server first, then queue to DB, and also send to webhooks
	err := notify.SendEventWithFallback(workspace, topic, source, dataType, runID, workflowName, data)
	if err != nil {
		zap.L().Debug("generateEvent: server delivery failed (event queued or webhook sent)",
			zap.String("topic", topic),
			zap.Error(err))
	}
	// Return true even if server was unavailable, as event is queued for later processing
	return vf.vm.ToValue(true)
}

// generateEventFromFile reads a file and generates an event for each non-empty line.
// It attempts to send to the server first, falls back to database queuing if unavailable,
// and also sends to configured webhooks.
// Usage: generate_event_from_file(workspace, topic, source, data_type, filePath)
// RunID and WorkflowName are automatically populated from the runtime context.
func (vf *vmFunc) generateEventFromFile(call goja.FunctionCall) goja.Value {
	workspace := call.Argument(0).String()
	topic := call.Argument(1).String()
	source := call.Argument(2).String()
	dataType := call.Argument(3).String()
	filePath := call.Argument(4).String()

	// Validate required fields
	if workspace == "" || filePath == "" || topic == "" || source == "" || dataType == "" {
		zap.L().Warn("generateEventFromFile: missing required fields",
			zap.String("workspace", workspace),
			zap.String("path", filePath),
			zap.String("topic", topic),
			zap.String("source", source),
			zap.String("data_type", dataType))
		return vf.vm.ToValue(0)
	}

	file, err := os.Open(filePath)
	if err != nil {
		zap.L().Warn("generateEventFromFile: failed to open file",
			zap.String("path", filePath),
			zap.Error(err))
		return vf.vm.ToValue(0)
	}
	defer func() { _ = file.Close() }()

	// Get runtime context for RunID and WorkflowName
	var runID, workflowName string
	if ctx := vf.getContext(); ctx != nil {
		runID = ctx.scanID
		workflowName = ctx.workflowName
	}

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Use SendEventWithFallback to try server first, then queue to DB, and also send to webhooks
		_ = notify.SendEventWithFallback(workspace, topic, source, dataType, runID, workflowName, line)
		count++ // Count all attempts since events are queued if server unavailable
	}

	if err := scanner.Err(); err != nil {
		zap.L().Warn("generateEventFromFile: error reading file",
			zap.String("path", filePath),
			zap.Error(err))
	}

	return vf.vm.ToValue(count)
}
