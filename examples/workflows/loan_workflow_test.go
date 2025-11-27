package workflows

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kshitiz1403/jsonjuggler/config"
	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
	"github.com/kshitiz1403/jsonjuggler/parser"
	"github.com/kshitiz1403/jsonjuggler/utils"
	"github.com/stretchr/testify/require"
)

func TestLoanWorkflow(t *testing.T) {
	engine, err := config.Initialize(
		config.WithDebug(true),
		config.WithLogger(zap.NewLogger(logger.DebugLevel)),
	)
	require.NoError(t, err)

	p := parser.NewParser(engine.GetRegistry())
	workflow, err := p.ParseFromFile("loan_workflow.json")
	require.NoError(t, err)

	testCases := []struct {
		name     string
		input    map[string]interface{}
		validate func(t *testing.T, result interface{}, debug map[string]interface{})
	}{
		{
			name: "High Risk Application",
			input: map[string]interface{}{
				"application": map[string]interface{}{
					"id": 12345,
					"user": map[string]interface{}{
						"id":     "USER123",
						"name":   "John Doe",
						"income": 50000,
					},
					"loan": map[string]interface{}{
						"amount":  150000,
						"term":    36,
						"purpose": "home_improvement",
					},
				},
			},
			validate: func(t *testing.T, result interface{}, debug map[string]interface{}) {
				states := debug["states"].([]interface{})
				// Verify state transitions
				require.Equal(t, "ExtractApplication", states[0].(map[string]interface{})["name"])
				require.Equal(t, "EnrichUserData", states[1].(map[string]interface{})["name"])
				require.Equal(t, "CalculateRiskScore", states[2].(map[string]interface{})["name"])
				require.Equal(t, "EvaluateApplication", states[3].(map[string]interface{})["name"])
				require.Equal(t, "RejectApplication", states[4].(map[string]interface{})["name"])

				// Verify final result
				resultMap := result.(map[string]interface{})
				require.Equal(t, "rejected", resultMap["decision"])
				require.NotNil(t, resultMap["riskScore"])
				require.NotNil(t, resultMap["factors"])
				require.NotEmpty(t, resultMap["timestamp"])
			},
		},
		{
			name: "Medium Risk Application",
			input: map[string]interface{}{
				"application": map[string]interface{}{
					"id": 12346,
					"user": map[string]interface{}{
						"id":     "USER124",
						"name":   "Jane Smith",
						"income": 80000,
					},
					"loan": map[string]interface{}{
						"amount":  150000,
						"term":    36,
						"purpose": "debt_consolidation",
					},
				},
			},
			validate: func(t *testing.T, result interface{}, debug map[string]interface{}) {
				states := debug["states"].([]interface{})
				lastState := states[len(states)-1].(map[string]interface{})
				require.Equal(t, "RequestAdditionalDocuments", lastState["name"])

				// Verify the request for additional documents
				actions := lastState["actions"].([]interface{})
				output := actions[0].(map[string]interface{})["output"]
				require.NotNil(t, output)

				// Verify risk score in headers
				args := actions[1].(map[string]interface{})["arguments"].(map[string]interface{})
				headers := args["headers"].(map[string]interface{})
				require.NotEmpty(t, headers["X-Risk-Score"])
				require.NotEmpty(t, headers["X-User-ID"])
			},
		},
		{
			name: "Low Risk Application",
			input: map[string]interface{}{
				"application": map[string]interface{}{
					"id": 12347,
					"user": map[string]interface{}{
						"id":     "USER125",
						"name":   "Alice Johnson",
						"income": 200000,
					},
					"loan": map[string]interface{}{
						"amount":  150000,
						"term":    36,
						"purpose": "business",
					},
				},
			},
			validate: func(t *testing.T, result interface{}, debug map[string]interface{}) {
				resultMap := result.(map[string]interface{})
				require.Equal(t, "approved", resultMap["decision"])
				require.NotNil(t, resultMap["riskScore"])
				require.NotNil(t, resultMap["approvedAmount"])
				require.NotEmpty(t, resultMap["timestamp"])

				// Verify the complete flow
				states := debug["states"].([]interface{})
				var stateNames []string
				for _, s := range states {
					stateNames = append(stateNames, s.(map[string]interface{})["name"].(string))
				}
				require.Equal(t, []string{
					"ExtractApplication",
					"EnrichUserData",
					"CalculateRiskScore",
					"EvaluateApplication",
					"ApproveApplication",
					"GetResponseBody",
				}, stateNames)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctx := logger.WithFields(context.Background(), logger.String("traceID", "48r909"))

			result, err := engine.Execute(ctx, workflow, tc.input, nil)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Write results to files
			fileName := "loan_workflow_" + strings.ToLower(strings.ReplaceAll(tc.name, " ", "_"))
			writeToFile(
				filepath.Join("outputs", fileName+"_result.json"),
				[]byte(utils.AnyToJSONStringPretty(result.Data)),
			)
			writeToFile(
				filepath.Join("outputs", fileName+"_debug.json"),
				[]byte(utils.AnyToJSONStringPretty(result.Debug)),
			)

			// Convert debug to map for easier testing
			debugMap := make(map[string]interface{})
			debugBytes, _ := json.Marshal(result.Debug)
			json.Unmarshal(debugBytes, &debugMap)

			// Run validation
			tc.validate(t, result.Data, debugMap)
		})
	}
}
