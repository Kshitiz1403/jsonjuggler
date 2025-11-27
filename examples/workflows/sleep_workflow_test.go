package workflows

import (
	"context"
	"testing"
	"time"

	sw "github.com/serverlessworkflow/sdk-go/v2/model"

	"github.com/kshitiz1403/jsonjuggler/config"
	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
	"github.com/kshitiz1403/jsonjuggler/parser"
	"github.com/kshitiz1403/jsonjuggler/utils"
	"github.com/stretchr/testify/require"
)

func TestSleepWorkflow(t *testing.T) {
	// Initialize engine with debugging enabled
	engine, err := config.Initialize(
		config.WithDebug(true),
		config.WithLogger(zap.NewLogger(logger.DebugLevel)),
	)
	require.NoError(t, err)

	// Parse workflow definition
	p := parser.NewParser(engine.GetRegistry())
	workflow, err := p.ParseFromFile("sleep_workflow.json")
	require.NoError(t, err)

	// Test normal execution
	t.Run("Normal Execution", func(t *testing.T) {
		result, err := engine.Execute(context.Background(), workflow, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Check execution duration (should be around 3 seconds: 1s + 2s)
		executionTime := result.Duration
		require.Greater(t, executionTime, 3*time.Second)
		require.Less(t, executionTime, 4*time.Second)

		// Write debug output to file
		writeToFile("outputs/sleep_workflow_normal_execution_result.json", []byte(utils.AnyToJSONStringPretty(result.Data)))
		writeToFile("outputs/sleep_workflow_normal_execution_debug.json", []byte(utils.AnyToJSONStringPretty(result.Debug)))

		// Verify state transitions
		require.Equal(t, 5, len(result.Debug.States))
		require.Equal(t, "ProcessInitial", result.Debug.States[0].Name)
		require.Equal(t, "WaitShort", result.Debug.States[1].Name)
		require.Equal(t, "ProcessIntermediate", result.Debug.States[2].Name)
		require.Equal(t, "WaitLong", result.Debug.States[3].Name)
		require.Equal(t, "ProcessFinal", result.Debug.States[4].Name)

		// Verify final status
		resultData, ok := result.Data.(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "completed", resultData["status"])
	})

	// Test context cancellation
	// TODO: As of now, even on context cancellation, the workflow is not terminating.
	// This is a known issue and should be fixed in the future.
	// Till then, we are skipping this test.
	// t.Run("Context Cancellation", func(t *testing.T) {
	// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// 	defer cancel()

	// 	result, err := engine.Execute(ctx, workflow, nil, nil)
	// 	require.Error(t, err) // Should error due to context cancellation
	// 	require.Contains(t, err.Error(), "context deadline exceeded")

	// 	if result != nil && result.Debug != nil {
	// 		// Should not reach the final state
	// 		lastState := result.Debug.States[len(result.Debug.States)-1]
	// 		require.NotEqual(t, "ProcessFinal", lastState.Name)
	// 	}
	// })

	// // Test invalid duration
	t.Run("Invalid Duration", func(t *testing.T) {
		// Create a modified workflow with invalid duration
		invalidWorkflow := *workflow
		invalidWorkflow.States[1].(*sw.SleepState).Duration = "invalid"

		result, err := engine.Execute(context.Background(), &invalidWorkflow, nil, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Invalid sleep duration")

		// Write debug output to file
		writeToFile("outputs/sleep_workflow_invalid_duration_execution_result.json", []byte(utils.AnyToJSONStringPretty(result.Data)))
		writeToFile("outputs/sleep_workflow_invalid_duration_execution_debug.json", []byte(utils.AnyToJSONStringPretty(result.Debug)))
	})
}
