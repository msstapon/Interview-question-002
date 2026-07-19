package dto

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

// FieldError is one failed validation rule, reported using the json field name.
type FieldError struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
}

// ValidationErrors converts a go-playground validator error into a serialisable
// slice for the {error.details} envelope field.
func ValidationErrors(err error) []FieldError {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil
	}
	out := make([]FieldError, 0, len(ve))
	for _, fe := range ve {
		out = append(out, FieldError{Field: fe.Field(), Rule: fe.Tag()})
	}
	return out
}
