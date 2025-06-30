package notification

import (
	"net/http"

	"github.com/karurosux/saas-go-kit/errors-go"
	"github.com/karurosux/saas-go-kit/response-go"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	notificationService NotificationService
	commonService       *CommonNotificationService
}

func NewHandlers(notificationService NotificationService, commonService *CommonNotificationService) *Handlers {
	return &Handlers{
		notificationService: notificationService,
		commonService:       commonService,
	}
}

// Send individual email
func (h *Handlers) SendEmail(c echo.Context) error {
	var req EmailRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
	}

	if err := h.notificationService.SendEmail(c.Request().Context(), &req); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Email sent successfully",
	})
}

// Send template-based email
func (h *Handlers) SendTemplateEmail(c echo.Context) error {
	var req TemplateEmailRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
	}

	if err := h.notificationService.SendTemplateEmail(c.Request().Context(), &req); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Template email sent successfully",
	})
}

// Send SMS
func (h *Handlers) SendSMS(c echo.Context) error {
	var req SMSRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
	}

	if err := h.notificationService.SendSMS(c.Request().Context(), &req); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "SMS sent successfully",
	})
}

// Send push notification
func (h *Handlers) SendPushNotification(c echo.Context) error {
	var req PushNotificationRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
	}

	if err := h.notificationService.SendPushNotification(c.Request().Context(), &req); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Push notification sent successfully",
	})
}

// Send bulk emails
func (h *Handlers) SendBulkEmails(c echo.Context) error {
	var req BulkEmailRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
	}

	if err := h.notificationService.SendBulkEmails(c.Request().Context(), &req); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Bulk emails sent successfully",
		"count":   string(len(req.Recipients)),
	})
}

// Verify email address
func (h *Handlers) VerifyEmail(c echo.Context) error {
	email := c.QueryParam("email")
	if email == "" {
		return response.Error(c, errors.BadRequest("Email parameter is required"))
	}

	if err := h.notificationService.VerifyEmail(c.Request().Context(), email); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Email is valid",
		"email":   email,
	})
}

// Verify phone number
func (h *Handlers) VerifyPhoneNumber(c echo.Context) error {
	phone := c.QueryParam("phone")
	if phone == "" {
		return response.Error(c, errors.BadRequest("Phone parameter is required"))
	}

	if err := h.notificationService.VerifyPhoneNumber(c.Request().Context(), phone); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Phone number is valid",
		"phone":   phone,
	})
}

// Common notification endpoints

func (h *Handlers) SendEmailVerification(c echo.Context) error {
	if h.commonService == nil {
		return response.Error(c, errors.Internal("Common notification service not configured"))
	}

	var req struct {
		Email string `json:"email" validate:"required,email"`
		Token string `json:"token" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
	}

	if err := h.commonService.SendEmailVerification(c.Request().Context(), req.Email, req.Token); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Email verification sent successfully",
	})
}

func (h *Handlers) SendPasswordReset(c echo.Context) error {
	if h.commonService == nil {
		return response.Error(c, errors.Internal("Common notification service not configured"))
	}

	var req struct {
		Email string `json:"email" validate:"required,email"`
		Token string `json:"token" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
	}

	if err := h.commonService.SendPasswordReset(c.Request().Context(), req.Email, req.Token); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Password reset email sent successfully",
	})
}

func (h *Handlers) SendTeamInvitation(c echo.Context) error {
	if h.commonService == nil {
		return response.Error(c, errors.Internal("Common notification service not configured"))
	}

	var req struct {
		Email       string `json:"email" validate:"required,email"`
		InviterName string `json:"inviter_name" validate:"required"`
		TeamName    string `json:"team_name" validate:"required"`
		Role        string `json:"role" validate:"required"`
		Token       string `json:"token" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
	}

	if err := h.commonService.SendTeamInvitation(c.Request().Context(), req.Email, req.InviterName, req.TeamName, req.Role, req.Token); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Team invitation sent successfully",
	})
}

func (h *Handlers) SendRoleChanged(c echo.Context) error {
	if h.commonService == nil {
		return response.Error(c, errors.Internal("Common notification service not configured"))
	}

	var req struct {
		Email    string `json:"email" validate:"required,email"`
		UserName string `json:"user_name" validate:"required"`
		TeamName string `json:"team_name" validate:"required"`
		OldRole  string `json:"old_role" validate:"required"`
		NewRole  string `json:"new_role" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
	}

	if err := h.commonService.SendRoleChanged(c.Request().Context(), req.Email, req.UserName, req.TeamName, req.OldRole, req.NewRole); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Role change notification sent successfully",
	})
}

func (h *Handlers) SendPaymentSucceeded(c echo.Context) error {
	if h.commonService == nil {
		return response.Error(c, errors.Internal("Common notification service not configured"))
	}

	var req struct {
		Email      string  `json:"email" validate:"required,email"`
		PlanName   string  `json:"plan_name" validate:"required"`
		Amount     float64 `json:"amount" validate:"required,min=0"`
		Currency   string  `json:"currency" validate:"required"`
		InvoiceURL string  `json:"invoice_url" validate:"required,url"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
	}

	if err := h.commonService.SendPaymentSucceeded(c.Request().Context(), req.Email, req.PlanName, req.Amount, req.Currency, req.InvoiceURL); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Payment success notification sent successfully",
	})
}

func (h *Handlers) SendPaymentFailed(c echo.Context) error {
	if h.commonService == nil {
		return response.Error(c, errors.Internal("Common notification service not configured"))
	}

	var req struct {
		Email    string  `json:"email" validate:"required,email"`
		PlanName string  `json:"plan_name" validate:"required"`
		Amount   float64 `json:"amount" validate:"required,min=0"`
		Currency string  `json:"currency" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
	}

	if err := h.commonService.SendPaymentFailed(c.Request().Context(), req.Email, req.PlanName, req.Amount, req.Currency); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Payment failure notification sent successfully",
	})
}

func (h *Handlers) SendTrialEnding(c echo.Context) error {
	if h.commonService == nil {
		return response.Error(c, errors.Internal("Common notification service not configured"))
	}

	var req struct {
		Email    string `json:"email" validate:"required,email"`
		DaysLeft int    `json:"days_left" validate:"required,min=1"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
	}

	if err := h.commonService.SendTrialEnding(c.Request().Context(), req.Email, req.DaysLeft); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Trial ending notification sent successfully",
	})
}

// Test endpoint for development
func (h *Handlers) TestEmail(c echo.Context) error {
	var req struct {
		To      string `json:"to" validate:"required,email"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request data"))
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

	emailReq := &EmailRequest{
		To:      []string{req.To},
		Subject: req.Subject,
		Body:    req.Body,
		IsHTML:  true,
	}

	if err := h.notificationService.SendEmail(c.Request().Context(), emailReq); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Test email sent successfully",
		"to":      req.To,
	})
}