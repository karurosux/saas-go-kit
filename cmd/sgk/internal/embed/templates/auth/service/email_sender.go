package authservice

import (
	"fmt"
	
	"{{.Project.GoModule}}/internal/auth/interface"
)

// MockEmailSender implements email sending (mock implementation)
type MockEmailSender struct {
	baseURL string
}

// NewMockEmailSender creates a new mock email sender
func NewMockEmailSender(baseURL string) authinterface.EmailSender {
	return &MockEmailSender{baseURL: baseURL}
}

// SendVerificationEmail sends a verification email
func (s *MockEmailSender) SendVerificationEmail(email, token string) error {
	// In production, implement actual email sending
	fmt.Printf("Sending verification email to %s\n", email)
	fmt.Printf("Verification link: %s/auth/verify-email?token=%s\n", s.baseURL, token)
	return nil
}

// SendPasswordResetEmail sends a password reset email
func (s *MockEmailSender) SendPasswordResetEmail(email, token string) error {
	// In production, implement actual email sending
	fmt.Printf("Sending password reset email to %s\n", email)
	fmt.Printf("Reset link: %s/auth/reset-password?token=%s\n", s.baseURL, token)
	return nil
}

// SendWelcomeEmail sends a welcome email
func (s *MockEmailSender) SendWelcomeEmail(email string) error {
	// In production, implement actual email sending
	fmt.Printf("Sending welcome email to %s\n", email)
	return nil
}