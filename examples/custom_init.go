package main

import (
	"context"

	"github.com/kshitiz1403/jsonjuggler/activities"
	"github.com/kshitiz1403/jsonjuggler/config"
	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
	"github.com/kshitiz1403/jsonjuggler/parser"
	"github.com/kshitiz1403/jsonjuggler/utils"
)

// CustomHelloWorldActivity is an example custom activity
type CustomHelloWorldActivity struct {
	activities.BaseActivity // Always inherit BaseActivity, this auto injects all required dependencies into the activity
}

func (a *CustomHelloWorldActivity) Execute(ctx context.Context, args map[string]any) (interface{}, error) {
	// Use the logger via GetLogger()
	a.GetLogger().InfoContext(ctx, "Starting custom operation")

	// Custom implementation
	result := "hello world"

	a.GetLogger().DebugContextf(ctx, "Custom operation completed with result: %v", result)
	return result, nil
}

func main() {
	// Initialize logger, defaults to info level
	customLogger := zap.NewLogger(logger.DebugLevel)

	// Initialize engine with logger
	engine := config.Initialize(
		config.WithLogger(customLogger),
		config.WithDebug(true),
		config.WithActivity("HelloWorld", &CustomHelloWorldActivity{}),
	)

	// Create context with fields
	ctx := logger.WithFields(context.Background(),
		logger.String("requestID", "123"),
		logger.String("userID", "456"),
	)

	p := parser.NewParser(engine.GetRegistry())
	workflow, err := p.ParseFromFile("workflows/custom_activity_workflow.json")
	if err != nil {
		customLogger.ErrorContextf(ctx, "Error parsing workflow: %v", err)
		return
	}

	// Example input data
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"message": "test input",
		},
	}

	result, err := engine.Execute(ctx, workflow, input, nil)
	if err != nil {
		ctx = logger.WithFields(ctx, logger.Error(err))
		customLogger.ErrorContext(ctx, "Error executing workflow")
		return
	}

	ctx = logger.WithFields(ctx,
		logger.Any("result", utils.AnyToJSONString(result.Data)),
		logger.Duration("duration", result.Duration),
	)
	customLogger.InfoContext(ctx, "Workflow completed successfully")

	// If debug is enabled, you can also access detailed execution information
	if result.Debug != nil {
		customLogger.DebugContextf(ctx, "Execution details: %+v", utils.AnyToJSONString(result.Debug))
	}
}
