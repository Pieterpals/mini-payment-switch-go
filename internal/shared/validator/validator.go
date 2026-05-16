package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps go-playground/validator for use with Echo's Validate() method.
type CustomValidator struct {
	validator *validator.Validate
}

// New creates a new CustomValidator instance.
func New() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

// Validate validates a struct based on its `validate` tags.
// Returns a human-readable error message listing all validation failures.
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		var messages []string
		for _, e := range err.(validator.ValidationErrors) {
			messages = append(messages, formatValidationError(e))
		}
		return fmt.Errorf("%s", strings.Join(messages, "; "))
	}
	return nil
}

// formatValidationError converts a validator.FieldError into a readable message.
func formatValidationError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("field '%s' is required", e.Field())
	case "gt":
		return fmt.Sprintf("field '%s' must be greater than %s", e.Field(), e.Param())
	case "min":
		return fmt.Sprintf("field '%s' must be at least %s characters", e.Field(), e.Param())
	case "max":
		return fmt.Sprintf("field '%s' must be at most %s characters", e.Field(), e.Param())
	default:
		return fmt.Sprintf("field '%s' failed on '%s' validation", e.Field(), e.Tag())
	}
}
