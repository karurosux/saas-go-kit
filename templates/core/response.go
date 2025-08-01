package core

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// Success sends a successful response
func Success(c echo.Context, data interface{}, message ...string) error {
	response := SuccessResponse{
		Success: true,
		Data:    data,
	}
	
	if len(message) > 0 {
		response.Message = message[0]
	}
	
	return c.JSON(http.StatusOK, response)
}

// Created sends a created response
func Created(c echo.Context, data interface{}, message ...string) error {
	response := SuccessResponse{
		Success: true,
		Data:    data,
	}
	
	if len(message) > 0 {
		response.Message = message[0]
	}
	
	return c.JSON(http.StatusCreated, response)
}

// Error sends an error response
func Error(c echo.Context, statusCode int, err error, code ...string) error {
	response := ErrorResponse{
		Success: false,
		Error:   err.Error(),
	}
	
	if len(code) > 0 {
		response.Code = code[0]
	}
	
	return c.JSON(statusCode, response)
}

// BadRequest sends a bad request error
func BadRequest(c echo.Context, err error, code ...string) error {
	return Error(c, http.StatusBadRequest, err, code...)
}

// Unauthorized sends an unauthorized error
func Unauthorized(c echo.Context, err error, code ...string) error {
	return Error(c, http.StatusUnauthorized, err, code...)
}

// NotFound sends a not found error
func NotFound(c echo.Context, err error, code ...string) error {
	return Error(c, http.StatusNotFound, err, code...)
}

// InternalServerError sends an internal server error
func InternalServerError(c echo.Context, err error, code ...string) error {
	return Error(c, http.StatusInternalServerError, err, code...)
}