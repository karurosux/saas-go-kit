package team

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/errors-go"
)

type teamService struct {
	teamMemberRepo  TeamMemberRepository
	userRepo        UserRepository
	tokenRepo       InvitationTokenRepository
	notificationSvc NotificationService
	usageSvc        UsageService
}

func NewTeamService(
	teamMemberRepo TeamMemberRepository,
	userRepo UserRepository,
	tokenRepo InvitationTokenRepository,
	notificationSvc NotificationService,
	usageSvc UsageService,
) TeamService {
	return &teamService{
		teamMemberRepo:  teamMemberRepo,
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		notificationSvc: notificationSvc,
		usageSvc:        usageSvc,
	}
}

func (s *teamService) ListMembers(ctx context.Context, accountID uuid.UUID) ([]TeamMember, error) {
	return s.teamMemberRepo.FindByAccountID(ctx, accountID)
}

func (s *teamService) GetMember(ctx context.Context, accountID uuid.UUID, memberID uuid.UUID) (*TeamMember, error) {
	member, err := s.teamMemberRepo.FindByID(ctx, memberID, "User")
	if err != nil {
		return nil, err
	}

	if member.AccountID != accountID {
		return nil, errors.Forbidden("access this team member")
	}

	return member, nil
}

func (s *teamService) InviteMember(ctx context.Context, req *InviteMemberRequest) (*TeamMember, error) {
	// Validate role
	if req.Role == RoleOwner {
		return nil, errors.Forbidden("invite another owner")
	}

	if !req.Role.IsValid() {
		return nil, errors.BadRequest("Invalid role specified")
	}

	// Check usage limits if usage service is available
	if s.usageSvc != nil {
		canAdd, reason, err := s.usageSvc.CanAddMember(ctx, req.AccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to check member limits: %w", err)
		}
		if !canAdd {
			return nil, errors.BadRequest(reason)
		}
	}

	// Check if user exists
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Create placeholder user if not exists
		user = &User{
			Email:    req.Email,
			IsActive: false,
		}
		// Generate temporary password
		if err := user.SetPassword(uuid.New().String()); err != nil {
			return nil, fmt.Errorf("failed to set temporary password: %w", err)
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Check if already a member
	existing, err := s.teamMemberRepo.FindByUserAndAccount(ctx, user.ID, req.AccountID)
	if err == nil && existing != nil {
		return nil, errors.Conflict("User is already a team member")
	}

	// Create team member
	member := &TeamMember{
		AccountID:  req.AccountID,
		UserID:     user.ID,
		Role:       req.Role,
		InvitedBy:  req.InviterID,
		InvitedAt:  time.Now(),
		AcceptedAt: nil,
	}

	if err := s.teamMemberRepo.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to create team member: %w", err)
	}

	// Generate invitation token
	tokenStr, err := GenerateInvitationToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	invitationToken := &InvitationToken{
		AccountID: req.AccountID,
		MemberID:  member.ID,
		Token:     tokenStr,
		Email:     req.Email,
		Role:      req.Role,
		InvitedBy: req.InviterID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.tokenRepo.Create(ctx, invitationToken); err != nil {
		return nil, fmt.Errorf("failed to create invitation token: %w", err)
	}

	// Send invitation notification
	if s.notificationSvc != nil {
		notification := &TeamInvitationNotification{
			Email:     req.Email,
			Role:      string(req.Role),
			Token:     tokenStr,
			ExpiresAt: invitationToken.ExpiresAt.Format(time.RFC3339),
		}
		if err := s.notificationSvc.SendTeamInvitation(ctx, notification); err != nil {
			// Log error but don't fail the invitation
			// TODO: Add proper logging
		}
	}

	// Track usage
	if s.usageSvc != nil {
		if err := s.usageSvc.TrackMemberAdded(ctx, req.AccountID); err != nil {
			// Log error but don't fail
			// TODO: Add proper logging
		}
	}

	// Load user data
	member.User = *user

	return member, nil
}

func (s *teamService) UpdateMemberRole(ctx context.Context, req *UpdateMemberRoleRequest) error {
	// Get member
	member, err := s.teamMemberRepo.FindByID(ctx, req.MemberID, "User")
	if err != nil {
		return errors.NotFound("Team member not found")
	}

	// Verify member belongs to account
	if member.AccountID != req.AccountID {
		return errors.Forbidden("access this team member")
	}

	// Cannot change owner role
	if member.Role == RoleOwner {
		return errors.Forbidden("change owner role")
	}

	// Cannot set someone as owner
	if req.NewRole == RoleOwner {
		return errors.Forbidden("assign owner role")
	}

	if !req.NewRole.IsValid() {
		return errors.BadRequest("Invalid role specified")
	}

	oldRole := member.Role

	// Update role
	member.Role = req.NewRole
	if err := s.teamMemberRepo.Update(ctx, member); err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	// Send notification
	if s.notificationSvc != nil {
		notification := &RoleChangedNotification{
			Email:    member.User.Email,
			UserName: member.User.FullName(),
			OldRole:  string(oldRole),
			NewRole:  string(req.NewRole),
		}
		if err := s.notificationSvc.SendRoleChanged(ctx, notification); err != nil {
			// Log error but don't fail
			// TODO: Add proper logging
		}
	}

	return nil
}

func (s *teamService) RemoveMember(ctx context.Context, req *RemoveMemberRequest) error {
	// Get member
	member, err := s.teamMemberRepo.FindByID(ctx, req.MemberID, "User")
	if err != nil {
		return errors.NotFound("Team member not found")
	}

	// Verify member belongs to account
	if member.AccountID != req.AccountID {
		return errors.Forbidden("access this team member")
	}

	// Cannot remove owner
	if member.Role == RoleOwner {
		return errors.Forbidden("remove owner")
	}

	// Delete member
	if err := s.teamMemberRepo.Delete(ctx, req.MemberID); err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}

	// Track usage
	if s.usageSvc != nil {
		if err := s.usageSvc.TrackMemberRemoved(ctx, req.AccountID); err != nil {
			// Log error but don't fail
			// TODO: Add proper logging
		}
	}

	// Send notification
	if s.notificationSvc != nil {
		notification := &MemberRemovedNotification{
			Email:    member.User.Email,
			UserName: member.User.FullName(),
		}
		if err := s.notificationSvc.SendMemberRemoved(ctx, notification); err != nil {
			// Log error but don't fail
			// TODO: Add proper logging
		}
	}

	return nil
}

func (s *teamService) AcceptInvitation(ctx context.Context, token string) error {
	// Find token
	invitationToken, err := s.tokenRepo.FindByToken(ctx, token)
	if err != nil {
		return errors.Unauthorized("Invalid or expired invitation token")
	}

	// Validate token
	if !invitationToken.IsValid() {
		return errors.Unauthorized("Invalid or expired invitation token")
	}

	// Get and update member
	member, err := s.teamMemberRepo.FindByID(ctx, invitationToken.MemberID)
	if err != nil {
		return fmt.Errorf("failed to find team member: %w", err)
	}

	// Mark as accepted
	now := time.Now()
	member.AcceptedAt = &now

	if err := s.teamMemberRepo.Update(ctx, member); err != nil {
		return fmt.Errorf("failed to update team member: %w", err)
	}

	// Mark token as used
	invitationToken.MarkAsUsed()
	if err := s.tokenRepo.Update(ctx, invitationToken); err != nil {
		return fmt.Errorf("failed to update invitation token: %w", err)
	}

	return nil
}

func (s *teamService) ResendInvitation(ctx context.Context, accountID uuid.UUID, memberID uuid.UUID) error {
	// Get member
	member, err := s.teamMemberRepo.FindByID(ctx, memberID, "User")
	if err != nil {
		return errors.NotFound("Team member not found")
	}

	// Verify member belongs to account
	if member.AccountID != accountID {
		return errors.Forbidden("access this team member")
	}

	// Member must be pending
	if member.IsActive() {
		return errors.BadRequest("Member has already accepted the invitation")
	}

	// Find existing token
	existingToken, err := s.tokenRepo.FindByMemberID(ctx, memberID)
	if err == nil && existingToken != nil {
		// Delete existing token
		if err := s.tokenRepo.Delete(ctx, existingToken.ID); err != nil {
			return fmt.Errorf("failed to delete existing token: %w", err)
		}
	}

	// Generate new invitation token
	tokenStr, err := GenerateInvitationToken()
	if err != nil {
		return fmt.Errorf("failed to generate invitation token: %w", err)
	}

	invitationToken := &InvitationToken{
		AccountID: accountID,
		MemberID:  memberID,
		Token:     tokenStr,
		Email:     member.User.Email,
		Role:      member.Role,
		InvitedBy: member.InvitedBy,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.tokenRepo.Create(ctx, invitationToken); err != nil {
		return fmt.Errorf("failed to create invitation token: %w", err)
	}

	// Send invitation notification
	if s.notificationSvc != nil {
		notification := &TeamInvitationNotification{
			Email:     member.User.Email,
			Role:      string(member.Role),
			Token:     tokenStr,
			ExpiresAt: invitationToken.ExpiresAt.Format(time.RFC3339),
		}
		if err := s.notificationSvc.SendTeamInvitation(ctx, notification); err != nil {
			// Log error but don't fail
			// TODO: Add proper logging
		}
	}

	return nil
}

func (s *teamService) CancelInvitation(ctx context.Context, accountID uuid.UUID, memberID uuid.UUID) error {
	// Get member
	member, err := s.teamMemberRepo.FindByID(ctx, memberID)
	if err != nil {
		return errors.NotFound("Team member not found")
	}

	// Verify member belongs to account
	if member.AccountID != accountID {
		return errors.Forbidden("access this team member")
	}

	// Member must be pending
	if member.IsActive() {
		return errors.BadRequest("Cannot cancel invitation for active member")
	}

	// Delete invitation token
	existingToken, err := s.tokenRepo.FindByMemberID(ctx, memberID)
	if err == nil && existingToken != nil {
		if err := s.tokenRepo.Delete(ctx, existingToken.ID); err != nil {
			return fmt.Errorf("failed to delete invitation token: %w", err)
		}
	}

	// Delete member
	if err := s.teamMemberRepo.Delete(ctx, memberID); err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}

	return nil
}

func (s *teamService) GetTeamStats(ctx context.Context, accountID uuid.UUID) (*TeamStats, error) {
	return s.teamMemberRepo.GetTeamStats(ctx, accountID)
}

func (s *teamService) CheckPermission(ctx context.Context, accountID uuid.UUID, userID uuid.UUID, permission string) (bool, error) {
	member, err := s.teamMemberRepo.FindByUserAndAccount(ctx, userID, accountID)
	if err != nil {
		return false, nil // User is not a member
	}

	if !member.IsActive() {
		return false, nil // Member hasn't accepted invitation
	}

	return member.Role.HasPermission(permission), nil
}

func (s *teamService) GetMemberRole(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) (MemberRole, error) {
	member, err := s.teamMemberRepo.FindByUserAndAccount(ctx, userID, accountID)
	if err != nil {
		return "", errors.NotFound("User is not a member of this team")
	}

	if !member.IsActive() {
		return "", errors.BadRequest("Member has not accepted invitation")
	}

	return member.Role, nil
}