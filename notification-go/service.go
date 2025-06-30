package notification

import (
	"context"
	"fmt"

	"github.com/karurosux/saas-go-kit/errors-go"
)

type notificationService struct {
	emailProvider EmailProvider
	smsProvider   SMSProvider
	pushProvider  PushProvider
}

func NewNotificationService(
	emailProvider EmailProvider,
	smsProvider SMSProvider,
	pushProvider PushProvider,
) NotificationService {
	return &notificationService{
		emailProvider: emailProvider,
		smsProvider:   smsProvider,
		pushProvider:  pushProvider,
	}
}

func (s *notificationService) SendEmail(ctx context.Context, req *EmailRequest) error {
	if s.emailProvider == nil {
		return errors.Internal("Email provider not configured")
	}

	if err := s.validateEmailRequest(req); err != nil {
		return err
	}

	return s.emailProvider.SendEmail(ctx, req)
}

func (s *notificationService) SendTemplateEmail(ctx context.Context, req *TemplateEmailRequest) error {
	if s.emailProvider == nil {
		return errors.Internal("Email provider not configured")
	}

	if err := s.validateTemplateEmailRequest(req); err != nil {
		return err
	}

	return s.emailProvider.SendTemplateEmail(ctx, req)
}

func (s *notificationService) SendSMS(ctx context.Context, req *SMSRequest) error {
	if s.smsProvider == nil {
		return errors.Internal("SMS provider not configured")
	}

	if err := s.validateSMSRequest(req); err != nil {
		return err
	}

	return s.smsProvider.SendSMS(ctx, req)
}

func (s *notificationService) SendPushNotification(ctx context.Context, req *PushNotificationRequest) error {
	if s.pushProvider == nil {
		return errors.Internal("Push notification provider not configured")
	}

	if err := s.validatePushRequest(req); err != nil {
		return err
	}

	return s.pushProvider.SendPushNotification(ctx, req)
}

func (s *notificationService) SendBulkEmails(ctx context.Context, req *BulkEmailRequest) error {
	if s.emailProvider == nil {
		return errors.Internal("Email provider not configured")
	}

	if err := s.validateBulkEmailRequest(req); err != nil {
		return err
	}

	return s.emailProvider.SendBulkEmails(ctx, req)
}

func (s *notificationService) VerifyEmail(ctx context.Context, email string) error {
	if s.emailProvider == nil {
		return errors.Internal("Email provider not configured")
	}

	return s.emailProvider.VerifyEmail(ctx, email)
}

func (s *notificationService) VerifyPhoneNumber(ctx context.Context, phone string) error {
	if s.smsProvider == nil {
		return errors.Internal("SMS provider not configured")
	}

	return s.smsProvider.VerifyPhoneNumber(ctx, phone)
}

// Validation methods

func (s *notificationService) validateEmailRequest(req *EmailRequest) error {
	if len(req.To) == 0 {
		return errors.BadRequest("At least one recipient is required")
	}

	if req.Subject == "" {
		return errors.BadRequest("Subject is required")
	}

	if req.Body == "" {
		return errors.BadRequest("Body is required")
	}

	return nil
}

func (s *notificationService) validateTemplateEmailRequest(req *TemplateEmailRequest) error {
	if len(req.To) == 0 {
		return errors.BadRequest("At least one recipient is required")
	}

	if req.TemplateID == "" {
		return errors.BadRequest("Template ID is required")
	}

	return nil
}

func (s *notificationService) validateSMSRequest(req *SMSRequest) error {
	if len(req.To) == 0 {
		return errors.BadRequest("At least one recipient is required")
	}

	if req.Message == "" {
		return errors.BadRequest("Message is required")
	}

	if len(req.Message) > 1600 {
		return errors.BadRequest("Message too long (max 1600 characters)")
	}

	return nil
}

func (s *notificationService) validatePushRequest(req *PushNotificationRequest) error {
	if len(req.Tokens) == 0 {
		return errors.BadRequest("At least one device token is required")
	}

	if req.Title == "" {
		return errors.BadRequest("Title is required")
	}

	if req.Body == "" {
		return errors.BadRequest("Body is required")
	}

	return nil
}

func (s *notificationService) validateBulkEmailRequest(req *BulkEmailRequest) error {
	if len(req.Recipients) == 0 {
		return errors.BadRequest("At least one recipient is required")
	}

	if req.Subject == "" {
		return errors.BadRequest("Subject is required")
	}

	if req.Body == "" {
		return errors.BadRequest("Body is required")
	}

	return nil
}

// Helper service for common notification patterns

type CommonNotificationService struct {
	notificationSvc NotificationService
	config          CommonNotificationConfig
}

type CommonNotificationConfig struct {
	AppName     string
	AppURL      string
	FromEmail   string
	FromName    string
	SupportEmail string
}

func NewCommonNotificationService(notificationSvc NotificationService, config CommonNotificationConfig) *CommonNotificationService {
	return &CommonNotificationService{
		notificationSvc: notificationSvc,
		config:          config,
	}
}

// Auth-related notifications

func (s *CommonNotificationService) SendEmailVerification(ctx context.Context, email, token string) error {
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", s.config.AppURL, token)
	
	req := &EmailRequest{
		To:      []string{email},
		Subject: fmt.Sprintf("Verify Your Email - %s", s.config.AppName),
		Body: fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome to %s!</h2>
			<p>Please click the link below to verify your email address:</p>
			<p><a href="%s" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Verify Email</a></p>
			<p>If you didn't create an account, please ignore this email.</p>
			<p>This link will expire in 24 hours.</p>
			<hr>
			<p><small>If you have any questions, contact us at %s</small></p>
		</body>
		</html>
		`, s.config.AppName, verificationURL, s.config.SupportEmail),
		IsHTML: true,
	}

	return s.notificationSvc.SendEmail(ctx, req)
}

func (s *CommonNotificationService) SendPasswordReset(ctx context.Context, email, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.config.AppURL, token)
	
	req := &EmailRequest{
		To:      []string{email},
		Subject: fmt.Sprintf("Reset Your Password - %s", s.config.AppName),
		Body: fmt.Sprintf(`
		<html>
		<body>
			<h2>Password Reset Request</h2>
			<p>You requested to reset your password for %s. Click the link below to set a new password:</p>
			<p><a href="%s" style="background-color: #dc3545; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Reset Password</a></p>
			<p>If you didn't request this, please ignore this email. Your password will remain unchanged.</p>
			<p>This link will expire in 1 hour.</p>
			<hr>
			<p><small>If you have any questions, contact us at %s</small></p>
		</body>
		</html>
		`, s.config.AppName, resetURL, s.config.SupportEmail),
		IsHTML: true,
	}

	return s.notificationSvc.SendEmail(ctx, req)
}

func (s *CommonNotificationService) SendLoginAlert(ctx context.Context, email, ipAddress, userAgent string) error {
	req := &EmailRequest{
		To:      []string{email},
		Subject: fmt.Sprintf("New Login to Your %s Account", s.config.AppName),
		Body: fmt.Sprintf(`
		<html>
		<body>
			<h2>New Login Detected</h2>
			<p>We detected a new login to your %s account:</p>
			<ul>
				<li><strong>IP Address:</strong> %s</li>
				<li><strong>Device/Browser:</strong> %s</li>
				<li><strong>Time:</strong> Just now</li>
			</ul>
			<p>If this was you, no action is needed.</p>
			<p>If you didn't authorize this login, please:</p>
			<ol>
				<li>Change your password immediately</li>
				<li>Review your account activity</li>
				<li>Contact our support team</li>
			</ol>
			<hr>
			<p><small>If you have any concerns, contact us at %s</small></p>
		</body>
		</html>
		`, s.config.AppName, ipAddress, userAgent, s.config.SupportEmail),
		IsHTML: true,
	}

	return s.notificationSvc.SendEmail(ctx, req)
}

// Team-related notifications

func (s *CommonNotificationService) SendTeamInvitation(ctx context.Context, email, inviterName, teamName, role, token string) error {
	inviteURL := fmt.Sprintf("%s/team/accept-invite?token=%s", s.config.AppURL, token)
	
	req := &EmailRequest{
		To:      []string{email},
		Subject: fmt.Sprintf("You've been invited to join %s", teamName),
		Body: fmt.Sprintf(`
		<html>
		<body>
			<h2>Team Invitation</h2>
			<p>%s has invited you to join the team "%s" on %s as a %s.</p>
			<p><a href="%s" style="background-color: #28a745; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Accept Invitation</a></p>
			<p>If you don't recognize this invitation, please ignore this email.</p>
			<p>This invitation will expire in 7 days.</p>
			<hr>
			<p><small>If you have any questions, contact us at %s</small></p>
		</body>
		</html>
		`, inviterName, teamName, s.config.AppName, role, inviteURL, s.config.SupportEmail),
		IsHTML: true,
	}

	return s.notificationSvc.SendEmail(ctx, req)
}

func (s *CommonNotificationService) SendRoleChanged(ctx context.Context, email, userName, teamName, oldRole, newRole string) error {
	req := &EmailRequest{
		To:      []string{email},
		Subject: fmt.Sprintf("Your role has been updated in %s", teamName),
		Body: fmt.Sprintf(`
		<html>
		<body>
			<h2>Role Update</h2>
			<p>Hi %s,</p>
			<p>Your role in the team "%s" has been updated:</p>
			<ul>
				<li><strong>Previous role:</strong> %s</li>
				<li><strong>New role:</strong> %s</li>
			</ul>
			<p>Your new permissions are now active. <a href="%s">Login to your account</a> to see what you can now access.</p>
			<hr>
			<p><small>If you have any questions, contact us at %s</small></p>
		</body>
		</html>
		`, userName, teamName, oldRole, newRole, s.config.AppURL, s.config.SupportEmail),
		IsHTML: true,
	}

	return s.notificationSvc.SendEmail(ctx, req)
}

// Billing-related notifications

func (s *CommonNotificationService) SendPaymentSucceeded(ctx context.Context, email, planName string, amount float64, currency, invoiceURL string) error {
	req := &EmailRequest{
		To:      []string{email},
		Subject: fmt.Sprintf("Payment Received - %s", s.config.AppName),
		Body: fmt.Sprintf(`
		<html>
		<body>
			<h2>Payment Successful</h2>
			<p>Thank you! We've successfully processed your payment:</p>
			<ul>
				<li><strong>Plan:</strong> %s</li>
				<li><strong>Amount:</strong> %.2f %s</li>
			</ul>
			<p>Your subscription is now active and you have full access to all features.</p>
			<p><a href="%s" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">View Invoice</a></p>
			<hr>
			<p><small>If you have any questions, contact us at %s</small></p>
		</body>
		</html>
		`, planName, amount, currency, invoiceURL, s.config.SupportEmail),
		IsHTML: true,
	}

	return s.notificationSvc.SendEmail(ctx, req)
}

func (s *CommonNotificationService) SendPaymentFailed(ctx context.Context, email, planName string, amount float64, currency string) error {
	req := &EmailRequest{
		To:      []string{email},
		Subject: fmt.Sprintf("Payment Failed - %s", s.config.AppName),
		Body: fmt.Sprintf(`
		<html>
		<body>
			<h2>Payment Failed</h2>
			<p>We were unable to process your payment:</p>
			<ul>
				<li><strong>Plan:</strong> %s</li>
				<li><strong>Amount:</strong> %.2f %s</li>
			</ul>
			<p>Please update your payment method to continue using %s without interruption.</p>
			<p><a href="%s/billing" style="background-color: #dc3545; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Update Payment Method</a></p>
			<p>Common reasons for payment failures:</p>
			<ul>
				<li>Expired credit card</li>
				<li>Insufficient funds</li>
				<li>Bank declined the transaction</li>
			</ul>
			<hr>
			<p><small>If you need help, contact us at %s</small></p>
		</body>
		</html>
		`, planName, amount, currency, s.config.AppName, s.config.AppURL, s.config.SupportEmail),
		IsHTML: true,
	}

	return s.notificationSvc.SendEmail(ctx, req)
}

func (s *CommonNotificationService) SendTrialEnding(ctx context.Context, email string, daysLeft int) error {
	req := &EmailRequest{
		To:      []string{email},
		Subject: fmt.Sprintf("Your %s trial ends in %d days", s.config.AppName, daysLeft),
		Body: fmt.Sprintf(`
		<html>
		<body>
			<h2>Trial Ending Soon</h2>
			<p>Your free trial of %s will end in %d days.</p>
			<p>To continue enjoying all the features, please choose a subscription plan:</p>
			<p><a href="%s/billing" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Choose a Plan</a></p>
			<p>What happens after your trial ends:</p>
			<ul>
				<li>Your account will be limited to free features</li>
				<li>You'll lose access to premium functionality</li>
				<li>Your data will be preserved for 30 days</li>
			</ul>
			<hr>
			<p><small>Questions? Contact us at %s</small></p>
		</body>
		</html>
		`, s.config.AppName, daysLeft, s.config.AppURL, s.config.SupportEmail),
		IsHTML: true,
	}

	return s.notificationSvc.SendEmail(ctx, req)
}