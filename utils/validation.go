package utils

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

/*
ParseAndValidateArgs parses a map of arguments into a struct and validates them
It checks:
1. Required fields are present
2. Field types match
3. Validation tags are satisfied

Usage:

	type RequestArgs struct {
		    URL         string            `arg:"url" required:"true"`
		    Method      string            `arg:"method" required:"true" validate:"oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"`
		    Headers     map[string]string `arg:"headers"`
		    Body        interface{}       `arg:"body"`
		    TimeoutSec  int               `arg:"timeoutSec" validate:"min=1,max=300"`
		    FailOnError bool              `arg:"failOnError"`
		}
*/
func ParseAndValidateArgs(ctx context.Context, args map[string]any, dest interface{}) error {
	// Configure mapstructure decoder
	config := &mapstructure.DecoderConfig{
		TagName:          "arg",
		Result:           dest,
		WeaklyTypedInput: false,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc("2006-01-02T15:04:05Z07:00"),
			mapstructure.StringToTimeDurationHookFunc(),
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return fmt.Errorf("failed to create decoder: %w", err)
	}

	// Create a copy of arguments to track which fields are present
	presentFields := make(map[string]bool)
	for k := range args {
		presentFields[k] = true
	}

	// Check required fields before decoding
	val := reflect.ValueOf(dest).Elem()
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		argName := field.Tag.Get("arg")
		if argName == "" {
			continue
		}

		if field.Tag.Get("required") == "true" {
			if _, ok := args[argName]; !ok {
				return fmt.Errorf("required argument '%s' is missing", argName)
			}
		}
	}

	// Decode arguments
	if err := decoder.Decode(args); err != nil {
		return fmt.Errorf("failed to decode arguments: %w", err)
	}

	// Create a custom validator that skips validation for absent optional fields
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("arg")
	})

	// Validate only fields that are present or required
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		argName := field.Tag.Get("arg")
		validateTag := field.Tag.Get("validate")

		if validateTag == "" {
			continue
		}

		// Skip validation for fields that aren't present in input
		// unless they are marked as required
		isRequired := field.Tag.Get("required") == "true"
		if !isRequired && !presentFields[argName] {
			continue
		}

		if err := validate.Var(val.Field(i).Interface(), validateTag); err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				for _, validationErr := range validationErrors {
					return fmt.Errorf(
						"validation failed for field '%s': expected %s=%s, got %v",
						field.Name,
						validationErr.Tag(),
						validationErr.Param(),
						validationErr.Value(),
					)
				}
			}
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	return nil
}

// IsValidJQTemplate checks if a string is a valid JQ template (wrapped in ${ and })
func IsValidJQTemplate(s string) bool {
	return strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}")
}

func ExtractJQTemplate(s string) (string, error) {
	if !IsValidJQTemplate(s) {
		return "", fmt.Errorf("invalid JQ template: %s", s)
	}
	return strings.TrimPrefix(strings.TrimSuffix(s, "}"), "${"), nil
}
