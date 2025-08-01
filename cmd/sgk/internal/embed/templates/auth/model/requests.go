package authmodel

import (
	"{{.Project.GoModule}}/internal/auth/constants"
	"{{.Project.GoModule}}/internal/core"
	"regexp"
	"strings"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
)

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// GetEmail returns the email
func (r *LoginRequest) GetEmail() string {
	return strings.ToLower(strings.TrimSpace(r.Email))
}

// GetPassword returns the password
func (r *LoginRequest) GetPassword() string {
	return r.Password
}

// Validate validates the login request
func (r *LoginRequest) Validate() error {
	if r.GetEmail() == "" {
		return core.BadRequest(authconstants.ErrInvalidEmail)
	}
	if !emailRegex.MatchString(r.GetEmail()) {
		return core.BadRequest(authconstants.ErrInvalidEmail)
	}
	if r.GetPassword() == "" {
		return core.BadRequest(authconstants.ErrInvalidPassword)
	}
	return nil
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,e164"`
	Password string `json:"password" validate:"required,min=8"`
}

// GetEmail returns the email
func (r *RegisterRequest) GetEmail() string {
	return strings.ToLower(strings.TrimSpace(r.Email))
}

// GetPhone returns the phone
func (r *RegisterRequest) GetPhone() string {
	return strings.TrimSpace(r.Phone)
}

// GetPassword returns the password
func (r *RegisterRequest) GetPassword() string {
	return r.Password
}

// Validate validates the registration request
func (r *RegisterRequest) Validate() error {
	if r.GetEmail() == "" {
		return core.BadRequest(authconstants.ErrInvalidEmail)
	}
	if !emailRegex.MatchString(r.GetEmail()) {
		return core.BadRequest(authconstants.ErrInvalidEmail)
	}
	if r.GetPhone() != "" && !phoneRegex.MatchString(r.GetPhone()) {
		return core.BadRequest(authconstants.ErrInvalidPhone)
	}
	if len(r.GetPassword()) < 8 {
		return core.BadRequest(authconstants.ErrPasswordTooWeak)
	}
	return nil
}