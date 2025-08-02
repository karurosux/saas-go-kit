package authservice

import (
	"fmt"
	
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
)

type MockSMSSender struct{}

func NewMockSMSSender() authinterface.SMSSender {
	return &MockSMSSender{}
}

func (s *MockSMSSender) SendVerificationSMS(phone, code string) error {
	fmt.Printf("Sending verification SMS to %s\n", phone)
	fmt.Printf("Verification code: %s\n", code)
	return nil
}