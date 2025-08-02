package emailservice

import (
	"context"
	"log"

	emailinterface "{{.Project.GoModule}}/internal/email/interface"
)

// MockSender implements EmailSender for development/testing
type MockSender struct {
	logEmails bool
}

// NewMockSender creates a new mock email sender
func NewMockSender(logEmails bool) *MockSender {
	return &MockSender{
		logEmails: logEmails,
	}
}

// Send logs the email instead of sending it
func (m *MockSender) Send(ctx context.Context, message *emailinterface.EmailMessage) error {
	if m.logEmails {
		log.Printf("📧 Mock Email Sent:\n")
		log.Printf("  To: %v\n", message.To)
		if len(message.CC) > 0 {
			log.Printf("  CC: %v\n", message.CC)
		}
		log.Printf("  Subject: %s\n", message.Subject)
		log.Printf("  Body: %s\n", message.Body)
		if message.HTML != "" {
			log.Printf("  HTML: %s\n", message.HTML[:min(100, len(message.HTML))]+"...")
		}
		if message.Template != "" {
			log.Printf("  Template: %s\n", message.Template)
			log.Printf("  Template Data: %+v\n", message.TemplateData)
		}
	}
	return nil
}

// SendBatch logs multiple emails
func (m *MockSender) SendBatch(ctx context.Context, messages []*emailinterface.EmailMessage) error {
	for _, message := range messages {
		if err := m.Send(ctx, message); err != nil {
			return err
		}
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}