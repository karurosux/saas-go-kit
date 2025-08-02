package emailinterface

import (
	"context"
	"time"
)

// EmailMessage represents an email to be sent
type EmailMessage struct {
	ID          uint                   `json:"id"`
	To          []string              `json:"to"`
	CC          []string              `json:"cc,omitempty"`
	BCC         []string              `json:"bcc,omitempty"`
	From        string                `json:"from"`
	Subject     string                `json:"subject"`
	Body        string                `json:"body"`
	HTML        string                `json:"html,omitempty"`
	Template    string                `json:"template,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
	Attachments []Attachment          `json:"attachments,omitempty"`
	Priority    EmailPriority         `json:"priority"`
	Status      EmailStatus           `json:"status"`
	Attempts    int                   `json:"attempts"`
	MaxAttempts int                   `json:"max_attempts"`
	SentAt      *time.Time            `json:"sent_at,omitempty"`
	Error       string                `json:"error,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
}

// EmailPriority defines email priority levels
type EmailPriority int

const (
	PriorityLow EmailPriority = iota
	PriorityNormal
	PriorityHigh
)

// EmailStatus defines email sending status
type EmailStatus string

const (
	StatusPending EmailStatus = "pending"
	StatusSending EmailStatus = "sending"
	StatusSent    EmailStatus = "sent"
	StatusFailed  EmailStatus = "failed"
)

// EmailSender defines the interface for sending emails
type EmailSender interface {
	Send(ctx context.Context, message *EmailMessage) error
	SendBatch(ctx context.Context, messages []*EmailMessage) error
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	HTML      string    `json:"html"`
	Variables []string  `json:"variables"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TemplateManager defines the interface for managing email templates
type TemplateManager interface {
	GetTemplate(ctx context.Context, name string) (*EmailTemplate, error)
	CreateTemplate(ctx context.Context, template *EmailTemplate) error
	UpdateTemplate(ctx context.Context, name string, template *EmailTemplate) error
	DeleteTemplate(ctx context.Context, name string) error
	ListTemplates(ctx context.Context) ([]*EmailTemplate, error)
	RenderTemplate(ctx context.Context, name string, data map[string]interface{}) (subject, body, html string, err error)
}

// EmailQueue defines the interface for email queue operations
type EmailQueue interface {
	Enqueue(ctx context.Context, message *EmailMessage) error
	Dequeue(ctx context.Context, limit int) ([]*EmailMessage, error)
	MarkAsSent(ctx context.Context, id uint) error
	MarkAsFailed(ctx context.Context, id uint, err error) error
	RetryFailed(ctx context.Context) error
	GetStatus(ctx context.Context, id uint) (*EmailMessage, error)
}

// TemplateRepository defines the interface for template storage
type TemplateRepository interface {
	GetTemplate(ctx context.Context, name string) (*EmailTemplate, error)
	CreateTemplate(ctx context.Context, template *EmailTemplate) error
	UpdateTemplate(ctx context.Context, name string, template *EmailTemplate) error
	DeleteTemplate(ctx context.Context, name string) error
	ListTemplates(ctx context.Context) ([]*EmailTemplate, error)
}

// EmailService defines the main email service interface
type EmailService interface {
	Send(ctx context.Context, to []string, subject, body string) error
	SendHTML(ctx context.Context, to []string, subject, body, html string) error
	SendTemplate(ctx context.Context, to []string, templateName string, data map[string]interface{}) error
	QueueEmail(ctx context.Context, message *EmailMessage) error
	ProcessQueue(ctx context.Context) error
	GetEmailStatus(ctx context.Context, id uint) (*EmailMessage, error)
}