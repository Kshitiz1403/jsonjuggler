package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kshitiz1403/jsonjuggler/activities"
	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/utils"
)

// RequestArgs represents the arguments for the HTTP request activity
type RequestArgs struct {
	URL         string            `arg:"url" required:"true"`
	Method      string            `arg:"method" required:"true" validate:"oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"`
	Headers     map[string]string `arg:"headers"`
	Body        interface{}       `arg:"body"`
	TimeoutSec  int               `arg:"timeoutSec" default:"30"`
	FailOnError bool              `arg:"failOnError"`
}

// Response represents an HTTP response
type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       interface{}       `json:"body"`
}

func (r *Response) ToMap() map[string]any {
	return map[string]any{
		"statusCode": r.StatusCode,
		"headers":    r.Headers,
		"body":       r.Body,
	}
}

// RequestActivity performs HTTP requests
type RequestActivity struct {
	*activities.BaseActivity
	client *http.Client
}

// New creates a new HTTP request activity
func New(activityName string, logger logger.Logger) *RequestActivity {
	return &RequestActivity{
		BaseActivity: &activities.BaseActivity{
			ActivityName: activityName,
			Logger:       logger,
		},
		client: &http.Client{},
	}
}

func (a *RequestActivity) Execute(ctx context.Context, arguments map[string]any) (interface{}, error) {
	var args RequestArgs
	if err := utils.ParseAndValidateArgs(ctx, arguments, &args); err != nil {
		a.GetLogger().ErrorContextf(ctx, "Invalid HTTP request arguments: %v", err)
		return nil, activities.NewActivityError(
			activities.ErrInvalidArguments,
			"Invalid HTTP request arguments",
			"HTTPRequest",
		).WithArguments(arguments).WithCause(err)
	}

	// Set timeout
	if args.TimeoutSec > 0 {
		a.client.Timeout = time.Duration(args.TimeoutSec) * time.Second
	}

	a.GetLogger().DebugContextf(ctx, "Making HTTP request to %s", args.URL)

	// Prepare request body
	var bodyReader io.Reader
	if args.Body != nil {
		bodyBytes, err := json.Marshal(args.Body)
		if err != nil {
			return nil, activities.NewActivityError(
				activities.ErrExecutionFailed,
				"Failed to marshal request body",
				"HTTPRequest",
			).WithCause(err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, args.Method, args.URL, bodyReader)
	if err != nil {
		return nil, activities.NewActivityError(
			activities.ErrHTTPRequestFailed,
			"Failed to create HTTP request",
			"HTTPRequest",
		).WithArguments(map[string]interface{}{
			"method": args.Method,
			"url":    args.URL,
		}).WithCause(err)
	}

	// Set headers
	for k, v := range args.Headers {
		req.Header.Set(k, v)
	}

	// Set default content-type if not provided and body exists
	if args.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, activities.NewActivityError(
			activities.ErrHTTPRequestFailed,
			"HTTP request failed",
			"HTTPRequest",
		).WithCause(err)
	}
	defer resp.Body.Close()

	a.GetLogger().DebugContextf(ctx, "HTTP request completed with status %d", resp.StatusCode)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, activities.NewActivityError(
			activities.ErrHTTPResponseFailed,
			"Failed to read response body",
			"HTTPRequest",
		).WithCause(err)
	}

	// Parse response body as JSON if content-type is application/json
	var parsedBody interface{}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "application/json" {
		if err := json.Unmarshal(respBody, &parsedBody); err != nil {
			// If JSON parsing fails, use raw body
			parsedBody = string(respBody)
		}
	} else {
		parsedBody = string(respBody)
	}

	// Convert headers to map
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	response := &Response{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       parsedBody,
	}

	// Check if we should fail on non-2xx status codes
	if args.FailOnError && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		return response, activities.NewActivityError(
			activities.ErrHTTPStatusError,
			fmt.Sprintf("Request failed with status code %d", resp.StatusCode),
			"HTTPRequest",
		).WithArguments(map[string]interface{}{
			"statusCode": resp.StatusCode,
			"response":   response.ToMap(),
		})
	}

	return response.ToMap(), nil
}
