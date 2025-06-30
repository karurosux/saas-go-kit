package team

import (
	"github.com/karurosux/saas-go-kit/core-go"
	"github.com/labstack/echo/v4"
)

type Module struct {
	config   ModuleConfig
	handlers *Handlers
}

type ModuleConfig struct {
	TeamService         TeamService
	RoutePrefix         string
	RequireAuth         bool // Whether to require authentication for all routes
	PermissionChecker   func(c echo.Context, permission string) bool // Custom permission checker
}

func NewModule(config ModuleConfig) core.Module {
	if config.RoutePrefix == "" {
		config.RoutePrefix = "/team"
	}

	handlers := NewHandlers(config.TeamService)

	return &Module{
		config:   config,
		handlers: handlers,
	}
}

func (m *Module) Name() string {
	return "team"
}

func (m *Module) Mount(e *echo.Echo) error {
	g := e.Group(m.config.RoutePrefix)

	// Public routes (no authentication required)
	g.POST("/accept-invite", m.handlers.AcceptInvitation)

	// Team management routes (require authentication)
	teamGroup := g.Group("")
	// Note: Authentication middleware should be added by the application
	if m.config.RequireAuth {
		// teamGroup.Use(authMiddleware) // This would be added by the application
	}

	// Member management
	teamGroup.GET("/members", m.handlers.ListMembers)
	teamGroup.GET("/members/:id", m.handlers.GetMember)
	teamGroup.POST("/members/invite", m.handlers.InviteMember)
	teamGroup.PUT("/members/:id/role", m.handlers.UpdateMemberRole)
	teamGroup.DELETE("/members/:id", m.handlers.RemoveMember)

	// Invitation management
	teamGroup.POST("/members/:id/resend-invite", m.handlers.ResendInvitation)
	teamGroup.DELETE("/members/:id/cancel-invite", m.handlers.CancelInvitation)

	// Team statistics and info
	teamGroup.GET("/stats", m.handlers.GetTeamStats)

	// Permission checking
	teamGroup.GET("/permissions/:permission", m.handlers.CheckPermission)
	teamGroup.GET("/roles/:role/permissions", m.handlers.GetRolePermissions)

	return nil
}

func (m *Module) Dependencies() []string {
	return []string{"auth"} // Typically depends on auth module for user context
}

func (m *Module) Priority() int {
	return 90 // High priority as it's often needed by other modules
}