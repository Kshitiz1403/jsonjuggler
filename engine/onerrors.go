package engine

import (
	"strings"

	"github.com/kshitiz1403/jsonjuggler/activities"
	sw "github.com/serverlessworkflow/sdk-go/v2/model"
)

// handleStateError processes the error against state's onErrors configuration
func (e *Engine) handleStateError(err *activities.ActivityError, onErrors []sw.OnError) (string, bool) {
	errMsg := err.Error()

	// First try to match specific error references
	for _, handler := range onErrors {
		if handler.ErrorRef != "DefaultErrorRef" && strings.Contains(errMsg, handler.ErrorRef) {
			return handler.Transition.NextState, true
		}
	}

	// Then look for DefaultErrorRef if none of the specific ones matched
	for _, handler := range onErrors {
		if handler.ErrorRef == "DefaultErrorRef" {
			return handler.Transition.NextState, true
		}
	}

	return "", false
}
