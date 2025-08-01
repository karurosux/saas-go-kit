package devprovider

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	
	"{{.Project.GoModule}}/internal/notification/constants"
	"{{.Project.GoModule}}/internal/notification/interface"
	smtpprovider "{{.Project.GoModule}}/internal/notification/provider/smtp"
)

// DevEmailProvider is a development-specific email provider that logs emails instead of sending them
type DevEmailProvider struct {
	*smtpprovider.SMTPProvider
	config notificationinterface.EmailConfig
}

// NewDevEmailProvider creates a new development email provider
func NewDevEmailProvider() notificationinterface.EmailProvider {
	return &DevEmailProvider{
		SMTPProvider: &smtpprovider.SMTPProvider{},
		config: notificationinterface.EmailConfig{
			Host:      "localhost",
			Port:      587,
			Username:  "dev",
			Password:  "dev",
			FromEmail: "dev@localhost",
			FromName:  "Development Server",
		},
	}
}

func (p *DevEmailProvider) Initialize(config notificationinterface.EmailConfig) error {
	// Override with development settings but preserve from email and name if provided
	p.config = notificationinterface.EmailConfig{
		Host:      "localhost",
		Port:      587,
		Username:  "dev",
		Password:  "dev",
		FromEmail: config.FromEmail,
		FromName:  config.FromName,
	}
	
	// Use default development email if not provided
	if p.config.FromEmail == "" {
		p.config.FromEmail = "dev@localhost"
	}
	if p.config.FromName == "" {
		p.config.FromName = "Development Server"
	}
	
	return nil
}

func (p *DevEmailProvider) GetProviderName() string {
	return "dev-smtp"
}

func (p *DevEmailProvider) SendEmail(ctx context.Context, req *notificationinterface.EmailRequest) error {
	return p.logEmailInDevelopment(strings.Join(req.To, ", "), req.Subject, req.Body)
}

func (p *DevEmailProvider) SendTemplateEmail(ctx context.Context, req *notificationinterface.TemplateEmailRequest) error {
	body := p.renderTemplate(req.TemplateID, req.Variables)
	subject := req.Subject
	if subject == "" {
		subject = p.getTemplateSubject(req.TemplateID)
		subject = p.replaceVariables(subject, req.Variables)
	}
	
	return p.logEmailInDevelopment(strings.Join(req.To, ", "), subject, body)
}

func (p *DevEmailProvider) SendBulkEmails(ctx context.Context, req *notificationinterface.BulkEmailRequest) error {
	log.Printf("=== BULK EMAIL (Development Mode) ===")
	log.Printf("Provider: %s", p.GetProviderName())
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

func (p *DevEmailProvider) VerifyEmail(ctx context.Context, email string) error {
	log.Printf("=== EMAIL VERIFICATION (Development Mode) ===")
	log.Printf("Provider: %s", p.GetProviderName())
	log.Printf("Verifying email: %s", email)
	log.Printf("Result: Valid (development mode always returns valid)")
	log.Printf("===============================================\n")
	return nil
}

func (p *DevEmailProvider) logEmailInDevelopment(to, subject, body string) error {
	log.Printf("=== EMAIL (Development Mode) ===")
	log.Printf("Provider: %s", p.GetProviderName())
	log.Printf("From: %s <%s>", p.config.FromName, p.config.FromEmail)
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

// Helper methods for template rendering (reuse from SMTP provider logic)

func (p *DevEmailProvider) renderTemplate(templateID string, variables map[string]interface{}) string {
	template, exists := p.getTemplate(templateID)
	if !exists {
		return fmt.Sprintf("Template %s not found", templateID)
	}

	return p.replaceVariables(template, variables)
}

func (p *DevEmailProvider) replaceVariables(content string, variables map[string]interface{}) string {
	result := content
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

func (p *DevEmailProvider) getTemplate(templateID string) (string, bool) {
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

func (p *DevEmailProvider) getTemplateSubject(templateID string) string {
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