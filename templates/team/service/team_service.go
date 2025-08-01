package teamservice

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/team/constants"
	"{{.Project.GoModule}}/internal/team/interface"
	"{{.Project.GoModule}}/internal/team/model"
	"github.com/google/uuid"
)

// TeamService implements the team service
type TeamService struct {
	userRepo         teaminterface.UserRepository
	memberRepo       teaminterface.TeamMemberRepository
	tokenRepo        teaminterface.InvitationTokenRepository
	notificationSvc  teaminterface.NotificationService
	maxTeamSize      int
}

// NewTeamService creates a new team service
func NewTeamService(
	userRepo teaminterface.UserRepository,
	memberRepo teaminterface.TeamMemberRepository,
	tokenRepo teaminterface.InvitationTokenRepository,
	notificationSvc teaminterface.NotificationService,
	maxTeamSize int,
) teaminterface.TeamService {
	if maxTeamSize <= 0 {
		maxTeamSize = teamconstants.DefaultMaxTeamSize
	}
	
	return &TeamService{
		userRepo:        userRepo,
		memberRepo:      memberRepo,
		tokenRepo:       tokenRepo,
		notificationSvc: notificationSvc,
		maxTeamSize:     maxTeamSize,
	}
}

// ListMembers lists all team members
func (s *TeamService) ListMembers(ctx context.Context, accountID uuid.UUID) ([]teaminterface.TeamMember, error) {
	return s.memberRepo.GetByAccountID(ctx, accountID)
}

// GetMember gets a specific team member
func (s *TeamService) GetMember(ctx context.Context, accountID uuid.UUID, memberID uuid.UUID) (teaminterface.TeamMember, error) {
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, core.NotFound(teamconstants.ErrMemberNotFound)
	}
	
	// Verify member belongs to the account
	if member.GetAccountID() != accountID {
		return nil, core.NotFound(teamconstants.ErrMemberNotFound)
	}
	
	return member, nil
}

// InviteMember invites a new team member
func (s *TeamService) InviteMember(ctx context.Context, req teaminterface.InviteMemberRequest) (teaminterface.TeamMember, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// Check team size limit
	count, err := s.memberRepo.CountByAccountID(ctx, req.GetAccountID())
	if err != nil {
		return nil, core.InternalServerError("failed to check team size")
	}
	if count >= int64(s.maxTeamSize) {
		return nil, core.BadRequest(teamconstants.ErrTeamLimitReached)
	}
	
	// Check if inviter has permission
	if !s.CanManageTeam(ctx, req.GetAccountID(), req.GetInviterID()) {
		return nil, core.Forbidden(teamconstants.ErrInsufficientPermission)
	}
	
	// Check if user already exists
	user, _ := s.userRepo.GetByEmail(ctx, strings.ToLower(req.GetEmail()))
	
	var userID uuid.UUID
	if user != nil {
		userID = user.GetID()
		
		// Check if already a member
		existingMember, _ := s.memberRepo.GetByUserAndAccount(ctx, userID, req.GetAccountID())
		if existingMember != nil {
			if existingMember.GetIsActive() {
				return nil, core.BadRequest(teamconstants.ErrAlreadyMember)
			}
			// Reactivate existing member
			if member, ok := existingMember.(*teammodel.TeamMember); ok {
				member.SetIsActive(true)
				member.SetRole(req.GetRole())
				member.InvitedAt = time.Now()
				member.InvitedByID = req.GetInviterID()
				
				if err := s.memberRepo.Update(ctx, member); err != nil {
					return nil, core.InternalServerError("failed to reactivate member")
				}
				
				return member, nil
			}
		}
	} else {
		// Create placeholder user
		userID = uuid.New()
		newUser := &teammodel.User{
			ID:    userID,
			Email: strings.ToLower(req.GetEmail()),
		}
		
		if err := s.userRepo.Create(ctx, newUser); err != nil {
			return nil, core.InternalServerError("failed to create user")
		}
		user = newUser
	}
	
	// Create team member
	member := &teammodel.TeamMember{
		ID:          uuid.New(),
		AccountID:   req.GetAccountID(),
		UserID:      userID,
		User:        user.(*teammodel.User),
		Role:        req.GetRole(),
		IsActive:    true,
		IsPending:   true,
		InvitedAt:   time.Now(),
		InvitedByID: req.GetInviterID(),
	}
	
	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, core.InternalServerError("failed to create team member")
	}
	
	// Create invitation token
	token := &teammodel.InvitationToken{
		ID:        uuid.New(),
		Token:     generateSecureToken(),
		MemberID:  member.ID,
		ExpiresAt: time.Now().Add(teamconstants.DefaultInvitationExpiration),
	}
	
	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return nil, core.InternalServerError("failed to create invitation token")
	}
	
	// Get inviter info for notification
	inviter, _ := s.memberRepo.GetByUserAndAccount(ctx, req.GetInviterID(), req.GetAccountID())
	inviterName := "Team Admin"
	if inviter != nil && inviter.GetUser() != nil {
		inviterName = inviter.GetUser().GetFullName()
	}
	
	// Send invitation notification
	go s.notificationSvc.SendTeamInvitation(
		context.Background(),
		req.GetEmail(),
		inviterName,
		"Your Team", // In production, get actual team/account name
		string(req.GetRole()),
		token.Token,
		token.ExpiresAt,
	)
	
	return member, nil
}

// UpdateMemberRole updates a member's role
func (s *TeamService) UpdateMemberRole(ctx context.Context, req teaminterface.UpdateMemberRoleRequest) error {
	// Validate request
	if err := req.Validate(); err != nil {
		return err
	}
	
	// Check if updater has permission
	if !s.CanManageTeam(ctx, req.GetAccountID(), req.GetUpdatedByID()) {
		return core.Forbidden(teamconstants.ErrInsufficientPermission)
	}
	
	// Prevent self role change
	memberToUpdate, err := s.memberRepo.GetByID(ctx, req.GetMemberID())
	if err != nil {
		return core.NotFound(teamconstants.ErrMemberNotFound)
	}
	
	if memberToUpdate.GetUserID() == req.GetUpdatedByID() {
		return core.BadRequest(teamconstants.ErrSelfRoleChange)
	}
	
	// Verify member belongs to the account
	if memberToUpdate.GetAccountID() != req.GetAccountID() {
		return core.NotFound(teamconstants.ErrMemberNotFound)
	}
	
	// Cannot change owner role
	if memberToUpdate.GetRole() == teaminterface.RoleOwner {
		return core.BadRequest(teamconstants.ErrCannotChangeOwnerRole)
	}
	
	// If changing to owner, ensure there's still an owner
	if req.GetNewRole() == teaminterface.RoleOwner {
		// Check if current user is owner
		updaterMember, err := s.memberRepo.GetByUserAndAccount(ctx, req.GetUpdatedByID(), req.GetAccountID())
		if err != nil || updaterMember.GetRole() != teaminterface.RoleOwner {
			return core.Forbidden("only owners can transfer ownership")
		}
	}
	
	oldRole := memberToUpdate.GetRole()
	
	// Update role
	if member, ok := memberToUpdate.(*teammodel.TeamMember); ok {
		member.SetRole(req.GetNewRole())
		if err := s.memberRepo.Update(ctx, member); err != nil {
			return core.InternalServerError("failed to update member role")
		}
	}
	
	// Get user info for notification
	userName := "Team Member"
	if memberToUpdate.GetUser() != nil {
		userName = memberToUpdate.GetUser().GetFullName()
	}
	
	updater, _ := s.memberRepo.GetByUserAndAccount(ctx, req.GetUpdatedByID(), req.GetAccountID())
	updaterName := "Team Admin"
	if updater != nil && updater.GetUser() != nil {
		updaterName = updater.GetUser().GetFullName()
	}
	
	// Send notification
	go s.notificationSvc.SendRoleChanged(
		context.Background(),
		memberToUpdate.GetUser().GetEmail(),
		userName,
		"Your Team",
		string(oldRole),
		string(req.GetNewRole()),
		updaterName,
	)
	
	return nil
}

// RemoveMember removes a team member
func (s *TeamService) RemoveMember(ctx context.Context, req teaminterface.RemoveMemberRequest) error {
	// Validate request
	if err := req.Validate(); err != nil {
		return err
	}
	
	// Check if remover has permission
	if !s.CanManageTeam(ctx, req.GetAccountID(), req.GetRemovedByID()) {
		return core.Forbidden(teamconstants.ErrInsufficientPermission)
	}
	
	// Get member to remove
	memberToRemove, err := s.memberRepo.GetByID(ctx, req.GetMemberID())
	if err != nil {
		return core.NotFound(teamconstants.ErrMemberNotFound)
	}
	
	// Verify member belongs to the account
	if memberToRemove.GetAccountID() != req.GetAccountID() {
		return core.NotFound(teamconstants.ErrMemberNotFound)
	}
	
	// Prevent self removal
	if memberToRemove.GetUserID() == req.GetRemovedByID() {
		return core.BadRequest(teamconstants.ErrSelfRemoval)
	}
	
	// Cannot remove owner
	if memberToRemove.GetRole() == teaminterface.RoleOwner {
		// Check if there are other owners
		members, err := s.memberRepo.GetByAccountID(ctx, req.GetAccountID())
		if err != nil {
			return core.InternalServerError("failed to check owners")
		}
		
		ownerCount := 0
		for _, m := range members {
			if m.GetRole() == teaminterface.RoleOwner && m.GetIsActive() {
				ownerCount++
			}
		}
		
		if ownerCount <= 1 {
			return core.BadRequest(teamconstants.ErrMustHaveOwner)
		}
	}
	
	// Soft delete (deactivate) the member
	if member, ok := memberToRemove.(*teammodel.TeamMember); ok {
		member.SetIsActive(false)
		if err := s.memberRepo.Update(ctx, member); err != nil {
			return core.InternalServerError("failed to remove member")
		}
	}
	
	// Get user info for notification
	userName := "Team Member"
	userEmail := ""
	if memberToRemove.GetUser() != nil {
		userName = memberToRemove.GetUser().GetFullName()
		userEmail = memberToRemove.GetUser().GetEmail()
	}
	
	remover, _ := s.memberRepo.GetByUserAndAccount(ctx, req.GetRemovedByID(), req.GetAccountID())
	removerName := "Team Admin"
	if remover != nil && remover.GetUser() != nil {
		removerName = remover.GetUser().GetFullName()
	}
	
	// Send notification
	if userEmail != "" {
		go s.notificationSvc.SendMemberRemoved(
			context.Background(),
			userEmail,
			userName,
			"Your Team",
			removerName,
		)
	}
	
	return nil
}

// AcceptInvitation accepts a team invitation
func (s *TeamService) AcceptInvitation(ctx context.Context, token string, acceptReq teaminterface.AcceptInvitationRequest) error {
	// Get invitation token
	inviteToken, err := s.tokenRepo.GetByToken(ctx, token)
	if err != nil {
		return core.BadRequest(teamconstants.ErrInvalidToken)
	}
	
	// Check if token is used
	if inviteToken.GetUsed() {
		return core.BadRequest(teamconstants.ErrTokenAlreadyUsed)
	}
	
	// Check if token is expired
	if inviteToken.IsExpired() {
		return core.BadRequest(teamconstants.ErrTokenExpired)
	}
	
	// Get team member
	member, err := s.memberRepo.GetByID(ctx, inviteToken.GetMemberID())
	if err != nil {
		return core.NotFound(teamconstants.ErrMemberNotFound)
	}
	
	// Update user if needed
	user, err := s.userRepo.GetByID(ctx, member.GetUserID())
	if err != nil {
		return core.NotFound(teamconstants.ErrUserNotFound)
	}
	
	// Update user details if provided
	if u, ok := user.(*teammodel.User); ok {
		if acceptReq.GetFirstName() != "" {
			u.FirstName = acceptReq.GetFirstName()
		}
		if acceptReq.GetLastName() != "" {
			u.LastName = acceptReq.GetLastName()
		}
		
		if err := s.userRepo.Update(ctx, u); err != nil {
			return core.InternalServerError("failed to update user")
		}
	}
	
	// Mark invitation as accepted
	now := time.Now()
	if m, ok := member.(*teammodel.TeamMember); ok {
		m.SetAcceptedAt(&now)
		m.IsPending = false
		
		if err := s.memberRepo.Update(ctx, m); err != nil {
			return core.InternalServerError("failed to update member")
		}
	}
	
	// Mark token as used
	if t, ok := inviteToken.(*teammodel.InvitationToken); ok {
		t.SetUsed(true)
		if err := s.tokenRepo.Update(ctx, t); err != nil {
			// Log error but don't fail
			fmt.Printf("Failed to mark token as used: %v\n", err)
		}
	}
	
	return nil
}

// ResendInvitation resends an invitation
func (s *TeamService) ResendInvitation(ctx context.Context, accountID uuid.UUID, memberID uuid.UUID) error {
	// Get member
	member, err := s.GetMember(ctx, accountID, memberID)
	if err != nil {
		return err
	}
	
	// Check if already accepted
	if !member.GetIsPending() {
		return core.BadRequest("member has already accepted the invitation")
	}
	
	// Get existing token
	existingToken, _ := s.tokenRepo.GetByMemberID(ctx, memberID)
	
	// Create new token
	token := &teammodel.InvitationToken{
		ID:        uuid.New(),
		Token:     generateSecureToken(),
		MemberID:  memberID,
		ExpiresAt: time.Now().Add(teamconstants.DefaultInvitationExpiration),
	}
	
	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return core.InternalServerError("failed to create invitation token")
	}
	
	// Delete old token if exists
	if existingToken != nil {
		s.tokenRepo.Delete(ctx, existingToken.GetID())
	}
	
	// Send invitation notification
	if member.GetUser() != nil {
		go s.notificationSvc.SendTeamInvitation(
			context.Background(),
			member.GetUser().GetEmail(),
			"Team Admin",
			"Your Team",
			string(member.GetRole()),
			token.Token,
			token.ExpiresAt,
		)
	}
	
	return nil
}

// CancelInvitation cancels a pending invitation
func (s *TeamService) CancelInvitation(ctx context.Context, accountID uuid.UUID, memberID uuid.UUID) error {
	// Get member
	member, err := s.GetMember(ctx, accountID, memberID)
	if err != nil {
		return err
	}
	
	// Check if already accepted
	if !member.GetIsPending() {
		return core.BadRequest("cannot cancel accepted invitation")
	}
	
	// Delete invitation token
	token, _ := s.tokenRepo.GetByMemberID(ctx, memberID)
	if token != nil {
		s.tokenRepo.Delete(ctx, token.GetID())
	}
	
	// Delete member
	return s.memberRepo.Delete(ctx, memberID)
}

// GetTeamStats gets team statistics
func (s *TeamService) GetTeamStats(ctx context.Context, accountID uuid.UUID) (teaminterface.TeamStats, error) {
	return s.memberRepo.GetTeamStats(ctx, accountID)
}

// CheckPermission checks if a user has a specific permission
func (s *TeamService) CheckPermission(ctx context.Context, accountID uuid.UUID, userID uuid.UUID, permission string) (bool, error) {
	member, err := s.memberRepo.GetByUserAndAccount(ctx, userID, accountID)
	if err != nil {
		return false, nil
	}
	
	if !member.GetIsActive() {
		return false, nil
	}
	
	// Get role permissions
	rolePerms, exists := teamconstants.RolePermissions[string(member.GetRole())]
	if !exists {
		return false, nil
	}
	
	// Check if permission exists in role
	for _, perm := range rolePerms {
		if perm == permission {
			return true, nil
		}
	}
	
	return false, nil
}

// GetMemberRole gets a member's role
func (s *TeamService) GetMemberRole(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) (teaminterface.MemberRole, error) {
	member, err := s.memberRepo.GetByUserAndAccount(ctx, userID, accountID)
	if err != nil {
		return "", core.NotFound(teamconstants.ErrMemberNotFound)
	}
	
	if !member.GetIsActive() {
		return "", core.Forbidden("member is not active")
	}
	
	return member.GetRole(), nil
}

// IsOwner checks if user is owner
func (s *TeamService) IsOwner(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) bool {
	role, err := s.GetMemberRole(ctx, accountID, userID)
	return err == nil && role == teaminterface.RoleOwner
}

// IsAdmin checks if user is admin or owner
func (s *TeamService) IsAdmin(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) bool {
	role, err := s.GetMemberRole(ctx, accountID, userID)
	return err == nil && (role == teaminterface.RoleOwner || role == teaminterface.RoleAdmin)
}

// CanManageTeam checks if user can manage team
func (s *TeamService) CanManageTeam(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) bool {
	hasPermission, _ := s.CheckPermission(ctx, accountID, userID, teamconstants.PermissionTeamManage)
	if hasPermission {
		return true
	}
	
	// Fallback to role check
	return s.IsAdmin(ctx, accountID, userID)
}

// generateSecureToken generates a secure random token
func generateSecureToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}