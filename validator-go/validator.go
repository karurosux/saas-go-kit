package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/saas-go-kit/errors-go"
)

// Validator wraps the go-playground validator
type Validator struct {
	validator *validator.Validate
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
}

// New creates a new validator instance
func New() *Validator {
	v := validator.New()
	
	// Use JSON tags as field names
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	
	// Register custom validators
	v.RegisterValidation("notblank", notBlank)
	
	return &Validator{
		validator: v,
	}
}

// Validate validates a struct
func (v *Validator) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		return v.formatValidationError(err)
	}
	return nil
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validator.Var(field, tag)
}

// RegisterValidation registers a custom validation function
func (v *Validator) RegisterValidation(tag string, fn validator.Func) error {
	return v.validator.RegisterValidation(tag, fn)
}

// formatValidationError formats validation errors into a structured format
func (v *Validator) formatValidationError(err error) error {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return errors.Internal("validation error")
	}
	
	fieldErrors := make(map[string]string)
	errorDetails := make([]ValidationError, 0, len(validationErrors))
	
	for _, e := range validationErrors {
		fieldName := e.Field()
		message := v.getErrorMessage(e)
		
		fieldErrors[fieldName] = message
		errorDetails = append(errorDetails, ValidationError{
			Field:   fieldName,
			Tag:     e.Tag(),
			Value:   fmt.Sprintf("%v", e.Value()),
			Message: message,
		})
	}
	
	return errors.ValidationError(fieldErrors)
}

// getErrorMessage returns a human-readable error message for a validation error
func (v *Validator) getErrorMessage(fe validator.FieldError) string {
	field := fe.Field()
	
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		if fe.Type().Kind() == reflect.String {
			return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
		}
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		if fe.Type().Kind() == reflect.String {
			return fmt.Sprintf("%s must not exceed %s characters", field, fe.Param())
		}
		return fmt.Sprintf("%s must not exceed %s", field, fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, fe.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "notblank":
		return fmt.Sprintf("%s cannot be blank", field)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, fe.Param())
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s failed %s validation", field, fe.Tag())
	}
}

// Custom validators

// notBlank validates that a string is not empty after trimming whitespace
func notBlank(fl validator.FieldLevel) bool {
	field := fl.Field()
	switch field.Kind() {
	case reflect.String:
		return len(strings.TrimSpace(field.String())) > 0
	default:
		return true
	}
}

// Common validation rules as constants
const (
	EmailRule    = "required,email"
	PasswordRule = "required,min=8,max=128"
	UUIDRule     = "required,uuid"
	URLRule      = "required,url"
	PhoneRule    = "required,e164"
)

// Default validator instance
var DefaultValidator = New()

// Package-level functions using the default validator

// Validate validates a struct using the default validator
func Validate(i interface{}) error {
	return DefaultValidator.Validate(i)
}

// ValidateVar validates a single variable using the default validator
func ValidateVar(field interface{}, tag string) error {
	return DefaultValidator.ValidateVar(field, tag)
}