package emailinterface

import (
	"context"
	"time"
)

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

type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
}

type EmailPriority int

const (
	PriorityLow EmailPriority = iota
	PriorityNormal
	PriorityHigh
)

type EmailStatus string

const (
	StatusPending EmailStatus = "pending"
	StatusSending EmailStatus = "sending"
	StatusSent    EmailStatus = "sent"
	StatusFailed  EmailStatus = "failed"
)

type EmailSender interface {
	Send(ctx context.Context, message *EmailMessage) error
	SendBatch(ctx context.Context, messages []*EmailMessage) error
}

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

type TemplateManager interface {
	GetTemplate(ctx context.Context, name string) (*EmailTemplate, error)
	CreateTemplate(ctx context.Context, template *EmailTemplate) error
	UpdateTemplate(ctx context.Context, name string, template *EmailTemplate) error
	DeleteTemplate(ctx context.Context, name string) error
	ListTemplates(ctx context.Context) ([]*EmailTemplate, error)
	RenderTemplate(ctx context.Context, name string, data map[string]interface{}) (subject, body, html string, err error)
}

type EmailQueue interface {
	Enqueue(ctx context.Context, message *EmailMessage) error
	Dequeue(ctx context.Context, limit int) ([]*EmailMessage, error)
	MarkAsSent(ctx context.Context, id uint) error
	MarkAsFailed(ctx context.Context, id uint, err error) error
	RetryFailed(ctx context.Context) error
	GetStatus(ctx context.Context, id uint) (*EmailMessage, error)
}

type TemplateRepository interface {
	GetTemplate(ctx context.Context, name string) (*EmailTemplate, error)
	CreateTemplate(ctx context.Context, template *EmailTemplate) error
	UpdateTemplate(ctx context.Context, name string, template *EmailTemplate) error
	DeleteTemplate(ctx context.Context, name string) error
	ListTemplates(ctx context.Context) ([]*EmailTemplate, error)
}

type EmailService interface {
	Send(ctx context.Context, to []string, subject, body string) error
	SendHTML(ctx context.Context, to []string, subject, body, html string) error
	SendTemplate(ctx context.Context, to []string, templateName string, data map[string]interface{}) error
	QueueEmail(ctx context.Context, message *EmailMessage) error
	ProcessQueue(ctx context.Context) error
	GetEmailStatus(ctx context.Context, id uint) (*EmailMessage, error)
}