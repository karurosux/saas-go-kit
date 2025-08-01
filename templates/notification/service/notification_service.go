package notificationservice

import (
	"context"
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/notification/constants"
	"{{.Project.GoModule}}/internal/notification/interface"
)

// notificationService implements the NotificationService interface
type notificationService struct {
	emailProvider notificationinterface.EmailProvider
	smsProvider   notificationinterface.SMSProvider
	pushProvider  notificationinterface.PushProvider
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	emailProvider notificationinterface.EmailProvider,
	smsProvider notificationinterface.SMSProvider,
	pushProvider notificationinterface.PushProvider,
) notificationinterface.NotificationService {
	return &notificationService{
		emailProvider: emailProvider,
		smsProvider:   smsProvider,
		pushProvider:  pushProvider,
	}
}

// Email methods

func (s *notificationService) SendEmail(ctx context.Context, req *notificationinterface.EmailRequest) error {
	if s.emailProvider == nil {
		return core.Internal(notificationconstants.ErrEmailProviderNotConfigured)
	}

	if err := s.validateEmailRequest(req); err != nil {
		return err
	}

	return s.emailProvider.SendEmail(ctx, req)
}

func (s *notificationService) SendTemplateEmail(ctx context.Context, req *notificationinterface.TemplateEmailRequest) error {
	if s.emailProvider == nil {
		return core.Internal(notificationconstants.ErrEmailProviderNotConfigured)
	}

	if err := s.validateTemplateEmailRequest(req); err != nil {
		return err
	}

	return s.emailProvider.SendTemplateEmail(ctx, req)
}

func (s *notificationService) SendBulkEmails(ctx context.Context, req *notificationinterface.BulkEmailRequest) error {
	if s.emailProvider == nil {
		return core.Internal(notificationconstants.ErrEmailProviderNotConfigured)
	}

	if err := s.validateBulkEmailRequest(req); err != nil {
		return err
	}

	return s.emailProvider.SendBulkEmails(ctx, req)
}

func (s *notificationService) VerifyEmail(ctx context.Context, email string) error {
	if s.emailProvider == nil {
		return core.Internal(notificationconstants.ErrEmailProviderNotConfigured)
	}

	return s.emailProvider.VerifyEmail(ctx, email)
}

// SMS methods

func (s *notificationService) SendSMS(ctx context.Context, req *notificationinterface.SMSRequest) error {
	if s.smsProvider == nil {
		return core.Internal(notificationconstants.ErrSMSProviderNotConfigured)
	}

	if err := s.validateSMSRequest(req); err != nil {
		return err
	}

	return s.smsProvider.SendSMS(ctx, req)
}

func (s *notificationService) VerifyPhoneNumber(ctx context.Context, phone string) error {
	if s.smsProvider == nil {
		return core.Internal(notificationconstants.ErrSMSProviderNotConfigured)
	}

	return s.smsProvider.VerifyPhoneNumber(ctx, phone)
}

// Push notification methods

func (s *notificationService) SendPushNotification(ctx context.Context, req *notificationinterface.PushNotificationRequest) error {
	if s.pushProvider == nil {
		return core.Internal(notificationconstants.ErrPushProviderNotConfigured)
	}

	if err := s.validatePushRequest(req); err != nil {
		return err
	}

	return s.pushProvider.SendPushNotification(ctx, req)
}

// Validation methods

func (s *notificationService) validateEmailRequest(req *notificationinterface.EmailRequest) error {
	if len(req.To) == 0 {
		return core.BadRequest(notificationconstants.ErrRecipientsRequired)
	}

	if req.Subject == "" {
		return core.BadRequest(notificationconstants.ErrSubjectRequired)
	}

	if req.Body == "" {
		return core.BadRequest(notificationconstants.ErrBodyRequired)
	}

	return nil
}

func (s *notificationService) validateTemplateEmailRequest(req *notificationinterface.TemplateEmailRequest) error {
	if len(req.To) == 0 {
		return core.BadRequest(notificationconstants.ErrRecipientsRequired)
	}

	if req.TemplateID == "" {
		return core.BadRequest(notificationconstants.ErrTemplateIDRequired)
	}

	return nil
}

func (s *notificationService) validateSMSRequest(req *notificationinterface.SMSRequest) error {
	if len(req.To) == 0 {
		return core.BadRequest(notificationconstants.ErrRecipientsRequired)
	}

	if req.Message == "" {
		return core.BadRequest(notificationconstants.ErrMessageRequired)
	}

	if len(req.Message) > 1600 {
		return core.BadRequest(notificationconstants.ErrMessageTooLong)
	}

	return nil
}

func (s *notificationService) validatePushRequest(req *notificationinterface.PushNotificationRequest) error {
	if len(req.Tokens) == 0 {
		return core.BadRequest(notificationconstants.ErrTokensRequired)
	}

	if req.Title == "" {
		return core.BadRequest(notificationconstants.ErrTitleRequired)
	}

	if req.Body == "" {
		return core.BadRequest(notificationconstants.ErrBodyRequired)
	}

	return nil
}

func (s *notificationService) validateBulkEmailRequest(req *notificationinterface.BulkEmailRequest) error {
	if len(req.Recipients) == 0 {
		return core.BadRequest(notificationconstants.ErrRecipientsRequired)
	}

	if req.Subject == "" {
		return core.BadRequest(notificationconstants.ErrSubjectRequired)
	}

	if req.Body == "" {
		return core.BadRequest(notificationconstants.ErrBodyRequired)
	}

	return nil
}