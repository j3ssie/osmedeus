package executor

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// parallelWorkItem represents a single parallel step to execute
type parallelWorkItem struct {
	index int
	step  *core.Step
}

// parallelWorkResult represents the result of executing a parallel step
type parallelWorkResult struct {
	index  int
	result *core.StepResult
	err    error
}

// ParallelExecutor executes parallel steps
type ParallelExecutor struct {
	dispatcher *StepDispatcher
}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor(dispatcher *StepDispatcher) *ParallelExecutor {
	return &ParallelExecutor{
		dispatcher: dispatcher,
	}
}

// Name returns the executor name for logging/debugging
func (e *ParallelExecutor) Name() string {
	return "parallel"
}

// StepTypes returns the step types this executor handles
func (e *ParallelExecutor) StepTypes() []core.StepType {
	return []core.StepType{core.StepTypeParallel}
}

// Execute executes a parallel step using a bounded worker pool
// to prevent resource exhaustion when executing many parallel steps
func (e *ParallelExecutor) Execute(ctx context.Context, step *core.Step, execCtx *core.ExecutionContext) (*core.StepResult, error) {
	result := &core.StepResult{
		StepName:  step.Name,
		Status:    core.StepStatusRunning,
		StartTime: time.Now(),
		Exports:   make(map[string]interface{}),
	}

	if len(step.ParallelSteps) == 0 {
		result.Status = core.StepStatusSuccess
		result.EndTime = time.Now()
		return result, nil
	}

	// Check if context is already cancelled
	if ctx.Err() != nil {
		result.Status = core.StepStatusFailed
		result.Error = ctx.Err()
		result.EndTime = time.Now()
		return result, ctx.Err()
	}

	numSteps := len(step.ParallelSteps)

	// Bounded worker pool: NumCPU * 2 workers (or fewer if less steps)
	numWorkers := runtime.NumCPU() * 2
	if numWorkers > numSteps {
		numWorkers = numSteps
	}

	// Bounded channels for backpressure
	workQueue := make(chan parallelWorkItem, numWorkers*2)
	results := make(chan parallelWorkResult, numWorkers*2)

	var workerWg sync.WaitGroup

	// Start fixed worker pool
	for i := 0; i < numWorkers; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for work := range workQueue {
				// Check context before execution
				if ctx.Err() != nil {
					results <- parallelWorkResult{index: work.index, err: ctx.Err()}
					continue
				}
				childCtx := execCtx.Clone()
				r, err := e.dispatcher.Dispatch(ctx, work.step, childCtx)
				results <- parallelWorkResult{index: work.index, result: r, err: err}
			}
		}()
	}

	// Producer: enqueue all work items
	go func() {
		defer close(workQueue)
		for i := range step.ParallelSteps {
			select {
			case workQueue <- parallelWorkItem{index: i, step: &step.ParallelSteps[i]}:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Collector: close results channel when all workers finish
	go func() {
		workerWg.Wait()
		close(results)
	}()

	// Collect results
	stepResults := make([]*core.StepResult, numSteps)
	var outputs []string
	var firstError error

	for r := range results {
		stepResults[r.index] = r.result
		if r.result != nil && r.result.Output != "" {
			outputs = append(outputs, r.result.Output)
		}
		if r.err != nil && firstError == nil {
			firstError = r.err
		}
		// Merge exports
		if r.result != nil && r.result.Exports != nil {
			for k, v := range r.result.Exports {
				result.Exports[k] = v
			}
		}
	}

	result.Output = strings.Join(outputs, "\n")
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if firstError != nil {
		result.Status = core.StepStatusFailed
		result.Error = firstError
		return result, firstError
	}

	// Check if any step failed
	for _, sr := range stepResults {
		if sr != nil && sr.Status == core.StepStatusFailed {
			result.Status = core.StepStatusFailed
			result.Error = fmt.Errorf("one or more parallel steps failed")
			return result, result.Error
		}
	}

	result.Status = core.StepStatusSuccess
	return result, nil
}

// CanHandle returns true if this executor can handle the given step type
func (e *ParallelExecutor) CanHandle(stepType core.StepType) bool {
	return stepType == core.StepTypeParallel
}
