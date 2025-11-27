package activities

import (
	"context"
	"fmt"
	"reflect"
)

// validateActivitySignature ensures the method has the correct signature:
// func(context.Context, map[string]any) (interface{}, error)
func validateActivitySignature(fnType reflect.Type) error {
	// Check if it's a function
	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("activity must be a function, got %s", fnType.Kind())
	}

	// Check number of input parameters (should be 3: receiver, context.Context, map[string]any)
	if fnType.NumIn() != 3 {
		return fmt.Errorf("activity function must have exactly 2 input parameters, got %d", fnType.NumIn()-1)
	}

	// Check second parameter is context.Context (first is receiver)
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if fnType.In(1) != contextType {
		return fmt.Errorf("first parameter must be context.Context, got %s", fnType.In(1))
	}

	// Check third parameter is map[string]any (second is receiver)
	mapType := reflect.TypeOf(map[string]any{})
	if fnType.In(2) != mapType {
		return fmt.Errorf("second parameter must be map[string]any, got %s", fnType.In(2))
	}

	// Check number of return values (should be 2: interface{}, error)
	if fnType.NumOut() != 2 {
		return fmt.Errorf("activity function must return exactly 2 values, got %d", fnType.NumOut())
	}

	// Check first return value is interface{}
	interfaceType := reflect.TypeOf((*interface{})(nil)).Elem()
	if fnType.Out(0) != interfaceType {
		return fmt.Errorf("first return value must be interface{}, got %s", fnType.Out(0))
	}

	// Check second return value is error
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if fnType.Out(1) != errorType {
		return fmt.Errorf("second return value must be error, got %s", fnType.Out(1))
	}

	return nil
}
