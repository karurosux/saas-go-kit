package emailservice

import (
	"context"
	"fmt"
	"log"
	"time"

	emailinterface "{{.Project.GoModule}}/internal/email/interface"
)

type EmailService struct {
	sender      emailinterface.EmailSender
	queue       emailinterface.EmailQueue
	templates   emailinterface.TemplateManager
	defaultFrom string
}

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

func (s *EmailService) Send(ctx context.Context, to []string, subject, body string) error {
	message := &emailinterface.EmailMessage{
		To:      to,
		From:    s.defaultFrom,
		Subject: subject,
		Body:    body,
	}
	
	return s.sender.Send(ctx, message)
}

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

func (s *EmailService) SendTemplate(ctx context.Context, to []string, templateName string, data map[string]interface{}) error {
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

func (s *EmailService) QueueEmail(ctx context.Context, message *emailinterface.EmailMessage) error {
	if message.From == "" {
		message.From = s.defaultFrom
	}
	
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

func (s *EmailService) ProcessQueue(ctx context.Context) error {
	messages, err := s.queue.Dequeue(ctx, 10)
	if err != nil {
		return fmt.Errorf("failed to dequeue messages: %w", err)
	}
	
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

func (s *EmailService) processMessage(ctx context.Context, message *emailinterface.EmailMessage) error {
	err := s.sender.Send(ctx, message)
	if err != nil {
		if queueErr := s.queue.MarkAsFailed(ctx, message.ID, err); queueErr != nil {
			log.Printf("Failed to mark email %d as failed: %v", message.ID, queueErr)
		}
		return err
	}
	
	if err := s.queue.MarkAsSent(ctx, message.ID); err != nil {
		log.Printf("Failed to mark email %d as sent: %v", message.ID, err)
	}
	
	return nil
}

func (s *EmailService) GetEmailStatus(ctx context.Context, id uint) (*emailinterface.EmailMessage, error) {
	return s.queue.GetStatus(ctx, id)
}

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