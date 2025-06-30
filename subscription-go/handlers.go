package subscription

import (
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/errors-go"
	"github.com/karurosux/saas-go-kit/response-go"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	subscriptionService SubscriptionService
	usageService        UsageService
	paymentService      PaymentService
}

func NewHandlers(
	subscriptionService SubscriptionService,
	usageService UsageService,
	paymentService PaymentService,
) *Handlers {
	return &Handlers{
		subscriptionService: subscriptionService,
		usageService:        usageService,
		paymentService:      paymentService,
	}
}

func (h *Handlers) GetUserSubscription(c echo.Context) error {
	accountIDStr := c.Get("account_id").(string)
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid account ID"))
	}

	subscription, err := h.subscriptionService.GetUserSubscription(c.Request().Context(), accountID)
	if err != nil {
		return response.Error(c, errors.NotFound("Subscription not found"))
	}

	return response.Success(c, subscription)
}

func (h *Handlers) GetAvailablePlans(c echo.Context) error {
	plans, err := h.subscriptionService.GetAvailablePlans(c.Request().Context())
	if err != nil {
		return response.Error(c, errors.Internal("Failed to get plans"))
	}

	return response.Success(c, plans)
}

func (h *Handlers) GetAllPlans(c echo.Context) error {
	plans, err := h.subscriptionService.GetAllPlans(c.Request().Context())
	if err != nil {
		return response.Error(c, errors.Internal("Failed to get plans"))
	}

	return response.Success(c, plans)
}

func (h *Handlers) CreateCheckoutSession(c echo.Context) error {
	var req CreateCheckoutRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request"))
	}

	accountIDStr := c.Get("account_id").(string)
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid account ID"))
	}
	req.AccountID = accountID

	session, err := h.paymentService.CreateCheckoutSession(c.Request().Context(), &req)
	if err != nil {
		return response.Error(c, errors.Internal("Failed to create checkout session"))
	}

	return response.Success(c, session)
}

func (h *Handlers) AssignCustomPlan(c echo.Context) error {
	planCode := c.Param("planCode")
	accountIDStr := c.Param("accountId")

	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid account ID"))
	}

	err = h.subscriptionService.AssignCustomPlan(c.Request().Context(), accountID, planCode)
	if err != nil {
		return response.Error(c, errors.Internal("Failed to assign plan"))
	}

	return response.Success(c, map[string]string{"message": "Plan assigned successfully"})
}

func (h *Handlers) CancelSubscription(c echo.Context) error {
	accountIDStr := c.Get("account_id").(string)
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid account ID"))
	}

	err = h.subscriptionService.CancelSubscription(c.Request().Context(), accountID)
	if err != nil {
		return response.Error(c, errors.Internal("Failed to cancel subscription"))
	}

	return response.Success(c, map[string]string{"message": "Subscription cancelled successfully"})
}

func (h *Handlers) CheckResourcePermission(c echo.Context) error {
	resourceType := c.Param("resourceType")
	accountIDStr := c.Get("account_id").(string)

	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid account ID"))
	}

	permission, err := h.subscriptionService.CanUserAccessResource(c.Request().Context(), accountID, resourceType)
	if err != nil {
		return response.Error(c, errors.Internal("Failed to check permissions"))
	}

	return response.Success(c, permission)
}

func (h *Handlers) GetCurrentUsage(c echo.Context) error {
	accountIDStr := c.Get("account_id").(string)
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid account ID"))
	}

	subscription, err := h.subscriptionService.GetUserSubscription(c.Request().Context(), accountID)
	if err != nil {
		return response.Error(c, errors.NotFound("Subscription not found"))
	}

	usage, err := h.usageService.GetCurrentUsage(c.Request().Context(), subscription.ID)
	if err != nil {
		return response.Error(c, errors.Internal("Failed to get usage"))
	}

	return response.Success(c, usage)
}

func (h *Handlers) CreatePortalSession(c echo.Context) error {
	var req struct {
		ReturnURL string `json:"return_url"`
	}
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request"))
	}

	accountIDStr := c.Get("account_id").(string)
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid account ID"))
	}

	session, err := h.paymentService.CreateCustomerPortalSession(c.Request().Context(), accountID, req.ReturnURL)
	if err != nil {
		return response.Error(c, errors.Internal("Failed to create portal session"))
	}

	return response.Success(c, session)
}

func (h *Handlers) GetPaymentMethods(c echo.Context) error {
	accountIDStr := c.Get("account_id").(string)
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid account ID"))
	}

	methods, err := h.paymentService.GetPaymentMethods(c.Request().Context(), accountID)
	if err != nil {
		return response.Error(c, errors.Internal("Failed to get payment methods"))
	}

	return response.Success(c, methods)
}

func (h *Handlers) GetInvoiceHistory(c echo.Context) error {
	accountIDStr := c.Get("account_id").(string)
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid account ID"))
	}

	limit := 20
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if err := echo.QueryParamsBinder(c).Int("limit", &limit).BindError(); err != nil {
			limit = 20
		}
	}

	invoices, err := h.paymentService.GetInvoiceHistory(c.Request().Context(), accountID, limit)
	if err != nil {
		return response.Error(c, errors.Internal("Failed to get invoice history"))
	}

	return response.Success(c, invoices)
}

func (h *Handlers) HandleWebhook(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid payload"))
	}

	signature := c.Request().Header.Get("Stripe-Signature")
	if signature == "" {
		return response.Error(c, errors.BadRequest("Missing signature"))
	}

	err = h.paymentService.HandleWebhookEvent(c.Request().Context(), body, signature)
	if err != nil {
		return response.Error(c, errors.Internal("Failed to process webhook"))
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handlers) GetFeatureRegistry(c echo.Context) error {
	return response.Success(c, GetAllFeatures())
}

func (h *Handlers) GetFeaturesByCategory(c echo.Context) error {
	category := c.Param("category")
	features := GetFeaturesByCategory(category)
	return response.Success(c, features)
}