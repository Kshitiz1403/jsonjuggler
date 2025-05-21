package html

import (
	"context"
	"html"

	"github.com/kshitiz1403/jsonjuggler/activities"
	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/utils"
)

// UnescapeArgs represents the arguments for the HTML unescape activity
type UnescapeArgs struct {
	Text string `arg:"text" required:"true"` // HTML-escaped text to unescape
}

// UnescapeActivity unescapes HTML-escaped text
type UnescapeActivity struct {
	activities.BaseActivity
}

// New creates a new HTML unescape activity
func New(activityName string, logger logger.Logger) *UnescapeActivity {
	return &UnescapeActivity{
		BaseActivity: activities.BaseActivity{
			ActivityName: activityName,
			Logger:       logger,
		},
	}
}

func (a *UnescapeActivity) Execute(ctx context.Context, arguments map[string]any) (interface{}, error) {
	var args UnescapeArgs
	if err := utils.ParseAndValidateArgs(ctx, arguments, &args); err != nil {
		a.GetLogger().ErrorContextf(ctx, "Invalid HTML unescape arguments: %v", err)
		return nil, activities.NewActivityError(
			activities.ErrInvalidArguments,
			"Invalid HTML unescape arguments",
			a.GetActivityName(),
		).WithArguments(arguments).WithCause(err)
	}

	a.GetLogger().DebugContext(ctx, "Unescaping HTML text")

	// Unescape the HTML text
	unescaped := html.UnescapeString(args.Text)

	a.GetLogger().DebugContextf(ctx, "Successfully unescaped HTML text: %s", unescaped)
	return unescaped, nil
}
