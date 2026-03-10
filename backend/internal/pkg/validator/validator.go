package validator

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Validate validates a struct based on its tags.
func Validate(s interface{}) error {
	return validate.Struct(s)
}

// ValidationErrors extracts human-readable error messages from validation errors.
func ValidationErrors(err error) []string {
	var messages []string
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			messages = append(messages, formatError(e))
		}
	}
	return messages
}

func formatError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "min":
		return e.Field() + " must be at least " + e.Param()
	case "max":
		return e.Field() + " must be at most " + e.Param()
	case "uuid":
		return e.Field() + " must be a valid UUID"
	case "oneof":
		return e.Field() + " must be one of: " + e.Param()
	default:
		return e.Field() + " is invalid"
	}
}
