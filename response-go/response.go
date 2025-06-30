package response

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/karurosux/saas-go-kit/errors-go"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorData  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorData contains error details
type ErrorData struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Meta contains metadata about the response
type Meta struct {
	Timestamp  time.Time   `json:"timestamp"`
	Version    string      `json:"version,omitempty"`
	RequestID  string      `json:"request_id,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination contains pagination information
type Pagination struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	Total     int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Handler provides response handling methods
type Handler struct {
	version    string
	prettyJSON bool
}

// NewHandler creates a new response handler
func NewHandler(version string, prettyJSON bool) *Handler {
	return &Handler{
		version:    version,
		prettyJSON: prettyJSON,
	}
}

// Default handler instance
var DefaultHandler = NewHandler("1.0", false)

// Helper function to send JSON response with optional pretty formatting
func (h *Handler) sendJSON(c echo.Context, code int, data interface{}) error {
	if h.prettyJSON {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		encoder := json.NewEncoder(c.Response())
		encoder.SetIndent("", "  ")
		c.Response().WriteHeader(code)
		return encoder.Encode(data)
	}
	return c.JSON(code, data)
}

// Success sends a successful response
func (h *Handler) Success(c echo.Context, data interface{}) error {
	response := Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Timestamp: time.Now(),
			Version:   h.version,
			RequestID: c.Response().Header().Get(echo.HeaderXRequestID),
		},
	}
	return h.sendJSON(c, http.StatusOK, response)
}

// SuccessWithPagination sends a successful response with pagination
func (h *Handler) SuccessWithPagination(c echo.Context, data interface{}, pagination *Pagination) error {
	response := Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Timestamp:  time.Now(),
			Version:    h.version,
			RequestID:  c.Response().Header().Get(echo.HeaderXRequestID),
			Pagination: pagination,
		},
	}
	return h.sendJSON(c, http.StatusOK, response)
}

// Created sends a successful creation response
func (h *Handler) Created(c echo.Context, data interface{}) error {
	response := Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Timestamp: time.Now(),
			Version:   h.version,
			RequestID: c.Response().Header().Get(echo.HeaderXRequestID),
		},
	}
	return h.sendJSON(c, http.StatusCreated, response)
}

// NoContent sends a successful no content response
func (h *Handler) NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// Error sends an error response
func (h *Handler) Error(c echo.Context, err error) error {
	appErr, ok := errors.IsAppError(err)
	if !ok {
		appErr = errors.ErrInternalServer
	}

	response := Response{
		Success: false,
		Error: &ErrorData{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
		Meta: &Meta{
			Timestamp: time.Now(),
			Version:   h.version,
			RequestID: c.Response().Header().Get(echo.HeaderXRequestID),
		},
	}
	return h.sendJSON(c, appErr.HTTPStatus(), response)
}

// Package-level functions using the default handler

// Success sends a successful response using the default handler
func Success(c echo.Context, data interface{}) error {
	return DefaultHandler.Success(c, data)
}

// SuccessWithPagination sends a successful response with pagination using the default handler
func SuccessWithPagination(c echo.Context, data interface{}, pagination *Pagination) error {
	return DefaultHandler.SuccessWithPagination(c, data, pagination)
}

// Created sends a successful creation response using the default handler
func Created(c echo.Context, data interface{}) error {
	return DefaultHandler.Created(c, data)
}

// NoContent sends a successful no content response using the default handler
func NoContent(c echo.Context) error {
	return DefaultHandler.NoContent(c)
}

// Error sends an error response using the default handler
func Error(c echo.Context, err error) error {
	return DefaultHandler.Error(c, err)
}

// CalculateTotalPages calculates the total number of pages
func CalculateTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}
	return pages
}

// NewPagination creates a new pagination object
func NewPagination(page, pageSize int, total int64) *Pagination {
	return &Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: CalculateTotalPages(total, pageSize),
	}
}