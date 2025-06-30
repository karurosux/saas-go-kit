package notification

import (
	"github.com/karurosux/saas-go-kit/core-go"
	"github.com/labstack/echo/v4"
)

type Module struct {
	config   ModuleConfig
	handlers *Handlers
}

type ModuleConfig struct {
	NotificationService NotificationService
	CommonService       *CommonNotificationService
	RoutePrefix         string
	EnableTestEndpoints bool // Enable test endpoints for development
	RequireAuth         bool // Whether to require authentication for endpoints
}

func NewModule(config ModuleConfig) core.Module {
	if config.RoutePrefix == "" {
		config.RoutePrefix = "/notifications"
	}

	handlers := NewHandlers(config.NotificationService, config.CommonService)

	return &Module{
		config:   config,
		handlers: handlers,
	}
}

func (m *Module) Name() string {
	return "notification"
}

func (m *Module) Routes() []core.Route {
	prefix := m.config.RoutePrefix
	routes := []core.Route{
		{Method: "POST", Path: prefix + "/email", Handler: m.handlers.SendEmail},
		{Method: "POST", Path: prefix + "/email/template", Handler: m.handlers.SendTemplateEmail},
		{Method: "POST", Path: prefix + "/sms", Handler: m.handlers.SendSMS},
		{Method: "POST", Path: prefix + "/push", Handler: m.handlers.SendPushNotification},
		{Method: "POST", Path: prefix + "/email/bulk", Handler: m.handlers.SendBulkEmails},
		{Method: "GET", Path: prefix + "/verify/email", Handler: m.handlers.VerifyEmail},
		{Method: "GET", Path: prefix + "/verify/phone", Handler: m.handlers.VerifyPhoneNumber},
		{Method: "POST", Path: prefix + "/common/auth/email-verification", Handler: m.handlers.SendEmailVerification},
		{Method: "POST", Path: prefix + "/common/auth/password-reset", Handler: m.handlers.SendPasswordReset},
		{Method: "POST", Path: prefix + "/common/team/invitation", Handler: m.handlers.SendTeamInvitation},
		{Method: "POST", Path: prefix + "/common/team/role-changed", Handler: m.handlers.SendRoleChanged},
		{Method: "POST", Path: prefix + "/common/billing/payment-succeeded", Handler: m.handlers.SendPaymentSucceeded},
		{Method: "POST", Path: prefix + "/common/billing/payment-failed", Handler: m.handlers.SendPaymentFailed},
		{Method: "POST", Path: prefix + "/common/billing/trial-ending", Handler: m.handlers.SendTrialEnding},
	}
	
	if m.config.EnableTestEndpoints {
		routes = append(routes, core.Route{Method: "POST", Path: prefix + "/test/email", Handler: m.handlers.TestEmail})
	}
	
	return routes
}

func (m *Module) Middleware() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{}
}

func (m *Module) Dependencies() []string {
	return []string{}
}

func (m *Module) Init(deps map[string]core.Module) error {
	return nil
}