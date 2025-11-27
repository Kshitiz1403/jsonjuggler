package activities

import (
	"fmt"
	"reflect"
)

// RegisterActivityStruct registers all public methods of a struct as activities
// Borrowed from - https://github.com/temporalio/sdk-go/blob/797e9aa584017cd0f8e4c20cfff20f09ef2292fb/internal/internal_worker.go#L679
func (r *Registry) RegisterActivityStruct(aStruct interface{}) error {
	r.Lock()
	defer r.Unlock()

	if aStruct == nil {
		return fmt.Errorf("cannot register nil struct")
	}

	structValue := reflect.ValueOf(aStruct)
	if !structValue.IsValid() {
		return fmt.Errorf("invalid struct value")
	}

	structType := structValue.Type()

	// Auto-initialize BaseActivity if it exists in the struct
	r.initializeBaseActivity(structValue)

	count := 0

	for i := 0; i < structValue.NumMethod(); i++ {
		methodValue := structValue.Method(i)
		method := structType.Method(i)
		// skip private methods
		if method.PkgPath != "" {
			continue
		}

		// Validate the method signature
		if err := validateActivitySignature(method.Type); err != nil {
			r.logger.Warnf("Skipping method %s: %v", method.Name, err)
			continue
		}

		r.logger.Debugf("Registering activity method: %s", method.Name)

		// Create a wrapper activity that calls the method
		activity := &methodActivity{
			baseActivity: BaseActivity{
				Logger: r.logger,
			},
			method:     methodValue,
			methodName: method.Name,
		}

		if err := r.RegisterActivity(method.Name, activity); err != nil {
			r.logger.Errorf("Failed to register activity %s: %v", method.Name, err)
			continue
		}
		count++
	}

	if count == 0 {
		return fmt.Errorf("no activities (public methods) found in struct %s", structType.Name())
	}

	return nil
}

// initializeBaseActivity automatically initializes BaseActivity fields in the struct
func (r *Registry) initializeBaseActivity(structValue reflect.Value) {
	// Ensure we're working with a pointer to a struct
	if structValue.Kind() != reflect.Ptr {
		return
	}

	structElem := structValue.Elem()
	if structElem.Kind() != reflect.Struct {
		return
	}

	// Look for BaseActivity field
	for i := 0; i < structElem.NumField(); i++ {
		field := structElem.Field(i)
		fieldType := structElem.Type().Field(i)

		// Check if this field is a BaseActivity
		if fieldType.Type == reflect.TypeOf(BaseActivity{}) && field.CanSet() {
			// Initialize the BaseActivity with the registry's logger
			field.Set(reflect.ValueOf(BaseActivity{
				Logger: r.logger,
			}))
			r.logger.Debugf("Auto-initialized BaseActivity for struct %s", structElem.Type().Name())
		}
	}
}
