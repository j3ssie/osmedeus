package lib

import (
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// RunResult holds the result of a workflow execution
type RunResult struct {
	// WorkflowName is the name of the executed workflow
	WorkflowName string

	// RunUUID is the unique identifier for this execution
	RunUUID string

	// Target is the scan target
	Target string

	// Status indicates the final execution status
	// Values: "completed", "failed", "cancelled", "skipped"
	Status string

	// StartTime is when execution started
	StartTime time.Time

	// EndTime is when execution finished
	EndTime time.Time

	// Duration is the total execution time
	Duration time.Duration

	// Steps contains results for each executed step
	Steps []*StepResult

	// Exports are variables exported by steps during execution
	Exports map[string]interface{}

	// Artifacts are file paths of generated outputs
	Artifacts []string

	// Error contains the error if execution failed
	Error error

	// OutputPath is the workspace directory path
	OutputPath string

	// Message contains optional status message (e.g., for skipped status)
	Message string
}

// IsSuccess returns true if the workflow completed successfully
func (r *RunResult) IsSuccess() bool {
	return r.Status == "completed"
}

// IsFailed returns true if the workflow failed
func (r *RunResult) IsFailed() bool {
	return r.Status == "failed"
}

// IsCancelled returns true if the workflow was cancelled
func (r *RunResult) IsCancelled() bool {
	return r.Status == "cancelled"
}

// IsSkipped returns true if the workflow was skipped
func (r *RunResult) IsSkipped() bool {
	return r.Status == "skipped"
}

// GetExport retrieves an exported variable by name
func (r *RunResult) GetExport(name string) (interface{}, bool) {
	if r.Exports == nil {
		return nil, false
	}
	v, ok := r.Exports[name]
	return v, ok
}

// GetExportString retrieves an exported variable as a string
func (r *RunResult) GetExportString(name string) (string, bool) {
	v, ok := r.GetExport(name)
	if !ok {
		return "", false
	}
	if s, ok := v.(string); ok {
		return s, true
	}
	return "", false
}

// GetExportBool retrieves an exported variable as a bool
func (r *RunResult) GetExportBool(name string) (bool, bool) {
	v, ok := r.GetExport(name)
	if !ok {
		return false, false
	}
	if b, ok := v.(bool); ok {
		return b, true
	}
	return false, false
}

// SuccessfulSteps returns the count of successful steps
func (r *RunResult) SuccessfulSteps() int {
	count := 0
	for _, step := range r.Steps {
		if step.Status == "success" {
			count++
		}
	}
	return count
}

// FailedSteps returns the count of failed steps
func (r *RunResult) FailedSteps() int {
	count := 0
	for _, step := range r.Steps {
		if step.Status == "failed" {
			count++
		}
	}
	return count
}

// SkippedSteps returns the count of skipped steps
func (r *RunResult) SkippedSteps() int {
	count := 0
	for _, step := range r.Steps {
		if step.Status == "skipped" {
			count++
		}
	}
	return count
}

// StepResult holds the result of a single step execution
type StepResult struct {
	// Name is the step name from the workflow
	Name string

	// Type is the step type (bash, function, etc.)
	Type string

	// Status indicates the step execution status
	// Values: "success", "failed", "skipped"
	Status string

	// Output is the captured stdout/stderr from the step
	Output string

	// Duration is how long the step took to execute
	Duration time.Duration

	// Error contains the error if the step failed
	Error error

	// Exports are variables exported by this step
	Exports map[string]interface{}
}

// IsSuccess returns true if the step completed successfully
func (s *StepResult) IsSuccess() bool {
	return s.Status == "success"
}

// IsFailed returns true if the step failed
func (s *StepResult) IsFailed() bool {
	return s.Status == "failed"
}

// IsSkipped returns true if the step was skipped
func (s *StepResult) IsSkipped() bool {
	return s.Status == "skipped"
}

// GetExport retrieves an exported variable by name
func (s *StepResult) GetExport(name string) (interface{}, bool) {
	if s.Exports == nil {
		return nil, false
	}
	v, ok := s.Exports[name]
	return v, ok
}

// fromWorkflowResult converts an internal WorkflowResult to a RunResult
func fromWorkflowResult(result *core.WorkflowResult, outputPath string) *RunResult {
	if result == nil {
		return nil
	}

	runResult := &RunResult{
		WorkflowName: result.WorkflowName,
		RunUUID:      result.RunUUID,
		Target:       result.Target,
		Status:       string(result.Status),
		StartTime:    result.StartTime,
		EndTime:      result.EndTime,
		Duration:     result.EndTime.Sub(result.StartTime),
		Exports:      result.Exports,
		Artifacts:    result.Artifacts,
		Error:        result.Error,
		OutputPath:   outputPath,
		Message:      result.Message,
	}

	// Convert step results
	if len(result.Steps) > 0 {
		runResult.Steps = make([]*StepResult, len(result.Steps))
		for i, step := range result.Steps {
			runResult.Steps[i] = fromStepResult(step)
		}
	}

	return runResult
}

// fromStepResult converts an internal StepResult to a lib StepResult
func fromStepResult(step *core.StepResult) *StepResult {
	if step == nil {
		return nil
	}

	return &StepResult{
		Name:     step.StepName,
		Type:     "", // Type is not stored in core.StepResult
		Status:   string(step.Status),
		Output:   step.Output,
		Duration: step.Duration,
		Error:    step.Error,
		Exports:  step.Exports,
	}
}
