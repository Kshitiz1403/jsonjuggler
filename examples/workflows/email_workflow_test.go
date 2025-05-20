package workflows

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kshitiz1403/jsonjuggler/config"
	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
	"github.com/kshitiz1403/jsonjuggler/parser"
	"github.com/kshitiz1403/jsonjuggler/utils"
	"github.com/stretchr/testify/require"
)

// writeToFile ensures the directory exists and writes data to file
func writeToFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func TestEmailWorkflow(t *testing.T) {
	// Initialize engine with debugging enabled
	engine := config.Initialize(
		config.WithDebug(true),
		config.WithLogger(zap.NewLogger(logger.DebugLevel)),
	)

	// Parse workflow definition
	p := parser.NewParser(engine.GetRegistry())
	workflow, err := p.ParseFromFile("email_workflow.json")
	require.NoError(t, err)

	// Test premium user flow
	premiumInput := map[string]interface{}{
		"user": map[string]interface{}{
			"type":  "premium",
			"email": "premium@example.com",
		},
	}

	result, err := engine.Execute(context.Background(), workflow, premiumInput, map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result.Data)

	// Write to file
	writeToFile("outputs/email_workflow_premium_result.json", []byte(utils.AnyToJSONStringPretty(result.Data)))
	writeToFile("outputs/email_workflow_premium_debug.json", []byte(utils.AnyToJSONStringPretty(result.Debug)))

	// Debug information available
	require.NotNil(t, result.Debug)
	require.Greater(t, len(result.Debug.States), 0)

	// Check execution path
	require.Equal(t, "ExtractUserData", result.Debug.States[0].Name)
	require.Equal(t, "CheckUserType", result.Debug.States[1].Name)
	require.Equal(t, "PremiumUserProcess", result.Debug.States[2].Name)

	// Test business user flow
	businessInput := map[string]interface{}{
		"user": map[string]interface{}{
			"type":  "business",
			"email": "business@example.com",
		},
	}

	ctx := logger.WithFields(context.Background(), logger.String("traceID", "123"))

	result, err = engine.Execute(ctx, workflow, businessInput, map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Write to file
	writeToFile("outputs/email_workflow_business_result.json", []byte(utils.AnyToJSONStringPretty(result.Data)))
	writeToFile("outputs/email_workflow_business_debug.json", []byte(utils.AnyToJSONStringPretty(result.Debug)))

	// Test standard user flow
	standardInput := map[string]interface{}{
		"user": map[string]interface{}{
			"type":  "standard",
			"email": "standard@example.com",
		},
	}

	result, err = engine.Execute(context.Background(), workflow, standardInput, map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Write to file
	writeToFile("outputs/email_workflow_standard_result.json", []byte(utils.AnyToJSONStringPretty(result.Data)))
	writeToFile("outputs/email_workflow_standard_debug.json", []byte(utils.AnyToJSONStringPretty(result.Debug)))
}
