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

// ValidActivityStruct has correct method signatures for testing RegisterActivityStruct
type ValidActivityStruct struct {
	BaseActivity
}

func (a *ValidActivityStruct) ProcessData(ctx context.Context, args map[string]any) (interface{}, error) {
	a.GetLogger().InfoContext(ctx, "Processing data")
	if data, ok := args["data"]; ok {
		return map[string]interface{}{
			"processed": data,
			"success":   true,
		}, nil
	}
	return map[string]interface{}{"success": false}, nil
}

func (a *ValidActivityStruct) ValidateInput(ctx context.Context, args map[string]any) (interface{}, error) {
	a.GetLogger().InfoContext(ctx, "Validating input")
	if input, ok := args["input"]; ok && input != nil {
		return map[string]interface{}{"valid": true}, nil
	}
	return nil, errors.New("invalid input")
}

func (a *ValidActivityStruct) CalculateSum(ctx context.Context, args map[string]any) (interface{}, error) {
	a.GetLogger().InfoContext(ctx, "Calculating sum")
	if numbers, ok := args["numbers"].([]interface{}); ok {
		sum := 0.0
		for _, num := range numbers {
			if val, ok := num.(float64); ok {
				sum += val
			}
		}
		return sum, nil
	}
	return 0, nil
}

// Private method should be ignored
func (a *ValidActivityStruct) privateMethod(ctx context.Context, args map[string]any) (interface{}, error) {
	return nil, nil
}

// InvalidActivityStruct has incorrect method signatures
type InvalidActivityStruct struct {
	BaseActivity
}

// Wrong parameter types
func (a *InvalidActivityStruct) WrongParams(ctx context.Context, args string) (interface{}, error) {
	return nil, nil
}

// Wrong return types
func (a *InvalidActivityStruct) WrongReturns(ctx context.Context, args map[string]any) (string, error) {
	return "", nil
}

// Wrong number of parameters
func (a *InvalidActivityStruct) TooManyParams(ctx context.Context, args map[string]any, extra string) (interface{}, error) {
	return nil, nil
}

// Wrong number of returns
func (a *InvalidActivityStruct) TooFewReturns(ctx context.Context, args map[string]any) interface{} {
	return nil
}

// Missing context
func (a *InvalidActivityStruct) NoContext(args map[string]any) (interface{}, error) {
	return nil, nil
}

// MixedActivityStruct has both valid and invalid methods
type MixedActivityStruct struct {
	BaseActivity
}

func (a *MixedActivityStruct) ValidMethod(ctx context.Context, args map[string]any) (interface{}, error) {
	return "valid", nil
}

func (a *MixedActivityStruct) InvalidMethod(ctx context.Context, args string) (interface{}, error) {
	return nil, nil
}

// EmptyActivityStruct has no activity methods (only BaseActivity methods)
type EmptyActivityStruct struct {
	BaseActivity
}

func TestActivityRegistry(t *testing.T) {
	logger := zap.NewLogger(logger.DebugLevel)

	t.Run("Basic Registration", func(t *testing.T) {
		registry := NewRegistry(logger)
		activity := &MockActivity{BaseActivity: &BaseActivity{}}

		// Register activity
		err := registry.RegisterActivity("TestActivity", activity)
		require.NoError(t, err)

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
		err := registry.RegisterActivity("DuplicateActivity", activity1)
		require.NoError(t, err)

		// Try to register second activity with same name - should return error
		err = registry.RegisterActivity("DuplicateActivity", activity2)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already registered")

		// Verify first registration remains
		retrieved, exists := registry.Get("DuplicateActivity")
		require.True(t, exists)

		// The retrieved activity is wrapped, so we need to check the underlying activity
		wrapper, ok := retrieved.(*activityWrapper)
		require.True(t, ok, "Retrieved activity should be wrapped")
		require.Equal(t, activity1, wrapper.activity)
	})

	t.Run("Logger Injection", func(t *testing.T) {
		registry := NewRegistry(logger)
		activity := &MockActivity{BaseActivity: &BaseActivity{}}

		err := registry.RegisterActivity("LoggerTest", activity)
		require.NoError(t, err)

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

		err := registry.RegisterActivity("MockActivity", activity)
		require.NoError(t, err)

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

		err := registry.RegisterActivity("MockActivityWithError", activity)
		require.NoError(t, err)

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

		err := registry.RegisterActivity("MockActivityWithArgs", activity)
		require.NoError(t, err)

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

		err := registry.RegisterActivity("MockActivityWithArgsAndError", activity)
		require.NoError(t, err)

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

func TestRegisterActivityStruct(t *testing.T) {
	logger := zap.NewLogger(logger.DebugLevel)

	t.Run("Valid Activity Struct Registration", func(t *testing.T) {
		registry := NewRegistry(logger)
		validActivity := &ValidActivityStruct{}

		// Register the struct
		err := registry.RegisterActivityStruct(validActivity)
		require.NoError(t, err)

		// Verify all valid methods were registered
		expectedMethods := []string{"ProcessData", "ValidateInput", "CalculateSum"}
		for _, methodName := range expectedMethods {
			activity, exists := registry.Get(methodName)
			require.True(t, exists, "Method %s should be registered", methodName)
			require.NotNil(t, activity)
			require.Equal(t, methodName, activity.GetActivityName())
			require.NotNil(t, activity.GetLogger())
		}

		// Verify private method was not registered
		_, exists := registry.Get("privateMethod")
		require.False(t, exists, "Private method should not be registered")

		// Verify BaseActivity methods were not registered
		baseActivityMethods := []string{"GetLogger", "GetActivityName", "setLogger", "setActivityName"}
		for _, methodName := range baseActivityMethods {
			_, exists := registry.Get(methodName)
			require.False(t, exists, "BaseActivity method %s should not be registered", methodName)
		}
	})

	t.Run("BaseActivity Auto-Initialization", func(t *testing.T) {
		registry := NewRegistry(logger)
		validActivity := &ValidActivityStruct{}

		// Verify BaseActivity is initially nil/empty
		require.Nil(t, validActivity.BaseActivity.Logger)

		// Register the struct
		err := registry.RegisterActivityStruct(validActivity)
		require.NoError(t, err)

		// Verify BaseActivity was auto-initialized
		require.NotNil(t, validActivity.BaseActivity.Logger)
		require.Equal(t, logger, validActivity.BaseActivity.Logger)
	})

	t.Run("Invalid Activity Struct Registration", func(t *testing.T) {
		registry := NewRegistry(logger)
		invalidActivity := &InvalidActivityStruct{}

		// Register the struct - should fail due to no valid methods found
		err := registry.RegisterActivityStruct(invalidActivity)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no activities (public methods) found")

		// Verify no methods were registered due to all being invalid
		invalidMethods := []string{"WrongParams", "WrongReturns", "TooManyParams", "TooFewReturns", "NoContext"}
		for _, methodName := range invalidMethods {
			_, exists := registry.Get(methodName)
			require.False(t, exists, "Invalid method %s should not be registered", methodName)
		}
	})

	t.Run("Mixed Valid/Invalid Activity Struct", func(t *testing.T) {
		registry := NewRegistry(logger)
		mixedActivity := &MixedActivityStruct{}

		// Register the struct - should succeed but skip invalid methods
		err := registry.RegisterActivityStruct(mixedActivity)
		require.NoError(t, err)

		// Verify valid method was registered
		activity, exists := registry.Get("ValidMethod")
		require.True(t, exists)
		require.NotNil(t, activity)

		// Verify invalid method was not registered
		_, exists = registry.Get("InvalidMethod")
		require.False(t, exists)
	})

	t.Run("Empty Activity Struct Registration", func(t *testing.T) {
		registry := NewRegistry(logger)
		emptyActivity := &EmptyActivityStruct{}

		// Register the struct - should fail as no valid activity methods found
		err := registry.RegisterActivityStruct(emptyActivity)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no activities (public methods) found")
	})

	t.Run("Activity Struct Method Execution", func(t *testing.T) {
		registry := NewRegistry(logger)
		validActivity := &ValidActivityStruct{}

		// Register the struct
		err := registry.RegisterActivityStruct(validActivity)
		require.NoError(t, err)

		// Test ProcessData method execution
		activity, exists := registry.Get("ProcessData")
		require.True(t, exists)

		ctx := context.Background()
		args := map[string]any{"data": "test input"}
		result, err := activity.Execute(ctx, args)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "test input", resultMap["processed"])
		require.Equal(t, true, resultMap["success"])
	})

	t.Run("Activity Struct Method Execution with Error", func(t *testing.T) {
		registry := NewRegistry(logger)
		validActivity := &ValidActivityStruct{}

		// Register the struct
		err := registry.RegisterActivityStruct(validActivity)
		require.NoError(t, err)

		// Test ValidateInput method execution with invalid input
		activity, exists := registry.Get("ValidateInput")
		require.True(t, exists)

		ctx := context.Background()
		args := map[string]any{} // No input provided
		result, err := activity.Execute(ctx, args)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "invalid input", err.Error())
	})

	t.Run("Activity Struct Method with Complex Logic", func(t *testing.T) {
		registry := NewRegistry(logger)
		validActivity := &ValidActivityStruct{}

		// Register the struct
		err := registry.RegisterActivityStruct(validActivity)
		require.NoError(t, err)

		// Test CalculateSum method execution
		activity, exists := registry.Get("CalculateSum")
		require.True(t, exists)

		ctx := context.Background()
		args := map[string]any{
			"numbers": []interface{}{1.0, 2.5, 3.7, 4.2},
		}
		result, err := activity.Execute(ctx, args)
		require.NoError(t, err)
		require.Equal(t, 11.4, result)
	})

	t.Run("Non-Pointer Struct Registration", func(t *testing.T) {
		registry := NewRegistry(logger)
		validActivity := ValidActivityStruct{} // Not a pointer

		// Should handle non-pointer struct gracefully
		err := registry.RegisterActivityStruct(validActivity)
		require.Error(t, err) // Should fail because we can't initialize BaseActivity on non-pointer
	})

	t.Run("Nil Struct Registration", func(t *testing.T) {
		registry := NewRegistry(logger)

		// Should handle nil gracefully
		err := registry.RegisterActivityStruct(nil)
		require.Error(t, err)
	})

	t.Run("Non-Struct Registration", func(t *testing.T) {
		registry := NewRegistry(logger)

		// Should handle non-struct types gracefully
		err := registry.RegisterActivityStruct("not a struct")
		require.Error(t, err)
	})
}
