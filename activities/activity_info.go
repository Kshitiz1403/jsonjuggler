package activities

import (
	"context"

	"github.com/kshitiz1403/jsonjuggler/logger"
)

// ActivityInfo contains information about the currently executing activity
type ActivityInfo struct {
	// ActivityName is the name of the activity
	ActivityName string

	// Logger is the logger instance for this activity
	Logger logger.Logger

	// Additional fields can be added here as needed
	// WorkflowID string
	// RunID string
	// StartedTime time.Time
	// etc.
}

// activityInfoContextKey is the context key for storing activity info
type activityInfoContextKey struct{}

// GetInfo retrieves the activity info from the context
// Similar to Temporal's activity.GetInfo(ctx)
func GetInfo(ctx context.Context) *ActivityInfo {
	if info, ok := ctx.Value(activityInfoContextKey{}).(*ActivityInfo); ok {
		return info
	}
	// Return a default ActivityInfo if not found in context
	return &ActivityInfo{
		ActivityName: "unknown",
		Logger:       nil,
	}
}

// withActivityInfo adds activity info to the context
func withActivityInfo(ctx context.Context, info *ActivityInfo) context.Context {
	return context.WithValue(ctx, activityInfoContextKey{}, info)
}

// HasActivityInfo checks if the context contains activity info
func HasActivityInfo(ctx context.Context) bool {
	_, ok := ctx.Value(activityInfoContextKey{}).(*ActivityInfo)
	return ok
}
