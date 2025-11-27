package activities

import (
	"context"
	"testing"

	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
	"github.com/stretchr/testify/require"
)

// TestActivity for testing activity info
type TestActivity struct {
	BaseActivity
	LastActivityInfo *ActivityInfo
}

func (a *TestActivity) Execute(ctx context.Context, args map[string]any) (interface{}, error) {
	// Capture activity info for verification
	a.LastActivityInfo = GetInfo(ctx)

	// Use activity info like in Temporal
	activityInfo := GetInfo(ctx)
	activityInfo.Logger.InfoContext(ctx, "Executing test activity")

	return map[string]interface{}{
		"activity_name": activityInfo.ActivityName,
		"has_logger":    activityInfo.Logger != nil,
	}, nil
}

// TestActivityStruct for testing struct-based registration with activity info
type TestActivityStruct struct {
	BaseActivity
	LastActivityInfo *ActivityInfo
}

func (a *TestActivityStruct) TestMethod(ctx context.Context, args map[string]any) (interface{}, error) {
	// Capture activity info for verification
	a.LastActivityInfo = GetInfo(ctx)

	activityInfo := GetInfo(ctx)
	return map[string]interface{}{
		"activity_name": activityInfo.ActivityName,
		"has_logger":    activityInfo.Logger != nil,
	}, nil
}

func TestActivityInfo(t *testing.T) {
	logger := zap.NewLogger(logger.DebugLevel)

	t.Run("GetInfo from Context", func(t *testing.T) {
		activityInfo := &ActivityInfo{
			ActivityName: "TestActivity",
			Logger:       logger,
		}

		ctx := withActivityInfo(context.Background(), activityInfo)

		// Test GetInfo retrieval
		retrievedInfo := GetInfo(ctx)
		require.NotNil(t, retrievedInfo)
		require.Equal(t, "TestActivity", retrievedInfo.ActivityName)
		require.Equal(t, logger, retrievedInfo.Logger)
	})

	t.Run("GetInfo without Context", func(t *testing.T) {
		// Test GetInfo with no activity info in context
		retrievedInfo := GetInfo(context.Background())
		require.NotNil(t, retrievedInfo)
		require.Equal(t, "unknown", retrievedInfo.ActivityName)
		require.Nil(t, retrievedInfo.Logger)
	})

	t.Run("HasActivityInfo", func(t *testing.T) {
		activityInfo := &ActivityInfo{
			ActivityName: "TestActivity",
			Logger:       logger,
		}

		// Context with activity info
		ctxWithInfo := withActivityInfo(context.Background(), activityInfo)
		require.True(t, HasActivityInfo(ctxWithInfo))

		// Context without activity info
		require.False(t, HasActivityInfo(context.Background()))
	})

	t.Run("Single Activity Registration with ActivityInfo", func(t *testing.T) {
		registry := NewRegistry(logger)
		activity := &TestActivity{BaseActivity: BaseActivity{}}

		// Register activity
		err := registry.RegisterActivity("TestSingleActivity", activity)
		require.NoError(t, err)

		// Execute activity
		retrievedActivity, exists := registry.Get("TestSingleActivity")
		require.True(t, exists)

		result, err := retrievedActivity.Execute(context.Background(), map[string]any{})
		require.NoError(t, err)

		// Verify activity info was injected
		require.NotNil(t, activity.LastActivityInfo)
		require.Equal(t, "TestSingleActivity", activity.LastActivityInfo.ActivityName)
		require.Equal(t, logger, activity.LastActivityInfo.Logger)

		// Verify result contains activity info
		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "TestSingleActivity", resultMap["activity_name"])
		require.Equal(t, true, resultMap["has_logger"])
	})

	t.Run("Struct Activity Registration with ActivityInfo", func(t *testing.T) {
		registry := NewRegistry(logger)
		activity := &TestActivityStruct{}

		// Register activity struct
		err := registry.RegisterActivityStruct(activity)
		require.NoError(t, err)

		// Execute activity method
		retrievedActivity, exists := registry.Get("TestMethod")
		require.True(t, exists)

		result, err := retrievedActivity.Execute(context.Background(), map[string]any{})
		require.NoError(t, err)

		// Verify activity info was injected
		require.NotNil(t, activity.LastActivityInfo)
		require.Equal(t, "TestMethod", activity.LastActivityInfo.ActivityName)
		require.Equal(t, logger, activity.LastActivityInfo.Logger)

		// Verify result contains activity info
		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "TestMethod", resultMap["activity_name"])
		require.Equal(t, true, resultMap["has_logger"])
	})

	t.Run("Activity Info Context Isolation", func(t *testing.T) {
		registry := NewRegistry(logger)
		activity1 := &TestActivity{BaseActivity: BaseActivity{}}
		activity2 := &TestActivity{BaseActivity: BaseActivity{}}

		// Register two different activities
		err := registry.RegisterActivity("Activity1", activity1)
		require.NoError(t, err)
		err = registry.RegisterActivity("Activity2", activity2)
		require.NoError(t, err)

		// Execute both activities
		retrievedActivity1, _ := registry.Get("Activity1")
		retrievedActivity2, _ := registry.Get("Activity2")

		_, err = retrievedActivity1.Execute(context.Background(), map[string]any{})
		require.NoError(t, err)
		_, err = retrievedActivity2.Execute(context.Background(), map[string]any{})
		require.NoError(t, err)

		// Verify each activity received its own correct info
		require.Equal(t, "Activity1", activity1.LastActivityInfo.ActivityName)
		require.Equal(t, "Activity2", activity2.LastActivityInfo.ActivityName)
	})

	t.Run("Activity Info with Custom Context Values", func(t *testing.T) {
		registry := NewRegistry(logger)
		activity := &TestActivity{BaseActivity: BaseActivity{}}

		err := registry.RegisterActivity("TestActivity", activity)
		require.NoError(t, err)

		// Create context with custom values
		ctx := context.WithValue(context.Background(), "custom_key", "custom_value")

		retrievedActivity, exists := registry.Get("TestActivity")
		require.True(t, exists)

		_, err = retrievedActivity.Execute(ctx, map[string]any{})
		require.NoError(t, err)

		// Verify activity info was injected while preserving original context
		require.NotNil(t, activity.LastActivityInfo)
		require.Equal(t, "TestActivity", activity.LastActivityInfo.ActivityName)

		// Verify original context values are still accessible (this would be tested in the activity)
		// For now, we just verify the activity info was properly injected
	})
}
