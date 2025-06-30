# Notification Module

A comprehensive notification system for SaaS applications supporting email, SMS, and push notifications with multiple providers and common notification patterns.

## Features

- **Multi-Channel Support**: Email, SMS, and push notifications
- **Multiple Providers**: Pluggable provider system (SMTP, SendGrid, Twilio, FCM, etc.)
- **Template System**: Built-in template engine with variable substitution
- **Bulk Notifications**: Efficient bulk sending with personalization
- **Common Patterns**: Pre-built templates for auth, team, and billing notifications
- **Development Mode**: Console logging for testing and development
- **Validation**: Request validation and email/phone verification

## Installation

```bash
go get github.com/karurosux/saas-go-kit/notification-go
```

## Quick Start

### 1. Basic Setup

```go
package main

import (
    "github.com/karurosux/saas-go-kit/core-go"
    "github.com/karurosux/saas-go-kit/notification-go"
)

func main() {
    // Initialize email provider
    emailProvider := notification.NewSMTPProvider()
    emailProvider.Initialize(notification.EmailConfig{
        Host:      "smtp.gmail.com",
        Port:      587,
        Username:  "your-email@gmail.com",
        Password:  "your-app-password",
        FromEmail: "noreply@yourapp.com",
        FromName:  "Your App",
        UseTLS:    true,
    })

    // Initialize notification service
    notificationService := notification.NewNotificationService(
        emailProvider,
        nil, // SMS provider (optional)
        nil, // Push provider (optional)
    )

    // Initialize common notification service
    commonService := notification.NewCommonNotificationService(
        notificationService,
        notification.CommonNotificationConfig{
            AppName:      "Your SaaS App",
            AppURL:       "https://yourapp.com",
            FromEmail:    "noreply@yourapp.com",
            FromName:     "Your App",
            SupportEmail: "support@yourapp.com",
        },
    )

    // Create and mount module
    module := notification.NewModule(notification.ModuleConfig{
        NotificationService: notificationService,
        CommonService:       commonService,
        RoutePrefix:         "/api/notifications",
        EnableTestEndpoints: true, // Enable for development
        RequireAuth:         true,
    })

    app := core.NewKit(nil, core.KitConfig{})
    app.Register(module)
    app.Mount()
}
```

### 2. Development Setup

For development, use the dev SMTP provider that logs emails to console:

```go
// Development setup
emailProvider := notification.NewDevSMTPProvider()
emailProvider.Initialize(notification.EmailConfig{
    FromEmail: "dev@yourapp.com",
    FromName:  "Your App (Dev)",
})
```

### 3. Send Basic Email

```go
req := &notification.EmailRequest{
    To:      []string{"user@example.com"},
    Subject: "Welcome to Our App",
    Body:    "<h1>Welcome!</h1><p>Thanks for joining us.</p>",
    IsHTML:  true,
}

err := notificationService.SendEmail(ctx, req)
if err != nil {
    log.Printf("Failed to send email: %v", err)
}
```

## Email Providers

### SMTP Provider

For standard SMTP servers (Gmail, Outlook, custom servers):

```go
emailProvider := notification.NewSMTPProvider()
emailProvider.Initialize(notification.EmailConfig{
    Host:      "smtp.gmail.com",
    Port:      587,
    Username:  "your-email@gmail.com",
    Password:  "your-app-password", // Use app password for Gmail
    FromEmail: "noreply@yourapp.com",
    FromName:  "Your App",
    UseTLS:    true,
})
```

### Development Provider

For development and testing:

```go
emailProvider := notification.NewDevSMTPProvider()
emailProvider.Initialize(notification.EmailConfig{
    FromEmail: "dev@yourapp.com",
    FromName:  "Your App (Dev)",
})
```

This will log all emails to the console with clickable links for easy testing.

## Common Notification Patterns

The module includes pre-built notification patterns for common SaaS use cases:

### Authentication Notifications

```go
// Email verification
err := commonService.SendEmailVerification(ctx, "user@example.com", "verification-token")

// Password reset
err := commonService.SendPasswordReset(ctx, "user@example.com", "reset-token")

// Login alert
err := commonService.SendLoginAlert(ctx, "user@example.com", "192.168.1.1", "Chrome/91.0")
```

### Team Notifications

```go
// Team invitation
err := commonService.SendTeamInvitation(ctx, 
    "newuser@example.com", 
    "John Doe", 
    "Acme Corp", 
    "Manager", 
    "invitation-token")

// Role changed
err := commonService.SendRoleChanged(ctx, 
    "user@example.com", 
    "Jane Smith", 
    "Acme Corp", 
    "Viewer", 
    "Admin")
```

### Billing Notifications

```go
// Payment succeeded
err := commonService.SendPaymentSucceeded(ctx, 
    "user@example.com", 
    "Pro Plan", 
    29.99, 
    "USD", 
    "https://app.com/invoice/123")

// Payment failed
err := commonService.SendPaymentFailed(ctx, 
    "user@example.com", 
    "Pro Plan", 
    29.99, 
    "USD")

// Trial ending
err := commonService.SendTrialEnding(ctx, "user@example.com", 3) // 3 days left
```

## Template System

### Built-in Templates

The SMTP provider includes several built-in templates:

- `welcome` - Welcome new users
- `password_reset` - Password reset emails
- `team_invitation` - Team invitation emails

### Using Templates

```go
req := &notification.TemplateEmailRequest{
    To:         []string{"user@example.com"},
    TemplateID: "welcome",
    Variables: map[string]interface{}{
        "user_name": "John Doe",
        "app_name":  "Your SaaS App",
        "app_url":   "https://yourapp.com",
    },
}

err := notificationService.SendTemplateEmail(ctx, req)
```

### Custom Templates

Templates use `{{variable}}` syntax for variable substitution:

```go
template := `
<html>
<body>
    <h2>Hello {{user_name}}!</h2>
    <p>Welcome to {{app_name}}. Click <a href="{{app_url}}">here</a> to get started.</p>
</body>
</html>`
```

## Bulk Notifications

Send personalized emails to multiple recipients:

```go
req := &notification.BulkEmailRequest{
    Recipients: []notification.BulkEmailRecipient{
        {
            Email: "user1@example.com",
            Variables: map[string]interface{}{
                "name": "John",
                "plan": "Pro",
            },
        },
        {
            Email: "user2@example.com",
            Variables: map[string]interface{}{
                "name": "Jane",
                "plan": "Enterprise",
            },
        },
    },
    Subject: "Your {{plan}} plan is ready!",
    Body:    "<h2>Hi {{name}}</h2><p>Your {{plan}} plan is now active.</p>",
    IsHTML:  true,
}

err := notificationService.SendBulkEmails(ctx, req)
```

## API Endpoints

### Core Notification Endpoints

- `POST /notifications/email` - Send single email
- `POST /notifications/email/template` - Send template email
- `POST /notifications/email/bulk` - Send bulk emails
- `POST /notifications/sms` - Send SMS (if SMS provider configured)
- `POST /notifications/push` - Send push notification (if push provider configured)

### Verification Endpoints

- `GET /notifications/verify/email?email=user@example.com` - Verify email format
- `GET /notifications/verify/phone?phone=+1234567890` - Verify phone number

### Common Pattern Endpoints

- `POST /notifications/common/auth/email-verification` - Send email verification
- `POST /notifications/common/auth/password-reset` - Send password reset
- `POST /notifications/common/team/invitation` - Send team invitation
- `POST /notifications/common/team/role-changed` - Send role change notification
- `POST /notifications/common/billing/payment-succeeded` - Send payment success
- `POST /notifications/common/billing/payment-failed` - Send payment failure
- `POST /notifications/common/billing/trial-ending` - Send trial ending notice

### Test Endpoints (Development Only)

- `POST /notifications/test/email` - Send test email

## Configuration

### Email Configuration

```go
type EmailConfig struct {
    Provider  string // "smtp", "sendgrid", "ses", etc.
    Host      string // SMTP host
    Port      int    // SMTP port (default: 587)
    Username  string // SMTP username
    Password  string // SMTP password
    FromEmail string // From email address
    FromName  string // From name (optional)
    APIKey    string // API key for cloud providers
    Region    string // Region for cloud providers
    UseTLS    bool   // Use TLS encryption
}
```

### Module Configuration

```go
type ModuleConfig struct {
    NotificationService NotificationService
    CommonService       *CommonNotificationService
    RoutePrefix         string // Default: "/notifications"
    EnableTestEndpoints bool   // Enable test endpoints
    RequireAuth         bool   // Require authentication
}
```

### Common Service Configuration

```go
type CommonNotificationConfig struct {
    AppName      string // Your application name
    AppURL       string // Your application URL
    FromEmail    string // Default from email
    FromName     string // Default from name
    SupportEmail string // Support email for help links
}
```

## Provider Interface

Implement custom providers by satisfying the provider interfaces:

### Email Provider

```go
type EmailProvider interface {
    Initialize(config EmailConfig) error
    GetProviderName() string
    SendEmail(ctx context.Context, req *EmailRequest) error
    SendTemplateEmail(ctx context.Context, req *TemplateEmailRequest) error
    SendBulkEmails(ctx context.Context, req *BulkEmailRequest) error
    VerifyEmail(ctx context.Context, email string) error
}
```

### SMS Provider

```go
type SMSProvider interface {
    Initialize(config SMSConfig) error
    GetProviderName() string
    SendSMS(ctx context.Context, req *SMSRequest) error
    VerifyPhoneNumber(ctx context.Context, phone string) error
}
```

### Push Provider

```go
type PushProvider interface {
    Initialize(config PushConfig) error
    GetProviderName() string
    SendPushNotification(ctx context.Context, req *PushNotificationRequest) error
    SendBulkPushNotifications(ctx context.Context, req *BulkPushRequest) error
}
```

## Error Handling

The module uses structured error responses:

```go
// Email validation error
{
    "success": false,
    "error": {
        "code": "BAD_REQUEST",
        "message": "Invalid email format"
    }
}

// Provider not configured error
{
    "success": false,
    "error": {
        "code": "INTERNAL_ERROR",
        "message": "Email provider not configured"
    }
}
```

## Development and Testing

### Console Logging

In development mode, emails are logged to the console with clickable links:

```
=== EMAIL (Development Mode) ===
To: user@example.com
Subject: Welcome to Your App
Body: <html>...</html>

🔗 CLICKABLE LINKS:
   👉 https://yourapp.com/verify-email?token=abc123
================================
```

### Test Endpoint

Use the test endpoint to verify your email configuration:

```bash
curl -X POST http://localhost:8080/api/notifications/test/email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "test@example.com",
    "subject": "Test Email",
    "body": "<h1>Hello!</h1><p>This is a test.</p>"
  }'
```

## Security Considerations

- Always validate email addresses and phone numbers
- Use environment variables for sensitive configuration (API keys, passwords)
- Implement rate limiting to prevent abuse
- Sanitize user input in email templates
- Use app passwords for Gmail instead of account passwords

## Environment Variables

```env
# SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=noreply@yourapp.com
SMTP_FROM_NAME=Your App
SMTP_USE_TLS=true

# Application Configuration
APP_NAME=Your SaaS App
APP_URL=https://yourapp.com
SUPPORT_EMAIL=support@yourapp.com

# Development
NOTIFICATION_DEV_MODE=true
```

## Dependencies

- `github.com/karurosux/saas-go-kit/core-go` - Core module system
- `github.com/karurosux/saas-go-kit/errors-go` - Error handling
- `github.com/labstack/echo/v4` - HTTP framework

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.