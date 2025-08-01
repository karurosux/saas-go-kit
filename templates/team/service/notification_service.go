package teamservice

import (
	"context"
	"fmt"
	"time"
	
	"{{.Project.GoModule}}/internal/team/interface"
)

// NotificationService implements team notifications
type NotificationService struct {
	emailSender teaminterface.EmailSender
	baseURL     string
}

// NewNotificationService creates a new notification service
func NewNotificationService(emailSender teaminterface.EmailSender, baseURL string) teaminterface.NotificationService {
	return &NotificationService{
		emailSender: emailSender,
		baseURL:     baseURL,
	}
}

// SendTeamInvitation sends a team invitation notification
func (s *NotificationService) SendTeamInvitation(ctx context.Context, email, inviterName, teamName, role, token string, expiresAt time.Time) error {
	inviteLink := fmt.Sprintf("%s/invitations/accept?token=%s", s.baseURL, token)
	return s.emailSender.SendInvitationEmail(email, inviterName, teamName, role, inviteLink, expiresAt)
}

// SendRoleChanged sends a role change notification
func (s *NotificationService) SendRoleChanged(ctx context.Context, email, userName, teamName, oldRole, newRole, changedBy string) error {
	return s.emailSender.SendRoleChangedEmail(email, userName, teamName, oldRole, newRole, changedBy)
}

// SendMemberRemoved sends a member removed notification
func (s *NotificationService) SendMemberRemoved(ctx context.Context, email, userName, teamName, removedBy string) error {
	return s.emailSender.SendMemberRemovedEmail(email, userName, teamName, removedBy)
}

// MockEmailSender implements email sending (mock implementation)
type MockEmailSender struct{}

// NewMockEmailSender creates a new mock email sender
func NewMockEmailSender() teaminterface.EmailSender {
	return &MockEmailSender{}
}

// SendInvitationEmail sends an invitation email
func (s *MockEmailSender) SendInvitationEmail(email, inviterName, teamName, role, inviteLink string, expiresAt time.Time) error {
	// In production, implement actual email sending
	fmt.Printf("Sending team invitation email to %s\n", email)
	fmt.Printf("Inviter: %s, Team: %s, Role: %s\n", inviterName, teamName, role)
	fmt.Printf("Invitation link: %s\n", inviteLink)
	fmt.Printf("Expires at: %s\n", expiresAt.Format(time.RFC3339))
	return nil
}

// SendRoleChangedEmail sends a role changed email
func (s *MockEmailSender) SendRoleChangedEmail(email, userName, teamName, oldRole, newRole, changedBy string) error {
	// In production, implement actual email sending
	fmt.Printf("Sending role change email to %s\n", email)
	fmt.Printf("User: %s, Team: %s\n", userName, teamName)
	fmt.Printf("Role changed from %s to %s by %s\n", oldRole, newRole, changedBy)
	return nil
}

// SendMemberRemovedEmail sends a member removed email
func (s *MockEmailSender) SendMemberRemovedEmail(email, userName, teamName, removedBy string) error {
	// In production, implement actual email sending
	fmt.Printf("Sending member removed email to %s\n", email)
	fmt.Printf("User: %s removed from team: %s by %s\n", userName, teamName, removedBy)
	return nil
}