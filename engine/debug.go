package engine

import (
	"time"
)

// ExecutionResult contains the final result and execution details
type ExecutionResult struct {
	Data     interface{}     `json:"data"`
	Debug    *ExecutionDebug `json:"debug,omitempty"`
	Duration time.Duration   `json:"duration"`
}

// ExecutionDebug contains debugging information about workflow execution
type ExecutionDebug struct {
	States []StateExecution `json:"states"`
}

// StateExecution represents a single state execution
type StateExecution struct {
	Name             string         `json:"name"`
	Type             string         `json:"type"`
	StartTime        time.Time      `json:"startTime"`
	EndTime          time.Time      `json:"endTime"`
	Input            interface{}    `json:"input,omitempty"`
	Output           interface{}    `json:"output,omitempty"`
	Error            string         `json:"error,omitempty"`
	Actions          []ActionResult `json:"actions,omitempty"`
	MatchedCondition string         `json:"matchedCondition,omitempty"` // For switch states
	SleepDuration    string         `json:"sleepDuration,omitempty"`    // For sleep states
}

// ActionResult represents the result of a single action execution
type ActionResult struct {
	ActivityName string      `json:"activityName"`
	Arguments    interface{} `json:"arguments"`
	StartTime    time.Time   `json:"startTime"`
	EndTime      time.Time   `json:"endTime"`
	Output       interface{} `json:"output"`
	Error        string      `json:"error,omitempty"`
}
