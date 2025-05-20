package activities

import (
	"github.com/kshitiz1403/jsonjuggler/logger"
)

// IBaseActivity is an interface implemented by base activity
type IBaseActivity interface {
	GetLogger() logger.Logger
	GetActivityName() string

	setLogger(logger logger.Logger)
	setActivityName(name string)
}

// BaseActivity provides common functionality for activities
type BaseActivity struct {
	ActivityName string
	Logger       logger.Logger
}

// SetLogger sets the logger for the activity
func (b *BaseActivity) setLogger(logger logger.Logger) {
	b.Logger = logger
}

// GetLogger returns the logger for the activity
func (b *BaseActivity) GetLogger() logger.Logger {
	return b.Logger
}

// setActivityName sets the name of the activity
func (b *BaseActivity) setActivityName(name string) {
	b.ActivityName = name
}

// GetActivityName returns the name of the activity
func (b *BaseActivity) GetActivityName() string {
	return b.ActivityName
}
