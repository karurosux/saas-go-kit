package smtpprovider

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"regexp"
	"strings"
	
	"{{.Project.GoModule}}/internal/notification/constants"
	"{{.Project.GoModule}}/internal/notification/interface"
)

// SMTPProvider implements the EmailProvider interface using SMTP
type SMTPProvider struct {
	config notificationinterface.EmailConfig
}

// NewSMTPProvider creates a new SMTP email provider
func NewSMTPProvider() notificationinterface.EmailProvider {
	return &SMTPProvider{}
}

func (p *SMTPProvider) Initialize(config notificationinterface.EmailConfig) error {
	if config.Host == "" {
		return fmt.Errorf(notificationconstants.ErrSMTPHostRequired)
	}
	if config.Port == 0 {
		config.Port = notificationconstants.DefaultSMTPPort
	}
	if config.FromEmail == "" {
		return fmt.Errorf(notificationconstants.ErrFromEmailRequired)
	}

	p.config = config
	return nil
}

func (p *SMTPProvider) GetProviderName() string {
	return notificationconstants.ProviderSMTP
}

func (p *SMTPProvider) SendEmail(ctx context.Context, req *notificationinterface.EmailRequest) error {
	return p.sendSingleEmail(req.To[0], req.Subject, req.Body, req.IsHTML)
}

func (p *SMTPProvider) SendTemplateEmail(ctx context.Context, req *notificationinterface.TemplateEmailRequest) error {
	// For SMTP provider, we need to resolve the template manually
	// In a real implementation, you might have a template engine
	body := p.renderTemplate(req.TemplateID, req.Variables)
	subject := req.Subject
	if subject == "" {
		subject = p.getTemplateSubject(req.TemplateID)
	}
	
	return p.sendSingleEmail(req.To[0], subject, body, true)
}

func (p *SMTPProvider) SendBulkEmails(ctx context.Context, req *notificationinterface.BulkEmailRequest) error {
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
		return fmt.Errorf(notificationconstants.ErrInvalidEmailFormat)
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
	contentType := notificationconstants.ContentTypeText
	if isHTML {
		contentType = notificationconstants.ContentTypeHTML
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
		return fmt.Errorf("%s: %w", notificationconstants.ErrFailedToSendEmail, err)
	}

	log.Printf("Email sent successfully to %s", to)
	return nil
}

func (p *SMTPProvider) logEmailInDevelopment(to, subject, body string) error {
	log.Printf("=== EMAIL (Development Mode) ===")
	log.Printf("Provider: %s", p.GetProviderName())
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
		notificationconstants.TemplateWelcome: `
		<html>
		<body>
			<h2>Welcome to {{app_name}}, {{user_name}}!</h2>
			<p>Thank you for joining us. We're excited to have you on board.</p>
			<p>Get started by <a href="{{app_url}}">exploring your dashboard</a>.</p>
		</body>
		</html>`,
		
		notificationconstants.TemplatePasswordReset: `
		<html>
		<body>
			<h2>Password Reset Request</h2>
			<p>Hi {{user_name}},</p>
			<p>You requested to reset your password. Click the link below:</p>
			<p><a href="{{reset_url}}">Reset Password</a></p>
			<p>If you didn't request this, please ignore this email.</p>
		</body>
		</html>`,
		
		notificationconstants.TemplateTeamInvitation: `
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
		notificationconstants.TemplateWelcome:         "Welcome to {{app_name}}!",
		notificationconstants.TemplatePasswordReset:  "Reset Your Password",
		notificationconstants.TemplateTeamInvitation: "You've been invited to join {{team_name}}",
	}

	subject, exists := subjects[templateID]
	if !exists {
		return "Notification"
	}
	return subject
}