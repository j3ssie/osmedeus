package executor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/functions"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/template"
	"go.uber.org/zap"
)

// ForeachExecutor executes foreach steps
type ForeachExecutor struct {
	dispatcher       *StepDispatcher
	templateEngine   template.TemplateEngine
	functionRegistry *functions.Registry
}

// NewForeachExecutor creates a new foreach executor
func NewForeachExecutor(dispatcher *StepDispatcher, engine template.TemplateEngine, registry *functions.Registry) *ForeachExecutor {
	return &ForeachExecutor{
		dispatcher:       dispatcher,
		templateEngine:   engine,
		functionRegistry: registry,
	}
}

// Name returns the executor name for logging/debugging
func (e *ForeachExecutor) Name() string {
	return "foreach"
}

// StepTypes returns the step types this executor handles
func (e *ForeachExecutor) StepTypes() []core.StepType {
	return []core.StepType{core.StepTypeForeach}
}

// Execute executes a foreach step
func (e *ForeachExecutor) Execute(ctx context.Context, step *core.Step, execCtx *core.ExecutionContext) (*core.StepResult, error) {
	result := &core.StepResult{
		StepName:  step.Name,
		Status:    core.StepStatusRunning,
		StartTime: time.Now(),
		Exports:   make(map[string]interface{}),
	}

	if step.Step == nil {
		result.Status = core.StepStatusFailed
		result.Error = fmt.Errorf("foreach step has no inner step")
		result.EndTime = time.Now()
		return result, result.Error
	}

	threads, err := step.Threads.Int()
	if err != nil {
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}
	if threads <= 0 {
		threads = 1
	}

	// Check for streaming output configuration via exports
	var streamingOutput string
	if step.Exports != nil {
		if so, ok := step.Exports["streaming_output"]; ok {
			// Render the template
			rendered, err := e.templateEngine.Render(so, execCtx.GetVariables())
			if err == nil {
				streamingOutput = rendered
			}
		}
	}

	// Auto-enable streaming for large inputs (>1000 lines) to avoid OOM
	autoStreaming := false
	var autoStreamFile string
	if streamingOutput == "" {
		lineCount, countErr := countInputLines(step.Input)
		if countErr == nil && lineCount > 1000 {
			autoStreaming = true
			tmpFile, tmpErr := os.CreateTemp("", "osmedeus-foreach-*.txt")
			if tmpErr == nil {
				autoStreamFile = tmpFile.Name()
				streamingOutput = autoStreamFile
				_ = tmpFile.Close()
			}
		}
	}
	// Clean up auto-stream temp file on exit
	if autoStreamFile != "" {
		defer func() { _ = os.Remove(autoStreamFile) }()
	}

	// Execute with streaming worker pool
	outputs, err := e.executeWithWorkerPoolStreaming(ctx, step, step.Input, threads, execCtx, streamingOutput)

	if autoStreaming && autoStreamFile != "" {
		// Read back a truncated version of the auto-streamed output
		result.Output = readTruncatedFile(autoStreamFile, 10*1024*1024) // 10MB max
	} else if streamingOutput != "" {
		// In streaming mode, output is written to file
		result.Output = fmt.Sprintf("Results streamed to: %s", streamingOutput)
	} else {
		result.Output = strings.Join(outputs, "\n")
	}
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if err != nil {
		result.Status = core.StepStatusFailed
		result.Error = err
		return result, err
	}

	result.Status = core.StepStatusSuccess
	return result, nil
}

// LineIterator provides streaming access to lines in a file
type LineIterator struct {
	file    *os.File
	scanner *bufio.Scanner
	current string
	err     error
}

// NewLineIterator creates an iterator for reading lines from a file
func NewLineIterator(path string) (*LineIterator, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	// Increase scanner buffer for long lines
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)

	return &LineIterator{
		file:    file,
		scanner: scanner,
	}, nil
}

// Next advances to the next non-empty line, returns false when done
func (it *LineIterator) Next() bool {
	for it.scanner.Scan() {
		line := strings.TrimSpace(it.scanner.Text())
		if line != "" {
			it.current = line
			return true
		}
	}
	it.err = it.scanner.Err()
	return false
}

// Value returns the current line
func (it *LineIterator) Value() string {
	return it.current
}

// Err returns any error encountered during iteration
func (it *LineIterator) Err() error {
	return it.err
}

// Close closes the underlying file
func (it *LineIterator) Close() error {
	if it.file != nil {
		return it.file.Close()
	}
	return nil
}

// countInputLines counts non-empty lines in an input file (for result slice allocation)
func countInputLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()

	count := 0
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			count++
		}
	}
	return count, scanner.Err()
}

// renderSecondaryTemplates clones the step and renders [[ ]] templates with loop context
func (e *ForeachExecutor) renderSecondaryTemplates(step *core.Step, execCtx *core.ExecutionContext) *core.Step {
	// Clone the step to avoid modifying the original
	cloned := step.Clone()
	ctx := execCtx.GetVariables()

	// Render Command if it has secondary variables
	if e.templateEngine.HasSecondaryVariable(cloned.Command) {
		rendered, err := e.templateEngine.RenderSecondary(cloned.Command, ctx)
		if err == nil {
			cloned.Command = rendered
		}
	}

	// Render Commands array
	for i, cmd := range cloned.Commands {
		if e.templateEngine.HasSecondaryVariable(cmd) {
			rendered, err := e.templateEngine.RenderSecondary(cmd, ctx)
			if err == nil {
				cloned.Commands[i] = rendered
			}
		}
	}

	// Render Input (for nested foreach)
	if e.templateEngine.HasSecondaryVariable(cloned.Input) {
		rendered, err := e.templateEngine.RenderSecondary(cloned.Input, ctx)
		if err == nil {
			cloned.Input = rendered
		}
	}

	return cloned
}

// ForeachIterationError wraps a step execution error with context about the
// failing foreach iteration (input value, rendered command, captured output).
type ForeachIterationError struct {
	Inner           error
	InputValue      string // Loop variable value that failed
	RenderedCommand string // Fully rendered command
	Output          string // stdout+stderr from failed step
}

func (e *ForeachIterationError) Error() string { return e.Inner.Error() }
func (e *ForeachIterationError) Unwrap() error { return e.Inner }

// workItem represents a single item to process in the worker pool
type workItem struct {
	index int
	value string
}

// workResult represents the result of processing a work item
type workResult struct {
	index  int
	output string
	err    error
}

// streamWriter handles concurrent writes to an output file
type streamWriter struct {
	file *os.File
	mu   sync.Mutex
}

func newStreamWriter(path string) (*streamWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &streamWriter{file: f}, nil
}

func (w *streamWriter) Write(line string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, err := w.file.WriteString(line + "\n")
	return err
}

func (w *streamWriter) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// executeWithWorkerPool executes the inner step using a streaming worker pool pattern
// This is memory-efficient: creates only 'threads' goroutines instead of N goroutines
// and streams input lines on-demand instead of loading all into memory
func (e *ForeachExecutor) executeWithWorkerPool(ctx context.Context, step *core.Step, inputPath string, threads int, execCtx *core.ExecutionContext) ([]string, error) {
	// Count lines first for result slice allocation (fast, O(n) with minimal memory)
	lineCount, err := countInputLines(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to count input lines: %w", err)
	}

	if lineCount == 0 {
		return nil, nil
	}

	// Create bounded work queue - buffer 2x thread count for smooth flow
	workQueue := make(chan workItem, threads*2)
	results := make(chan workResult, threads*2)

	// Track completion
	var workerWg sync.WaitGroup
	var producerErr error

	// Start fixed worker pool (only 'threads' goroutines, not N)
	for i := 0; i < threads; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for work := range workQueue {
				// Check context cancellation
				if ctx.Err() != nil {
					results <- workResult{index: work.index, err: ctx.Err()}
					continue
				}

				// Apply variable pre-processing if configured
				loopValue := work.value
				if step.VariablePreProcess != "" {
					processedValue, err := e.preProcessVariable(step.VariablePreProcess, step.Variable, work.value, execCtx)
					if err != nil {
						// Log warning but continue with original value (fail-safe)
						logger.Get().Warn("variable pre-process failed, using original value",
							zap.String("expression", step.VariablePreProcess),
							zap.String("original_value", work.value),
							zap.Error(err))
					} else {
						loopValue = processedValue
					}
				}

				// Create optimized child context with loop variables pre-set
				childCtx := execCtx.CloneForLoop(step.Variable, loopValue, work.index+1)

				// Clone inner step and render secondary templates [[ ]]
				innerStep := e.renderSecondaryTemplates(step.Step, childCtx)

				// Execute inner step
				stepResult, err := e.dispatcher.Dispatch(ctx, innerStep, childCtx)

				var output string
				if stepResult != nil {
					output = stepResult.Output
				}

				// Wrap error with foreach iteration context
				if err != nil {
					renderedCmd := getInnerStepCommand(innerStep)
					if rendered, renderErr := e.templateEngine.Render(renderedCmd, childCtx.GetVariables()); renderErr == nil {
						renderedCmd = rendered
					}
					err = &ForeachIterationError{
						Inner:           err,
						InputValue:      loopValue,
						RenderedCommand: renderedCmd,
						Output:          output,
					}
				}

				results <- workResult{index: work.index, output: output, err: err}
			}
		}()
	}

	// Producer: stream lines into work queue (separate goroutine)
	go func() {
		defer close(workQueue)

		iter, err := NewLineIterator(inputPath)
		if err != nil {
			producerErr = err
			return
		}
		defer func() { _ = iter.Close() }()

		idx := 0
		for iter.Next() {
			select {
			case workQueue <- workItem{index: idx, value: iter.Value()}:
				idx++
			case <-ctx.Done():
				producerErr = ctx.Err()
				return
			}
		}

		if iter.Err() != nil {
			producerErr = iter.Err()
		}
	}()

	// Collector: close results when all workers done
	go func() {
		workerWg.Wait()
		close(results)
	}()

	// Collect results in order
	outputs := make([]string, lineCount)
	var firstError error

	for r := range results {
		if r.index < len(outputs) {
			outputs[r.index] = r.output
		}
		if r.err != nil && firstError == nil {
			firstError = r.err
		}
	}

	// Check for producer error
	if producerErr != nil && firstError == nil {
		firstError = producerErr
	}

	return outputs, firstError
}

// executeWithWorkerPoolStreaming is like executeWithWorkerPool but supports streaming output to file.
// When streamingOutput is set, results are written directly to the file instead of being collected in memory.
// This enables O(1) memory usage for million-line inputs.
func (e *ForeachExecutor) executeWithWorkerPoolStreaming(ctx context.Context, step *core.Step, inputPath string, threads int, execCtx *core.ExecutionContext, streamingOutput string) ([]string, error) {
	// If no streaming, delegate to original implementation
	if streamingOutput == "" {
		return e.executeWithWorkerPool(ctx, step, inputPath, threads, execCtx)
	}

	log := logger.Get()
	log.Debug("Foreach streaming mode enabled",
		zap.String("input", inputPath),
		zap.String("output", streamingOutput),
		zap.Int("threads", threads),
	)

	// Create streaming writer
	writer, err := newStreamWriter(streamingOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to create streaming output file: %w", err)
	}
	defer func() { _ = writer.Close() }()

	// Create bounded work queue
	workQueue := make(chan workItem, threads*2)
	done := make(chan struct{})

	// Track completion
	var workerWg sync.WaitGroup
	var producerErr error
	var writeErr error
	var writeErrMu sync.Mutex
	var firstErr error
	var firstErrMu sync.Mutex

	// Start fixed worker pool
	for i := 0; i < threads; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for work := range workQueue {
				// Check context cancellation
				if ctx.Err() != nil {
					continue
				}

				// Apply variable pre-processing if configured
				loopValue := work.value
				if step.VariablePreProcess != "" {
					processedValue, err := e.preProcessVariable(step.VariablePreProcess, step.Variable, work.value, execCtx)
					if err != nil {
						logger.Get().Warn("variable pre-process failed, using original value",
							zap.String("expression", step.VariablePreProcess),
							zap.String("original_value", work.value),
							zap.Error(err))
					} else {
						loopValue = processedValue
					}
				}

				// Create optimized child context with loop variables pre-set
				childCtx := execCtx.CloneForLoop(step.Variable, loopValue, work.index+1)

				// Clone inner step and render secondary templates [[ ]]
				innerStep := e.renderSecondaryTemplates(step.Step, childCtx)

				// Execute inner step
				stepResult, dispatchErr := e.dispatcher.Dispatch(ctx, innerStep, childCtx)

				// Track first error with foreach iteration context
				if dispatchErr != nil {
					firstErrMu.Lock()
					if firstErr == nil {
						renderedCmd := getInnerStepCommand(innerStep)
						if rendered, renderErr := e.templateEngine.Render(renderedCmd, childCtx.GetVariables()); renderErr == nil {
							renderedCmd = rendered
						}
						var output string
						if stepResult != nil {
							output = stepResult.Output
						}
						firstErr = &ForeachIterationError{
							Inner:           dispatchErr,
							InputValue:      loopValue,
							RenderedCommand: renderedCmd,
							Output:          output,
						}
					}
					firstErrMu.Unlock()
				}

				// Stream output directly to file (no memory collection)
				if stepResult != nil && stepResult.Output != "" {
					if err := writer.Write(stepResult.Output); err != nil {
						writeErrMu.Lock()
						if writeErr == nil {
							writeErr = err
						}
						writeErrMu.Unlock()
					}
				}
			}
		}()
	}

	// Producer: stream lines into work queue
	go func() {
		defer close(workQueue)

		iter, err := NewLineIterator(inputPath)
		if err != nil {
			producerErr = err
			return
		}
		defer func() { _ = iter.Close() }()

		idx := 0
		for iter.Next() {
			select {
			case workQueue <- workItem{index: idx, value: iter.Value()}:
				idx++
			case <-ctx.Done():
				producerErr = ctx.Err()
				return
			}
		}

		if iter.Err() != nil {
			producerErr = iter.Err()
		}
	}()

	// Wait for all workers to complete
	go func() {
		workerWg.Wait()
		close(done)
	}()

	<-done

	// Check for errors (priority: producer > write > dispatch)
	if producerErr != nil {
		return nil, producerErr
	}
	if writeErr != nil {
		return nil, fmt.Errorf("streaming write error: %w", writeErr)
	}
	if firstErr != nil {
		return nil, firstErr
	}

	// Return empty slice since results were streamed
	return nil, nil
}

// CanHandle returns true if this executor can handle the given step type
func (e *ForeachExecutor) CanHandle(stepType core.StepType) bool {
	return stepType == core.StepTypeForeach
}

// autoQuoteForJS wraps string values in single quotes for JS function calls,
// escaping any internal single quotes.
func autoQuoteForJS(value string) string {
	// Escape single quotes: ' -> \'
	escaped := strings.ReplaceAll(value, "'", "\\'")
	return "'" + escaped + "'"
}

// preProcessVariable evaluates the pre-process expression with the loop variable set.
// The expression can use [[variable]] syntax to reference the current value.
// For example: "get_parent_url([[url]])" with url="http://example.com/path"
// becomes: "get_parent_url('http://example.com/path')"
func (e *ForeachExecutor) preProcessVariable(expr, varName, varValue string, execCtx *core.ExecutionContext) (string, error) {
	// Build context with parent variables
	ctx := execCtx.GetVariables()

	// For pre-process expressions, auto-quote the loop variable value
	// This allows clean syntax: get_parent_url([[url]]) instead of get_parent_url('[[url]]')
	ctx[varName] = autoQuoteForJS(varValue)

	// Render secondary templates [[var]] -> 'quoted_value'
	renderedExpr, err := e.templateEngine.RenderSecondary(expr, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to render pre-process expression: %w", err)
	}

	// Execute the function expression
	result, err := e.functionRegistry.Execute(renderedExpr, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to execute pre-process function: %w", err)
	}

	// Convert result to string
	return fmt.Sprintf("%v", result), nil
}

// readTruncatedFile reads up to maxBytes from a file, appending "[output truncated]" if the file is larger.
func readTruncatedFile(path string, maxBytes int) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()

	buf := make([]byte, maxBytes+1) // read 1 extra byte to detect truncation
	n, _ := f.Read(buf)
	if n == 0 {
		return ""
	}
	if n > maxBytes {
		return string(buf[:maxBytes]) + "\n[output truncated]"
	}
	return string(buf[:n])
}
