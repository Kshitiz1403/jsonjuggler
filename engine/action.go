package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kshitiz1403/jsonjuggler/activities"
	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/utils"
	sw "github.com/serverlessworkflow/sdk-go/v2/model"
)

// executeAction executes a single action and returns its result
func (e *Engine) executeAction(ctx context.Context, action sw.Action, data *WorkflowData, stateExec *StateExecution) (interface{}, error) {
	ctx = context.WithValue(ctx, logger.ActivityNameKey, action.FunctionRef.RefName)
	e.logger.InfoContextf(ctx, "Executing activity: %s", action.FunctionRef.RefName)

	var actionResult *ActionResult
	if e.debugEnabled && stateExec != nil {
		actionResult = &ActionResult{
			ActivityName: action.FunctionRef.RefName,
			StartTime:    time.Now(),
		}
		defer func() {
			actionResult.EndTime = time.Now()
			stateExec.Actions = append(stateExec.Actions, *actionResult)
			e.logger.DebugContextf(ctx, "Activity execution completed in %v", actionResult.EndTime.Sub(actionResult.StartTime))
		}()
	}

	// Get activity
	activity, ok := e.registry.Get(action.FunctionRef.RefName)
	if !ok {
		err := NewWorkflowError(ErrActivityNotFound, "Activity not found").
			WithActivity(action.FunctionRef.RefName)
		e.logger.ErrorContextf(ctx, "Activity not found: %s", action.FunctionRef.RefName)
		if actionResult != nil {
			actionResult.Error = err.Error()
		}
		return nil, err
	}

	// Parse and evaluate arguments
	e.logger.DebugContext(ctx, "Evaluating activity arguments")
	arguments, err := e.evaluateArguments(ctx, action.FunctionRef.Arguments, data)
	if err != nil {
		err = NewWorkflowError(ErrExpressionEval, "Failed to evaluate arguments").
			WithActivity(action.FunctionRef.RefName).
			WithCause(err)
		e.logger.ErrorContextf(ctx, "Failed to evaluate arguments: %v", err)
		if actionResult != nil {
			actionResult.Error = err.Error()
		}
		return nil, err // Not an activity error. TODO: @Pranithkampelly what are your thoughts on this? Should we enable onErrors for all type of errors or only ActivityErrors? Moreover, getting the arguments is not an activity error, so this would not be handled by onErrors. False negative.
	}

	if actionResult != nil {
		actionResult.Arguments = arguments
	}

	e.logger.DebugContextf(ctx, "Executing activity with arguments: %+v", arguments)

	// Execute activity
	result, err := activity.Execute(ctx, arguments)
	if err != nil {
		e.logger.ErrorContextf(ctx, "Activity execution failed: %v", err)
		if actionResult != nil {
			actionResult.Error = err.Error()
		}

		// If it's already an ActivityError, return it as is
		if actErr, ok := err.(*activities.ActivityError); ok {
			return nil, actErr
		}

		// Otherwise wrap it in an ActivityError
		return nil, activities.NewActivityError(
			activities.ErrExecutionFailed,
			"Activity execution failed",
			action.FunctionRef.RefName,
		).WithCause(err)
	}

	e.logger.InfoContext(ctx, "Activity executed successfully")
	e.logger.DebugContextf(ctx, "Activity result: %+v", result)

	if actionResult != nil {
		actionResult.Output = result
	}

	return result, nil
}

// evaluateArguments evaluates all arguments in a map recursively
func (e *Engine) evaluateArguments(ctx context.Context, args map[string]sw.Object, data *WorkflowData) (map[string]any, error) {
	arguments := convertToAnyMap(args)
	return utils.EvaluateArgumentMap(arguments, data.ToMap())
}

// executeActions executes a list of actions
func (e *Engine) executeActions(ctx context.Context, actions []sw.Action, data *WorkflowData, stateExec *StateExecution) (*StateResult, error) {
	result := data.Current

	for _, action := range actions {
		var err error
		result, err = e.executeAction(ctx, action, data, stateExec)
		if err != nil {
			return &StateResult{
				Data:  result,
				Error: err,
			}, err
		}
		// Update current data
		data.Current = result
	}

	return &StateResult{
		Data: result,
	}, nil
}

// convertToAnyMap converts model.Object map to map[string]any
func convertToAnyMap(m map[string]sw.Object) map[string]any {
	// TODO: This is a temporary solution to convert the map to a map[string]any
	result, err := json.Marshal(m)
	if err != nil {
		fmt.Println("error marshalling map:", err)
		return nil
	}

	var resultMap map[string]any
	if err := json.Unmarshal(result, &resultMap); err != nil {
		fmt.Println("error unmarshalling map:", err)
		return nil
	}

	return resultMap
}
