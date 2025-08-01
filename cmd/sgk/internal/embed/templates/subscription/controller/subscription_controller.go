package subscriptioncontroller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"{{.Project.GoModule}}/internal/core"
	"time"

	subscriptioninterface "{{.Project.GoModule}}/internal/subscription/interface"
	subscriptionmiddleware "{{.Project.GoModule}}/internal/subscription/middleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type SubscriptionController struct {
	service subscriptioninterface.SubscriptionService
}

func NewSubscriptionController(service subscriptioninterface.SubscriptionService) *SubscriptionController {
	return &SubscriptionController{
		service: service,
	}
}

func (sc *SubscriptionController) handleError(c echo.Context, err error) error {
	var appErr *core.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case core.ErrCodeNotFound:
			return core.NotFound(c, err)
		case core.ErrCodeUnauthorized:
			return core.Unauthorized(c, err)
		case core.ErrCodeForbidden:
			return core.Forbidden(c, err)
		case core.ErrCodeBadRequest, core.ErrCodeValidation:
			return core.BadRequest(c, err)
		case core.ErrCodeConflict:
			return core.Error(c, http.StatusConflict, err)
		default:
			return core.InternalServerError(c, err)
		}
	}
	return core.InternalServerError(c, err)
}

func (sc *SubscriptionController) RegisterRoutes(e *echo.Echo, basePath string, subMiddleware *subscriptionmiddleware.SubscriptionMiddleware) {
	group := e.Group(basePath)

	group.GET("/plans", sc.GetPlans)
	group.GET("/plans/:planId", sc.GetPlan)

	group.POST("/webhook/stripe", sc.HandleStripeWebhook)

	accountGroup := group.Group("/accounts/:accountId")
	accountGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accountIDStr := c.Param("accountId")
			accountID, err := uuid.Parse(accountIDStr)
			if err != nil {
				return core.BadRequest(c, fmt.Errorf("invalid account ID"))
			}
			subscriptionmiddleware.SetAccountIDInContext(c, accountID)
			return next(c)
		}
	})

	accountGroup.GET("/subscription", sc.GetSubscription)
	accountGroup.POST("/subscription", sc.CreateSubscription)
	accountGroup.PUT("/subscription", sc.UpdateSubscription, subMiddleware.RequireActiveSubscription())
	accountGroup.DELETE("/subscription", sc.CancelSubscription, subMiddleware.RequireActiveSubscription())
	accountGroup.POST("/subscription/reactivate", sc.ReactivateSubscription)

	accountGroup.GET("/usage", sc.GetUsageReport, subMiddleware.RequireActiveSubscription())
	accountGroup.GET("/usage/:resource", sc.GetResourceUsage, subMiddleware.RequireActiveSubscription())

	accountGroup.GET("/invoices", sc.GetInvoices, subMiddleware.RequireActiveSubscription())
	accountGroup.POST("/checkout", sc.CreateCheckoutSession)
	accountGroup.POST("/billing-portal", sc.CreateBillingPortalSession, subMiddleware.RequireActiveSubscription())
}

// GetPlans godoc
// @Summary Get subscription plans
// @Description Get all available subscription plans
// @Tags subscription
// @Accept json
// @Produce json
// @Success 200 {array} subscriptionmodel.SubscriptionPlan
// @Failure 500 {object} core.ErrorResponse
// @Router /subscriptions/plans [get]
func (sc *SubscriptionController) GetPlans(c echo.Context) error {
	plans, err := sc.service.GetPlans(c.Request().Context())
	if err != nil {
		return sc.handleError(c, err)
	}

	return core.Success(c, plans)
}

// GetPlan godoc
// @Summary Get subscription plan
// @Description Get a specific subscription plan
// @Tags subscription
// @Accept json
// @Produce json
// @Param planId path string true "Plan ID"
// @Success 200 {object} subscriptionmodel.SubscriptionPlan
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /subscriptions/plans/{planId} [get]
func (sc *SubscriptionController) GetPlan(c echo.Context) error {
	planID, err := uuid.Parse(c.Param("planId"))
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid plan ID"))
	}

	plan, err := sc.service.GetPlan(c.Request().Context(), planID)
	if err != nil {
		return sc.handleError(c, err)
	}

	return core.Success(c, plan)
}

// GetSubscription godoc
// @Summary Get account subscription
// @Description Get the current subscription for an account
// @Tags subscription
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Success 200 {object} subscriptionmodel.Subscription
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /subscriptions/accounts/{accountId}/subscription [get]
func (sc *SubscriptionController) GetSubscription(c echo.Context) error {
	accountID, err := subscriptionmiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid account ID"))
	}

	subscription, err := sc.service.GetSubscription(c.Request().Context(), accountID)
	if err != nil {
		return sc.handleError(c, err)
	}

	return core.Success(c, subscription)
}

type CreateSubscriptionRequest struct {
	PlanID          uuid.UUID                           `json:"plan_id" validate:"required"`
	BillingPeriod   subscriptioninterface.BillingPeriod `json:"billing_period" validate:"required,oneof=monthly yearly"`
	PaymentMethodID string                              `json:"payment_method_id,omitempty"`
	CustomerEmail   string                              `json:"customer_email" validate:"required,email"`
}

// CreateSubscription godoc
// @Summary Create subscription
// @Description Create a new subscription for an account
// @Tags subscription
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param request body CreateSubscriptionRequest true "Subscription details"
// @Success 201 {object} subscriptionmodel.Subscription
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /subscriptions/accounts/{accountId}/subscription [post]
func (sc *SubscriptionController) CreateSubscription(c echo.Context) error {
	accountID, err := subscriptionmiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid account ID"))
	}

	var req CreateSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid request body"))
	}

	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}

	createReq := subscriptioninterface.ServiceCreateSubscriptionRequest{
		AccountID:       accountID,
		PlanID:          req.PlanID,
		BillingPeriod:   req.BillingPeriod,
		PaymentMethodID: req.PaymentMethodID,
		CustomerEmail:   req.CustomerEmail,
	}

	subscription, err := sc.service.CreateSubscription(c.Request().Context(), createReq)
	if err != nil {
		return sc.handleError(c, err)
	}

	return core.Created(c, subscription)
}

type UpdateSubscriptionRequest struct {
	PlanID        uuid.UUID                           `json:"plan_id" validate:"required"`
	BillingPeriod subscriptioninterface.BillingPeriod `json:"billing_period" validate:"required,oneof=monthly yearly"`
}

// UpdateSubscription godoc
// @Summary Update subscription
// @Description Update an existing subscription (change plan)
// @Tags subscription
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param request body UpdateSubscriptionRequest true "Update details"
// @Success 200 {object} subscriptionmodel.Subscription
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /subscriptions/accounts/{accountId}/subscription [put]
func (sc *SubscriptionController) UpdateSubscription(c echo.Context) error {
	accountID, err := subscriptionmiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid account ID"))
	}

	var req UpdateSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid request body"))
	}

	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}

	updateReq := subscriptioninterface.ServiceUpdateSubscriptionRequest{
		NewPlanID:     req.PlanID,
		BillingPeriod: req.BillingPeriod,
	}

	subscription, err := sc.service.UpdateSubscription(c.Request().Context(), accountID, updateReq)
	if err != nil {
		return sc.handleError(c, err)
	}

	return core.Success(c, subscription)
}

type CancelSubscriptionRequest struct {
	Immediately bool `json:"immediately"`
}

// CancelSubscription godoc
// @Summary Cancel subscription
// @Description Cancel an active subscription
// @Tags subscription
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param request body CancelSubscriptionRequest false "Cancellation options"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /subscriptions/accounts/{accountId}/subscription [delete]
func (sc *SubscriptionController) CancelSubscription(c echo.Context) error {
	accountID, err := subscriptionmiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid account ID"))
	}

	var req CancelSubscriptionRequest
	c.Bind(&req) // Optional body

	if err := sc.service.CancelSubscription(c.Request().Context(), accountID, req.Immediately); err != nil {
		return sc.handleError(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Subscription cancelled successfully",
	})
}

// ReactivateSubscription godoc
// @Summary Reactivate subscription
// @Description Reactivate a cancelled subscription
// @Tags subscription
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /subscriptions/accounts/{accountId}/subscription/reactivate [post]
func (sc *SubscriptionController) ReactivateSubscription(c echo.Context) error {
	accountID, err := subscriptionmiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid account ID"))
	}

	if err := sc.service.ReactivateSubscription(c.Request().Context(), accountID); err != nil {
		return sc.handleError(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "Subscription reactivated successfully",
	})
}

// GetUsageReportRequest represents usage report request
type GetUsageReportRequest struct {
	StartDate time.Time `query:"start_date" validate:"required"`
	EndDate   time.Time `query:"end_date" validate:"required"`
}

// GetUsageReport godoc
// @Summary Get usage report
// @Description Get usage report for an account
// @Tags subscription
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param start_date query string true "Start date (RFC3339)"
// @Param end_date query string true "End date (RFC3339)"
// @Success 200 {object} subscriptioninterface.UsageReport
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /subscriptions/accounts/{accountId}/usage [get]
func (sc *SubscriptionController) GetUsageReport(c echo.Context) error {
	accountID, err := subscriptionmiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid account ID"))
	}

	var req GetUsageReportRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid request parameters"))
	}

	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}

	report, err := sc.service.GetUsageReport(c.Request().Context(), accountID, req.StartDate, req.EndDate)
	if err != nil {
		return sc.handleError(c, err)
	}

	return core.Success(c, report)
}

// GetResourceUsage godoc
// @Summary Get resource usage
// @Description Get current usage for a specific resource
// @Tags subscription
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param resource path string true "Resource name"
// @Success 200 {object} map[string]any
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /subscriptions/accounts/{accountId}/usage/{resource} [get]
func (sc *SubscriptionController) GetResourceUsage(c echo.Context) error {
	accountID, err := subscriptionmiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid account ID"))
	}

	resource := c.Param("resource")
	if resource == "" {
		return core.BadRequest(c, fmt.Errorf("resource name required"))
	}

	usage, err := sc.service.GetUsage(c.Request().Context(), accountID, resource)
	if err != nil {
		return sc.handleError(c, err)
	}

	// Also check limit
	withinLimit, remaining, _ := sc.service.CheckLimit(c.Request().Context(), accountID, resource)

	return core.Success(c, map[string]any{
		"resource":     resource,
		"usage":        usage,
		"remaining":    remaining,
		"within_limit": withinLimit,
	})
}

// GetInvoices godoc
// @Summary Get invoices
// @Description Get account invoices
// @Tags subscription
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Success 200 {array} subscriptionmodel.Invoice
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /subscriptions/accounts/{accountId}/invoices [get]
func (sc *SubscriptionController) GetInvoices(c echo.Context) error {
	accountID, err := subscriptionmiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid account ID"))
	}

	invoices, err := sc.service.GetInvoices(c.Request().Context(), accountID)
	if err != nil {
		return sc.handleError(c, err)
	}

	return core.Success(c, invoices)
}

// CreateCheckoutSessionRequest represents checkout session request
type CreateCheckoutSessionRequest struct {
	PlanID        uuid.UUID                           `json:"plan_id" validate:"required"`
	BillingPeriod subscriptioninterface.BillingPeriod `json:"billing_period" validate:"required,oneof=monthly yearly"`
	SuccessURL    string                              `json:"success_url" validate:"required,url"`
	CancelURL     string                              `json:"cancel_url" validate:"required,url"`
}

// CreateCheckoutSession godoc
// @Summary Create checkout session
// @Description Create a Stripe checkout session
// @Tags subscription
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param request body CreateCheckoutSessionRequest true "Checkout details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /subscriptions/accounts/{accountId}/checkout [post]
func (sc *SubscriptionController) CreateCheckoutSession(c echo.Context) error {
	accountID, err := subscriptionmiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid account ID"))
	}

	var req CreateCheckoutSessionRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid request body"))
	}

	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}

	checkoutReq := subscriptioninterface.CheckoutRequest{
		PlanID:        req.PlanID,
		BillingPeriod: req.BillingPeriod,
		SuccessURL:    req.SuccessURL,
		CancelURL:     req.CancelURL,
	}

	sessionURL, err := sc.service.CreateCheckoutSession(c.Request().Context(), accountID, checkoutReq)
	if err != nil {
		return sc.handleError(c, err)
	}

	return core.Success(c, map[string]string{
		"checkout_url": sessionURL,
	})
}

// CreateBillingPortalSession godoc
// @Summary Create billing portal session
// @Description Create a Stripe billing portal session
// @Tags subscription
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /subscriptions/accounts/{accountId}/billing-portal [post]
func (sc *SubscriptionController) CreateBillingPortalSession(c echo.Context) error {
	accountID, err := subscriptionmiddleware.GetAccountIDFromContext(c)
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid account ID"))
	}

	portalURL, err := sc.service.CreateBillingPortalSession(c.Request().Context(), accountID)
	if err != nil {
		return sc.handleError(c, err)
	}

	return core.Success(c, map[string]string{
		"portal_url": portalURL,
	})
}

// HandleStripeWebhook godoc
// @Summary Handle Stripe webhook
// @Description Handle incoming Stripe webhook events
// @Tags subscription
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /subscriptions/webhook/stripe [post]
func (sc *SubscriptionController) HandleStripeWebhook(c echo.Context) error {
	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("failed to read request body"))
	}

	signature := c.Request().Header.Get("Stripe-Signature")
	if signature == "" {
		return core.BadRequest(c, fmt.Errorf("missing stripe signature"))
	}

	if err := sc.service.HandleStripeWebhook(c.Request().Context(), payload, signature); err != nil {
		return sc.handleError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "success",
	})
}
