package activities

import (
	"fmt"
)

// ErrorCode represents a unique identifier for each type of activity error
type ErrorCode string

const (
	// Common activity errors
	ErrInvalidArguments ErrorCode = "INVALID_ARGUMENTS"
	ErrExecutionFailed  ErrorCode = "EXECUTION_FAILED"
	ErrPanic            ErrorCode = "PANIC"

	// HTTP specific errors
	ErrHTTPRequestFailed  ErrorCode = "HTTP_REQUEST_FAILED"
	ErrHTTPResponseFailed ErrorCode = "HTTP_RESPONSE_FAILED"
	ErrHTTPStatusError    ErrorCode = "HTTP_STATUS_ERROR"

	// JQ specific errors
	ErrJQParseError   ErrorCode = "JQ_PARSE_ERROR"
	ErrJQExecuteError ErrorCode = "JQ_EXECUTE_ERROR"

	// JWE specific errors
	ErrJWEEncryptError ErrorCode = "JWE_ENCRYPT_ERROR"
)

// ActivityError represents a structured error from an activity
type ActivityError struct {
	// Code is a unique identifier for the error type
	Code ErrorCode `json:"code"`

	// Message is a human-readable description of the error
	Message string `json:"message"`

	// ActivityName is the name of the activity that generated the error
	ActivityName string `json:"activityName"`

	// Arguments contains any relevant arguments that led to the error
	Arguments map[string]interface{} `json:"arguments,omitempty"`

	// Cause is the underlying error that caused this error (if any)
	Cause error `json:"cause,omitempty"`
}

func (e *ActivityError) Error() string {
	msg := fmt.Sprintf("[%s] %s (Activity: %s)", e.Code, e.Message, e.ActivityName)
	if e.Cause != nil {
		msg += fmt.Sprintf(" - caused by: %v", e.Cause)
	}
	return msg
}

// NewActivityError creates a new ActivityError
func NewActivityError(code ErrorCode, message string, activityName string) *ActivityError {
	return &ActivityError{
		Code:         code,
		Message:      message,
		ActivityName: activityName,
	}
}

// WithCause adds an underlying cause to the error
func (e *ActivityError) WithCause(err error) *ActivityError {
	e.Cause = err
	return e
}

// WithArguments adds argument context to the error
func (e *ActivityError) WithArguments(args map[string]interface{}) *ActivityError {
	e.Arguments = args
	return e
}
