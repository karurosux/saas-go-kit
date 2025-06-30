package notification

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"regexp"
	"strings"
)

type SMTPProvider struct {
	config EmailConfig
}

func NewSMTPProvider() *SMTPProvider {
	return &SMTPProvider{}
}

func (p *SMTPProvider) Initialize(config EmailConfig) error {
	if config.Host == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if config.Port == 0 {
		config.Port = 587 // Default SMTP port
	}
	if config.FromEmail == "" {
		return fmt.Errorf("from email is required")
	}

	p.config = config
	return nil
}

func (p *SMTPProvider) GetProviderName() string {
	return "smtp"
}

func (p *SMTPProvider) SendEmail(ctx context.Context, req *EmailRequest) error {
	return p.sendSingleEmail(req.To[0], req.Subject, req.Body, req.IsHTML)
}

func (p *SMTPProvider) SendTemplateEmail(ctx context.Context, req *TemplateEmailRequest) error {
	// For SMTP provider, we need to resolve the template manually
	// In a real implementation, you might have a template engine
	body := p.renderTemplate(req.TemplateID, req.Variables)
	subject := req.Subject
	if subject == "" {
		subject = p.getTemplateSubject(req.TemplateID)
	}
	
	return p.sendSingleEmail(req.To[0], subject, body, true)
}

func (p *SMTPProvider) SendBulkEmails(ctx context.Context, req *BulkEmailRequest) error {
	for _, recipient := range req.Recipients {
		body := p.replaceVariables(req.Body, recipient.Variables)
		if err := p.sendSingleEmail(recipient.Email, req.Subject, body, req.IsHTML); err != nil {
			log.Printf("Failed to send email to %s: %v", recipient.Email, err)
			// Continue with other recipients
		}
	}
	return nil
}

func (p *SMTPProvider) VerifyEmail(ctx context.Context, email string) error {
	// Basic email format validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func (p *SMTPProvider) sendSingleEmail(to, subject, body string, isHTML bool) error {
	// Check if we're in development mode (simplified check)
	if p.config.Host == "localhost" || p.config.Host == "127.0.0.1" || p.config.Username == "dev" {
		return p.logEmailInDevelopment(to, subject, body)
	}

	// Production email sending
	if p.config.Username == "" || p.config.Password == "" {
		log.Printf("SMTP credentials not configured, skipping email to %s", to)
		return nil
	}

	// Set content type based on isHTML
	contentType := "text/plain; charset=UTF-8"
	if isHTML {
		contentType = "text/html; charset=UTF-8"
	}

	// Compose email message
	fromHeader := p.config.FromEmail
	if p.config.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", p.config.FromName, p.config.FromEmail)
	}

	headers := map[string]string{
		"From":         fromHeader,
		"To":           to,
		"Subject":      subject,
		"Content-Type": contentType,
		"MIME-Version": "1.0",
	}

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Setup authentication
	auth := smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.Host)

	// Send email
	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)
	err := smtp.SendMail(addr, auth, p.config.FromEmail, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent successfully to %s", to)
	return nil
}

func (p *SMTPProvider) logEmailInDevelopment(to, subject, body string) error {
	log.Printf("=== EMAIL (Development Mode) ===")
	log.Printf("To: %s", to)
	log.Printf("Subject: %s", subject)
	log.Printf("Body: %s", body)

	// Extract and highlight any links for easy clicking in terminal
	if strings.Contains(body, "href=") {
		linkRegex := regexp.MustCompile(`href="([^"]+)"`)
		matches := linkRegex.FindAllStringSubmatch(body, -1)
		if len(matches) > 0 {
			log.Println("\n🔗 CLICKABLE LINKS:")
			for _, match := range matches {
				if len(match) > 1 {
					log.Printf("   👉 %s", match[1])
				}
			}
		}
	}
	log.Printf("================================\n")
	return nil
}

func (p *SMTPProvider) renderTemplate(templateID string, variables map[string]interface{}) string {
	// This is a simplified template rendering
	// In a real implementation, you would use a proper template engine
	template, exists := p.getTemplate(templateID)
	if !exists {
		return fmt.Sprintf("Template %s not found", templateID)
	}

	return p.replaceVariables(template, variables)
}

func (p *SMTPProvider) replaceVariables(content string, variables map[string]interface{}) string {
	result := content
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

func (p *SMTPProvider) getTemplate(templateID string) (string, bool) {
	// Predefined templates - in a real implementation, these would come from a database
	templates := map[string]string{
		"welcome": `
		<html>
		<body>
			<h2>Welcome to {{app_name}}, {{user_name}}!</h2>
			<p>Thank you for joining us. We're excited to have you on board.</p>
			<p>Get started by <a href="{{app_url}}">exploring your dashboard</a>.</p>
		</body>
		</html>`,
		
		"password_reset": `
		<html>
		<body>
			<h2>Password Reset Request</h2>
			<p>Hi {{user_name}},</p>
			<p>You requested to reset your password. Click the link below:</p>
			<p><a href="{{reset_url}}">Reset Password</a></p>
			<p>If you didn't request this, please ignore this email.</p>
		</body>
		</html>`,
		
		"team_invitation": `
		<html>
		<body>
			<h2>Team Invitation</h2>
			<p>{{inviter_name}} has invited you to join {{team_name}}.</p>
			<p><a href="{{invite_url}}">Accept Invitation</a></p>
			<p>This invitation expires in 7 days.</p>
		</body>
		</html>`,
	}

	template, exists := templates[templateID]
	return template, exists
}

func (p *SMTPProvider) getTemplateSubject(templateID string) string {
	subjects := map[string]string{
		"welcome":         "Welcome to {{app_name}}!",
		"password_reset":  "Reset Your Password",
		"team_invitation": "You've been invited to join {{team_name}}",
	}

	subject, exists := subjects[templateID]
	if !exists {
		return "Notification"
	}
	return subject
}

// Development-specific SMTP provider for easier testing
type DevSMTPProvider struct {
	*SMTPProvider
}

func NewDevSMTPProvider() *DevSMTPProvider {
	return &DevSMTPProvider{
		SMTPProvider: &SMTPProvider{
			config: EmailConfig{
				Host:      "localhost",
				Port:      587,
				Username:  "dev",
				Password:  "dev",
				FromEmail: "dev@localhost",
				FromName:  "Development Server",
			},
		},
	}
}

func (p *DevSMTPProvider) Initialize(config EmailConfig) error {
	// Override with development settings
	p.config = EmailConfig{
		Host:      "localhost",
		Port:      587,
		Username:  "dev",
		Password:  "dev",
		FromEmail: config.FromEmail,
		FromName:  config.FromName,
	}
	return nil
}

func (p *DevSMTPProvider) SendEmail(ctx context.Context, req *EmailRequest) error {
	return p.logEmailInDevelopment(strings.Join(req.To, ", "), req.Subject, req.Body)
}

func (p *DevSMTPProvider) SendTemplateEmail(ctx context.Context, req *TemplateEmailRequest) error {
	body := p.renderTemplate(req.TemplateID, req.Variables)
	subject := req.Subject
	if subject == "" {
		subject = p.getTemplateSubject(req.TemplateID)
		subject = p.replaceVariables(subject, req.Variables)
	}
	
	return p.logEmailInDevelopment(strings.Join(req.To, ", "), subject, body)
}

func (p *DevSMTPProvider) SendBulkEmails(ctx context.Context, req *BulkEmailRequest) error {
	log.Printf("=== BULK EMAIL (Development Mode) ===")
	log.Printf("Subject: %s", req.Subject)
	log.Printf("Recipients: %d", len(req.Recipients))
	
	for i, recipient := range req.Recipients {
		body := p.replaceVariables(req.Body, recipient.Variables)
		log.Printf("\n--- Recipient %d ---", i+1)
		log.Printf("To: %s", recipient.Email)
		log.Printf("Body: %s", body)
	}
	log.Printf("=====================================\n")
	return nil
}