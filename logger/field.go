package logger

import (
	"time"
)

// FieldType represents the type of a log field
type FieldType int

const (
	StringType FieldType = iota
	IntType
	Int64Type
	Float64Type
	BoolType
	TimeType
	DurationType
	ErrorType
	AnyType
)

// Field represents a structured log field
type Field struct {
	Key       string
	Type      FieldType
	String    string
	Int64     int64
	Float64   float64
	Interface interface{}
}

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Type: StringType, String: value}
}

// Int creates an integer field
func Int(key string, value int) Field {
	return Field{Key: key, Type: IntType, Int64: int64(value)}
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{Key: key, Type: Int64Type, Int64: value}
}

// Float64 creates a float64 field
func Float64(key string, value float64) Field {
	return Field{Key: key, Type: Float64Type, Float64: value}
}

// Bool creates a bool field
func Bool(key string, value bool) Field {
	return Field{Key: key, Type: BoolType, Int64: btoi(value)}
}

// Error creates an error field
func Error(err error) Field {
	return Field{Key: "error", Type: ErrorType, Interface: err}
}

// Time creates a time field
func Time(key string, value time.Time) Field {
	return Field{Key: key, Type: TimeType, Interface: value}
}

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Type: DurationType, Int64: int64(value)}
}

// Any creates a field with any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Type: AnyType, Interface: value}
}

// btoi converts a bool to an int64
func btoi(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
