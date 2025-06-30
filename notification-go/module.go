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

func (m *Module) Mount(e *echo.Echo) error {
	g := e.Group(m.config.RoutePrefix)

	// Core notification endpoints
	if m.config.RequireAuth {
		// Note: Authentication middleware should be added by the application
		// g.Use(authMiddleware)
	}

	// Basic notification endpoints
	g.POST("/email", m.handlers.SendEmail)
	g.POST("/email/template", m.handlers.SendTemplateEmail)
	g.POST("/sms", m.handlers.SendSMS)
	g.POST("/push", m.handlers.SendPushNotification)
	g.POST("/email/bulk", m.handlers.SendBulkEmails)

	// Verification endpoints
	g.GET("/verify/email", m.handlers.VerifyEmail)
	g.GET("/verify/phone", m.handlers.VerifyPhoneNumber)

	// Common notification patterns
	commonGroup := g.Group("/common")
	commonGroup.POST("/auth/email-verification", m.handlers.SendEmailVerification)
	commonGroup.POST("/auth/password-reset", m.handlers.SendPasswordReset)
	commonGroup.POST("/team/invitation", m.handlers.SendTeamInvitation)
	commonGroup.POST("/team/role-changed", m.handlers.SendRoleChanged)
	commonGroup.POST("/billing/payment-succeeded", m.handlers.SendPaymentSucceeded)
	commonGroup.POST("/billing/payment-failed", m.handlers.SendPaymentFailed)
	commonGroup.POST("/billing/trial-ending", m.handlers.SendTrialEnding)

	// Test endpoints (only in development)
	if m.config.EnableTestEndpoints {
		testGroup := g.Group("/test")
		testGroup.POST("/email", m.handlers.TestEmail)
	}

	return nil
}

func (m *Module) Dependencies() []string {
	return []string{} // No dependencies on other modules
}

func (m *Module) Priority() int {
	return 50 // Medium priority
}