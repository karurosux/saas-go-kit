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
	Strategy    string         `json:"strategy"`
	Credentials map[string]any `json:"credentials"`
}

func (r *LoginRequest) GetStrategy() string {
	if r.Strategy == "" {
		return "email_password"
	}
	return r.Strategy
}

func (r *LoginRequest) GetCredentials() map[string]any {
	if r.Credentials == nil {
		return make(map[string]any)
	}
	return r.Credentials
}

func (r *LoginRequest) Validate() error {
	strategy := r.GetStrategy()
	creds := r.GetCredentials()
	
	switch strategy {
	case "email_password":
		email, ok := creds["email"].(string)
		if !ok || email == "" {
			return fmt.Errorf(authconstants.ErrInvalidEmail)
		}
		if !emailRegex.MatchString(strings.ToLower(strings.TrimSpace(email))) {
			return fmt.Errorf(authconstants.ErrInvalidEmail)
		}
		password, ok := creds["password"].(string)
		if !ok || password == "" {
			return fmt.Errorf(authconstants.ErrInvalidPassword)
		}
	case "google", "github", "facebook":
		code, ok := creds["code"].(string)
		if !ok || code == "" {
			return fmt.Errorf("authorization code is required")
		}
	default:
		return fmt.Errorf("unsupported authentication strategy: %s", strategy)
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