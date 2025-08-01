package notificationconstants

// Error messages for the notification module
const (
	ErrEmailProviderNotConfigured      = "email provider not configured"
	ErrSMSProviderNotConfigured        = "SMS provider not configured"
	ErrPushProviderNotConfigured       = "push notification provider not configured"
	ErrCommonNotificationNotConfigured = "common notification service not configured"
	
	ErrInvalidEmailFormat     = "invalid email format"
	ErrInvalidPhoneFormat     = "invalid phone number format"
	ErrEmailRequired          = "email is required"
	ErrSubjectRequired        = "subject is required"
	ErrBodyRequired           = "body is required"
	ErrMessageRequired        = "message is required"
	ErrTitleRequired          = "title is required"
	ErrTemplateIDRequired     = "template ID is required"
	ErrRecipientsRequired     = "at least one recipient is required"
	ErrTokensRequired         = "at least one device token is required"
	
	ErrMessageTooLong         = "message too long (max 1600 characters)"
	ErrTemplateNotFound       = "template not found"
	ErrFailedToSendEmail      = "failed to send email"
	ErrFailedToSendSMS        = "failed to send SMS"
	ErrFailedToSendPush       = "failed to send push notification"
	
	ErrSMTPHostRequired       = "SMTP host is required"
	ErrFromEmailRequired      = "from email is required"
	ErrSMTPCredentialsRequired = "SMTP credentials not configured"
)