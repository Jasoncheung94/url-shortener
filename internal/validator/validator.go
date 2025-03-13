package validator

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	// Validate is a Singleton of validator.
	Validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())
	// ValidateOnce ensures validator is created once.
	ValidateOnce sync.Once
)

// SetupValidator returns the singleton validator instance.
func SetupValidator() {
	ValidateOnce.Do(func() {
		// Default logger setup
		Validate = validator.New(validator.WithRequiredStructEnabled())
	})
}
