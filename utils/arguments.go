package utils

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/itchyny/gojq"
	"github.com/mitchellh/mapstructure"
)

// EvaluateArgument evaluates an argument which can be a static value or a JQ expression
func EvaluateArgument(arg interface{}, data interface{}) (interface{}, error) {
	// If not a string, decode using mapstructure
	strArg, ok := arg.(string)
	if !ok {
		var result interface{}
		if err := decodeToJSON(arg, &result); err != nil {
			return nil, fmt.Errorf("failed to decode argument: %w", err)
		}
		return result, nil
	}

	// Check if it's a JQ expression
	if IsValidJQTemplate(strArg) {
		// Extract JQ query
		query, err := ExtractJQTemplate(strArg)
		if err != nil {
			return nil, fmt.Errorf("invalid JQ template: %w", err)
		}

		// Parse and run JQ query
		q, err := gojq.Parse(query)
		if err != nil {
			return nil, fmt.Errorf("invalid JQ query '%s': %w", query, err)
		}

		iter := q.Run(data)
		result, ok := iter.Next()
		if !ok {
			return nil, fmt.Errorf("no result for JQ query '%s'", query)
		}
		if err, ok := result.(error); ok {
			return nil, fmt.Errorf("JQ query '%s' failed: %w", query, err)
		}

		// Decode the result using mapstructure
		var output interface{}
		if err := decodeToJSON(result, &output); err != nil {
			return nil, fmt.Errorf("failed to decode JQ result: %w", err)
		}
		return output, nil
	}

	return strArg, nil
}

// decodeToJSON decodes input to output using mapstructure with JSON tags
func decodeToJSON(input interface{}, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		Result:  output,
		TagName: "json",
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
		),
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

// EvaluateStringArgument evaluates an argument and ensures it's a string
func EvaluateStringArgument(arg interface{}, data interface{}) (string, error) {
	result, err := EvaluateArgument(arg, data)
	if err != nil {
		return "", err
	}

	str, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("argument must be a string, got %T", result)
	}

	return str, nil
}

func AnyToJSONString(arg interface{}) string {
	jsonString, err := json.Marshal(arg)
	if err != nil {
		return fmt.Sprintf("failed to marshal argument: %v", err)
	}
	return string(jsonString)
}

func AnyToJSONStringPretty(arg interface{}) string {
	jsonString, err := json.MarshalIndent(arg, "", "  ")
	if err != nil {
		return fmt.Sprintf("failed to marshal argument: %v", err)
	}
	return string(jsonString)
}

// EvaluateArgumentsRecursively evaluates all arguments in a map recursively
func EvaluateArgumentsRecursively(value interface{}, data map[string]interface{}) (interface{}, error) {
	// Handle string templates
	if strVal, ok := value.(string); ok && IsValidJQTemplate(strVal) {
		return EvaluateArgument(strVal, data)
	}

	// Handle arrays
	if arr, ok := value.([]interface{}); ok {
		result := make([]interface{}, len(arr))
		for i, item := range arr {
			evaluated, err := EvaluateArgumentsRecursively(item, data)
			if err != nil {
				return nil, err
			}
			result[i] = evaluated
		}
		return result, nil
	}

	// Handle nested maps
	if m, ok := value.(map[string]interface{}); ok {
		result := make(map[string]interface{})
		for k, v := range m {
			evaluated, err := EvaluateArgumentsRecursively(v, data)
			if err != nil {
				return nil, err
			}
			result[k] = evaluated
		}
		return result, nil
	}

	// Return primitive values as is
	return value, nil
}

// EvaluateArgumentMap evaluates all arguments in a map recursively
func EvaluateArgumentMap(args map[string]interface{}, data map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range args {
		evaluated, err := EvaluateArgumentsRecursively(value, data)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate argument '%s': %w", key, err)
		}
		result[key] = evaluated
	}

	return result, nil
}
