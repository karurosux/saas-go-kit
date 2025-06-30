package main

import (
	"context"
	"log"
	"time"
)

// ConsoleEmailProvider implements auth.EmailProvider for development
type ConsoleEmailProvider struct{}

func NewConsoleEmailProvider() *ConsoleEmailProvider {
	return &ConsoleEmailProvider{}
}

func (p *ConsoleEmailProvider) SendVerificationEmail(ctx context.Context, email, token string) error {
	log.Printf(`
======================================
EMAIL: Verification Email
TO: %s
--------------------------------------
Please verify your email by clicking the link below:

http://localhost:8080/api/v1/auth/verify-email?token=%s

This link will expire in 24 hours.
======================================
`, email, token)
	return nil
}

func (p *ConsoleEmailProvider) SendPasswordResetEmail(ctx context.Context, email, token string) error {
	log.Printf(`
======================================
EMAIL: Password Reset
TO: %s
--------------------------------------
You requested a password reset. Click the link below to reset your password:

http://localhost:8080/api/v1/auth/reset-password?token=%s

This link will expire in 1 hour.

If you didn't request this, please ignore this email.
======================================
`, email, token)
	return nil
}

func (p *ConsoleEmailProvider) SendEmailChangeConfirmation(ctx context.Context, oldEmail, newEmail, token string) error {
	log.Printf(`
======================================
EMAIL: Email Change Confirmation
TO: %s
NEW EMAIL: %s
--------------------------------------
Please confirm your email change by clicking the link below:

http://localhost:8080/api/v1/auth/confirm-email-change?token=%s

This link will expire in 24 hours.
======================================
`, oldEmail, newEmail, token)
	return nil
}

func (p *ConsoleEmailProvider) SendWelcomeEmail(ctx context.Context, email string) error {
	log.Printf(`
======================================
EMAIL: Welcome!
TO: %s
--------------------------------------
Welcome to SaaS Go Kit!

Your email has been verified and your account is now active.

Get started by exploring our API documentation.
======================================
`, email)
	return nil
}

// AppConfig implements auth.ConfigProvider
type AppConfig struct {
	jwtSecret      string
	jwtExpiration  time.Duration
	appURL         string
	appName        string
	isDev          bool
}

func (c *AppConfig) GetJWTSecret() string {
	return c.jwtSecret
}

func (c *AppConfig) GetJWTExpiration() time.Duration {
	return c.jwtExpiration
}

func (c *AppConfig) GetRefreshExpiration() time.Duration {
	return 7 * 24 * time.Hour
}

func (c *AppConfig) GetAppURL() string {
	return c.appURL
}

func (c *AppConfig) GetAppName() string {
	return c.appName
}

func (c *AppConfig) IsDevMode() bool {
	return c.isDev
}