package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

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

// Execute executes a parallel step
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

	type stepResult struct {
		index  int
		result *core.StepResult
		err    error
	}

	results := make(chan stepResult, len(step.ParallelSteps))
	var wg sync.WaitGroup

	for i := range step.ParallelSteps {
		wg.Add(1)
		go func(idx int, s *core.Step) {
			defer wg.Done()

			// Check if context is cancelled before starting
			select {
			case <-ctx.Done():
				results <- stepResult{index: idx, err: ctx.Err()}
				return
			default:
			}

			// Clone context for parallel execution
			childCtx := execCtx.Clone()
			r, err := e.dispatcher.Dispatch(ctx, s, childCtx)

			// Send result (use select to handle cancelled context)
			select {
			case results <- stepResult{index: idx, result: r, err: err}:
			case <-ctx.Done():
				// Context cancelled, still need to send a result
				results <- stepResult{index: idx, result: r, err: ctx.Err()}
			}
		}(i, &step.ParallelSteps[i])
	}

	// Wait for all steps to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results with context awareness
	stepResults := make([]*core.StepResult, len(step.ParallelSteps))
	var outputs []string
	var firstError error
	collected := 0

	for collected < len(step.ParallelSteps) {
		select {
		case r, ok := <-results:
			if !ok {
				// Channel closed
				goto done
			}
			collected++
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
		case <-ctx.Done():
			// Context cancelled - set error and wait for remaining results
			if firstError == nil {
				firstError = ctx.Err()
			}
		}
	}

done:
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
