package jq

import (
	"context"

	"github.com/itchyny/gojq"
	"github.com/kshitiz1403/jsonjuggler/activities"
	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/utils"
)

type TransformArgs struct {
	Query string `arg:"query" required:"true"`
	Data  any    `arg:"data" required:"true"` // JQ expression to select input data
}

type TransformActivity struct {
	*activities.BaseActivity
}

func New(activityName string, logger logger.Logger) *TransformActivity {
	return &TransformActivity{
		BaseActivity: &activities.BaseActivity{
			ActivityName: activityName,
			Logger:       logger,
		},
	}
}

func (a *TransformActivity) Execute(ctx context.Context, arguments map[string]any) (interface{}, error) {
	var args TransformArgs
	if err := utils.ParseAndValidateArgs(ctx, arguments, &args); err != nil {
		a.GetLogger().ErrorContextf(ctx, "Invalid JQ transform arguments: %v", err)
		return nil, activities.NewActivityError(
			activities.ErrInvalidArguments,
			"Invalid JQ transform arguments",
			"JQ",
		).WithArguments(arguments).WithCause(err)
	}

	a.GetLogger().DebugContextf(ctx, "Executing JQ query: %s", args.Query)

	// Parse and run the query
	q, err := gojq.Parse(args.Query)
	if err != nil {
		return nil, activities.NewActivityError(
			activities.ErrJQParseError,
			"Failed to parse JQ query",
			"JQ",
		).WithArguments(map[string]interface{}{
			"query": args.Query,
		}).WithCause(err)
	}

	iter := q.RunWithContext(ctx, args.Data)
	result, ok := iter.Next()
	if !ok {
		return nil, activities.NewActivityError(
			activities.ErrJQExecuteError,
			"JQ query returned no results",
			"JQ",
		).WithArguments(map[string]interface{}{
			"query": args.Query,
			"data":  args.Data,
		})
	}
	if err, ok := result.(error); ok {
		return nil, activities.NewActivityError(
			activities.ErrJQExecuteError,
			"JQ query execution failed",
			"JQ",
		).WithArguments(map[string]interface{}{
			"query": args.Query,
			"data":  args.Data,
		}).WithCause(err)
	}

	return result, nil
}
