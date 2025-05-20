package logger

// LogLevel represents the log level
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

type customLogContextKeys string

const (
	StateNameKey    customLogContextKeys = "stateName"
	StateTypeKey    customLogContextKeys = "stateType"
	ActivityNameKey customLogContextKeys = "activityName"
	WorkflowIDKey   customLogContextKeys = "workflowId"
)
