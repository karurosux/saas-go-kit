package notificationinterface

import (
	"context"
	"time"
)

// NotificationService defines the main interface for sending notifications
type NotificationService interface {
	// Email notifications
	SendEmail(ctx context.Context, req *EmailRequest) error
	SendTemplateEmail(ctx context.Context, req *TemplateEmailRequest) error
	
	// SMS notifications (if SMS provider is configured)
	SendSMS(ctx context.Context, req *SMSRequest) error
	
	// Push notifications (if push provider is configured)
	SendPushNotification(ctx context.Context, req *PushNotificationRequest) error
	
	// Bulk notifications
	SendBulkEmails(ctx context.Context, req *BulkEmailRequest) error
	
	// Notification verification
	VerifyEmail(ctx context.Context, email string) error
	VerifyPhoneNumber(ctx context.Context, phone string) error
}

// EmailProvider defines the interface for email providers
type EmailProvider interface {
	Initialize(config EmailConfig) error
	GetProviderName() string
	SendEmail(ctx context.Context, req *EmailRequest) error
	SendTemplateEmail(ctx context.Context, req *TemplateEmailRequest) error
	SendBulkEmails(ctx context.Context, req *BulkEmailRequest) error
	VerifyEmail(ctx context.Context, email string) error
}

// SMSProvider defines the interface for SMS providers
type SMSProvider interface {
	Initialize(config SMSConfig) error
	GetProviderName() string
	SendSMS(ctx context.Context, req *SMSRequest) error
	VerifyPhoneNumber(ctx context.Context, phone string) error
}

// PushProvider defines the interface for push notification providers
type PushProvider interface {
	Initialize(config PushConfig) error
	GetProviderName() string
	SendPushNotification(ctx context.Context, req *PushNotificationRequest) error
	SendBulkPushNotifications(ctx context.Context, req *BulkPushRequest) error
}

// CommonNotificationService provides high-level notification methods for common use cases
type CommonNotificationService interface {
	// Auth-related notifications
	SendEmailVerification(ctx context.Context, email, token string) error
	SendPasswordReset(ctx context.Context, email, token string) error
	SendLoginAlert(ctx context.Context, email, ipAddress, userAgent string) error
	
	// Team-related notifications
	SendTeamInvitation(ctx context.Context, email, inviterName, teamName, role, token string) error
	SendRoleChanged(ctx context.Context, email, userName, teamName, oldRole, newRole string) error
	
	// Billing-related notifications
	SendPaymentSucceeded(ctx context.Context, email, planName string, amount float64, currency, invoiceURL string) error
	SendPaymentFailed(ctx context.Context, email, planName string, amount float64, currency string) error
	SendTrialEnding(ctx context.Context, email string, daysLeft int) error
}

// Request types

type EmailRequest struct {
	To          []string          `json:"to" validate:"required,min=1,dive,email"`
	CC          []string          `json:"cc,omitempty" validate:"omitempty,dive,email"`
	BCC         []string          `json:"bcc,omitempty" validate:"omitempty,dive,email"`
	Subject     string            `json:"subject" validate:"required"`
	Body        string            `json:"body" validate:"required"`
	IsHTML      bool              `json:"is_html"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	ReplyTo     string            `json:"reply_to,omitempty" validate:"omitempty,email"`
	Priority    Priority          `json:"priority"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
}

type TemplateEmailRequest struct {
	To          []string               `json:"to" validate:"required,min=1,dive,email"`
	CC          []string               `json:"cc,omitempty" validate:"omitempty,dive,email"`
	BCC         []string               `json:"bcc,omitempty" validate:"omitempty,dive,email"`
	TemplateID  string                 `json:"template_id" validate:"required"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Subject     string                 `json:"subject,omitempty"` // Override template subject
	Headers     map[string]string      `json:"headers,omitempty"`
	ReplyTo     string                 `json:"reply_to,omitempty" validate:"omitempty,email"`
	Priority    Priority               `json:"priority"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
}

type BulkEmailRequest struct {
	Recipients []BulkEmailRecipient `json:"recipients" validate:"required,min=1"`
	Subject    string               `json:"subject" validate:"required"`
	Body       string               `json:"body" validate:"required"`
	IsHTML     bool                 `json:"is_html"`
	Headers    map[string]string    `json:"headers,omitempty"`
	Priority   Priority             `json:"priority"`
}

type BulkEmailRecipient struct {
	Email     string                 `json:"email" validate:"required,email"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type SMSRequest struct {
	To          []string   `json:"to" validate:"required,min=1"`
	Message     string     `json:"message" validate:"required,max=1600"`
	Priority    Priority   `json:"priority"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

type PushNotificationRequest struct {
	Tokens      []string          `json:"tokens" validate:"required,min=1"`
	Title       string            `json:"title" validate:"required"`
	Body        string            `json:"body" validate:"required"`
	Data        map[string]string `json:"data,omitempty"`
	ImageURL    string            `json:"image_url,omitempty"`
	Sound       string            `json:"sound,omitempty"`
	Badge       int               `json:"badge,omitempty"`
	Priority    Priority          `json:"priority"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
}

type BulkPushRequest struct {
	Recipients []PushRecipient `json:"recipients" validate:"required,min=1"`
	Title      string          `json:"title" validate:"required"`
	Body       string          `json:"body" validate:"required"`
	Priority   Priority        `json:"priority"`
}

type PushRecipient struct {
	Token     string                 `json:"token" validate:"required"`
	Data      map[string]string      `json:"data,omitempty"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// Configuration types

type EmailConfig struct {
	Provider      string                 `json:"provider"`       // smtp, sendgrid, ses, etc.
	Host          string                 `json:"host,omitempty"`
	Port          int                    `json:"port,omitempty"`
	Username      string                 `json:"username,omitempty"`
	Password      string                 `json:"password,omitempty"`
	FromEmail     string                 `json:"from_email"`
	FromName      string                 `json:"from_name,omitempty"`
	APIKey        string                 `json:"api_key,omitempty"`
	Region        string                 `json:"region,omitempty"`
	UseTLS        bool                   `json:"use_tls"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}

type SMSConfig struct {
	Provider      string                 `json:"provider"`       // twilio, aws-sns, etc.
	APIKey        string                 `json:"api_key"`
	APISecret     string                 `json:"api_secret,omitempty"`
	FromNumber    string                 `json:"from_number"`
	Region        string                 `json:"region,omitempty"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}

type PushConfig struct {
	Provider      string                 `json:"provider"`       // fcm, apns, etc.
	APIKey        string                 `json:"api_key"`
	ProjectID     string                 `json:"project_id,omitempty"`
	BundleID      string                 `json:"bundle_id,omitempty"`
	TeamID        string                 `json:"team_id,omitempty"`
	KeyID         string                 `json:"key_id,omitempty"`
	PrivateKey    string                 `json:"private_key,omitempty"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}

// Supporting types

type Attachment struct {
	Filename    string `json:"filename" validate:"required"`
	Content     []byte `json:"content" validate:"required"`
	ContentType string `json:"content_type"`
}

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
)

// Response types

type EmailResponse struct {
	MessageID string    `json:"message_id"`
	Status    string    `json:"status"`
	SentAt    time.Time `json:"sent_at"`
}

type SMSResponse struct {
	MessageID string    `json:"message_id"`
	Status    string    `json:"status"`
	SentAt    time.Time `json:"sent_at"`
}

type PushResponse struct {
	MessageID    string    `json:"message_id"`
	Status       string    `json:"status"`
	SentAt       time.Time `json:"sent_at"`
	FailedTokens []string  `json:"failed_tokens,omitempty"`
}

type BulkEmailResponse struct {
	TotalSent   int           `json:"total_sent"`
	TotalFailed int           `json:"total_failed"`
	Results     []EmailResponse `json:"results"`
	Failures    []BulkFailure   `json:"failures,omitempty"`
}

type BulkFailure struct {
	Recipient string `json:"recipient"`
	Error     string `json:"error"`
}

// Common notification configuration
type CommonNotificationConfig struct {
	AppName      string
	AppURL       string
	FromEmail    string
	FromName     string
	SupportEmail string
}