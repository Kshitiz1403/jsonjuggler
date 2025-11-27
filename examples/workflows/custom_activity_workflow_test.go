package workflows

import (
	"context"
	"testing"

	"github.com/kshitiz1403/jsonjuggler/activities"
	"github.com/kshitiz1403/jsonjuggler/config"
	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
	"github.com/kshitiz1403/jsonjuggler/parser"
	"github.com/kshitiz1403/jsonjuggler/utils"
	"github.com/stretchr/testify/require"
)

// CustomHelloWorldActivity is an example custom activity
type CustomHelloWorldActivity struct {
	activities.BaseActivity
}

func (a *CustomHelloWorldActivity) Execute(ctx context.Context, args map[string]any) (interface{}, error) {
	a.GetLogger().InfoContext(ctx, "Starting custom operation")
	result := "hello world"
	a.GetLogger().DebugContextf(ctx, "Custom operation completed with result: %v", result)
	return result, nil
}

func TestHelloWorldWorkflow(t *testing.T) {
	// Initialize custom logger
	customLogger := zap.NewLogger(logger.DebugLevel)

	// Initialize engine with custom activity
	engine, err := config.Initialize(
		config.WithLogger(customLogger),
		config.WithDebug(true),
		config.WithActivity("HelloWorld", &CustomHelloWorldActivity{}),
	)
	require.NoError(t, err)

	// Create context with fields
	ctx := logger.WithFields(context.Background(),
		logger.String("requestID", "test-123"),
		logger.String("userID", "test-456"),
	)

	// Parse workflow
	p := parser.NewParser(engine.GetRegistry())
	workflow, err := p.ParseFromFile("custom_activity_workflow.json")
	require.NoError(t, err)

	// Test input
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"message": "test input",
		},
	}

	// Execute workflow
	result, err := engine.Execute(ctx, workflow, input, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Write debug output to file
	writeToFile("outputs/custom_activity_workflow_result.json", []byte(utils.AnyToJSONStringPretty(result.Data)))
	writeToFile("outputs/custom_activity_workflow_debug.json", []byte(utils.AnyToJSONStringPretty(result.Debug)))

	// Verify execution path through debug info
	require.NotNil(t, result.Debug)
	require.Equal(t, 3, len(result.Debug.States))
	require.Equal(t, "PrepareInput", result.Debug.States[0].Name)
	require.Equal(t, "ExecuteHelloWorld", result.Debug.States[1].Name)
	require.Equal(t, "FormatOutput", result.Debug.States[2].Name)

	// Verify result of hello world activity
	require.Equal(t, "hello world", result.Debug.States[1].Output)
}
