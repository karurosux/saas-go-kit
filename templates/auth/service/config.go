package authservice

import (
	"time"
	
	"{{.Project.GoModule}}/internal/auth/interface"
	"golang.org/x/crypto/bcrypt"
)

// DefaultAuthConfig implements default auth configuration
type DefaultAuthConfig struct {
	jwtSecret                      string
	jwtExpiration                  time.Duration
	refreshTokenExpiration         time.Duration
	verificationTokenExpiration    time.Duration
	passwordResetTokenExpiration   time.Duration
	bcryptCost                     int
	emailVerificationRequired      bool
	phoneVerificationRequired      bool
}

// NewDefaultAuthConfig creates a new default auth config
func NewDefaultAuthConfig() authinterface.AuthConfig {
	return &DefaultAuthConfig{
		jwtSecret:                      "your-secret-key", // Should be loaded from env
		jwtExpiration:                  15 * time.Minute,
		refreshTokenExpiration:         7 * 24 * time.Hour,
		verificationTokenExpiration:    24 * time.Hour,
		passwordResetTokenExpiration:   1 * time.Hour,
		bcryptCost:                     bcrypt.DefaultCost,
		emailVerificationRequired:      true,
		phoneVerificationRequired:      false,
	}
}

func (c *DefaultAuthConfig) GetJWTSecret() string {
	return c.jwtSecret
}

func (c *DefaultAuthConfig) GetJWTExpiration() time.Duration {
	return c.jwtExpiration
}

func (c *DefaultAuthConfig) GetRefreshTokenExpiration() time.Duration {
	return c.refreshTokenExpiration
}

func (c *DefaultAuthConfig) GetVerificationTokenExpiration() time.Duration {
	return c.verificationTokenExpiration
}

func (c *DefaultAuthConfig) GetPasswordResetTokenExpiration() time.Duration {
	return c.passwordResetTokenExpiration
}

func (c *DefaultAuthConfig) GetBcryptCost() int {
	return c.bcryptCost
}

func (c *DefaultAuthConfig) IsEmailVerificationRequired() bool {
	return c.emailVerificationRequired
}

func (c *DefaultAuthConfig) IsPhoneVerificationRequired() bool {
	return c.phoneVerificationRequired
}