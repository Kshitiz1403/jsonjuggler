package jq

import (
	"context"
	"testing"

	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
	"github.com/stretchr/testify/require"
)

func TestJQTransform(t *testing.T) {
	activity := New("JQ", zap.NewLogger(logger.DebugLevel))

	tests := []struct {
		name        string
		args        map[string]any
		expected    interface{}
		expectError bool
	}{
		{
			name: "Simple Object Transform",
			args: map[string]any{
				"query": "{ name: .user.name, email: .user.email }",
				"data": map[string]interface{}{
					"user": map[string]interface{}{
						"name":  "John Doe",
						"email": "john@example.com",
						"age":   30,
					},
				},
			},
			expected: map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
			},
		},
		{
			name: "Array Manipulation",
			args: map[string]any{
				"query": ".items | map(select(.price > 10)) | map({name: .name, price: (.price | tonumber)})",
				"data": map[string]interface{}{
					"items": []interface{}{
						map[string]interface{}{"name": "Item 1", "price": 5},
						map[string]interface{}{"name": "Item 2", "price": 15},
						map[string]interface{}{"name": "Item 3", "price": 25},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{"name": "Item 2", "price": int(15)},
				map[string]interface{}{"name": "Item 3", "price": int(25)},
			},
		},
		{
			name: "Invalid Query",
			args: map[string]any{
				"query": "invalid query",
				"data":  map[string]interface{}{},
			},
			expectError: true,
		},
		{
			name: "Missing Required Arguments",
			args: map[string]any{
				"data": map[string]interface{}{},
			},
			expectError: true,
		},
		{
			name: "Complex Transformation",
			args: map[string]any{
				"query": `{
					users: [.users[] | {name: .name, active: (.status == "active")}],
					totalActive: ([.users[] | select(.status == "active")] | length | tonumber)
				}`,
				"data": map[string]interface{}{
					"users": []interface{}{
						map[string]interface{}{"name": "John", "status": "active"},
						map[string]interface{}{"name": "Jane", "status": "inactive"},
						map[string]interface{}{"name": "Bob", "status": "active"},
					},
				},
			},
			expected: map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{"name": "John", "active": true},
					map[string]interface{}{"name": "Jane", "active": false},
					map[string]interface{}{"name": "Bob", "active": true},
				},
				"totalActive": int(2),
			},
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
