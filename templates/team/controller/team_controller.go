package teamcontroller

import (
	"net/http"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/team/interface"
	"{{.Project.GoModule}}/internal/team/middleware"
	"{{.Project.GoModule}}/internal/team/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// TeamController handles team management requests
type TeamController struct {
	service teaminterface.TeamService
}

// NewTeamController creates a new team controller
func NewTeamController(service teaminterface.TeamService) *TeamController {
	return &TeamController{
		service: service,
	}
}

// RegisterRoutes registers all team-related routes
func (tc *TeamController) RegisterRoutes(e *echo.Echo, basePath string, teamMiddleware *teammiddleware.TeamMiddleware) {
	group := e.Group(basePath)
	
	// Account-scoped routes
	accountGroup := group.Group("/accounts/:accountId")
	accountGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract and validate account ID
			accountIDStr := c.Param("accountId")
			accountID, err := uuid.Parse(accountIDStr)
			if err != nil {
				return core.Error(c, core.BadRequest("invalid account ID"))
			}
			teammiddleware.SetAccountIDInContext(c, accountID)
			return next(c)
		}
	})
	
	// Team member routes
	members := accountGroup.Group("/members")
	members.Use(teamMiddleware.RequireTeamMember())
	
	members.GET("", tc.ListMembers)
	members.GET("/:memberId", tc.GetMember)
	members.POST("/invite", tc.InviteMember, teamMiddleware.RequirePermission("team:invite"))
	members.PUT("/:memberId/role", tc.UpdateMemberRole, teamMiddleware.RequirePermission("team:update"))
	members.DELETE("/:memberId", tc.RemoveMember, teamMiddleware.RequirePermission("team:remove"))
	members.POST("/:memberId/resend-invitation", tc.ResendInvitation, teamMiddleware.RequirePermission("team:invite"))
	members.DELETE("/:memberId/invitation", tc.CancelInvitation, teamMiddleware.RequirePermission("team:invite"))
	
	// Team statistics
	accountGroup.GET("/stats", tc.GetTeamStats, teamMiddleware.RequireTeamMember())
	
	// Public invitation routes
	group.POST("/invitations/accept", tc.AcceptInvitation)
}

// ListMembers godoc
// @Summary List team members
// @Description Get all team members for an account
// @Tags team
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Success 200 {array} TeamMemberResponse
// @Failure 400 {object} core.ErrorResponse
// @Failure 403 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /teams/accounts/{accountId}/members [get]
func (tc *TeamController) ListMembers(c echo.Context) error {
	accountID, err := teammiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.Error(c, core.BadRequest("invalid account ID"))
	}
	
	members, err := tc.service.ListMembers(c.Request().Context(), accountID)
	if err != nil {
		return core.Error(c, err)
	}
	
	// Convert to response format
	response := make([]TeamMemberResponse, len(members))
	for i, member := range members {
		response[i] = toTeamMemberResponse(member)
	}
	
	return core.Success(c, response)
}

// GetMember godoc
// @Summary Get team member
// @Description Get a specific team member
// @Tags team
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param memberId path string true "Member ID"
// @Success 200 {object} TeamMemberResponse
// @Failure 400 {object} core.ErrorResponse
// @Failure 403 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /teams/accounts/{accountId}/members/{memberId} [get]
func (tc *TeamController) GetMember(c echo.Context) error {
	accountID, err := teammiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.Error(c, core.BadRequest("invalid account ID"))
	}
	
	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		return core.Error(c, core.BadRequest("invalid member ID"))
	}
	
	member, err := tc.service.GetMember(c.Request().Context(), accountID, memberID)
	if err != nil {
		return core.Error(c, err)
	}
	
	return core.Success(c, toTeamMemberResponse(member))
}

// InviteMemberRequest represents the invitation request
type InviteMemberRequest struct {
	Email string                   `json:"email" validate:"required,email"`
	Role  teaminterface.MemberRole `json:"role" validate:"required"`
}

// InviteMember godoc
// @Summary Invite team member
// @Description Invite a new team member
// @Tags team
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param request body InviteMemberRequest true "Invitation details"
// @Success 201 {object} TeamMemberResponse
// @Failure 400 {object} core.ErrorResponse
// @Failure 403 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /teams/accounts/{accountId}/members/invite [post]
func (tc *TeamController) InviteMember(c echo.Context) error {
	accountID, err := teammiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.Error(c, core.BadRequest("invalid account ID"))
	}
	
	inviterID, err := teammiddleware.GetUserIDFromContext(c)
	if err != nil {
		return core.Error(c, core.Unauthorized("user not authenticated"))
	}
	
	var req InviteMemberRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("invalid request body"))
	}
	
	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}
	
	inviteReq := &teammodel.InviteMemberRequest{
		AccountID: accountID,
		InviterID: inviterID,
		Email:     req.Email,
		Role:      req.Role,
	}
	
	member, err := tc.service.InviteMember(c.Request().Context(), inviteReq)
	if err != nil {
		return core.Error(c, err)
	}
	
	return core.Created(c, toTeamMemberResponse(member))
}

// UpdateMemberRoleRequest represents the role update request
type UpdateMemberRoleRequest struct {
	Role teaminterface.MemberRole `json:"role" validate:"required"`
}

// UpdateMemberRole godoc
// @Summary Update member role
// @Description Update a team member's role
// @Tags team
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param memberId path string true "Member ID"
// @Param request body UpdateMemberRoleRequest true "New role"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 403 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /teams/accounts/{accountId}/members/{memberId}/role [put]
func (tc *TeamController) UpdateMemberRole(c echo.Context) error {
	accountID, err := teammiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.Error(c, core.BadRequest("invalid account ID"))
	}
	
	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		return core.Error(c, core.BadRequest("invalid member ID"))
	}
	
	updatedByID, err := teammiddleware.GetUserIDFromContext(c)
	if err != nil {
		return core.Error(c, core.Unauthorized("user not authenticated"))
	}
	
	var req UpdateMemberRoleRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("invalid request body"))
	}
	
	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}
	
	updateReq := &teammodel.UpdateMemberRoleRequest{
		AccountID:   accountID,
		MemberID:    memberID,
		NewRole:     req.Role,
		UpdatedByID: updatedByID,
	}
	
	if err := tc.service.UpdateMemberRole(c.Request().Context(), updateReq); err != nil {
		return core.Error(c, err)
	}
	
	return core.Success(c, map[string]string{
		"message": "Member role updated successfully",
	})
}

// RemoveMember godoc
// @Summary Remove team member
// @Description Remove a team member
// @Tags team
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param memberId path string true "Member ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 403 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /teams/accounts/{accountId}/members/{memberId} [delete]
func (tc *TeamController) RemoveMember(c echo.Context) error {
	accountID, err := teammiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.Error(c, core.BadRequest("invalid account ID"))
	}
	
	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		return core.Error(c, core.BadRequest("invalid member ID"))
	}
	
	removedByID, err := teammiddleware.GetUserIDFromContext(c)
	if err != nil {
		return core.Error(c, core.Unauthorized("user not authenticated"))
	}
	
	removeReq := &teammodel.RemoveMemberRequest{
		AccountID:   accountID,
		MemberID:    memberID,
		RemovedByID: removedByID,
	}
	
	if err := tc.service.RemoveMember(c.Request().Context(), removeReq); err != nil {
		return core.Error(c, err)
	}
	
	return core.Success(c, map[string]string{
		"message": "Member removed successfully",
	})
}

// ResendInvitation godoc
// @Summary Resend invitation
// @Description Resend invitation to a pending member
// @Tags team
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param memberId path string true "Member ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 403 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /teams/accounts/{accountId}/members/{memberId}/resend-invitation [post]
func (tc *TeamController) ResendInvitation(c echo.Context) error {
	accountID, err := teammiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.Error(c, core.BadRequest("invalid account ID"))
	}
	
	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		return core.Error(c, core.BadRequest("invalid member ID"))
	}
	
	if err := tc.service.ResendInvitation(c.Request().Context(), accountID, memberID); err != nil {
		return core.Error(c, err)
	}
	
	return core.Success(c, map[string]string{
		"message": "Invitation resent successfully",
	})
}

// CancelInvitation godoc
// @Summary Cancel invitation
// @Description Cancel a pending invitation
// @Tags team
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param memberId path string true "Member ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 403 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /teams/accounts/{accountId}/members/{memberId}/invitation [delete]
func (tc *TeamController) CancelInvitation(c echo.Context) error {
	accountID, err := teammiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.Error(c, core.BadRequest("invalid account ID"))
	}
	
	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		return core.Error(c, core.BadRequest("invalid member ID"))
	}
	
	if err := tc.service.CancelInvitation(c.Request().Context(), accountID, memberID); err != nil {
		return core.Error(c, err)
	}
	
	return core.Success(c, map[string]string{
		"message": "Invitation cancelled successfully",
	})
}

// AcceptInvitationRequest represents the accept invitation request
type AcceptInvitationRequest struct {
	Token     string `json:"token" validate:"required"`
	Password  string `json:"password,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// AcceptInvitation godoc
// @Summary Accept invitation
// @Description Accept a team invitation
// @Tags team
// @Accept json
// @Produce json
// @Param request body AcceptInvitationRequest true "Acceptance details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /teams/invitations/accept [post]
func (tc *TeamController) AcceptInvitation(c echo.Context) error {
	var req AcceptInvitationRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("invalid request body"))
	}
	
	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}
	
	acceptReq := &teammodel.AcceptInvitationRequest{
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}
	
	if err := tc.service.AcceptInvitation(c.Request().Context(), req.Token, acceptReq); err != nil {
		return core.Error(c, err)
	}
	
	return core.Success(c, map[string]string{
		"message": "Invitation accepted successfully",
	})
}

// GetTeamStats godoc
// @Summary Get team statistics
// @Description Get team statistics for an account
// @Tags team
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Success 200 {object} teammodel.TeamStats
// @Failure 400 {object} core.ErrorResponse
// @Failure 403 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /teams/accounts/{accountId}/stats [get]
func (tc *TeamController) GetTeamStats(c echo.Context) error {
	accountID, err := teammiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.Error(c, core.BadRequest("invalid account ID"))
	}
	
	stats, err := tc.service.GetTeamStats(c.Request().Context(), accountID)
	if err != nil {
		return core.Error(c, err)
	}
	
	return core.Success(c, stats)
}

// Response types

type TeamMemberResponse struct {
	ID         uuid.UUID                `json:"id"`
	User       UserResponse             `json:"user"`
	Role       teaminterface.MemberRole `json:"role"`
	IsActive   bool                     `json:"is_active"`
	IsPending  bool                     `json:"is_pending"`
	InvitedAt  string                   `json:"invited_at"`
	AcceptedAt *string                  `json:"accepted_at,omitempty"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	FullName  string    `json:"full_name"`
}

func toTeamMemberResponse(member teaminterface.TeamMember) TeamMemberResponse {
	resp := TeamMemberResponse{
		ID:        member.GetID(),
		Role:      member.GetRole(),
		IsActive:  member.GetIsActive(),
		IsPending: member.GetIsPending(),
		InvitedAt: member.GetInvitedAt().Format("2006-01-02T15:04:05Z"),
	}
	
	if member.GetAcceptedAt() != nil {
		acceptedAt := member.GetAcceptedAt().Format("2006-01-02T15:04:05Z")
		resp.AcceptedAt = &acceptedAt
	}
	
	if user := member.GetUser(); user != nil {
		resp.User = UserResponse{
			ID:        user.GetID(),
			Email:     user.GetEmail(),
			FirstName: user.GetFirstName(),
			LastName:  user.GetLastName(),
			FullName:  user.GetFullName(),
		}
	}
	
	return resp
}