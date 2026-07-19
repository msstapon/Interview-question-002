package http

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct{ v *validator.Validate }

func NewValidator() *Validator {
	v := validator.New()
	// Report errors using the json field name (e.g. "confirm_password") rather than
	// the Go struct field, so {field, rule} details match the wire contract.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return &Validator{v: v}
}

func (v *Validator) Validate(i any) error { return v.v.Struct(i) }
