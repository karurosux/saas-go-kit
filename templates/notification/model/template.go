package notificationmodel

import (
	"time"
	
	"{{.Project.GoModule}}/internal/notification/interface"
)

// EmailTemplate represents an email template in the system
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

// TemplateVariable represents a variable that can be used in templates
type TemplateVariable struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // string, number, boolean, date
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

// Predefined notification types for common SaaS scenarios

// AuthNotification represents authentication-related notifications
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

// TeamNotification represents team-related notifications
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
	TeamInvitation    TeamNotificationType = "invitation"
	TeamRoleChanged   TeamNotificationType = "role_changed"
	TeamMemberAdded   TeamNotificationType = "member_added"
	TeamMemberRemoved TeamNotificationType = "member_removed"
)

// BillingNotification represents billing-related notifications
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

// CreateEmailTemplateRequest represents the request to create an email template
type CreateEmailTemplateRequest struct {
	Name        string                 `json:"name" validate:"required"`
	Subject     string                 `json:"subject" validate:"required"`
	Body        string                 `json:"body" validate:"required"`
	IsHTML      bool                   `json:"is_html"`
	Variables   []TemplateVariable     `json:"variables,omitempty"`
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
}

// UpdateEmailTemplateRequest represents the request to update an email template
type UpdateEmailTemplateRequest struct {
	Name        *string                `json:"name,omitempty"`
	Subject     *string                `json:"subject,omitempty"`
	Body        *string                `json:"body,omitempty"`
	IsHTML      *bool                  `json:"is_html,omitempty"`
	Variables   *[]TemplateVariable    `json:"variables,omitempty"`
	Category    *string                `json:"category,omitempty"`
	Description *string                `json:"description,omitempty"`
}

// TestEmailRequest represents a request to send a test email
type TestEmailRequest struct {
	To      string `json:"to" validate:"required,email"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// Common notification request types with validation

type EmailVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
	Token string `json:"token" validate:"required"`
}

type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
	Token string `json:"token" validate:"required"`
}

type TeamInvitationRequest struct {
	Email       string `json:"email" validate:"required,email"`
	InviterName string `json:"inviter_name" validate:"required"`
	TeamName    string `json:"team_name" validate:"required"`
	Role        string `json:"role" validate:"required"`
	Token       string `json:"token" validate:"required"`
}

type RoleChangedRequest struct {
	Email    string `json:"email" validate:"required,email"`
	UserName string `json:"user_name" validate:"required"`
	TeamName string `json:"team_name" validate:"required"`
	OldRole  string `json:"old_role" validate:"required"`
	NewRole  string `json:"new_role" validate:"required"`
}

type PaymentSucceededRequest struct {
	Email      string  `json:"email" validate:"required,email"`
	PlanName   string  `json:"plan_name" validate:"required"`
	Amount     float64 `json:"amount" validate:"required,min=0"`
	Currency   string  `json:"currency" validate:"required"`
	InvoiceURL string  `json:"invoice_url" validate:"required,url"`
}

type PaymentFailedRequest struct {
	Email    string  `json:"email" validate:"required,email"`
	PlanName string  `json:"plan_name" validate:"required"`
	Amount   float64 `json:"amount" validate:"required,min=0"`
	Currency string  `json:"currency" validate:"required"`
}

type TrialEndingRequest struct {
	Email    string `json:"email" validate:"required,email"`
	DaysLeft int    `json:"days_left" validate:"required,min=1"`
}

// Implement interface methods for model types

func (e *EmailTemplate) GetID() string {
	return e.ID
}

func (e *EmailTemplate) GetName() string {
	return e.Name
}

func (e *EmailTemplate) GetSubject() string {
	return e.Subject
}

func (e *EmailTemplate) GetBody() string {
	return e.Body
}

func (e *EmailTemplate) IsHTMLTemplate() bool {
	return e.IsHTML
}

func (e *EmailTemplate) GetVariables() []TemplateVariable {
	return e.Variables
}

func (e *EmailTemplate) GetCategory() string {
	return e.Category
}