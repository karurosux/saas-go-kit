package notificationconstants

// Default configuration values
const (
	DefaultSMTPPort = 587
	DefaultTimeout  = 30 // seconds
)

// Provider names
const (
	ProviderSMTP     = "smtp"
	ProviderSendGrid = "sendgrid"
	ProviderSES      = "ses"
	ProviderTwilio   = "twilio"
	ProviderFCM      = "fcm"
	ProviderAPNS     = "apns"
)

// Email template IDs for common notifications
const (
	TemplateWelcome         = "welcome"
	TemplatePasswordReset   = "password_reset"
	TemplateEmailVerification = "email_verification"
	TemplateTeamInvitation  = "team_invitation"
	TemplatePaymentSucceeded = "payment_succeeded"
	TemplatePaymentFailed   = "payment_failed"
	TemplateTrialEnding     = "trial_ending"
	TemplateLoginAlert      = "login_alert"
	TemplateRoleChanged     = "role_changed"
)

// Content types
const (
	ContentTypeHTML = "text/html; charset=UTF-8"
	ContentTypeText = "text/plain; charset=UTF-8"
)

// Environment variable keys
const (
	EnvEmailProvider = "EMAIL_PROVIDER"
	EnvSMTPHost      = "SMTP_HOST"
	EnvSMTPPort      = "SMTP_PORT"
	EnvSMTPUsername  = "SMTP_USERNAME"
	EnvSMTPPassword  = "SMTP_PASSWORD"
	EnvFromEmail     = "FROM_EMAIL"
	EnvFromName      = "FROM_NAME"
	EnvAppName       = "APP_NAME"
	EnvAppURL        = "APP_URL"
	EnvSupportEmail  = "SUPPORT_EMAIL"
	
	EnvSMSProvider  = "SMS_PROVIDER"
	EnvSMSAPIKey    = "SMS_API_KEY"
	EnvSMSAPISecret = "SMS_API_SECRET"
	EnvSMSFromNumber = "SMS_FROM_NUMBER"
	
	EnvPushProvider  = "PUSH_PROVIDER"
	EnvPushAPIKey    = "PUSH_API_KEY"
	EnvPushProjectID = "PUSH_PROJECT_ID"
)