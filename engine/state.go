package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/itchyny/gojq"
	"github.com/kshitiz1403/jsonjuggler/logger"
	sw "github.com/serverlessworkflow/sdk-go/v2/model"
)

// ErrorHandler defines how to handle specific errors
type ErrorHandler struct {
	ErrorRef   string `json:"errorRef"`
	Transition string `json:"transition"`
}

// StateResult represents the result of a state execution
type StateResult struct {
	Data      interface{}
	NextState string
	Error     error
}

func (e *Engine) executeState(ctx context.Context, state sw.State, data *WorkflowData) (*StateResult, error) {
	// Add state context to logger
	ctx = context.WithValue(ctx, logger.StateNameKey, state.GetName())
	ctx = context.WithValue(ctx, logger.StateTypeKey, string(state.GetType()))

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

	var result *StateResult
	var err error

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
			nextState, handled := e.handleStateError(actErr, state.OnErrors)
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

func (e *Engine) executeDataBasedSwitch(ctx context.Context, state *sw.SwitchState, data *WorkflowData, stateExec *StateExecution) (*StateResult, error) {
	for _, condition := range state.DataConditions {
		// Parse the JQ query
		query, err := gojq.Parse(condition.Condition)
		if err != nil {
			e.logger.ErrorContextf(ctx, "failed to parse condition '%s': %v", condition.Condition, err)
			return nil, err
		}

		// Run the query
		iter := query.Run(data.ToMap())
		result, ok := iter.Next()
		if !ok {
			e.logger.ErrorContextf(ctx, "no result from condition '%s'", condition.Condition)
			return nil, fmt.Errorf("no result from condition '%s'", condition.Condition)
		}
		// Check for error
		if err, isErr := result.(error); isErr {
			e.logger.ErrorContextf(ctx, "failed to evaluate condition '%s': %v", condition.Condition, err)
			return nil, fmt.Errorf("failed to evaluate condition '%s': %v", condition.Condition, err)
		}

		isTrue, ok := result.(bool)
		if !ok {
			e.logger.ErrorContextf(ctx, "condition '%s' did not evaluate to boolean, got %T", condition.Condition, result)
			return nil, fmt.Errorf("condition '%s' did not evaluate to boolean, got %T", condition.Condition, result)
		}

		if isTrue {
			if stateExec != nil {
				stateExec.MatchedCondition = condition.Name
			}
			return &StateResult{
				Data:      data.Current,
				NextState: condition.Transition.NextState,
			}, nil
		}
	}

	// If no condition matched or evaluation failed, use default transition
	if stateExec != nil {
		stateExec.MatchedCondition = "default"
	}
	return &StateResult{
		Data:      data.Current,
		NextState: state.DefaultCondition.Transition.NextState,
	}, nil
}

func (e *Engine) executeEventBasedSwitch(ctx context.Context, state *sw.SwitchState, data *WorkflowData, stateExec *StateExecution) (*StateResult, error) {
	// TODO: Implement event-based switching when needed
	return nil, fmt.Errorf("event-based switch not implemented")
}
