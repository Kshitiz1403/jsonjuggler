package engine

import (
	"context"
	"time"

	"github.com/kshitiz1403/jsonjuggler/utils"
	sw "github.com/serverlessworkflow/sdk-go/v2/model"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// executeSleepState executes a sleep state with a timer which is cancelled if the context is done
func (e *Engine) executeSleepState(ctx context.Context, state *sw.SleepState, data *WorkflowData, stateExec *StateExecution) (result *StateResult, err error) {
	// Add sleep span first, before any operations
	var sleepSpan trace.Span
	if e.telemetry != nil {
		ctx, sleepSpan = e.telemetry.StartSleepSpan(ctx, state.GetName(), state.Duration)
		defer func() {
			if err != nil {
				sleepSpan.RecordError(err)
				sleepSpan.SetStatus(codes.Error, err.Error())
			} else {
				sleepSpan.SetStatus(codes.Ok, "")
			}
			sleepSpan.End()
		}()
	}

	e.logger.DebugContext(ctx, "Sleeping for %d seconds", state.Duration)
	duration, err := utils.ParseISODuration(state.Duration)
	if err != nil {
		e.logger.ErrorContextf(ctx, "Invalid sleep duration: %v", err)
		return nil, NewWorkflowError(ErrExpressionInvalid, "Invalid sleep duration").WithState(state.GetName()).WithCause(err)
	}

	if stateExec != nil {
		stateExec.SleepDuration = duration.String()
	}

	timer := time.NewTimer(duration)
	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
	case <-timer.C:
	}

	result = &StateResult{Data: data.Current, NextState: state.Transition.NextState}
	return result, nil
}
