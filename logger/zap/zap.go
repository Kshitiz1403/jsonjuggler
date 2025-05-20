package zap

import (
	"context"
	"time"

	"github.com/kshitiz1403/jsonjuggler/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger implements Logger interface with zap
type zapLogger struct {
	*zap.SugaredLogger
	baseLogger *zap.Logger
}

// NewLogger creates a new Zap-based logger
func NewLogger(level logger.LogLevel) logger.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Enable caller tracking in logs
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Set the log level
	switch level {
	case logger.DebugLevel:
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case logger.InfoLevel:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case logger.WarnLevel:
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case logger.ErrorLevel:
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case logger.FatalLevel:
		config.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	baseLogger, _ := config.Build()
	sugaredLogger := baseLogger.Sugar()

	return &zapLogger{
		SugaredLogger: sugaredLogger,
		baseLogger:    baseLogger,
	}
}

// Implementation of context-aware logging methods
func (l *zapLogger) DebugContext(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Debug(args...)
}

func (l *zapLogger) DebugContextf(ctx context.Context, format string, args ...interface{}) {
	l.withContext(ctx).Debugf(format, args...)
}

func (l *zapLogger) InfoContext(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Info(args...)
}

func (l *zapLogger) InfoContextf(ctx context.Context, format string, args ...interface{}) {
	l.withContext(ctx).Infof(format, args...)
}

func (l *zapLogger) WarnContext(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Warn(args...)
}

func (l *zapLogger) WarnContextf(ctx context.Context, format string, args ...interface{}) {
	l.withContext(ctx).Warnf(format, args...)
}

func (l *zapLogger) ErrorContext(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Error(args...)
}

func (l *zapLogger) ErrorContextf(ctx context.Context, format string, args ...interface{}) {
	l.withContext(ctx).Errorf(format, args...)
}

func (l *zapLogger) FatalContext(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Fatal(args...)
}

func (l *zapLogger) FatalContextf(ctx context.Context, format string, args ...interface{}) {
	l.withContext(ctx).Fatalf(format, args...)
}

// withContext creates a new logger with fields from context
func (l *zapLogger) withContext(ctx context.Context) *zapLogger {
	fields := logger.GetFields(ctx)
	zapFields := toZapFields(fields)

	// AddCallerSkip(1) is crucial - it tells zap to skip 1 stack frame when determining the caller.
	// Without this, logs would show this file (logger.go) as the caller instead of the actual calling code.
	// We skip 1 frame to account for our wrapper methods (Debug, Info, Error etc.)

	// Use baseLogger and add caller skip for context methods
	newLogger := l.baseLogger.WithOptions(zap.AddCallerSkip(1)).With(zapFields...)
	return &zapLogger{
		SugaredLogger: newLogger.Sugar(),
		baseLogger:    newLogger,
	}
}

// toZapFields converts our Fields to zap fields
func toZapFields(fields []logger.Field) []zapcore.Field {
	if len(fields) == 0 {
		return nil
	}

	zapFields := make([]zapcore.Field, len(fields))
	for i, f := range fields {
		switch f.Type {
		case logger.StringType:
			zapFields[i] = zap.String(f.Key, f.String)
		case logger.IntType, logger.Int64Type:
			zapFields[i] = zap.Int64(f.Key, f.Int64)
		case logger.Float64Type:
			zapFields[i] = zap.Float64(f.Key, f.Float64)
		case logger.BoolType:
			zapFields[i] = zap.Bool(f.Key, f.Int64 == 1)
		case logger.TimeType:
			zapFields[i] = zap.Time(f.Key, f.Interface.(time.Time))
		case logger.DurationType:
			zapFields[i] = zap.Duration(f.Key, time.Duration(f.Int64))
		case logger.ErrorType:
			zapFields[i] = zap.Error(f.Interface.(error))
		default:
			zapFields[i] = zap.Any(f.Key, f.Interface)
		}
	}
	return zapFields
}
