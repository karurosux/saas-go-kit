package core

import (
	"fmt"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func NewAppError(code, message string, cause ...error) *AppError {
	err := &AppError{
		Code:    code,
		Message: message,
	}
	if len(cause) > 0 {
		err.Cause = cause[0]
	}
	return err
}

const (
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeInternalServer = "INTERNAL_SERVER_ERROR"
	ErrCodeBadRequest     = "BAD_REQUEST"
)

func NewValidationError(message string) *AppError {
	return NewAppError(ErrCodeValidation, message)
}

func NewNotFoundError(resource string) *AppError {
	return NewAppError(ErrCodeNotFound, fmt.Sprintf("%s not found", resource))
}

