package config

import (
	"github.com/kshitiz1403/jsonjuggler/activities"
	"github.com/kshitiz1403/jsonjuggler/activities/http"
	"github.com/kshitiz1403/jsonjuggler/activities/jq"
	"github.com/kshitiz1403/jsonjuggler/engine"
	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
	"github.com/kshitiz1403/jsonjuggler/telemetry"
)

// Config holds the configuration for JsonJuggler
type Config struct {
	// DebugEnabled is a flag to enable debug mode
	DebugEnabled bool
	// CustomActivities is a map of activity name to activity implementation
	CustomActivities map[string]activities.Activity
	// ActivityStructs is a slice of activity structs to register
	ActivityStructs []interface{}
	// Logger is the logger implementation to use
	Logger logger.Logger
	// Telemetry configuration
	TelemetryConfig *telemetry.Config
}

// Option is a function that modifies Config
type Option func(*Config)

// WithActivity adds a custom activity to the configuration
func WithActivity(name string, activity activities.Activity) Option {
	return func(c *Config) {
		c.CustomActivities[name] = activity
	}
}

// WithDebug adds debug enabled to the configuration
func WithDebug(enabled bool) Option {
	return func(c *Config) {
		c.DebugEnabled = enabled
	}
}

// WithLogger sets the logger implementation
func WithLogger(log logger.Logger) Option {
	return func(c *Config) {
		c.Logger = log
	}
}

// WithActivityStruct adds all activities from a struct to the configuration
func WithActivityStruct(activityStruct interface{}) Option {
	return func(c *Config) {
		c.ActivityStructs = append(c.ActivityStructs, activityStruct)
	}
}

// WithTelemetry adds telemetry configuration
func WithTelemetry(cfg *telemetry.Config) Option {
	return func(c *Config) {
		c.TelemetryConfig = cfg
	}
}

// Initialize creates a new JSONJuggler engine with the given configuration options
func Initialize(opts ...Option) (*engine.Engine, error) {
	config := &Config{
		CustomActivities: make(map[string]activities.Activity),
		Logger:           zap.NewLogger(logger.InfoLevel), // Default logger
	}

	// Apply all options
	for _, opt := range opts {
		opt(config)
	}

	// Initialize telemetry
	var tel *telemetry.Telemetry
	if config.TelemetryConfig != nil {
		var err error
		tel, err = telemetry.New(config.TelemetryConfig)
		if err != nil {
			return nil, err
		}
	}

	// Create registry and register activities
	registry := activities.NewRegistry(config.Logger)

	// Register default activities
	registerBuiltInActivities(registry)

	// Register custom activities
	for name, activity := range config.CustomActivities {
		if err := registry.RegisterActivity(name, activity); err != nil {
			return nil, err
		}
	}

	// Register activity structs
	for _, activityStruct := range config.ActivityStructs {
		if err := registry.RegisterActivityStruct(activityStruct); err != nil {
			return nil, err
		}
	}

	return engine.NewEngine(registry, config.DebugEnabled, config.Logger, tel), nil
}

func registerBuiltInActivities(registry *activities.Registry) {
	// Here we'll register all the built-in activities
	// For example:
	registry.RegisterActivity("JQ", jq.New("JQ", registry.GetLogger()))
	registry.RegisterActivity("HTTPRequest", http.New("HTTPRequest", registry.GetLogger()))
}
