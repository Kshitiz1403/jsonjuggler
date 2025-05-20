package engine

import (
	"context"

	"github.com/kshitiz1403/jsonjuggler/utils"
	sw "github.com/serverlessworkflow/sdk-go/v2/model"

	"time"
)

// executeSleepState executes a sleep state with a timer which is cancelled if the context is done
func (e *Engine) executeSleepState(ctx context.Context, state *sw.SleepState, data *WorkflowData, stateExec *StateExecution) (*StateResult, error) {
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

	return &StateResult{Data: data.Current, NextState: state.Transition.NextState}, nil
}
