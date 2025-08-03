package authservice

import (
	"context"
	"fmt"
	
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
	emailinterface "{{.Project.GoModule}}/internal/email/interface"
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

type EmailSenderAdapter struct {
	emailService emailinterface.EmailService
	baseURL      string
}

func NewEmailSenderAdapter(emailService emailinterface.EmailService, baseURL string) authinterface.EmailSender {
	return &EmailSenderAdapter{
		emailService: emailService,
		baseURL:      baseURL,
	}
}

func (a *EmailSenderAdapter) SendVerificationEmail(email, token string) error {
	ctx := context.Background()
	
	verificationLink := fmt.Sprintf("%s/auth/verify-email?token=%s", a.baseURL, token)
	
	data := map[string]interface{}{
		"email":             email,
		"verificationLink":  verificationLink,
		"baseURL":          a.baseURL,
	}
	
	err := a.emailService.SendTemplate(ctx, []string{email}, "email-verification", data)
	if err != nil {
		subject := "Verify Your Email Address"
		body := fmt.Sprintf(
			"Hello,\n\n"+
			"Please verify your email address by clicking the link below:\n\n"+
			"%s\n\n"+
			"This link will expire in 24 hours.\n\n"+
			"If you didn't create an account, please ignore this email.\n\n"+
			"Best regards,\n"+
			"The Team",
			verificationLink,
		)
		
		return a.emailService.Send(ctx, []string{email}, subject, body)
	}
	
	return nil
}

func (a *EmailSenderAdapter) SendPasswordResetEmail(email, token string) error {
	ctx := context.Background()
	
	resetLink := fmt.Sprintf("%s/auth/reset-password?token=%s", a.baseURL, token)
	
	data := map[string]interface{}{
		"email":      email,
		"resetLink":  resetLink,
		"baseURL":    a.baseURL,
	}
	
	err := a.emailService.SendTemplate(ctx, []string{email}, "password-reset", data)
	if err != nil {
		subject := "Reset Your Password"
		body := fmt.Sprintf(
			"Hello,\n\n"+
			"You requested to reset your password. Click the link below to proceed:\n\n"+
			"%s\n\n"+
			"This link will expire in 1 hour.\n\n"+
			"If you didn't request this, please ignore this email.\n\n"+
			"Best regards,\n"+
			"The Team",
			resetLink,
		)
		
		return a.emailService.Send(ctx, []string{email}, subject, body)
	}
	
	return nil
}

func (a *EmailSenderAdapter) SendWelcomeEmail(email string) error {
	ctx := context.Background()
	
	data := map[string]interface{}{
		"email":   email,
		"baseURL": a.baseURL,
	}
	
	err := a.emailService.SendTemplate(ctx, []string{email}, "welcome", data)
	if err != nil {
		subject := "Welcome to Our Platform!"
		body := fmt.Sprintf(
			"Hello,\n\n"+
			"Welcome to our platform! We're excited to have you on board.\n\n"+
			"Your account has been successfully created with the email: %s\n\n"+
			"If you have any questions, feel free to reach out to our support team.\n\n"+
			"Best regards,\n"+
			"The Team",
			email,
		)
		
		return a.emailService.Send(ctx, []string{email}, subject, body)
	}
	
	return nil
}