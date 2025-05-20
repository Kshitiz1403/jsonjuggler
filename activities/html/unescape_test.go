package html

import (
	"context"
	"testing"

	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
	"github.com/stretchr/testify/require"
)

func TestHTMLUnescapeActivity(t *testing.T) {
	activity := New("HTMLUnescape", zap.NewLogger(logger.DebugLevel))

	tests := []struct {
		name        string
		args        map[string]any
		expected    string
		expectError bool
	}{
		{
			name: "Basic HTML Entities",
			args: map[string]any{
				"text": "&lt;div&gt;Hello &amp; World&lt;/div&gt;",
			},
			expected: "<div>Hello & World</div>",
		},
		{
			name: "Special Characters",
			args: map[string]any{
				"text": "Copyright &copy; 2024 &reg; &trade;",
			},
			expected: "Copyright © 2024 ® ™",
		},
		{
			name: "Mixed Content",
			args: map[string]any{
				"text": "Hello &quot;World&quot; &amp; Universe!",
			},
			expected: "Hello \"World\" & Universe!",
		},
		{
			name: "No Entities",
			args: map[string]any{
				"text": "Plain text without entities",
			},
			expected: "Plain text without entities",
		},
		{
			name:        "Missing Required Argument",
			args:        map[string]any{},
			expectError: true,
		},
		{
			name: "Invalid Argument Type",
			args: map[string]any{
				"text": 123,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := activity.Execute(context.Background(), tt.args)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, result)
		})
	}
}
