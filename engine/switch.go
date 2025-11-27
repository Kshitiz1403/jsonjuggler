package engine

import (
	"context"
	"fmt"

	"github.com/itchyny/gojq"
	sw "github.com/serverlessworkflow/sdk-go/v2/model"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (e *Engine) executeDataBasedSwitch(ctx context.Context, state *sw.SwitchState, data *WorkflowData, stateExec *StateExecution) (result *StateResult, err error) {
	for _, condition := range state.DataConditions {
		// Evaluate each condition with its own span
		conditionResult, conditionErr := e.evaluateCondition(ctx, state.GetName(), condition, data)
		if conditionErr != nil {
			return nil, conditionErr
		}

		if conditionResult {
			if stateExec != nil {
				stateExec.MatchedCondition = condition.Condition
			}
			result = &StateResult{
				Data:      data.Current,
				NextState: condition.Transition.NextState,
			}
			return result, nil
		}
	}

	// If no condition matched or evaluation failed, use default transition
	if e.telemetry != nil {
		var defaultSpan trace.Span
		ctx, defaultSpan = e.telemetry.StartSwitchConditionSpan(ctx, state.GetName(), "default", "default_condition")
		defer func() {
			if err != nil {
				defaultSpan.RecordError(err)
				defaultSpan.SetStatus(codes.Error, err.Error())
			} else {
				defaultSpan.SetStatus(codes.Ok, "")
			}
			defaultSpan.End()
		}()
	}

	if stateExec != nil {
		stateExec.MatchedCondition = "default"
	}
	result = &StateResult{
		Data:      data.Current,
		NextState: state.DefaultCondition.Transition.NextState,
	}
	return result, nil
}

// evaluateCondition evaluates a single condition with proper telemetry
func (e *Engine) evaluateCondition(ctx context.Context, stateName string, condition sw.DataCondition, data *WorkflowData) (isTrue bool, err error) {
	// Add switch condition span for each condition evaluation
	var switchSpan trace.Span
	if e.telemetry != nil {
		ctx, switchSpan = e.telemetry.StartSwitchConditionSpan(ctx, stateName, condition.Name, condition.Condition)
		defer func() {
			if err != nil {
				switchSpan.RecordError(err)
				switchSpan.SetStatus(codes.Error, err.Error())
			} else {
				switchSpan.SetStatus(codes.Ok, "")
			}
			switchSpan.End()
		}()
	}

	// Parse the JQ query
	query, err := gojq.Parse(condition.Condition)
	if err != nil {
		e.logger.ErrorContextf(ctx, "failed to parse condition '%s': %v", condition.Condition, err)
		return false, err
	}

	// Run the query
	iter := query.Run(data.ToMap())
	queryResult, hasNext := iter.Next()
	if !hasNext {
		e.logger.ErrorContextf(ctx, "no result from condition '%s'", condition.Condition)
		return false, fmt.Errorf("no result from condition '%s'", condition.Condition)
	}
	// Check for error
	if err, isErr := queryResult.(error); isErr {
		e.logger.ErrorContextf(ctx, "failed to evaluate condition '%s': %v", condition.Condition, err)
		return false, fmt.Errorf("failed to evaluate condition '%s': %v", condition.Condition, err)
	}

	var isBool bool
	isTrue, isBool = queryResult.(bool)
	if !isBool {
		e.logger.ErrorContextf(ctx, "condition '%s' did not evaluate to boolean, got %T", condition.Condition, queryResult)
		return false, fmt.Errorf("condition '%s' did not evaluate to boolean, got %T", condition.Condition, queryResult)
	}

	return isTrue, nil
}
