package config

import (
	"github.com/kshitiz1403/jsonjuggler/activities"
	"github.com/kshitiz1403/jsonjuggler/activities/html"
	"github.com/kshitiz1403/jsonjuggler/activities/http"
	"github.com/kshitiz1403/jsonjuggler/activities/jq"
	"github.com/kshitiz1403/jsonjuggler/activities/jwe"
	"github.com/kshitiz1403/jsonjuggler/engine"
	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
)

// Config holds the configuration for JSONJuggler
type Config struct {
	// DebugEnabled is a flag to enable debug mode
	DebugEnabled bool
	// CustomActivities is a map of activity name to activity implementation
	CustomActivities map[string]activities.Activity
	// Logger is the logger implementation to use
	Logger logger.Logger
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

// Initialize creates a new JSONJuggler engine with the given configuration options
func Initialize(opts ...Option) *engine.Engine {
	config := &Config{
		CustomActivities: make(map[string]activities.Activity),
		Logger:           zap.NewLogger(logger.InfoLevel), // Default logger
	}

	// Apply all options
	for _, opt := range opts {
		opt(config)
	}

	// Create registry and register activities
	registry := activities.NewRegistry(config.Logger)

	// Register default activities
	registerBuiltInActivities(registry)

	// Register custom activities
	for name, activity := range config.CustomActivities {
		registry.Register(name, activity)
	}

	return engine.NewEngine(registry, config.DebugEnabled, config.Logger)
}

func registerBuiltInActivities(registry *activities.Registry) {
	// Here we'll register all the built-in activities
	// For example:
	registry.Register("JQ", jq.New("JQ", registry.GetLogger()))
	registry.Register("JWEEncrypt", jwe.New("JWEEncrypt", registry.GetLogger()))
	registry.Register("HTTPRequest", http.New("HTTPRequest", registry.GetLogger()))
	registry.Register("HTMLUnescape", html.New("HTMLUnescape", registry.GetLogger()))
	// registry.Register("JQTransform", jq.NewTransformActivity())
	// registry.Register("JWEEncrypt", jwe.NewEncryptActivity())
	// etc.
}
