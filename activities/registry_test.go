package activities

import (
	"context"
	"errors"
	"testing"

	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
	"github.com/stretchr/testify/require"
)

// MockActivity implements a simple test activity
type MockActivity struct {
	*BaseActivity
	ExecuteFunc func(ctx context.Context, args map[string]any) (interface{}, error)
}

func (a *MockActivity) Execute(ctx context.Context, args map[string]any) (interface{}, error) {
	if a.ExecuteFunc != nil {
		return a.ExecuteFunc(ctx, args)
	}
	return "mock result", nil
}

func TestActivityRegistry(t *testing.T) {
	logger := zap.NewLogger(logger.DebugLevel)

	t.Run("Basic Registration", func(t *testing.T) {
		registry := NewRegistry(logger)
		activity := &MockActivity{BaseActivity: &BaseActivity{}}

		// Register activity
		registry.Register("TestActivity", activity)

		// Verify registration
		retrieved, exists := registry.Get("TestActivity")
		require.True(t, exists)
		require.NotNil(t, retrieved)
		require.Equal(t, "TestActivity", retrieved.GetActivityName())
		require.NotNil(t, retrieved.GetLogger())
	})

	t.Run("Duplicate Registration", func(t *testing.T) {
		registry := NewRegistry(logger)
		activity1 := &MockActivity{BaseActivity: &BaseActivity{}}
		activity2 := &MockActivity{BaseActivity: &BaseActivity{}}

		// Register first activity
		registry.Register("DuplicateActivity", activity1)

		// Try to register second activity with same name
		registry.Register("DuplicateActivity", activity2)

		// Verify first registration remains
		retrieved, exists := registry.Get("DuplicateActivity")
		require.True(t, exists)
		require.Equal(t, activity1, retrieved)
	})

	t.Run("Logger Injection", func(t *testing.T) {
		registry := NewRegistry(logger)
		activity := &MockActivity{BaseActivity: &BaseActivity{}}

		registry.Register("LoggerTest", activity)

		retrieved, exists := registry.Get("LoggerTest")
		require.True(t, exists)
		require.NotNil(t, retrieved.GetLogger())
		require.Equal(t, logger, retrieved.GetLogger())
	})

	t.Run("Activity Execution", func(t *testing.T) {
		registry := NewRegistry(logger)
		activity := &MockActivity{BaseActivity: &BaseActivity{}, ExecuteFunc: func(ctx context.Context, args map[string]any) (interface{}, error) {
			return "mock result", nil
		}}

		registry.Register("MockActivity", activity)

		retrieved, exists := registry.Get("MockActivity")
		require.True(t, exists)

		result, err := retrieved.Execute(context.Background(), map[string]any{"key": "value"})
		require.NoError(t, err)
		require.Equal(t, "mock result", result)
	})

	t.Run("Activity Execution with Error", func(t *testing.T) {
		registry := NewRegistry(logger)
		activity := &MockActivity{BaseActivity: &BaseActivity{}, ExecuteFunc: func(ctx context.Context, args map[string]any) (interface{}, error) {
			return nil, errors.New("mock error")
		}}

		registry.Register("MockActivityWithError", activity)

		retrieved, exists := registry.Get("MockActivityWithError")
		require.True(t, exists)

		result, err := retrieved.Execute(context.Background(), map[string]any{"key": "value"})
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "mock error", err.Error())
	})

	t.Run("Activity Execution with some args", func(t *testing.T) {
		registry := NewRegistry(logger)
		activity := &MockActivity{BaseActivity: &BaseActivity{}, ExecuteFunc: func(ctx context.Context, args map[string]any) (interface{}, error) {
			return args["key"], nil
		}}

		registry.Register("MockActivityWithArgs", activity)

		retrieved, exists := registry.Get("MockActivityWithArgs")
		require.True(t, exists)

		result, err := retrieved.Execute(context.Background(), map[string]any{"key": "value"})
		require.NoError(t, err)
		require.Equal(t, "value", result)
	})

	t.Run("Activity Execution with some args and error", func(t *testing.T) {
		registry := NewRegistry(logger)
		activity := &MockActivity{BaseActivity: &BaseActivity{}, ExecuteFunc: func(ctx context.Context, args map[string]any) (interface{}, error) {
			if args["key"] == "error" {
				return nil, errors.New("mock error")
			}
			return args["key"], nil
		}}

		registry.Register("MockActivityWithArgsAndError", activity)

		retrieved, exists := registry.Get("MockActivityWithArgsAndError")
		require.True(t, exists)

		result, err := retrieved.Execute(context.Background(), map[string]any{"key": "error"})
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "mock error", err.Error())

		result, err = retrieved.Execute(context.Background(), map[string]any{"key": "value"})
		require.NoError(t, err)
		require.Equal(t, "value", result)
	})

}
