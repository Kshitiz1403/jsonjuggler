package activities

import (
	"context"
	"fmt"

	"github.com/kshitiz1403/jsonjuggler/logger"
)

// activityWrapper wraps a regular Activity to inject ActivityInfo into context
type activityWrapper struct {
	activity Activity
}

func (aw *activityWrapper) Execute(ctx context.Context, arguments map[string]any) (interface{}, error) {
	// Inject activity info into context
	activityInfo := &ActivityInfo{
		ActivityName: aw.activity.GetActivityName(),
		Logger:       aw.activity.GetLogger(),
	}
	ctxWithInfo := withActivityInfo(ctx, activityInfo)

	return aw.activity.Execute(ctxWithInfo, arguments)
}

func (aw *activityWrapper) GetLogger() logger.Logger {
	return aw.activity.GetLogger()
}

func (aw *activityWrapper) GetActivityName() string {
	return aw.activity.GetActivityName()
}

func (aw *activityWrapper) setLogger(logger logger.Logger) {
	aw.activity.setLogger(logger)
}

func (aw *activityWrapper) setActivityName(name string) {
	aw.activity.setActivityName(name)
}

// RegisterActivity adds an activity to the registry
func (r *Registry) RegisterActivity(name string, activity Activity) error {
	// r.Lock()
	// defer r.Unlock()
	activity.setLogger(r.logger)
	if r.isAlreadyRegistered(name) {
		r.logger.Warnf("Activity %s is already registered", name)
		return fmt.Errorf("activity %s is already registered", name)
	}
	activity.setActivityName(name)

	// Wrap the activity to inject ActivityInfo into context
	wrapper := &activityWrapper{activity: activity}

	r.activities[name] = &activityRegistry{
		fn:   wrapper,
		name: name,
	}
	r.logger.Debugf("Successfully registered activity: %s", name)
	return nil
}

func (r *Registry) isAlreadyRegistered(name string) bool {
	_, ok := r.activities[name]
	return ok
}
