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
	"github.com/j3ssie/osmedeus/v5/internal/template"
)

// ForeachExecutor executes foreach steps
type ForeachExecutor struct {
	dispatcher     *StepDispatcher
	templateEngine *template.Engine
}

// NewForeachExecutor creates a new foreach executor
func NewForeachExecutor(dispatcher *StepDispatcher, engine *template.Engine) *ForeachExecutor {
	return &ForeachExecutor{
		dispatcher:     dispatcher,
		templateEngine: engine,
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

	// Execute with streaming worker pool
	outputs, err := e.executeWithWorkerPool(ctx, step, step.Input, threads, execCtx)

	result.Output = strings.Join(outputs, "\n")
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

				// Create optimized child context with loop variables pre-set
				childCtx := execCtx.CloneForLoop(step.Variable, work.value, work.index+1)

				// Clone inner step and render secondary templates [[ ]]
				innerStep := e.renderSecondaryTemplates(step.Step, childCtx)

				// Execute inner step
				stepResult, err := e.dispatcher.Dispatch(ctx, innerStep, childCtx)

				var output string
				if stepResult != nil {
					output = stepResult.Output
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

// CanHandle returns true if this executor can handle the given step type
func (e *ForeachExecutor) CanHandle(stepType core.StepType) bool {
	return stepType == core.StepTypeForeach
}
