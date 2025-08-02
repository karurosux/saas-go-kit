package authmodel

import (
	"fmt"
	"regexp"
	"strings"
	
	authconstants "{{.Project.GoModule}}/internal/auth/constants"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (r *LoginRequest) GetEmail() string {
	return strings.ToLower(strings.TrimSpace(r.Email))
}

func (r *LoginRequest) GetPassword() string {
	return r.Password
}

func (r *LoginRequest) Validate() error {
	if r.GetEmail() == "" {
		return fmt.Errorf(authconstants.ErrInvalidEmail)
	}
	if !emailRegex.MatchString(r.GetEmail()) {
		return fmt.Errorf(authconstants.ErrInvalidEmail)
	}
	if r.GetPassword() == "" {
		return fmt.Errorf(authconstants.ErrInvalidPassword)
	}
	return nil
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,e164"`
	Password string `json:"password" validate:"required,min=8"`
}

func (r *RegisterRequest) GetEmail() string {
	return strings.ToLower(strings.TrimSpace(r.Email))
}

func (r *RegisterRequest) GetPhone() string {
	return strings.TrimSpace(r.Phone)
}

func (r *RegisterRequest) GetPassword() string {
	return r.Password
}

func (r *RegisterRequest) Validate() error {
	if r.GetEmail() == "" {
		return fmt.Errorf(authconstants.ErrInvalidEmail)
	}
	if !emailRegex.MatchString(r.GetEmail()) {
		return fmt.Errorf(authconstants.ErrInvalidEmail)
	}
	if r.GetPhone() != "" && !phoneRegex.MatchString(r.GetPhone()) {
		return fmt.Errorf(authconstants.ErrInvalidPhone)
	}
	if len(r.GetPassword()) < 8 {
		return fmt.Errorf(authconstants.ErrPasswordTooWeak)
	}
	return nil
}