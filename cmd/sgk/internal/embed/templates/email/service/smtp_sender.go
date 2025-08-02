package emailservice

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	emailinterface "{{.Project.GoModule}}/internal/email/interface"
)

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	From       string
	FromName   string
	TLS        bool
	SkipVerify bool
}

// SMTPSender implements EmailSender using SMTP
type SMTPSender struct {
	config SMTPConfig
}

// NewSMTPSender creates a new SMTP email sender
func NewSMTPSender(config SMTPConfig) *SMTPSender {
	return &SMTPSender{
		config: config,
	}
}

// Send sends a single email message
func (s *SMTPSender) Send(ctx context.Context, message *emailinterface.EmailMessage) error {
	// Set default from address if not provided
	if message.From == "" {
		message.From = s.config.From
	}

	// Build the email
	emailBody := s.buildEmail(message)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	
	var auth smtp.Auth
	if s.config.Username != "" && s.config.Password != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	// Send email
	if s.config.TLS {
		return s.sendTLS(addr, auth, message, emailBody)
	}
	
	return smtp.SendMail(addr, auth, message.From, message.To, []byte(emailBody))
}

// SendBatch sends multiple email messages
func (s *SMTPSender) SendBatch(ctx context.Context, messages []*emailinterface.EmailMessage) error {
	for _, message := range messages {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := s.Send(ctx, message); err != nil {
				return fmt.Errorf("failed to send email to %v: %w", message.To, err)
			}
		}
	}
	return nil
}

// buildEmail constructs the email message
func (s *SMTPSender) buildEmail(message *emailinterface.EmailMessage) string {
	var builder strings.Builder

	// Headers
	builder.WriteString(fmt.Sprintf("From: %s\r\n", s.formatFrom()))
	builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(message.To, ", ")))
	
	if len(message.CC) > 0 {
		builder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(message.CC, ", ")))
	}
	
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", message.Subject))
	
	// MIME headers for HTML email
	if message.HTML != "" {
		builder.WriteString("MIME-Version: 1.0\r\n")
		builder.WriteString("Content-Type: multipart/alternative; boundary=\"boundary\"\r\n")
		builder.WriteString("\r\n")
		
		// Plain text part
		builder.WriteString("--boundary\r\n")
		builder.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(message.Body)
		builder.WriteString("\r\n")
		
		// HTML part
		builder.WriteString("--boundary\r\n")
		builder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(message.HTML)
		builder.WriteString("\r\n")
		
		builder.WriteString("--boundary--\r\n")
	} else {
		builder.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(message.Body)
	}

	return builder.String()
}

// formatFrom formats the from address with name if available
func (s *SMTPSender) formatFrom() string {
	if s.config.FromName != "" {
		return fmt.Sprintf("%s <%s>", s.config.FromName, s.config.From)
	}
	return s.config.From
}

// sendTLS sends email using TLS
func (s *SMTPSender) sendTLS(addr string, auth smtp.Auth, message *emailinterface.EmailMessage, emailBody string) error {
	tlsConfig := &tls.Config{
		ServerName:         s.config.Host,
		InsecureSkipVerify: s.config.SkipVerify,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	if err := client.Mail(message.From); err != nil {
		return err
	}

	recipients := append(message.To, message.CC...)
	recipients = append(recipients, message.BCC...)
	
	for _, addr := range recipients {
		if err := client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(emailBody))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}