package authservice

import (
	"fmt"
	
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
)

type MockEmailSender struct {
	baseURL string
}

func NewMockEmailSender(baseURL string) authinterface.EmailSender {
	return &MockEmailSender{baseURL: baseURL}
}

func (s *MockEmailSender) SendVerificationEmail(email, token string) error {
	fmt.Printf("Sending verification email to %s\n", email)
	fmt.Printf("Verification link: %s/auth/verify-email?token=%s\n", s.baseURL, token)
	return nil
}

func (s *MockEmailSender) SendPasswordResetEmail(email, token string) error {
	fmt.Printf("Sending password reset email to %s\n", email)
	fmt.Printf("Reset link: %s/auth/reset-password?token=%s\n", s.baseURL, token)
	return nil
}

func (s *MockEmailSender) SendWelcomeEmail(email string) error {
	fmt.Printf("Sending welcome email to %s\n", email)
	return nil
}