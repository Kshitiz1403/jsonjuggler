package activities

import (
	"context"
	"reflect"

	"github.com/kshitiz1403/jsonjuggler/logger"
)

// methodActivity wraps a struct method as an Activity
type methodActivity struct {
	baseActivity BaseActivity
	method       reflect.Value
	methodName   string
}

func (a *methodActivity) Execute(ctx context.Context, arguments map[string]any) (interface{}, error) {
	// Inject activity info into context
	activityInfo := &ActivityInfo{
		ActivityName: a.GetActivityName(),
		Logger:       a.GetLogger(),
	}
	ctxWithInfo := withActivityInfo(ctx, activityInfo)

	return a.method.Interface().(func(context.Context, map[string]any) (interface{}, error))(ctxWithInfo, arguments)
}

func (a *methodActivity) GetLogger() logger.Logger {
	return a.baseActivity.GetLogger()
}

func (a *methodActivity) GetActivityName() string {
	return a.methodName
}

func (a *methodActivity) setLogger(logger logger.Logger) {
	a.baseActivity.setLogger(logger)
}

func (a *methodActivity) setActivityName(name string) {
	a.baseActivity.setActivityName(name)
}
