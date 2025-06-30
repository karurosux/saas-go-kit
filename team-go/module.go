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

func (m *Module) Routes() []core.Route {
	prefix := m.config.RoutePrefix
	return []core.Route{
		{Method: "POST", Path: prefix + "/accept-invite", Handler: m.handlers.AcceptInvitation},
		{Method: "GET", Path: prefix + "/members", Handler: m.handlers.ListMembers},
		{Method: "GET", Path: prefix + "/members/:id", Handler: m.handlers.GetMember},
		{Method: "POST", Path: prefix + "/members/invite", Handler: m.handlers.InviteMember},
		{Method: "PUT", Path: prefix + "/members/:id/role", Handler: m.handlers.UpdateMemberRole},
		{Method: "DELETE", Path: prefix + "/members/:id", Handler: m.handlers.RemoveMember},
		{Method: "POST", Path: prefix + "/members/:id/resend-invite", Handler: m.handlers.ResendInvitation},
		{Method: "DELETE", Path: prefix + "/members/:id/cancel-invite", Handler: m.handlers.CancelInvitation},
		{Method: "GET", Path: prefix + "/stats", Handler: m.handlers.GetTeamStats},
		{Method: "GET", Path: prefix + "/permissions/:permission", Handler: m.handlers.CheckPermission},
		{Method: "GET", Path: prefix + "/roles/:role/permissions", Handler: m.handlers.GetRolePermissions},
	}
}

func (m *Module) Middleware() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{}
}

func (m *Module) Dependencies() []string {
	return []string{"auth"}
}

func (m *Module) Init(deps map[string]core.Module) error {
	return nil
}