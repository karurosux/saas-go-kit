package team

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/errors-go"
	"github.com/karurosux/saas-go-kit/response-go"
	"github.com/karurosux/saas-go-kit/validator-go"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	teamService TeamService
	validator   *validator.Validator
}

func NewHandlers(teamService TeamService) *Handlers {
	return &Handlers{
		teamService: teamService,
		validator:   validator.New(),
	}
}

// Request DTOs for handlers
type InviteMemberRequestDTO struct {
	Email string     `json:"email" validate:"required,email"`
	Role  MemberRole `json:"role" validate:"required"`
}

type UpdateRoleRequestDTO struct {
	Role MemberRole `json:"role" validate:"required"`
}

type AcceptInviteRequestDTO struct {
	Token     string `json:"token" validate:"required"`
	Password  string `json:"password,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

func (h *Handlers) ListMembers(c echo.Context) error {
	accountID, err := h.getAccountID(c)
	if err != nil {
		return response.Error(c, err)
	}

	members, err := h.teamService.ListMembers(c.Request().Context(), accountID)
	if err != nil {
		return response.Error(c, errors.Internal("Failed to list team members"))
	}

	// Convert to response format
	memberResponses := make([]TeamMemberResponse, len(members))
	for i, member := range members {
		memberResponses[i] = h.convertToMemberResponse(member)
	}

	return response.Success(c, memberResponses)
}

func (h *Handlers) GetMember(c echo.Context) error {
	accountID, err := h.getAccountID(c)
	if err != nil {
		return response.Error(c, err)
	}

	memberID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid member ID"))
	}

	member, err := h.teamService.GetMember(c.Request().Context(), accountID, memberID)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, h.convertToMemberResponse(*member))
}

func (h *Handlers) InviteMember(c echo.Context) error {
	accountID, err := h.getAccountID(c)
	if err != nil {
		return response.Error(c, err)
	}

	userID, err := h.getUserID(c)
	if err != nil {
		return response.Error(c, err)
	}

	// Check permissions
	hasPermission, err := h.teamService.CheckPermission(c.Request().Context(), accountID, userID, "invite_members")
	if err != nil {
		return response.Error(c, errors.Internal("Failed to check permissions"))
	}
	if !hasPermission {
		return response.Error(c, errors.Forbidden("invite team members"))
	}

	var req InviteMemberRequestDTO
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
	}

	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, err)
	}

	// Validate role
	if !req.Role.IsValid() || req.Role == RoleOwner {
		return response.Error(c, errors.BadRequest("Invalid role specified"))
	}

	inviteReq := &InviteMemberRequest{
		AccountID: accountID,
		InviterID: userID,
		Email:     req.Email,
		Role:      req.Role,
	}

	member, err := h.teamService.InviteMember(c.Request().Context(), inviteReq)
	if err != nil {
		return response.Error(c, err)
	}

	return c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Data:    h.convertToMemberResponse(*member),
	})
}

func (h *Handlers) UpdateMemberRole(c echo.Context) error {
	accountID, err := h.getAccountID(c)
	if err != nil {
		return response.Error(c, err)
	}

	userID, err := h.getUserID(c)
	if err != nil {
		return response.Error(c, err)
	}

	// Check permissions
	hasPermission, err := h.teamService.CheckPermission(c.Request().Context(), accountID, userID, "change_roles")
	if err != nil {
		return response.Error(c, errors.Internal("Failed to check permissions"))
	}
	if !hasPermission {
		return response.Error(c, errors.Forbidden("change member roles"))
	}

	memberID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid member ID"))
	}

	var req UpdateRoleRequestDTO
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
	}

	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, err)
	}

	updateReq := &UpdateMemberRoleRequest{
		AccountID:   accountID,
		MemberID:    memberID,
		NewRole:     req.Role,
		UpdatedByID: userID,
	}

	if err := h.teamService.UpdateMemberRole(c.Request().Context(), updateReq); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Member role updated successfully",
	})
}

func (h *Handlers) RemoveMember(c echo.Context) error {
	accountID, err := h.getAccountID(c)
	if err != nil {
		return response.Error(c, err)
	}

	userID, err := h.getUserID(c)
	if err != nil {
		return response.Error(c, err)
	}

	// Check permissions
	hasPermission, err := h.teamService.CheckPermission(c.Request().Context(), accountID, userID, "remove_members")
	if err != nil {
		return response.Error(c, errors.Internal("Failed to check permissions"))
	}
	if !hasPermission {
		return response.Error(c, errors.Forbidden("remove team members"))
	}

	memberID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid member ID"))
	}

	removeReq := &RemoveMemberRequest{
		AccountID:   accountID,
		MemberID:    memberID,
		RemovedByID: userID,
	}

	if err := h.teamService.RemoveMember(c.Request().Context(), removeReq); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Member removed successfully",
	})
}

func (h *Handlers) AcceptInvitation(c echo.Context) error {
	var req AcceptInviteRequestDTO
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
	}

	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, err)
	}

	if err := h.teamService.AcceptInvitation(c.Request().Context(), req.Token); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Invitation accepted successfully",
	})
}

func (h *Handlers) ResendInvitation(c echo.Context) error {
	accountID, err := h.getAccountID(c)
	if err != nil {
		return response.Error(c, err)
	}

	userID, err := h.getUserID(c)
	if err != nil {
		return response.Error(c, err)
	}

	// Check permissions
	hasPermission, err := h.teamService.CheckPermission(c.Request().Context(), accountID, userID, "invite_members")
	if err != nil {
		return response.Error(c, errors.Internal("Failed to check permissions"))
	}
	if !hasPermission {
		return response.Error(c, errors.Forbidden("manage team invitations"))
	}

	memberID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid member ID"))
	}

	if err := h.teamService.ResendInvitation(c.Request().Context(), accountID, memberID); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Invitation resent successfully",
	})
}

func (h *Handlers) CancelInvitation(c echo.Context) error {
	accountID, err := h.getAccountID(c)
	if err != nil {
		return response.Error(c, err)
	}

	userID, err := h.getUserID(c)
	if err != nil {
		return response.Error(c, err)
	}

	// Check permissions
	hasPermission, err := h.teamService.CheckPermission(c.Request().Context(), accountID, userID, "invite_members")
	if err != nil {
		return response.Error(c, errors.Internal("Failed to check permissions"))
	}
	if !hasPermission {
		return response.Error(c, errors.Forbidden("manage team invitations"))
	}

	memberID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid member ID"))
	}

	if err := h.teamService.CancelInvitation(c.Request().Context(), accountID, memberID); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Invitation cancelled successfully",
	})
}

func (h *Handlers) GetTeamStats(c echo.Context) error {
	accountID, err := h.getAccountID(c)
	if err != nil {
		return response.Error(c, err)
	}

	stats, err := h.teamService.GetTeamStats(c.Request().Context(), accountID)
	if err != nil {
		return response.Error(c, errors.Internal("Failed to get team statistics"))
	}

	return response.Success(c, stats)
}

func (h *Handlers) CheckPermission(c echo.Context) error {
	accountID, err := h.getAccountID(c)
	if err != nil {
		return response.Error(c, err)
	}

	userID, err := h.getUserID(c)
	if err != nil {
		return response.Error(c, err)
	}

	permission := c.Param("permission")
	if permission == "" {
		return response.Error(c, errors.BadRequest("Permission parameter is required"))
	}

	hasPermission, err := h.teamService.CheckPermission(c.Request().Context(), accountID, userID, permission)
	if err != nil {
		return response.Error(c, errors.Internal("Failed to check permission"))
	}

	role, _ := h.teamService.GetMemberRole(c.Request().Context(), accountID, userID)

	result := PermissionCheckResponse{
		HasPermission: hasPermission,
		Role:          string(role),
	}

	if !hasPermission {
		result.Reason = "Insufficient permissions"
	}

	return response.Success(c, result)
}

func (h *Handlers) GetRolePermissions(c echo.Context) error {
	roleStr := c.Param("role")
	role := MemberRole(roleStr)

	if !role.IsValid() {
		return response.Error(c, errors.BadRequest("Invalid role specified"))
	}

	permissions := role.GetPermissions()
	return response.Success(c, permissions)
}

// Helper methods

func (h *Handlers) getAccountID(c echo.Context) (uuid.UUID, error) {
	accountIDStr, ok := c.Get("account_id").(string)
	if !ok {
		return uuid.Nil, errors.Unauthorized("Account ID not found in context")
	}
	
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return uuid.Nil, errors.BadRequest("Invalid account ID")
	}
	
	return accountID, nil
}

func (h *Handlers) getUserID(c echo.Context) (uuid.UUID, error) {
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return uuid.Nil, errors.Unauthorized("User ID not found in context")
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.BadRequest("Invalid user ID")
	}
	
	return userID, nil
}

func (h *Handlers) convertToMemberResponse(member TeamMember) TeamMemberResponse {
	var acceptedAt *string
	if member.AcceptedAt != nil {
		acceptedAtStr := member.AcceptedAt.Format("2006-01-02T15:04:05Z07:00")
		acceptedAt = &acceptedAtStr
	}

	return TeamMemberResponse{
		ID: member.ID,
		User: UserResponse{
			ID:        member.User.ID,
			Email:     member.User.Email,
			FirstName: member.User.FirstName,
			LastName:  member.User.LastName,
			FullName:  member.User.FullName(),
		},
		Role:       member.Role,
		IsActive:   member.IsActive(),
		IsPending:  member.IsPending(),
		InvitedAt:  member.InvitedAt.Format("2006-01-02T15:04:05Z07:00"),
		AcceptedAt: acceptedAt,
	}
}