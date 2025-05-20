package engine

import (
	"errors"
	"fmt"

	"github.com/kshitiz1403/jsonjuggler/activities"
)

// ErrorCode represents a unique identifier for each type of error
type ErrorCode string

const (
	// Workflow Errors
	ErrWorkflowInvalid ErrorCode = "WORKFLOW_INVALID"

	// State Errors
	ErrStateNotFound       ErrorCode = "STATE_NOT_FOUND"
	ErrStateInvalid        ErrorCode = "STATE_INVALID"
	ErrStateExecutionFail  ErrorCode = "STATE_EXECUTION_FAILED"
	ErrStateTransitionFail ErrorCode = "STATE_TRANSITION_FAILED"

	// Activity Errors
	ErrActivityNotFound   ErrorCode = "ACTIVITY_NOT_FOUND"
	ErrActivityArgInvalid ErrorCode = "ACTIVITY_ARGS_INVALID"
	ErrActivityExecution  ErrorCode = "ACTIVITY_EXECUTION_FAILED"

	// Data Errors
	ErrDataTransform      ErrorCode = "DATA_TRANSFORM_FAILED"
	ErrDataValidation     ErrorCode = "DATA_VALIDATION_FAILED"
	ErrDataTypeConversion ErrorCode = "DATA_TYPE_CONVERSION_FAILED"

	// Expression Errors
	ErrExpressionInvalid ErrorCode = "EXPRESSION_INVALID"
	ErrExpressionEval    ErrorCode = "EXPRESSION_EVAL_FAILED"
)

// WorkflowError represents a structured error in the workflow engine
type WorkflowError struct {
	// Code is a unique identifier for the error type
	Code ErrorCode `json:"code"`

	// Message is a human-readable description of the error
	Message string `json:"message"`

	// Context contains additional information about where/why the error occurred
	Context ErrorContext `json:"context"`

	// Cause is the underlying error that caused this error (if any)
	Cause error `json:"cause,omitempty"`
}

// ErrorContext provides additional context about where the error occurred
type ErrorContext struct {
	// WorkflowID is the ID of the workflow where the error occurred
	WorkflowID string `json:"workflowId,omitempty"`

	// StateName is the name of the state where the error occurred
	StateName string `json:"stateName,omitempty"`

	// ActivityName is the name of the activity where the error occurred
	ActivityName string `json:"activityName,omitempty"`

	// Expression is the expression that failed (if applicable)
	Expression string `json:"expression,omitempty"`

	// Arguments contains any relevant arguments that led to the error
	Arguments map[string]interface{} `json:"arguments,omitempty"`

	// AdditionalInfo contains any other relevant information
	AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
}

func (e *WorkflowError) Error() string {
	msg := fmt.Sprintf("[%s] %s", e.Code, e.Message)
	if e.Context.WorkflowID != "" {
		msg += fmt.Sprintf(" (Workflow: %s)", e.Context.WorkflowID)
	}
	if e.Context.StateName != "" {
		msg += fmt.Sprintf(" (State: %s)", e.Context.StateName)
	}
	if e.Context.ActivityName != "" {
		msg += fmt.Sprintf(" (Activity: %s)", e.Context.ActivityName)
	}
	if e.Cause != nil {
		msg += fmt.Sprintf(" - caused by: %v", e.Cause)
	}
	return msg
}

// NewWorkflowError creates a new WorkflowError with the given code and message
func NewWorkflowError(code ErrorCode, message string) *WorkflowError {
	return &WorkflowError{
		Code:    code,
		Message: message,
		Context: ErrorContext{},
	}
}

// WithState adds state context to the error
func (e *WorkflowError) WithState(stateName string) *WorkflowError {
	e.Context.StateName = stateName
	return e
}

// WithActivity adds activity context to the error
func (e *WorkflowError) WithActivity(activityName string) *WorkflowError {
	e.Context.ActivityName = activityName
	return e
}

// WithWorkflow adds workflow context to the error
func (e *WorkflowError) WithWorkflow(workflowID string) *WorkflowError {
	e.Context.WorkflowID = workflowID
	return e
}

// WithCause adds an underlying cause to the error
func (e *WorkflowError) WithCause(err error) *WorkflowError {
	e.Cause = err
	return e
}

// WithContext adds additional context to the error
func (e *WorkflowError) WithContext(key string, value interface{}) *WorkflowError {
	if e.Context.AdditionalInfo == nil {
		e.Context.AdditionalInfo = make(map[string]interface{})
	}
	e.Context.AdditionalInfo[key] = value
	return e
}

// WithArguments adds argument context to the error
func (e *WorkflowError) WithArguments(args map[string]interface{}) *WorkflowError {
	e.Context.Arguments = args
	return e
}

// IsWorkflowError checks if an error is a WorkflowError
func IsWorkflowError(err error) bool {
	_, ok := err.(*WorkflowError)
	return ok
}

// GetErrorCode extracts the error code from an error if it's a WorkflowError
func GetErrorCode(err error) (ErrorCode, bool) {
	if wErr, ok := err.(*WorkflowError); ok {
		return wErr.Code, true
	}
	return "", false
}

// UnwrapActivityError extracts the ActivityError if present in the error chain
func UnwrapActivityError(err error) (*activities.ActivityError, bool) {
	var actErr *activities.ActivityError
	if errors.As(err, &actErr) {
		return actErr, true
	}
	return nil, false
}

// WithActivityError wraps an ActivityError in a WorkflowError while preserving the activity context
func NewWorkflowErrorFromActivity(code ErrorCode, message string, actErr *activities.ActivityError) *WorkflowError {
	return &WorkflowError{
		Code:    code,
		Message: message,
		Context: ErrorContext{
			ActivityName: actErr.ActivityName,
			Arguments:    actErr.Arguments,
		},
		Cause: actErr,
	}
}
