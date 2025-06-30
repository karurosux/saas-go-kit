package notification

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
	To          []string          `json:"to" validate:"required,min=1,dive,email"`
	CC          []string          `json:"cc,omitempty" validate:"omitempty,dive,email"`
	BCC         []string          `json:"bcc,omitempty" validate:"omitempty,dive,email"`
	TemplateID  string            `json:"template_id" validate:"required"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Subject     string            `json:"subject,omitempty"` // Override template subject
	Headers     map[string]string `json:"headers,omitempty"`
	ReplyTo     string            `json:"reply_to,omitempty" validate:"omitempty,email"`
	Priority    Priority          `json:"priority"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
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
	Token       string            `json:"token" validate:"required"`
	Data        map[string]string `json:"data,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
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
	MessageID     string    `json:"message_id"`
	Status        string    `json:"status"`
	SentAt        time.Time `json:"sent_at"`
	FailedTokens  []string  `json:"failed_tokens,omitempty"`
}

type BulkEmailResponse struct {
	TotalSent   int                   `json:"total_sent"`
	TotalFailed int                   `json:"total_failed"`
	Results     []EmailResponse       `json:"results"`
	Failures    []BulkFailure         `json:"failures,omitempty"`
}

type BulkFailure struct {
	Recipient string `json:"recipient"`
	Error     string `json:"error"`
}

// Predefined templates for common notifications

type EmailTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Subject     string                 `json:"subject"`
	Body        string                 `json:"body"`
	IsHTML      bool                   `json:"is_html"`
	Variables   []TemplateVariable     `json:"variables"`
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type TemplateVariable struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // string, number, boolean, date
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

// Common notification types

type AuthNotification struct {
	Type           AuthNotificationType `json:"type"`
	Email          string               `json:"email"`
	Token          string               `json:"token,omitempty"`
	AppName        string               `json:"app_name"`
	AppURL         string               `json:"app_url"`
	ExpirationTime *time.Time           `json:"expiration_time,omitempty"`
}

type AuthNotificationType string

const (
	AuthEmailVerification AuthNotificationType = "email_verification"
	AuthPasswordReset     AuthNotificationType = "password_reset"
	AuthLoginAlert        AuthNotificationType = "login_alert"
	AuthPasswordChanged   AuthNotificationType = "password_changed"
)

type TeamNotification struct {
	Type        TeamNotificationType `json:"type"`
	Email       string               `json:"email"`
	TeamName    string               `json:"team_name"`
	InviterName string               `json:"inviter_name,omitempty"`
	Role        string               `json:"role,omitempty"`
	Token       string               `json:"token,omitempty"`
	AppURL      string               `json:"app_url"`
}

type TeamNotificationType string

const (
	TeamInvitation   TeamNotificationType = "invitation"
	TeamRoleChanged  TeamNotificationType = "role_changed"
	TeamMemberAdded  TeamNotificationType = "member_added"
	TeamMemberRemoved TeamNotificationType = "member_removed"
)

type BillingNotification struct {
	Type         BillingNotificationType `json:"type"`
	Email        string                  `json:"email"`
	Amount       float64                 `json:"amount,omitempty"`
	Currency     string                  `json:"currency,omitempty"`
	InvoiceURL   string                  `json:"invoice_url,omitempty"`
	DueDate      *time.Time              `json:"due_date,omitempty"`
	PlanName     string                  `json:"plan_name,omitempty"`
}

type BillingNotificationType string

const (
	BillingPaymentSucceeded    BillingNotificationType = "payment_succeeded"
	BillingPaymentFailed       BillingNotificationType = "payment_failed"
	BillingInvoiceCreated      BillingNotificationType = "invoice_created"
	BillingSubscriptionChanged BillingNotificationType = "subscription_changed"
	BillingTrialEnding         BillingNotificationType = "trial_ending"
)