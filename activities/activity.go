package activities

import (
	"context"
	"reflect"
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

// Register adds an activity to the registry
func (r *Registry) Register(name string, activity Activity) {
	r.Lock()
	defer r.Unlock()
	activity.setLogger(r.logger)
	if r.isAlreadyRegistered(name) {
		r.logger.Warnf("Activity %s is already registered", name)
		return
	}
	activity.setActivityName(name)
	r.activities[name] = &activityRegistry{
		fn:   activity,
		name: name,
	}
}

func (r *Registry) isAlreadyRegistered(name string) bool {
	_, ok := r.activities[name]
	return ok
}

// // Borrowed from - https://github.com/temporalio/sdk-go/blob/797e9aa584017cd0f8e4c20cfff20f09ef2292fb/internal/internal_worker.go#L679
// func (r *Registry) RegisterActivityStruct(aStruct interface{}) error {
// 	r.Lock()
// 	defer r.Unlock()
// 	structValue := reflect.ValueOf(aStruct)
// 	structType := structValue.Type()
// 	count := 0
// 	for i := 0; i < structValue.NumMethod(); i++ {
// 		methodValue := structValue.Method(i)
// 		method := structType.Method(i)
// 		// skip private method
// 		if method.PkgPath != "" {
// 			continue
// 		}
// 		name := method.Name
// 		if err := validateActivitySignature(method.Type); err != nil {
// 			return err
// 		}
// 		r.Register(name, methodValue.Interface().(Activity))
// 		count++
// 	}
// 	if count == 0 {
// 		return fmt.Errorf("no activities (public methods) found in struct %s", structType.Name())
// 	}
// 	return nil
// }

// TODO
func validateActivitySignature(fnType reflect.Type) error {
	return nil
}

// Get retrieves an activity from the registry
func (r *Registry) Get(name string) (Activity, bool) {
	activityRegistry, ok := r.activities[name]
	return activityRegistry.fn, ok
}

// GetLogger returns the logger for the registry
func (r *Registry) GetLogger() logger.Logger {
	return r.logger
}
