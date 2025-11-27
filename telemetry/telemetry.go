package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	serviceName = "jsonjuggler"
)

// Config holds telemetry configuration
type Config struct {
	// TracerProvider is the OpenTelemetry TracerProvider to use
	TracerProvider trace.TracerProvider
	// MeterProvider is the OpenTelemetry MeterProvider to use
	MeterProvider metric.MeterProvider
	// Enabled indicates whether telemetry is enabled
	Enabled bool
}

// Telemetry provides telemetry functionality for JsonJuggler
type Telemetry struct {
	tracer trace.Tracer
	meter  metric.Meter

	// Metrics
	workflowDuration      metric.Float64Histogram
	activityDuration      metric.Float64Histogram
	workflowErrorCount    metric.Int64Counter
	activityErrorCount    metric.Int64Counter
	workflowStateCount    metric.Int64Counter
	workflowActivityCount metric.Int64Counter

	enabled bool
}

// New creates a new Telemetry instance
func New(cfg *Config) (*Telemetry, error) {
	if cfg == nil || !cfg.Enabled {
		return &Telemetry{enabled: false}, nil
	}

	// Use provided providers or default ones
	tp := cfg.TracerProvider
	if tp == nil {
		tp = otel.GetTracerProvider()
	}

	mp := cfg.MeterProvider
	if mp == nil {
		mp = otel.GetMeterProvider()
	}

	meter := mp.Meter(serviceName)

	// Initialize metrics
	workflowDuration, err := meter.Float64Histogram("workflow.duration",
		metric.WithDescription("Duration of workflow executions"),
		metric.WithUnit("s"))
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow duration metric: %w", err)
	}

	activityDuration, err := meter.Float64Histogram("activity.duration",
		metric.WithDescription("Duration of activity executions"),
		metric.WithUnit("s"))
	if err != nil {
		return nil, fmt.Errorf("failed to create activity duration metric: %w", err)
	}

	workflowErrorCount, err := meter.Int64Counter("workflow.errors",
		metric.WithDescription("Number of workflow errors"))
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow error counter: %w", err)
	}

	activityErrorCount, err := meter.Int64Counter("activity.errors",
		metric.WithDescription("Number of activity errors"))
	if err != nil {
		return nil, fmt.Errorf("failed to create activity error counter: %w", err)
	}

	workflowStateCount, err := meter.Int64Counter("workflow.states",
		metric.WithDescription("Number of workflow states executed"))
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow state counter: %w", err)
	}

	workflowActivityCount, err := meter.Int64Counter("workflow.activities",
		metric.WithDescription("Number of workflow activities executed"))
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow activity counter: %w", err)
	}

	return &Telemetry{
		tracer:                tp.Tracer(serviceName),
		meter:                 meter,
		workflowDuration:      workflowDuration,
		activityDuration:      activityDuration,
		workflowErrorCount:    workflowErrorCount,
		activityErrorCount:    activityErrorCount,
		workflowStateCount:    workflowStateCount,
		workflowActivityCount: workflowActivityCount,
		enabled:               true,
	}, nil
}

// StartWorkflowSpan starts a new workflow span
func (t *Telemetry) StartWorkflowSpan(ctx context.Context, workflowID string) (context.Context, trace.Span) {
	if !t.enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := t.tracer.Start(ctx, "workflow.execute",
		trace.WithAttributes(
			attribute.String("workflow.id", workflowID),
		))

	// Add workflow ID to span
	span.SetAttributes(attribute.String("workflow.id", workflowID))

	return ctx, span
}

// StartStateSpan starts a new state span as child of current span
func (t *Telemetry) StartStateSpan(ctx context.Context, stateName, stateType string) (context.Context, trace.Span) {
	if !t.enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := t.tracer.Start(ctx, "workflow.state",
		trace.WithAttributes(
			attribute.String("state.name", stateName),
			attribute.String("state.type", stateType),
		))

	return ctx, span
}

// StartActivitySpan starts a new activity span as child of current span
func (t *Telemetry) StartActivitySpan(ctx context.Context, activityName string) (context.Context, trace.Span) {
	if !t.enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := t.tracer.Start(ctx, "workflow.activity",
		trace.WithAttributes(
			attribute.String("activity.name", activityName),
		))

	return ctx, span
}

// StartActivityLookupSpan starts a new span for activity lookup
func (t *Telemetry) StartActivityLookupSpan(ctx context.Context, activityName string) (context.Context, trace.Span) {
	if !t.enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := t.tracer.Start(ctx, "workflow.activity.lookup",
		trace.WithAttributes(
			attribute.String("activity.name", activityName),
		))

	return ctx, span
}

// StartActivityArgsSpan starts a new span for activity argument evaluation
func (t *Telemetry) StartActivityArgsSpan(ctx context.Context, activityName string) (context.Context, trace.Span) {
	if !t.enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := t.tracer.Start(ctx, "workflow.activity.args",
		trace.WithAttributes(
			attribute.String("activity.name", activityName),
		))

	return ctx, span
}

// StartActivityExecutionSpan starts a new span for activity execution
func (t *Telemetry) StartActivityExecutionSpan(ctx context.Context, activityName string) (context.Context, trace.Span) {
	if !t.enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := t.tracer.Start(ctx, "workflow.activity.execution",
		trace.WithAttributes(
			attribute.String("activity.name", activityName),
		))

	return ctx, span
}

// StartSwitchConditionSpan starts a new span for switch condition evaluation
func (t *Telemetry) StartSwitchConditionSpan(ctx context.Context, stateName, conditionName, condition string) (context.Context, trace.Span) {
	if !t.enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := t.tracer.Start(ctx, "workflow.switch.condition",
		trace.WithAttributes(
			attribute.String("state.name", stateName),
			attribute.String("condition.name", conditionName),
			attribute.String("condition.expression", condition),
		))

	return ctx, span
}

// StartActionGroupSpan starts a new span for a group of actions in an operation state
func (t *Telemetry) StartActionGroupSpan(ctx context.Context, stateName string, actionCount int) (context.Context, trace.Span) {
	if !t.enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := t.tracer.Start(ctx, "workflow.operation.actions",
		trace.WithAttributes(
			attribute.String("state.name", stateName),
			attribute.Int("action.count", actionCount),
		))

	return ctx, span
}

// StartErrorHandlingSpan starts a new span for error handling
func (t *Telemetry) StartErrorHandlingSpan(ctx context.Context, stateName, errorType, errString, handlerAction string) (context.Context, trace.Span) {
	if !t.enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := t.tracer.Start(ctx, "workflow.error.handling",
		trace.WithAttributes(
			attribute.String("state.name", stateName),
			attribute.String("error.type", errorType),
			attribute.String("error.string", errString),
			attribute.String("handler.action", handlerAction),
		))

	return ctx, span
}

// StartSleepSpan starts a new span for sleep operations
func (t *Telemetry) StartSleepSpan(ctx context.Context, stateName, duration string) (context.Context, trace.Span) {
	if !t.enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := t.tracer.Start(ctx, "workflow.sleep",
		trace.WithAttributes(
			attribute.String("state.name", stateName),
			attribute.String("sleep.duration", duration),
		))

	return ctx, span
}

// RecordWorkflowDuration records workflow execution duration
func (t *Telemetry) RecordWorkflowDuration(ctx context.Context, duration float64, workflowID string) {
	if !t.enabled {
		return
	}

	t.workflowDuration.Record(ctx, duration,
		metric.WithAttributes(
			attribute.String("workflow.id", workflowID),
		))
}

// RecordActivityDuration records activity execution duration
func (t *Telemetry) RecordActivityDuration(ctx context.Context, duration float64, activityName string) {
	if !t.enabled {
		return
	}

	t.activityDuration.Record(ctx, duration,
		metric.WithAttributes(
			attribute.String("activity.name", activityName),
		))
}

// RecordWorkflowError records a workflow error
func (t *Telemetry) RecordWorkflowError(ctx context.Context, workflowID string, errorCode string) {
	if !t.enabled {
		return
	}

	t.workflowErrorCount.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("workflow.id", workflowID),
			attribute.String("error.code", errorCode),
		))
}

// RecordActivityError records an activity error
func (t *Telemetry) RecordActivityError(ctx context.Context, activityName string, errorCode string) {
	if !t.enabled {
		return
	}

	t.activityErrorCount.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("activity.name", activityName),
			attribute.String("error.code", errorCode),
		))
}

// RecordState records a state execution
func (t *Telemetry) RecordState(ctx context.Context, stateName, stateType string) {
	if !t.enabled {
		return
	}

	t.workflowStateCount.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("state.name", stateName),
			attribute.String("state.type", stateType),
		))
}

// RecordActivity records an activity execution
func (t *Telemetry) RecordActivity(ctx context.Context, activityName string) {
	if !t.enabled {
		return
	}

	t.workflowActivityCount.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("activity.name", activityName),
		))
}
