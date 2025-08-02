package emailservice

import (
	"context"
	"fmt"
	"log"
	"time"

	emailinterface "{{.Project.GoModule}}/internal/email/interface"
)

// EmailService implements the main email service
type EmailService struct {
	sender      emailinterface.EmailSender
	queue       emailinterface.EmailQueue
	templates   emailinterface.TemplateManager
	defaultFrom string
}

// NewEmailService creates a new email service
func NewEmailService(
	sender emailinterface.EmailSender,
	queue emailinterface.EmailQueue,
	templates emailinterface.TemplateManager,
	defaultFrom string,
) *EmailService {
	return &EmailService{
		sender:      sender,
		queue:       queue,
		templates:   templates,
		defaultFrom: defaultFrom,
	}
}

// Send sends a plain text email immediately
func (s *EmailService) Send(ctx context.Context, to []string, subject, body string) error {
	message := &emailinterface.EmailMessage{
		To:      to,
		From:    s.defaultFrom,
		Subject: subject,
		Body:    body,
	}
	
	return s.sender.Send(ctx, message)
}

// SendHTML sends an HTML email immediately
func (s *EmailService) SendHTML(ctx context.Context, to []string, subject, body, html string) error {
	message := &emailinterface.EmailMessage{
		To:      to,
		From:    s.defaultFrom,
		Subject: subject,
		Body:    body,
		HTML:    html,
	}
	
	return s.sender.Send(ctx, message)
}

// SendTemplate sends an email using a template
func (s *EmailService) SendTemplate(ctx context.Context, to []string, templateName string, data map[string]interface{}) error {
	// Render template
	subject, body, html, err := s.templates.RenderTemplate(ctx, templateName, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}
	
	message := &emailinterface.EmailMessage{
		To:           to,
		From:         s.defaultFrom,
		Subject:      subject,
		Body:         body,
		HTML:         html,
		Template:     templateName,
		TemplateData: data,
	}
	
	return s.sender.Send(ctx, message)
}

// QueueEmail adds an email to the queue for async sending
func (s *EmailService) QueueEmail(ctx context.Context, message *emailinterface.EmailMessage) error {
	if message.From == "" {
		message.From = s.defaultFrom
	}
	
	// If template is specified, render it
	if message.Template != "" && s.templates != nil {
		subject, body, html, err := s.templates.RenderTemplate(ctx, message.Template, message.TemplateData)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}
		message.Subject = subject
		message.Body = body
		message.HTML = html
	}
	
	return s.queue.Enqueue(ctx, message)
}

// ProcessQueue processes pending emails in the queue
func (s *EmailService) ProcessQueue(ctx context.Context) error {
	// Get batch of pending emails
	messages, err := s.queue.Dequeue(ctx, 10)
	if err != nil {
		return fmt.Errorf("failed to dequeue messages: %w", err)
	}
	
	// Process each message
	for _, message := range messages {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := s.processMessage(ctx, message); err != nil {
				log.Printf("Failed to send email %d: %v", message.ID, err)
			}
		}
	}
	
	return nil
}

// processMessage processes a single email message
func (s *EmailService) processMessage(ctx context.Context, message *emailinterface.EmailMessage) error {
	// Send the email
	err := s.sender.Send(ctx, message)
	if err != nil {
		// Mark as failed
		if queueErr := s.queue.MarkAsFailed(ctx, message.ID, err); queueErr != nil {
			log.Printf("Failed to mark email %d as failed: %v", message.ID, queueErr)
		}
		return err
	}
	
	// Mark as sent
	if err := s.queue.MarkAsSent(ctx, message.ID); err != nil {
		log.Printf("Failed to mark email %d as sent: %v", message.ID, err)
	}
	
	return nil
}

// GetEmailStatus retrieves the status of a queued email
func (s *EmailService) GetEmailStatus(ctx context.Context, id uint) (*emailinterface.EmailMessage, error) {
	return s.queue.GetStatus(ctx, id)
}

// StartQueueProcessor starts a background worker to process the email queue
func (s *EmailService) StartQueueProcessor(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.ProcessQueue(ctx); err != nil {
					log.Printf("Error processing email queue: %v", err)
				}
			}
		}
	}()
}