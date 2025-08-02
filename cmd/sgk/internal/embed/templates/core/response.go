package core

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type SuccessResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

func Success(c echo.Context, data any, message ...string) error {
	response := SuccessResponse{
		Success: true,
		Data:    data,
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	return c.JSON(http.StatusOK, response)
}

func Created(c echo.Context, data any, message ...string) error {
	response := SuccessResponse{
		Success: true,
		Data:    data,
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	return c.JSON(http.StatusCreated, response)
}

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

func BadRequest(c echo.Context, err error, code ...string) error {
	return Error(c, http.StatusBadRequest, err, code...)
}

func Unauthorized(c echo.Context, err error, code ...string) error {
	return Error(c, http.StatusUnauthorized, err, code...)
}

func NotFound(c echo.Context, err error, code ...string) error {
	return Error(c, http.StatusNotFound, err, code...)
}

func InternalServerError(c echo.Context, err error, code ...string) error {
	return Error(c, http.StatusInternalServerError, err, code...)
}

func PaymentRequired(c echo.Context, err error, code ...string) error {
	return Error(c, http.StatusPaymentRequired, err, code...)
}

func Forbidden(c echo.Context, err error, code ...string) error {
	return Error(c, http.StatusForbidden, err, code...)
}

func ServiceUnavailable(c echo.Context, err error, code ...string) error {
	return Error(c, http.StatusServiceUnavailable, err, code...)
}
