package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/kshitiz1403/jsonjuggler/activities"
	"github.com/kshitiz1403/jsonjuggler/logger"
	sw "github.com/serverlessworkflow/sdk-go/v2/model"
)

// Engine executes serverless workflows
type Engine struct {
	registry     *activities.Registry
	workflow     *sw.Workflow
	debugEnabled bool
	currentDebug *ExecutionDebug
	logger       logger.Logger
	statesMemo   map[string]sw.State
}

// NewEngine creates a new workflow engine
func NewEngine(registry *activities.Registry, debugEnabled bool, log logger.Logger) *Engine {
	return &Engine{
		registry:     registry,
		debugEnabled: debugEnabled,
		logger:       log,
	}
}

// GetRegistry returns the activity registry
func (e *Engine) GetRegistry() *activities.Registry {
	return e.registry
}

// Execute runs a workflow with the given input
func (e *Engine) Execute(ctx context.Context, workflow *sw.Workflow, input interface{}, globals map[string]interface{}) (executionResult *ExecutionResult, err error) {
	executionResult = &ExecutionResult{}
	startTime := time.Now()

	if workflow == nil {
		e.logger.ErrorContext(ctx, "Workflow cannot be nil")
		return nil, NewWorkflowError(ErrWorkflowInvalid, "workflow cannot be nil")
	}

	// Add workflow ID to context for logging
	ctx = context.WithValue(ctx, logger.WorkflowIDKey, workflow.ID)

	e.logger.InfoContextf(ctx, "Starting workflow execution. ID: %s", workflow.ID)

	if e.debugEnabled {
		e.currentDebug = &ExecutionDebug{
			States: make([]StateExecution, 0),
		}
		e.logger.DebugContext(ctx, "Debug mode enabled")
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}

		executionTime := time.Since(startTime)
		e.logger.InfoContextf(ctx, "Workflow execution completed in %v", executionTime)

		executionResult.Duration = executionTime
		if e.debugEnabled {
			executionResult.Debug = e.currentDebug
		}
	}()

	e.workflow = workflow

	if workflow.Start == nil {
		e.logger.ErrorContext(ctx, "Workflow must have a start state")
		return executionResult, fmt.Errorf("workflow must have a start state")
	}

	// Initialize workflow data
	workflowData := NewWorkflowData(input, globals)
	e.logger.DebugContextf(ctx, "Initialized workflow data with input: %+v", input)

	// Set initial data
	workflowData.Initial = input

	state := e.findState(workflow, workflow.Start.StateName)
	if state == nil {
		e.logger.ErrorContextf(ctx, "Start state '%s' not found", workflow.Start.StateName)
		return executionResult, NewWorkflowError(ErrStateNotFound, fmt.Sprintf("start state '%s' not found", workflow.Start.StateName))
	}

	for state != nil {
		e.logger.InfoContextf(ctx, "Executing state: %s (Type: %s)", state.GetName(), state.GetType())

		stateResult, err := e.executeState(ctx, state, workflowData)
		if err != nil {
			e.logger.ErrorContextf(ctx, "Error executing state %s: %v", state.GetName(), err)
			return executionResult, NewWorkflowError(ErrStateExecutionFail, fmt.Errorf("error executing state %s: %w", state.GetName(), err).Error()).WithCause(err)
		}

		// Store state result in the States map
		workflowData.States[state.GetName()] = stateResult.Data
		// Update current data
		workflowData.Current = stateResult.Data

		if stateResult.NextState == "" {
			e.logger.InfoContext(ctx, "Workflow execution completed - reached end state")
			break
		}

		e.logger.DebugContextf(ctx, "Transitioning from state '%s' to '%s'", state.GetName(), stateResult.NextState)
		state = e.findState(workflow, stateResult.NextState)
		if state == nil {
			err := NewWorkflowError(ErrStateTransitionFail, fmt.Sprintf("transition state %s not found", stateResult.NextState))
			e.logger.ErrorContext(ctx, err)
			return executionResult, err
		}
	}

	executionResult.Data = workflowData.Current

	return executionResult, nil
}

func (e *Engine) findState(workflow *sw.Workflow, name string) sw.State {
	if e.statesMemo == nil {
		e.statesMemo = make(map[string]sw.State)
		for _, state := range workflow.States {
			e.statesMemo[state.GetName()] = state
		}
	}

	if state, ok := e.statesMemo[name]; ok {
		return state
	}

	return nil
}
