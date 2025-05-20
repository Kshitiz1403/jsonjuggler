package utils

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateArgs(t *testing.T) {
	type testStruct struct {
		URL         string            `arg:"url" required:"true"`
		Method      string            `arg:"method" required:"true" validate:"oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"`
		Headers     map[string]string `arg:"headers"`
		Body        interface{}       `arg:"body"`
		TimeoutSec  int               `arg:"timeoutSec" validate:"min=1,max=300"`
		FailOnError bool              `arg:"failOnError"`
	}

	tests := []struct {
		name      string
		args      map[string]any
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid arguments",
			args: map[string]any{
				"url":         "https://example.com",
				"method":      "GET",
				"timeoutSec":  60,
				"failOnError": true,
			},
			wantError: false,
		},
		{
			name: "missing required url",
			args: map[string]any{
				"method": "GET",
			},
			wantError: true,
			errorMsg:  "required argument 'url' is missing",
		},
		{
			name: "invalid method",
			args: map[string]any{
				"url":    "https://example.com",
				"method": "INVALID",
			},
			wantError: true,
			errorMsg:  "validation failed for field 'Method': expected oneof=GET POST PUT DELETE PATCH HEAD OPTIONS",
		},
		{
			name: "timeout too low",
			args: map[string]any{
				"url":        "https://example.com",
				"method":     "GET",
				"timeoutSec": 0,
			},
			wantError: true,
			errorMsg:  "validation failed for field 'TimeoutSec': expected min=1",
		},
		{
			name: "timeout too high",
			args: map[string]any{
				"url":        "https://example.com",
				"method":     "GET",
				"timeoutSec": 301,
			},
			wantError: true,
			errorMsg:  "validation failed for field 'TimeoutSec': expected max=300",
		},
		{
			name: "with optional fields",
			args: map[string]any{
				"url":     "https://example.com",
				"method":  "POST",
				"headers": map[string]string{"Content-Type": "application/json"},
				"body":    map[string]string{"key": "value"},
			},
			wantError: false,
		},
		{
			name: "wrong type for headers",
			args: map[string]any{
				"url":     "https://example.com",
				"method":  "GET",
				"headers": "invalid-headers-type",
			},
			wantError: true,
			errorMsg:  "failed to decode arguments",
		},
		{
			name: "int field passed in a string field",
			args: map[string]any{
				"url":         123,
				"method":      "GET",
				"timeoutSec":  60,
				"failOnError": true,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s testStruct
			err := ParseAndValidateArgs(context.Background(), tt.args, &s)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateArgs() error = nil, want error containing %q", tt.errorMsg)
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateArgs() error = %v, want error containing %q", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateArgs() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateArgsRequiredVsValidate(t *testing.T) {
	type testStruct struct {
		// Field must be present in input args
		MustBePresentOnly string `arg:"present" required:"true"`

		// Field must be present and non-empty
		MustBePresentAndNonEmpty string `arg:"both" required:"true" validate:"required"`

		// Field can be missing but if present must be non-empty
		OptionalButNonEmpty string `arg:"optional" validate:"required"`

		// Field can be missing and can be empty
		TotallyOptional string `arg:"whatever"`
	}

	tests := []struct {
		name      string
		args      map[string]any
		wantError bool
		errorMsg  string
	}{
		{
			name: "missing required field",
			args: map[string]any{
				"both": "value",
			},
			wantError: true,
			errorMsg:  "required argument 'present' is missing",
		},
		{
			name: "present but empty required field",
			args: map[string]any{
				"present": "",
				"both":    "value",
			},
			wantError: false, // OK because only presence is required
		},
		{
			name: "empty field with validate:required",
			args: map[string]any{
				"present": "value",
				"both":    "", // This will fail validation
			},
			wantError: true,
			errorMsg:  "validation failed",
		},
		{
			name: "missing optional field with validate:required",
			args: map[string]any{
				"present": "value",
				"both":    "value",
				// "optional" is missing - which is OK
			},
			wantError: false,
		},
		{
			name: "present but empty optional field with validate:required",
			args: map[string]any{
				"present":  "value",
				"both":     "value",
				"optional": "", // This will fail validation because it's present but empty
			},
			wantError: true,
			errorMsg:  "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s testStruct
			err := ParseAndValidateArgs(context.Background(), tt.args, &s)

			if tt.wantError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
