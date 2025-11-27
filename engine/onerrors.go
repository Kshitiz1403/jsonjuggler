package engine

import (
	"context"
	"strings"

	"github.com/kshitiz1403/jsonjuggler/activities"
	"github.com/kshitiz1403/jsonjuggler/logger"
	sw "github.com/serverlessworkflow/sdk-go/v2/model"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel/trace"
)

// handleStateError processes the error against state's onErrors configuration
func (e *Engine) handleStateError(ctx context.Context, err *activities.ActivityError, onErrors []sw.OnError) (string, bool) {
	errMsg := err.Error()

	// Add error handling span
	var errorHandlingSpan trace.Span
	if e.telemetry != nil {
		// Extract state name from context
		stateName := ""
		if stateNameValue := ctx.Value(logger.StateNameKey); stateNameValue != nil {
			stateName = cast.ToString(stateNameValue)
		}
		ctx, errorHandlingSpan = e.telemetry.StartErrorHandlingSpan(ctx, stateName, string(err.Code), errMsg, "evaluate_handlers")
		defer errorHandlingSpan.End()
	}

	// First try to match specific error references
	for _, handler := range onErrors {
		if handler.ErrorRef != "DefaultErrorRef" && strings.Contains(errMsg, handler.ErrorRef) {
			e.logger.DebugContextf(ctx, "Matched error handler for ref '%s', transitioning to state '%s'", handler.ErrorRef, handler.Transition.NextState)
			return handler.Transition.NextState, true
		}
	}

	// Then look for DefaultErrorRef if none of the specific ones matched
	for _, handler := range onErrors {
		if handler.ErrorRef == "DefaultErrorRef" {
			e.logger.DebugContextf(ctx, "Using default error handler, transitioning to state '%s'", handler.Transition.NextState)
			return handler.Transition.NextState, true
		}
	}

	e.logger.DebugContext(ctx, "No error handler matched")
	return "", false
}
