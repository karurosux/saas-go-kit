package authservice

import (
	"fmt"
	
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
)

// MockSMSSender implements SMS sending (mock implementation)
type MockSMSSender struct{}

// NewMockSMSSender creates a new mock SMS sender
func NewMockSMSSender() authinterface.SMSSender {
	return &MockSMSSender{}
}

// SendVerificationSMS sends a verification SMS
func (s *MockSMSSender) SendVerificationSMS(phone, code string) error {
	// In production, implement actual SMS sending
	fmt.Printf("Sending verification SMS to %s\n", phone)
	fmt.Printf("Verification code: %s\n", code)
	return nil
}