package logger

import "context"

// contextKey is used to store logger fields in context
type contextKey struct{}

// WithFields stores fields in context for later logging
func WithFields(ctx context.Context, fields ...Field) context.Context {
	if existingFields, ok := ctx.Value(contextKey{}).([]Field); ok {
		fields = append(existingFields, fields...)
	}
	return context.WithValue(ctx, contextKey{}, fields)
}

// GetFields retrieves fields from context
func GetFields(ctx context.Context) []Field {
	if ctx == nil {
		return nil
	}

	var fields []Field
	if contextFields, ok := ctx.Value(contextKey{}).([]Field); ok {
		fields = append(fields, contextFields...)
	}

	// Add state name if present
	if stateName, ok := ctx.Value(StateNameKey).(string); ok {
		fields = append(fields, String(string(StateNameKey), stateName))
	}

	// Add activity name if present
	if activityName, ok := ctx.Value(ActivityNameKey).(string); ok {
		fields = append(fields, String(string(ActivityNameKey), activityName))
	}

	// Add state type if present
	if stateType, ok := ctx.Value(StateTypeKey).(string); ok {
		fields = append(fields, String(string(StateTypeKey), stateType))
	}

	// Add workflow name if present
	if workflowID, ok := ctx.Value(WorkflowIDKey).(string); ok {
		fields = append(fields, String(string(WorkflowIDKey), workflowID))
	}

	return fields
}
