package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/kshitiz1403/jsonjuggler/logger"
	sw "github.com/serverlessworkflow/sdk-go/v2/model"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (e *Engine) executeState(ctx context.Context, state sw.State, data *WorkflowData) (result *StateResult, err error) {
	// Add state context to logger
	ctx = context.WithValue(ctx, logger.StateNameKey, state.GetName())
	ctx = context.WithValue(ctx, logger.StateTypeKey, string(state.GetType()))

	if e.telemetry != nil {
		var span trace.Span
		ctx, span = e.telemetry.StartStateSpan(ctx, state.GetName(), string(state.GetType()))
		defer func() {
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "")
			}
			span.End()
		}()
		// Record state execution
		e.telemetry.RecordState(ctx, state.GetName(), string(state.GetType()))
	}

	e.logger.DebugContextf(ctx, "Starting state execution. Type: %s", state.GetType())
	e.logger.DebugContextf(ctx, "State input data: %+v", data.Current)

	var stateExec *StateExecution
	if e.debugEnabled {
		stateExec = &StateExecution{
			Name:      state.GetName(),
			Type:      string(state.GetType()),
			StartTime: time.Now(),
			Input:     data.Current,
		}
		defer func() {
			stateExec.EndTime = time.Now()
			e.currentDebug.States = append(e.currentDebug.States, *stateExec)
			e.logger.DebugContextf(ctx, "State execution completed in %v", stateExec.EndTime.Sub(stateExec.StartTime))
		}()
	}

	switch state.GetType() {
	case sw.StateTypeOperation:
		e.logger.DebugContext(ctx, "Executing operation state")
		result, err = e.executeOperationState(ctx, state.(*sw.OperationState), data, stateExec)

	case sw.StateTypeSwitch:
		e.logger.DebugContext(ctx, "Executing switch state")
		result, err = e.executeSwitchState(ctx, state.(*sw.SwitchState), data, stateExec)

	case sw.StateTypeSleep:
		e.logger.DebugContext(ctx, "Executing sleep state")
		result, err = e.executeSleepState(ctx, state.(*sw.SleepState), data, stateExec)

	default:
		err = NewWorkflowError(ErrStateInvalid, "Unsupported state type").
			WithState(state.GetName()).
			WithContext("stateType", state.GetType())
		e.logger.ErrorContextf(ctx, "Unsupported state type: %s", state.GetType())
		return nil, err
	}

	if err != nil {
		if stateExec != nil {
			stateExec.Error = err.Error()
		}
		e.logger.ErrorContextf(ctx, "State execution failed: %v", err)
		return nil, err
	}

	if stateExec != nil {
		stateExec.Output = result.Data
	}

	e.logger.InfoContextf(ctx, "State '%s' completed successfully", state.GetName())
	e.logger.DebugContextf(ctx, "State output data: %+v", result.Data)

	return result, nil
}

func (e *Engine) executeOperationState(ctx context.Context, state *sw.OperationState, data *WorkflowData, stateExec *StateExecution) (*StateResult, error) {
	result, err := e.executeActions(ctx, state.Actions, data, stateExec)
	if err != nil {
		// Check for activity error anywhere in the error chain
		if actErr, ok := UnwrapActivityError(err); ok && len(state.OnErrors) > 0 {
			nextState, handled := e.handleStateError(ctx, actErr, state.OnErrors)
			if handled {
				return &StateResult{
					Data:      data.Current,
					NextState: nextState,
					Error:     err, // Preserve the full error chain
				}, nil
			}
		}
		return nil, err
	}

	var nextState string
	if state.GetTransition() != nil {
		nextState = state.GetTransition().NextState
	}

	result.NextState = nextState
	return result, nil
}

func (e *Engine) executeSwitchState(ctx context.Context, state *sw.SwitchState, data *WorkflowData, stateExec *StateExecution) (*StateResult, error) {

	if state.DataConditions != nil {
		return e.executeDataBasedSwitch(ctx, state, data, stateExec)
	}

	if state.EventConditions != nil {
		return e.executeEventBasedSwitch(ctx, state, data, stateExec)
	}

	return nil, fmt.Errorf("switch state must have either data conditions or event conditions")
}
