package activities

import (
	"context"
	"sync"

	"github.com/kshitiz1403/jsonjuggler/logger"
)

// Activity represents a workflow activity
type Activity interface {
	IBaseActivity // Always inherit IBaseActivity, this auto injects all required dependencies into the activity
	// Execute runs the activity with the given arguments
	Execute(ctx context.Context, arguments map[string]any) (interface{}, error)
}

type activityRegistry struct {
	fn   Activity
	name string
}

// Registry maintains a mapping of activity names to their implementations
type Registry struct {
	sync.Mutex
	activities map[string]*activityRegistry
	logger     logger.Logger
}

// NewRegistry creates a new activity registry
func NewRegistry(logger logger.Logger) *Registry {
	return &Registry{
		activities: make(map[string]*activityRegistry),
		logger:     logger,
	}
}

// Get retrieves an activity from the registry
func (r *Registry) Get(name string) (Activity, bool) {
	activityRegistry, ok := r.activities[name]
	if !ok {
		return nil, false
	}
	return activityRegistry.fn, true
}

// GetLogger returns the logger for the registry
func (r *Registry) GetLogger() logger.Logger {
	return r.logger
}

func (r *Registry) GetAllActivities() []string {
	activityNames := make([]string, 0, len(r.activities))
	for name := range r.activities {
		activityNames = append(activityNames, name)
	}
	return activityNames
}
