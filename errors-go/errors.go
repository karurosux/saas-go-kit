package errors

import (
	"fmt"
	"net/http"
)

// AppError represents a structured application error with HTTP status
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Details any    `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("code: %s, message: %s", e.Code, e.Message)
}

// HTTPStatus returns the HTTP status code
func (e *AppError) HTTPStatus() int {
	return e.Status
}

// WithDetails adds details to an error
func (e *AppError) WithDetails(details any) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Status:  e.Status,
		Details: details,
	}
}

// New creates a new AppError
func New(code string, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// NewWithDetails creates a new AppError with details
func NewWithDetails(code string, message string, status int, details any) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Details: details,
	}
}

// Common errors
var (
	// HTTP Status Errors
	ErrBadRequest     = New("BAD_REQUEST", "The request could not be understood or was missing required parameters", http.StatusBadRequest)
	ErrUnauthorized   = New("UNAUTHORIZED", "Authentication is required to access this resource", http.StatusUnauthorized)
	ErrForbidden      = New("FORBIDDEN", "You don't have permission to access this resource", http.StatusForbidden)
	ErrNotFound       = New("NOT_FOUND", "The requested resource was not found", http.StatusNotFound)
	ErrConflict       = New("CONFLICT", "The request could not be completed due to a conflict with the current state", http.StatusConflict)
	ErrInternalServer = New("INTERNAL_SERVER_ERROR", "An unexpected error occurred. Please try again later", http.StatusInternalServerError)

	// Validation Errors
	ErrValidation     = New("VALIDATION_ERROR", "The provided data failed validation", http.StatusBadRequest)
	ErrInvalidRequest = New("INVALID_REQUEST", "The request format is invalid", http.StatusBadRequest)
	ErrMissingField   = New("MISSING_FIELD", "Required field is missing", http.StatusBadRequest)

	// Authentication & Authorization Errors
	ErrInvalidCredentials = New("INVALID_CREDENTIALS", "The provided credentials are incorrect", http.StatusUnauthorized)
	ErrTokenExpired       = New("TOKEN_EXPIRED", "Your session has expired. Please sign in again", http.StatusUnauthorized)
	ErrTokenInvalid       = New("TOKEN_INVALID", "The provided authentication token is invalid", http.StatusUnauthorized)
	ErrPermissionDenied   = New("PERMISSION_DENIED", "You don't have permission to perform this action", http.StatusForbidden)

	// Resource Errors
	ErrResourceNotFound    = New("RESOURCE_NOT_FOUND", "The requested resource does not exist", http.StatusNotFound)
	ErrResourceExists      = New("RESOURCE_EXISTS", "A resource with the same identifier already exists", http.StatusConflict)
	ErrResourceLimit       = New("RESOURCE_LIMIT", "You've reached the maximum number of resources allowed", http.StatusForbidden)
	ErrResourceUnavailable = New("RESOURCE_UNAVAILABLE", "The requested resource is temporarily unavailable", http.StatusServiceUnavailable)

	// Rate Limit Errors
	ErrRateLimitExceeded = New("RATE_LIMIT_EXCEEDED", "Too many requests. Please try again later", http.StatusTooManyRequests)
)

// IsAppError checks if an error is an AppError and returns it
func IsAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code string, message string, status int) *AppError {
	if appErr, ok := IsAppError(err); ok {
		// If it's already an AppError, preserve the original details
		return &AppError{
			Code:    code,
			Message: message,
			Status:  status,
			Details: map[string]interface{}{
				"original_code":    appErr.Code,
				"original_message": appErr.Message,
				"original_details": appErr.Details,
			},
		}
	}
	// For other errors, create a new AppError
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Details: err.Error(),
	}
}

// Helper functions for creating common errors

// NotFound creates a resource not found error with a custom message
func NotFound(resource string) *AppError {
	return New("RESOURCE_NOT_FOUND", fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

// Forbidden creates a permission denied error with a custom message
func Forbidden(action string) *AppError {
	return New("PERMISSION_DENIED", fmt.Sprintf("You don't have permission to %s", action), http.StatusForbidden)
}

// BadRequest creates a bad request error with a custom message
func BadRequest(message string) *AppError {
	return New("BAD_REQUEST", message, http.StatusBadRequest)
}

// Unauthorized creates an unauthorized error with a custom message
func Unauthorized(message string) *AppError {
	return New("UNAUTHORIZED", message, http.StatusUnauthorized)
}

// Conflict creates a conflict error with a custom message
func Conflict(message string) *AppError {
	return New("CONFLICT", message, http.StatusConflict)
}

// Internal creates an internal server error with a generic user message
// The actual error details should be logged server-side
func Internal(details string) *AppError {
	return NewWithDetails("INTERNAL_SERVER_ERROR", "An unexpected error occurred. Please try again later", http.StatusInternalServerError, details)
}

// ValidationError creates a validation error with field details
func ValidationError(fields map[string]string) *AppError {
	return NewWithDetails("VALIDATION_ERROR", "The provided data failed validation", http.StatusBadRequest, fields)
}