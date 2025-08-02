package core

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{validator: v}
}

func (v *Validator) Validate(i any) error {
	if err := v.validator.Struct(i); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return &ValidationErrors{Errors: validationErrors}
		}
		return err
	}
	return nil
}

type ValidationErrors struct {
	Errors validator.ValidationErrors
}

func (ve *ValidationErrors) Error() string {
	return ve.Errors.Error()
}

func (ve *ValidationErrors) ToMap() map[string][]string {
	errors := make(map[string][]string)

	for _, err := range ve.Errors {
		field := err.Field()
		message := getValidationMessage(err)

		if _, exists := errors[field]; !exists {
			errors[field] = []string{}
		}
		errors[field] = append(errors[field], message)
	}

	return errors
}

func getValidationMessage(err validator.FieldError) string {
	field := err.Field()

	switch err.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " must be at least " + err.Param() + " characters long"
	case "max":
		return field + " must be at most " + err.Param() + " characters long"
	default:
		return field + " is invalid"
	}
}
