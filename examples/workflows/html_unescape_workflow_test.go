package workflows

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kshitiz1403/jsonjuggler/config"
	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
	"github.com/kshitiz1403/jsonjuggler/parser"
	"github.com/kshitiz1403/jsonjuggler/utils"
	"github.com/stretchr/testify/require"
)

func TestHTMLUnescapeWorkflow(t *testing.T) {
	// Initialize engine with debug mode
	engine := config.Initialize(
		config.WithDebug(true),
		config.WithLogger(zap.NewLogger(logger.DebugLevel)),
	)

	// Parse workflow
	p := parser.NewParser(engine.GetRegistry())
	workflow, err := p.ParseFromFile("html_unescape_workflow.json")
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    map[string]interface{}
		validate func(*testing.T, interface{})
	}{
		{
			name: "Basic HTML Entities",
			input: map[string]interface{}{
				"htmlText": "&lt;div&gt;Hello &amp; World&lt;/div&gt;",
			},
			validate: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				require.Equal(t, "<div>Hello & World</div>", resultMap["unescaped"])
			},
		},
		{
			name: "Special Characters",
			input: map[string]interface{}{
				"htmlText": "Copyright &copy; 2024 &reg; &trade;",
			},
			validate: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				require.Equal(t, "Copyright © 2024 ® ™", resultMap["unescaped"])
			},
		},
		{
			name: "Mixed Content",
			input: map[string]interface{}{
				"htmlText": "Hello &quot;World&quot; &amp; Universe!",
			},
			validate: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				require.Equal(t, "Hello \"World\" & Universe!", resultMap["unescaped"])
			},
		},
		{
			name: "No Entities",
			input: map[string]interface{}{
				"htmlText": "Plain text without entities",
			},
			validate: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				require.Equal(t, "Plain text without entities", resultMap["unescaped"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context with test name
			ctx := logger.WithFields(context.Background(),
				logger.String("testName", tt.name),
			)

			// Execute workflow
			result, err := engine.Execute(ctx, workflow, tt.input, nil)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Write to file
			writeToFile(fmt.Sprintf("outputs/html_unescape_workflow_result_%s.json", strings.ToLower(strings.ReplaceAll(tt.name, " ", "_"))), []byte(utils.AnyToJSONStringPretty(result.Data)))
			writeToFile(fmt.Sprintf("outputs/html_unescape_workflow_debug_%s.json", strings.ToLower(strings.ReplaceAll(tt.name, " ", "_"))), []byte(utils.AnyToJSONStringPretty(result.Debug)))

			// Validate result
			tt.validate(t, result.Data)

			// Verify debug information is present
			require.NotNil(t, result.Debug)
			require.Len(t, result.Debug.States, 2) // Should have two states: UnescapeHTML and FormatOutput
			require.Equal(t, "UnescapeHTML", result.Debug.States[0].Name)
			require.Equal(t, "FormatOutput", result.Debug.States[1].Name)
		})
	}
}
