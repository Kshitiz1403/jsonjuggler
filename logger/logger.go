package logger

import (
	"context"
)

// Logger interface defines the logging methods
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})

	// Context-aware logging methods
	DebugContext(ctx context.Context, args ...interface{})
	DebugContextf(ctx context.Context, format string, args ...interface{})
	InfoContext(ctx context.Context, args ...interface{})
	InfoContextf(ctx context.Context, format string, args ...interface{})
	WarnContext(ctx context.Context, args ...interface{})
	WarnContextf(ctx context.Context, format string, args ...interface{})
	ErrorContext(ctx context.Context, args ...interface{})
	ErrorContextf(ctx context.Context, format string, args ...interface{})
	FatalContext(ctx context.Context, args ...interface{})
	FatalContextf(ctx context.Context, format string, args ...interface{})
}
