package notificationcontroller

import (
	"net/http"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/notification/interface"
	"{{.Project.GoModule}}/internal/notification/model"
	"github.com/labstack/echo/v4"
)

// NotificationController handles notification requests
type NotificationController struct {
	notificationService notificationinterface.NotificationService
	commonService       notificationinterface.CommonNotificationService
}

// NewNotificationController creates a new notification controller
func NewNotificationController(
	notificationService notificationinterface.NotificationService,
	commonService notificationinterface.CommonNotificationService,
) *NotificationController {
	return &NotificationController{
		notificationService: notificationService,
		commonService:       commonService,
	}
}

// RegisterRoutes registers all notification-related routes
func (nc *NotificationController) RegisterRoutes(e *echo.Echo, basePath string) {
	group := e.Group(basePath)
	
	// Basic notification endpoints
	group.POST("/email", nc.SendEmail)
	group.POST("/email/template", nc.SendTemplateEmail)
	group.POST("/email/bulk", nc.SendBulkEmails)
	group.POST("/sms", nc.SendSMS)
	group.POST("/push", nc.SendPushNotification)
	
	// Verification endpoints
	group.GET("/verify/email", nc.VerifyEmail)
	group.GET("/verify/phone", nc.VerifyPhoneNumber)
	
	// Common notification endpoints
	group.POST("/auth/email-verification", nc.SendEmailVerification)
	group.POST("/auth/password-reset", nc.SendPasswordReset)
	group.POST("/auth/login-alert", nc.SendLoginAlert)
	group.POST("/team/invitation", nc.SendTeamInvitation)
	group.POST("/team/role-changed", nc.SendRoleChanged)
	group.POST("/billing/payment-succeeded", nc.SendPaymentSucceeded)
	group.POST("/billing/payment-failed", nc.SendPaymentFailed)
	group.POST("/billing/trial-ending", nc.SendTrialEnding)
	
	// Development/testing endpoints
	group.POST("/test/email", nc.TestEmail)
}

// Basic notification endpoints

// SendEmail godoc
// @Summary Send an email
// @Description Send a single email to one or more recipients
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body notificationinterface.EmailRequest true "Email details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/email [post]
func (nc *NotificationController) SendEmail(c echo.Context) error {
	var req notificationinterface.EmailRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	if err := nc.notificationService.SendEmail(c.Request().Context(), &req); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Email sent successfully",
	})
}

// SendTemplateEmail godoc
// @Summary Send a template-based email
// @Description Send an email using a predefined template
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body notificationinterface.TemplateEmailRequest true "Template email details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/email/template [post]
func (nc *NotificationController) SendTemplateEmail(c echo.Context) error {
	var req notificationinterface.TemplateEmailRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	if err := nc.notificationService.SendTemplateEmail(c.Request().Context(), &req); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Template email sent successfully",
	})
}

// SendBulkEmails godoc
// @Summary Send bulk emails
// @Description Send emails to multiple recipients
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body notificationinterface.BulkEmailRequest true "Bulk email details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/email/bulk [post]
func (nc *NotificationController) SendBulkEmails(c echo.Context) error {
	var req notificationinterface.BulkEmailRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	if err := nc.notificationService.SendBulkEmails(c.Request().Context(), &req); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]interface{}{
		"message": "Bulk emails sent successfully",
		"count":   len(req.Recipients),
	})
}

// SendSMS godoc
// @Summary Send an SMS
// @Description Send an SMS message to one or more phone numbers
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body notificationinterface.SMSRequest true "SMS details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/sms [post]
func (nc *NotificationController) SendSMS(c echo.Context) error {
	var req notificationinterface.SMSRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	if err := nc.notificationService.SendSMS(c.Request().Context(), &req); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "SMS sent successfully",
	})
}

// SendPushNotification godoc
// @Summary Send a push notification
// @Description Send a push notification to one or more devices
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body notificationinterface.PushNotificationRequest true "Push notification details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/push [post]
func (nc *NotificationController) SendPushNotification(c echo.Context) error {
	var req notificationinterface.PushNotificationRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	if err := nc.notificationService.SendPushNotification(c.Request().Context(), &req); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Push notification sent successfully",
	})
}

// Verification endpoints

// VerifyEmail godoc
// @Summary Verify email address
// @Description Validate an email address format
// @Tags notifications
// @Accept json
// @Produce json
// @Param email query string true "Email address to verify"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/verify/email [get]
func (nc *NotificationController) VerifyEmail(c echo.Context) error {
	email := c.QueryParam("email")
	if email == "" {
		return core.Error(c, core.BadRequest("Email parameter is required"))
	}

	if err := nc.notificationService.VerifyEmail(c.Request().Context(), email); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Email is valid",
		"email":   email,
	})
}

// VerifyPhoneNumber godoc
// @Summary Verify phone number
// @Description Validate a phone number format
// @Tags notifications
// @Accept json
// @Produce json
// @Param phone query string true "Phone number to verify"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/verify/phone [get]
func (nc *NotificationController) VerifyPhoneNumber(c echo.Context) error {
	phone := c.QueryParam("phone")
	if phone == "" {
		return core.Error(c, core.BadRequest("Phone parameter is required"))
	}

	if err := nc.notificationService.VerifyPhoneNumber(c.Request().Context(), phone); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Phone number is valid",
		"phone":   phone,
	})
}

// Common notification endpoints

// SendEmailVerification godoc
// @Summary Send email verification
// @Description Send an email verification message
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body notificationmodel.EmailVerificationRequest true "Email verification details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/auth/email-verification [post]
func (nc *NotificationController) SendEmailVerification(c echo.Context) error {
	if nc.commonService == nil {
		return core.Error(c, core.Internal("Common notification service not configured"))
	}

	var req notificationmodel.EmailVerificationRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	if err := nc.commonService.SendEmailVerification(c.Request().Context(), req.Email, req.Token); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Email verification sent successfully",
	})
}

// SendPasswordReset godoc
// @Summary Send password reset email
// @Description Send a password reset email
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body notificationmodel.PasswordResetRequest true "Password reset details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/auth/password-reset [post]
func (nc *NotificationController) SendPasswordReset(c echo.Context) error {
	if nc.commonService == nil {
		return core.Error(c, core.Internal("Common notification service not configured"))
	}

	var req notificationmodel.PasswordResetRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	if err := nc.commonService.SendPasswordReset(c.Request().Context(), req.Email, req.Token); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Password reset email sent successfully",
	})
}

// SendLoginAlert godoc
// @Summary Send login alert
// @Description Send a login alert notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body map[string]string true "Login alert details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/auth/login-alert [post]
func (nc *NotificationController) SendLoginAlert(c echo.Context) error {
	if nc.commonService == nil {
		return core.Error(c, core.Internal("Common notification service not configured"))
	}

	var req struct {
		Email     string `json:"email" validate:"required,email"`
		IPAddress string `json:"ip_address" validate:"required"`
		UserAgent string `json:"user_agent" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	if err := nc.commonService.SendLoginAlert(c.Request().Context(), req.Email, req.IPAddress, req.UserAgent); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Login alert sent successfully",
	})
}

// SendTeamInvitation godoc
// @Summary Send team invitation
// @Description Send a team invitation email
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body notificationmodel.TeamInvitationRequest true "Team invitation details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/team/invitation [post]
func (nc *NotificationController) SendTeamInvitation(c echo.Context) error {
	if nc.commonService == nil {
		return core.Error(c, core.Internal("Common notification service not configured"))
	}

	var req notificationmodel.TeamInvitationRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	if err := nc.commonService.SendTeamInvitation(c.Request().Context(), req.Email, req.InviterName, req.TeamName, req.Role, req.Token); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Team invitation sent successfully",
	})
}

// SendRoleChanged godoc
// @Summary Send role changed notification
// @Description Send a role change notification email
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body notificationmodel.RoleChangedRequest true "Role change details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/team/role-changed [post]
func (nc *NotificationController) SendRoleChanged(c echo.Context) error {
	if nc.commonService == nil {
		return core.Error(c, core.Internal("Common notification service not configured"))
	}

	var req notificationmodel.RoleChangedRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	if err := nc.commonService.SendRoleChanged(c.Request().Context(), req.Email, req.UserName, req.TeamName, req.OldRole, req.NewRole); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Role change notification sent successfully",
	})
}

// SendPaymentSucceeded godoc
// @Summary Send payment success notification
// @Description Send a payment successful notification email
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body notificationmodel.PaymentSucceededRequest true "Payment success details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/billing/payment-succeeded [post]
func (nc *NotificationController) SendPaymentSucceeded(c echo.Context) error {
	if nc.commonService == nil {
		return core.Error(c, core.Internal("Common notification service not configured"))
	}

	var req notificationmodel.PaymentSucceededRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	if err := nc.commonService.SendPaymentSucceeded(c.Request().Context(), req.Email, req.PlanName, req.Amount, req.Currency, req.InvoiceURL); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Payment success notification sent successfully",
	})
}

// SendPaymentFailed godoc
// @Summary Send payment failure notification
// @Description Send a payment failed notification email
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body notificationmodel.PaymentFailedRequest true "Payment failure details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/billing/payment-failed [post]
func (nc *NotificationController) SendPaymentFailed(c echo.Context) error {
	if nc.commonService == nil {
		return core.Error(c, core.Internal("Common notification service not configured"))
	}

	var req notificationmodel.PaymentFailedRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	if err := nc.commonService.SendPaymentFailed(c.Request().Context(), req.Email, req.PlanName, req.Amount, req.Currency); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Payment failure notification sent successfully",
	})
}

// SendTrialEnding godoc
// @Summary Send trial ending notification
// @Description Send a trial ending notification email
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body notificationmodel.TrialEndingRequest true "Trial ending details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/billing/trial-ending [post]
func (nc *NotificationController) SendTrialEnding(c echo.Context) error {
	if nc.commonService == nil {
		return core.Error(c, core.Internal("Common notification service not configured"))
	}

	var req notificationmodel.TrialEndingRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	if err := nc.commonService.SendTrialEnding(c.Request().Context(), req.Email, req.DaysLeft); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Trial ending notification sent successfully",
	})
}

// TestEmail godoc
// @Summary Send a test email
// @Description Send a test email for development/testing purposes
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body notificationmodel.TestEmailRequest true "Test email details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notifications/test/email [post]
func (nc *NotificationController) TestEmail(c echo.Context) error {
	var req notificationmodel.TestEmailRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	if req.Subject == "" {
		req.Subject = "Test Email from Notification Service"
	}

	if req.Body == "" {
		req.Body = `
		<html>
		<body>
			<h2>Test Email</h2>
			<p>This is a test email from the notification service.</p>
			<p>If you received this, the email system is working correctly!</p>
		</body>
		</html>`
	}

	emailReq := &notificationinterface.EmailRequest{
		To:      []string{req.To},
		Subject: req.Subject,
		Body:    req.Body,
		IsHTML:  true,
	}

	if err := nc.notificationService.SendEmail(c.Request().Context(), emailReq); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Test email sent successfully",
		"to":      req.To,
	})
}