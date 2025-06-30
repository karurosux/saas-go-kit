package saasgokit

import (
	authGo "github.com/karurosux/saas-go-kit/auth-go"
	coreGo "github.com/karurosux/saas-go-kit/core-go"
	errorsGo "github.com/karurosux/saas-go-kit/errors-go"
	ratelimitGo "github.com/karurosux/saas-go-kit/ratelimit-go"
	responseGo "github.com/karurosux/saas-go-kit/response-go"
	validatorGo "github.com/karurosux/saas-go-kit/validator-go"
)

// Re-export core types and functions
type (
	Kit = coreGo.Kit
)

var (
	NewKit = coreGo.NewKit
)

// Re-export auth types and functions
type (
	AuthService = authGo.AuthService
	User        = authGo.User
	LoginRequest = authGo.LoginRequest
	RegisterRequest = authGo.RegisterRequest
	JWTClaims   = authGo.JWTClaims
)

var (
	NewAuthService = authGo.NewAuthService
)

// Re-export error types and functions
type (
	Error = errorsGo.Error
)

var (
	NewError = errorsGo.NewError
	NotFound = errorsGo.NotFound
	BadRequest = errorsGo.BadRequest
	Unauthorized = errorsGo.Unauthorized
	Forbidden = errorsGo.Forbidden
	InternalServerError = errorsGo.InternalServerError
)

// Re-export response types and functions
type (
	Response = responseGo.Response
)

var (
	JSON = responseGo.JSON
	ResponseError = responseGo.Error
	Success = responseGo.Success
)

// Re-export validator functions
var (
	Validate = validatorGo.Validate
)

// Re-export ratelimit types and functions
var (
	RateLimit = ratelimitGo.RateLimit
)