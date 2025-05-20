package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
	"github.com/stretchr/testify/require"
)

func TestHTTPRequest(t *testing.T) {
	activity := New("HTTPRequest", zap.NewLogger(logger.DebugLevel))

	tests := []struct {
		name        string
		server      func() *httptest.Server
		args        map[string]any
		validate    func(*testing.T, interface{})
		expectError bool
	}{
		{
			name: "Successful GET Request",
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, "GET", r.Method)
					json.NewEncoder(w).Encode(map[string]string{"message": "success"})
				}))
			},
			args: map[string]any{
				"method": "GET",
				"url":    "placeholder", // Will be replaced with server URL
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
			},
			validate: func(t *testing.T, result interface{}) {
				resp := result.(map[string]interface{})
				require.Equal(t, int(200), resp["statusCode"])
				body := resp["body"].(string)
				// parse the body as json
				var bodyMap map[string]interface{}
				err := json.Unmarshal([]byte(body), &bodyMap)
				require.NoError(t, err)
				require.Equal(t, "success", bodyMap["message"])
			},
		},
		{
			name: "POST Request with Body",
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, "POST", r.Method)
					var body map[string]interface{}
					json.NewDecoder(r.Body).Decode(&body)
					require.Equal(t, "test data", body["data"])
					w.WriteHeader(http.StatusCreated)
				}))
			},
			args: map[string]any{
				"method": "POST",
				"url":    "placeholder",
				"body": map[string]interface{}{
					"data": "test data",
				},
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
			},
			validate: func(t *testing.T, result interface{}) {
				resp := result.(map[string]interface{})
				require.Equal(t, int(201), resp["statusCode"])
			},
		},
		{
			name: "Request Timeout",
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(2 * time.Second)
				}))
			},
			args: map[string]any{
				"method":     "GET",
				"url":        "placeholder",
				"timeoutSec": 1,
			},
			expectError: true,
		},
		{
			name: "Invalid URL",
			args: map[string]any{
				"method": "GET",
				"url":    "invalid-url",
			},
			expectError: true,
		},
		{
			name: "Error Status with FailOnError",
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			args: map[string]any{
				"method":      "GET",
				"url":         "placeholder",
				"failOnError": true,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.server != nil {
				server = tt.server()
				defer server.Close()
				// Replace placeholder URL with actual server URL
				tt.args["url"] = server.URL
			}

			result, err := activity.Execute(context.Background(), tt.args)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			tt.validate(t, result)
		})
	}
}
