package activities

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
)

// ValidActivity has correct signature
type ValidActivity struct {
	BaseActivity
}

func (a *ValidActivity) ValidMethod(ctx context.Context, args map[string]any) (interface{}, error) {
	return "valid", nil
}

// InvalidActivity has incorrect signatures
type InvalidActivity struct {
	BaseActivity
}

// Wrong parameter types
func (a *InvalidActivity) WrongParams(ctx context.Context, args string) (interface{}, error) {
	return nil, nil
}

// Wrong return types
func (a *InvalidActivity) WrongReturns(ctx context.Context, args map[string]any) (string, error) {
	return "", nil
}

// Wrong number of parameters
func (a *InvalidActivity) TooManyParams(ctx context.Context, args map[string]any, extra string) (interface{}, error) {
	return nil, nil
}

// Wrong number of returns
func (a *InvalidActivity) TooFewReturns(ctx context.Context, args map[string]any) interface{} {
	return nil
}

// Missing context
func (a *InvalidActivity) NoContext(args map[string]any) (interface{}, error) {
	return nil, nil
}

func TestValidateActivitySignature(t *testing.T) {
	tests := []struct {
		name        string
		methodName  string
		structType  reflect.Type
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Valid signature",
			methodName:  "ValidMethod",
			structType:  reflect.TypeOf(&ValidActivity{}),
			shouldError: false,
		},
		{
			name:        "Wrong parameter type",
			methodName:  "WrongParams",
			structType:  reflect.TypeOf(&InvalidActivity{}),
			shouldError: true,
			errorMsg:    "second parameter must be map[string]any",
		},
		{
			name:        "Wrong return type",
			methodName:  "WrongReturns",
			structType:  reflect.TypeOf(&InvalidActivity{}),
			shouldError: true,
			errorMsg:    "first return value must be interface{}",
		},
		{
			name:        "Too many parameters",
			methodName:  "TooManyParams",
			structType:  reflect.TypeOf(&InvalidActivity{}),
			shouldError: true,
			errorMsg:    "activity function must have exactly 2 input parameters",
		},
		{
			name:        "Too few returns",
			methodName:  "TooFewReturns",
			structType:  reflect.TypeOf(&InvalidActivity{}),
			shouldError: true,
			errorMsg:    "activity function must return exactly 2 values",
		},
		{
			name:        "Missing context",
			methodName:  "NoContext",
			structType:  reflect.TypeOf(&InvalidActivity{}),
			shouldError: true,
			errorMsg:    "activity function must have exactly 2 input parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, found := tt.structType.MethodByName(tt.methodName)
			if !found {
				t.Fatalf("Method %s not found in struct", tt.methodName)
			}

			err := validateActivitySignature(method.Type)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.name)
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for %s, but got: %v", tt.name, err)
				}
			}
		})
	}
}

func TestRegisterActivityStructValidation(t *testing.T) {
	logger := zap.NewLogger(logger.DebugLevel)
	registry := NewRegistry(logger)

	// Test valid activities
	validActivity := &ValidActivity{}
	err := registry.RegisterActivityStruct(validActivity)
	if err != nil {
		t.Errorf("Expected no error registering valid activity, got: %v", err)
	}

	// Verify the valid method was registered
	_, found := registry.Get("ValidMethod")
	if !found {
		t.Error("Expected ValidMethod to be registered")
	}

	// Test invalid activities - should return error since no valid methods found
	invalidActivity := &InvalidActivity{}
	err = registry.RegisterActivityStruct(invalidActivity)
	if err == nil {
		t.Error("Expected error when registering struct with no valid methods")
	}

	// Verify invalid methods were not registered
	invalidMethods := []string{"WrongParams", "WrongReturns", "TooManyParams", "TooFewReturns", "NoContext"}
	for _, methodName := range invalidMethods {
		_, found := registry.Get(methodName)
		if found {
			t.Errorf("Expected invalid method %s to not be registered", methodName)
		}
	}
}
