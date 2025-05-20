package workflows

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kshitiz1403/jsonjuggler/config"
	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
	"github.com/kshitiz1403/jsonjuggler/parser"
	"github.com/kshitiz1403/jsonjuggler/utils"
	"github.com/stretchr/testify/require"
)

func TestErrorHandlingWorkflow(t *testing.T) {
	// Initialize engine with debugging enabled
	engine := config.Initialize(
		config.WithDebug(true),
		config.WithLogger(zap.NewLogger(logger.DebugLevel)),
	)

	// Parse workflow definition
	p := parser.NewParser(engine.GetRegistry())
	workflow, err := p.ParseFromFile("error_handling_workflow.json")
	require.NoError(t, err)

	tests := []struct {
		name          string
		input         map[string]interface{}
		globals       map[string]interface{}
		mockServer    func() *httptest.Server
		expectedState string
		expectError   bool
	}{
		{
			name: "Successful Order Processing",
			input: map[string]interface{}{
				"order": map[string]interface{}{
					"id":     "ORD123",
					"amount": 100.50,
					"items":  []string{"item1", "item2"},
				},
			},
			globals: map[string]interface{}{},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"status":"success"}`))
				}))
			},
			expectedState: "CompleteOrder",
			expectError:   false,
		},
		{
			name: "Invalid Order Amount",
			input: map[string]interface{}{
				"order": map[string]interface{}{
					"id":     "ORD124",
					"amount": -50.00,
					"items":  []string{"item1"},
				},
			},
			globals: map[string]interface{}{},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
			},
			expectError: true, // This should fail at validation
		},
		{
			name: "Connection Error",
			input: map[string]interface{}{
				"order": map[string]interface{}{
					"id":     "ORD125",
					"amount": 75.00,
					"items":  []string{"item1"},
				},
			},
			globals: map[string]interface{}{},
			mockServer: func() *httptest.Server {
				// Return a server that's immediately closed
				server := httptest.NewServer(nil)
				server.Close()
				return server
			},
			expectedState: "HandleConnectionError",
			expectError:   false,
		},
		{
			name: "Timeout Error",
			input: map[string]interface{}{
				"order": map[string]interface{}{
					"id":     "ORD126",
					"amount": 200.00,
					"items":  []string{"item1", "item2", "item3"},
				},
			},
			globals: map[string]interface{}{
				"timeoutSeconds": 1,
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(3 * time.Second) // Force timeout
					w.WriteHeader(http.StatusOK)
				}))
			},
			expectedState: "HandleTimeout",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start mock server
			mockServer := tt.mockServer()
			defer mockServer.Close()

			// Execute workflow

			globals := tt.globals
			globals["serverURL"] = mockServer.URL

			ctx := logger.WithFields(context.Background(), logger.String("traceID", "3283721"))

			result, err := engine.Execute(ctx, workflow, tt.input, globals)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			// Write results to files
			writeToFile(
				"outputs/error_handling_"+tt.name+"_result.json",
				[]byte(utils.AnyToJSONStringPretty(result.Data)),
			)
			writeToFile(
				"outputs/error_handling_"+tt.name+"_debug.json",
				[]byte(utils.AnyToJSONStringPretty(result.Debug)),
			)

			// Verify the execution path
			if tt.expectedState != "" {
				lastState := result.Debug.States[len(result.Debug.States)-1]
				require.Equal(t, tt.expectedState, lastState.Name)
			}
		})
	}
}
