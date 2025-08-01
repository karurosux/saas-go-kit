package notificationservice

import (
	"context"
	"fmt"
	
	"{{.Project.GoModule}}/internal/notification/interface"
)

// commonNotificationService implements the CommonNotificationService interface
type commonNotificationService struct {
	notificationSvc notificationinterface.NotificationService
	config          notificationinterface.CommonNotificationConfig
}

// NewCommonNotificationService creates a new common notification service
func NewCommonNotificationService(
	notificationSvc notificationinterface.NotificationService, 
	config notificationinterface.CommonNotificationConfig,
) notificationinterface.CommonNotificationService {
	return &commonNotificationService{
		notificationSvc: notificationSvc,
		config:          config,
	}
}

// Auth-related notifications

func (s *commonNotificationService) SendEmailVerification(ctx context.Context, email, token string) error {
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", s.config.AppURL, token)
	
	req := &notificationinterface.EmailRequest{
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

func (s *commonNotificationService) SendPasswordReset(ctx context.Context, email, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.config.AppURL, token)
	
	req := &notificationinterface.EmailRequest{
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

func (s *commonNotificationService) SendLoginAlert(ctx context.Context, email, ipAddress, userAgent string) error {
	req := &notificationinterface.EmailRequest{
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

func (s *commonNotificationService) SendTeamInvitation(ctx context.Context, email, inviterName, teamName, role, token string) error {
	inviteURL := fmt.Sprintf("%s/team/accept-invite?token=%s", s.config.AppURL, token)
	
	req := &notificationinterface.EmailRequest{
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

func (s *commonNotificationService) SendRoleChanged(ctx context.Context, email, userName, teamName, oldRole, newRole string) error {
	req := &notificationinterface.EmailRequest{
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

func (s *commonNotificationService) SendPaymentSucceeded(ctx context.Context, email, planName string, amount float64, currency, invoiceURL string) error {
	req := &notificationinterface.EmailRequest{
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

func (s *commonNotificationService) SendPaymentFailed(ctx context.Context, email, planName string, amount float64, currency string) error {
	req := &notificationinterface.EmailRequest{
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

func (s *commonNotificationService) SendTrialEnding(ctx context.Context, email string, daysLeft int) error {
	req := &notificationinterface.EmailRequest{
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