// Package lib provides the Osmedeus SDK for programmatic workflow execution.
package lib

import (
	"errors"
	"fmt"
)

// Sentinel errors for common validation failures
var (
	// ErrEmptyTarget indicates that the target cannot be empty
	ErrEmptyTarget = errors.New("target cannot be empty")

	// ErrEmptyWorkflow indicates that the workflow content cannot be empty
	ErrEmptyWorkflow = errors.New("workflow content cannot be empty")

	// ErrEmptyExpression indicates that the expression cannot be empty
	ErrEmptyExpression = errors.New("expression cannot be empty")

	// ErrNotModule indicates that the workflow must be of kind 'module'
	ErrNotModule = errors.New("workflow must be of kind 'module' (flows not supported in library mode)")
)

// ParseError wraps YAML parsing errors with additional context
type ParseError struct {
	Message string
	Err     error
}

// Error implements the error interface
func (e *ParseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("parse error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("parse error: %s", e.Message)
}

// Unwrap returns the underlying error
func (e *ParseError) Unwrap() error {
	return e.Err
}

// NewParseError creates a new ParseError
func NewParseError(message string, err error) *ParseError {
	return &ParseError{
		Message: message,
		Err:     err,
	}
}

// ValidationError wraps workflow validation errors
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// Unwrap returns the underlying error
func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewValidationErrorWithCause creates a new ValidationError with an underlying cause
func NewValidationErrorWithCause(field, message string, err error) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// ExecutionError wraps execution errors with step information
type ExecutionError struct {
	StepName string
	StepType string
	Message  string
	Err      error
}

// Error implements the error interface
func (e *ExecutionError) Error() string {
	if e.StepName != "" {
		return fmt.Sprintf("execution error at step '%s' (%s): %s", e.StepName, e.StepType, e.Message)
	}
	return fmt.Sprintf("execution error: %s", e.Message)
}

// Unwrap returns the underlying error
func (e *ExecutionError) Unwrap() error {
	return e.Err
}

// NewExecutionError creates a new ExecutionError
func NewExecutionError(stepName, stepType, message string) *ExecutionError {
	return &ExecutionError{
		StepName: stepName,
		StepType: stepType,
		Message:  message,
	}
}

// NewExecutionErrorWithCause creates a new ExecutionError with an underlying cause
func NewExecutionErrorWithCause(stepName, stepType, message string, err error) *ExecutionError {
	return &ExecutionError{
		StepName: stepName,
		StepType: stepType,
		Message:  message,
		Err:      err,
	}
}

// IsParseError checks if the error is a ParseError
func IsParseError(err error) bool {
	var parseErr *ParseError
	return errors.As(err, &parseErr)
}

// IsValidationError checks if the error is a ValidationError
func IsValidationError(err error) bool {
	var validErr *ValidationError
	return errors.As(err, &validErr)
}

// IsExecutionError checks if the error is an ExecutionError
func IsExecutionError(err error) bool {
	var execErr *ExecutionError
	return errors.As(err, &execErr)
}
