package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/functions"
	"github.com/j3ssie/osmedeus/v5/internal/template"
)

// FunctionExecutor executes function steps
type FunctionExecutor struct {
	templateEngine   template.TemplateEngine
	functionRegistry *functions.Registry
}

// NewFunctionExecutor creates a new function executor
func NewFunctionExecutor(engine template.TemplateEngine, registry *functions.Registry) *FunctionExecutor {
	return &FunctionExecutor{
		templateEngine:   engine,
		functionRegistry: registry,
	}
}

// Name returns the executor name for logging/debugging
func (e *FunctionExecutor) Name() string {
	return "function"
}

// StepTypes returns the step types this executor handles
func (e *FunctionExecutor) StepTypes() []core.StepType {
	return []core.StepType{core.StepTypeFunction}
}

// Execute executes a function step
func (e *FunctionExecutor) Execute(ctx context.Context, step *core.Step, execCtx *core.ExecutionContext) (*core.StepResult, error) {
	result := &core.StepResult{
		StepName:  step.Name,
		Status:    core.StepStatusRunning,
		StartTime: time.Now(),
		Exports:   make(map[string]interface{}),
	}

	vars := execCtx.GetVariables()
	if step.SuppressDetails {
		vars["SuppressDetails"] = true
	}
	var outputs []interface{}
	var err error

	// Determine execution mode
	if len(step.ParallelFunctions) > 0 {
		outputs, err = e.executeParallel(ctx, step.ParallelFunctions, vars)
	} else if len(step.Functions) > 0 {
		outputs, err = e.executeSequential(ctx, step.Functions, vars)
	} else if step.Function != "" {
		var output interface{}
		output, err = e.executeFunction(ctx, step.Function, vars)
		outputs = []interface{}{output}
	} else {
		err = fmt.Errorf("no function specified")
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if err != nil {
		result.Status = core.StepStatusFailed
		result.Error = err
		return result, err
	}

	// Convert outputs to string
	if len(outputs) > 0 {
		result.Output = fmt.Sprintf("%v", outputs[0])
	}

	result.Status = core.StepStatusSuccess
	return result, nil
}

// executeFunction executes a single function
func (e *FunctionExecutor) executeFunction(ctx context.Context, expr string, vars map[string]interface{}) (interface{}, error) {
	return e.functionRegistry.Execute(expr, vars)
}

// executeSequential executes functions sequentially
func (e *FunctionExecutor) executeSequential(ctx context.Context, funcs []string, vars map[string]interface{}) ([]interface{}, error) {
	var outputs []interface{}

	for _, fn := range funcs {
		select {
		case <-ctx.Done():
			return outputs, ctx.Err()
		default:
		}

		output, err := e.executeFunction(ctx, fn, vars)
		if err != nil {
			return outputs, err
		}
		outputs = append(outputs, output)
	}

	return outputs, nil
}

// executeParallel executes functions in parallel
func (e *FunctionExecutor) executeParallel(ctx context.Context, funcs []string, vars map[string]interface{}) ([]interface{}, error) {
	type result struct {
		index  int
		output interface{}
		err    error
	}

	results := make(chan result, len(funcs))
	var wg sync.WaitGroup

	for i, fn := range funcs {
		wg.Add(1)
		go func(idx int, expr string) {
			defer wg.Done()
			output, err := e.executeFunction(ctx, expr, vars)
			results <- result{index: idx, output: output, err: err}
		}(i, fn)
	}

	// Wait for all functions to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	outputs := make([]interface{}, len(funcs))
	var firstError error

	for r := range results {
		outputs[r.index] = r.output
		if r.err != nil && firstError == nil {
			firstError = r.err
		}
	}

	return outputs, firstError
}

// CanHandle returns true if this executor can handle the given step type
func (e *FunctionExecutor) CanHandle(stepType core.StepType) bool {
	return stepType == core.StepTypeFunction
}
