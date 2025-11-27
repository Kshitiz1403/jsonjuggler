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
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// executeAction executes a single action and returns its result
func (e *Engine) executeAction(ctx context.Context, action sw.Action, data *WorkflowData, stateExec *StateExecution) (result interface{}, err error) {
	ctx = context.WithValue(ctx, logger.ActivityNameKey, action.FunctionRef.RefName)

	startTime := time.Now()
	var activitySpan trace.Span
	if e.telemetry != nil {
		ctx, activitySpan = e.telemetry.StartActivitySpan(ctx, action.FunctionRef.RefName)
		defer func() {
			if err != nil {
				activitySpan.RecordError(err)
				activitySpan.SetStatus(codes.Error, err.Error())
			} else {
				activitySpan.SetStatus(codes.Ok, "")
			}
			activitySpan.End()
		}()
		// Record activity execution
		defer func() {
			e.telemetry.RecordActivityDuration(ctx, time.Since(startTime).Seconds(), action.FunctionRef.RefName)
			e.telemetry.RecordActivity(ctx, action.FunctionRef.RefName)
		}()
	}

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

	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			e.logger.ErrorContextf(ctx, "Panic recovered in activity execution: %v", r)
			err = activities.NewActivityError(
				activities.ErrPanic,
				fmt.Sprintf("Panic in activity execution: %v", r),
				action.FunctionRef.RefName,
			)
			if actionResult != nil {
				actionResult.Error = err.Error()
			}
		}
	}()

	// Phase 1: Activity lookup with its own span
	activity, err := e.lookupActivityWithTelemetry(ctx, action.FunctionRef.RefName)
	if err != nil {
		if actionResult != nil {
			actionResult.Error = err.Error()
		}
		return nil, err
	}

	// Phase 2: Argument resolution with its own span
	arguments, err := e.resolveArgumentsWithTelemetry(ctx, action.FunctionRef.RefName, action.FunctionRef.Arguments, data)
	if err != nil {
		if actionResult != nil {
			actionResult.Error = err.Error()
		}
		return nil, err
	}

	if actionResult != nil {
		actionResult.Arguments = arguments
	}

	// Phase 3: Activity execution with its own span
	result, err = e.executeActivityWithTelemetry(ctx, action.FunctionRef.RefName, activity, arguments)
	if err != nil {
		if actionResult != nil {
			actionResult.Error = err.Error()
		}
		return nil, err
	}

	e.logger.InfoContext(ctx, "Activity executed successfully")
	e.logger.DebugContextf(ctx, "Activity result: %+v", result)

	if actionResult != nil {
		actionResult.Output = result
	}

	// Record activity error if any
	if err != nil && e.telemetry != nil {
		if actErr, ok := err.(*activities.ActivityError); ok {
			e.telemetry.RecordActivityError(ctx, action.FunctionRef.RefName, string(actErr.Code))
		}
	}

	return result, nil
}

// evaluateArguments evaluates all arguments in a map recursively
func (e *Engine) evaluateArguments(ctx context.Context, args map[string]sw.Object, data *WorkflowData) (map[string]any, error) {
	arguments := convertToAnyMap(args)
	return utils.EvaluateArgumentMap(arguments, data.ToMap())
}

// executeActions executes a list of actions
func (e *Engine) executeActions(ctx context.Context, actions []sw.Action, data *WorkflowData, stateExec *StateExecution) (result *StateResult, err error) {
	currentResult := data.Current

	// Add action group span for operation states
	var actionGroupSpan trace.Span
	if e.telemetry != nil && len(actions) > 0 {
		// Extract state name from context
		stateName := ""
		if stateNameValue := ctx.Value(logger.StateNameKey); stateNameValue != nil {
			stateName = cast.ToString(stateNameValue)
		}
		ctx, actionGroupSpan = e.telemetry.StartActionGroupSpan(ctx, stateName, len(actions))
		defer func() {
			if err != nil {
				actionGroupSpan.RecordError(err)
				actionGroupSpan.SetStatus(codes.Error, err.Error())
			} else {
				actionGroupSpan.SetStatus(codes.Ok, "")
			}
			actionGroupSpan.End()
		}()
	}

	for _, action := range actions {
		var actionResult interface{}
		actionResult, err = e.executeAction(ctx, action, data, stateExec)
		if err != nil {
			result = &StateResult{
				Data:  currentResult,
				Error: err,
			}
			return result, err
		}
		// Update current data
		data.Current = actionResult
		currentResult = actionResult
	}

	result = &StateResult{
		Data: currentResult,
	}
	return result, nil
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

// executeActivityWithTelemetry executes an activity with proper telemetry
func (e *Engine) executeActivityWithTelemetry(ctx context.Context, activityName string, activity activities.Activity, arguments map[string]any) (result interface{}, err error) {
	// Add activity execution span with defer for proper error handling
	var execSpan trace.Span
	if e.telemetry != nil {
		ctx, execSpan = e.telemetry.StartActivityExecutionSpan(ctx, activityName)
		defer func() {
			if err != nil {
				execSpan.RecordError(err)
				execSpan.SetStatus(codes.Error, err.Error())
			} else {
				execSpan.SetStatus(codes.Ok, "")
			}
			execSpan.End()
		}()
	}

	e.logger.DebugContextf(ctx, "Executing activity %s with arguments: %+v", activityName, arguments)

	// Execute activity
	result, err = activity.Execute(ctx, arguments)
	if err != nil {
		e.logger.ErrorContextf(ctx, "Activity execution failed: %v", err)

		// If it's already an ActivityError, return it as is
		if actErr, ok := err.(*activities.ActivityError); ok {
			return nil, actErr
		}

		// Otherwise wrap it in an ActivityError
		return nil, activities.NewActivityError(
			activities.ErrExecutionFailed,
			"Activity execution failed",
			activityName,
		).WithCause(err)
	}

	e.logger.InfoContext(ctx, "Activity executed successfully")
	e.logger.DebugContextf(ctx, "Activity result: %+v", result)

	return result, nil
}

// lookupActivityWithTelemetry looks up an activity with proper telemetry
func (e *Engine) lookupActivityWithTelemetry(ctx context.Context, activityName string) (activity activities.Activity, err error) {
	// Add activity lookup span with defer for proper error handling
	var lookupSpan trace.Span
	if e.telemetry != nil {
		ctx, lookupSpan = e.telemetry.StartActivityLookupSpan(ctx, activityName)
		defer func() {
			if err != nil {
				lookupSpan.RecordError(err)
				lookupSpan.SetStatus(codes.Error, err.Error())
			} else {
				lookupSpan.SetStatus(codes.Ok, "")
			}
			lookupSpan.End()
		}()
	}

	activity, ok := e.registry.Get(activityName)
	if !ok {
		err := NewWorkflowError(ErrActivityNotFound, "Activity not found").
			WithActivity(activityName)
		e.logger.ErrorContextf(ctx, "Activity not found: %s", activityName)
		return nil, err
	}

	return activity, nil
}

// resolveArgumentsWithTelemetry resolves activity arguments with proper telemetry
func (e *Engine) resolveArgumentsWithTelemetry(ctx context.Context, activityName string, args map[string]sw.Object, data *WorkflowData) (arguments map[string]any, err error) {
	// Add activity args span with defer for proper error handling
	var argsSpan trace.Span
	if e.telemetry != nil {
		ctx, argsSpan = e.telemetry.StartActivityArgsSpan(ctx, activityName)
		defer func() {
			if err != nil {
				argsSpan.RecordError(err)
				argsSpan.SetStatus(codes.Error, err.Error())
			} else {
				argsSpan.SetStatus(codes.Ok, "")
			}
			argsSpan.End()
		}()
	}

	e.logger.DebugContext(ctx, "Evaluating activity arguments")
	arguments, err = e.evaluateArguments(ctx, args, data)
	if err != nil {
		err = NewWorkflowError(ErrExpressionEval, "Failed to evaluate arguments").
			WithActivity(activityName).
			WithCause(err)
		e.logger.ErrorContextf(ctx, "Failed to evaluate arguments: %v", err)
		return nil, err // Not an activity error. TODO: @Pranithkampelly what are your thoughts on this? Should we enable onErrors for all type of errors or only ActivityErrors? Moreover, getting the arguments is not an activity error, so this would not be handled by onErrors. False negative.
	}

	return arguments, nil
}
