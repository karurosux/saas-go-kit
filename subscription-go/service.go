package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type subscriptionService struct {
	subscriptionRepo SubscriptionRepository
	planRepo         SubscriptionPlanRepository
	usageRepo        UsageRepository
}

func NewSubscriptionService(
	subscriptionRepo SubscriptionRepository,
	planRepo SubscriptionPlanRepository,
	usageRepo UsageRepository,
) SubscriptionService {
	return &subscriptionService{
		subscriptionRepo: subscriptionRepo,
		planRepo:         planRepo,
		usageRepo:        usageRepo,
	}
}

func (s *subscriptionService) GetUserSubscription(ctx context.Context, accountID uuid.UUID) (*Subscription, error) {
	return s.subscriptionRepo.FindByAccountID(ctx, accountID)
}

func (s *subscriptionService) GetAvailablePlans(ctx context.Context) ([]SubscriptionPlan, error) {
	return s.planRepo.FindAll(ctx)
}

func (s *subscriptionService) GetAllPlans(ctx context.Context) ([]SubscriptionPlan, error) {
	return s.planRepo.FindAllIncludingHidden(ctx)
}

func (s *subscriptionService) CreateSubscription(ctx context.Context, req *CreateSubscriptionRequest) (*Subscription, error) {
	plan, err := s.planRepo.FindByID(ctx, req.PlanID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	now := time.Now()
	subscription := &Subscription{
		AccountID:          req.AccountID,
		PlanID:             req.PlanID,
		Status:             SubscriptionActive,
		Plan:               *plan,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0),
	}

	err = s.subscriptionRepo.Create(ctx, subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return subscription, nil
}

func (s *subscriptionService) AssignCustomPlan(ctx context.Context, accountID uuid.UUID, planCode string) error {
	plan, err := s.planRepo.FindByCode(ctx, planCode)
	if err != nil {
		return fmt.Errorf("plan not found: %w", err)
	}

	existingSubscription, err := s.subscriptionRepo.FindByAccountID(ctx, accountID)
	if err == nil && existingSubscription != nil {
		existingSubscription.PlanID = plan.ID
		existingSubscription.Plan = *plan
		existingSubscription.Status = SubscriptionActive
		return s.subscriptionRepo.Update(ctx, existingSubscription)
	}

	now := time.Now()
	subscription := &Subscription{
		AccountID:          accountID,
		PlanID:             plan.ID,
		Status:             SubscriptionActive,
		Plan:               *plan,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0),
	}

	return s.subscriptionRepo.Create(ctx, subscription)
}

func (s *subscriptionService) CancelSubscription(ctx context.Context, accountID uuid.UUID) error {
	subscription, err := s.subscriptionRepo.FindByAccountID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	subscription.Status = SubscriptionCanceled
	return s.subscriptionRepo.Update(ctx, subscription)
}

func (s *subscriptionService) CanUserAccessResource(ctx context.Context, accountID uuid.UUID, resourceType string) (*PermissionResponse, error) {
	return s.CanUserAccessResourceWithLimitKey(ctx, accountID, resourceType, "")
}

func (s *subscriptionService) CanUserAccessResourceWithLimitKey(ctx context.Context, accountID uuid.UUID, resourceType string, limitKey string) (*PermissionResponse, error) {
	subscription, err := s.subscriptionRepo.FindByAccountID(ctx, accountID)
	if err != nil {
		return &PermissionResponse{
			CanCreate:          false,
			Reason:             "No active subscription found",
			SubscriptionStatus: "none",
		}, nil
	}

	if !subscription.IsActive() {
		return &PermissionResponse{
			CanCreate:          false,
			Reason:             "Subscription is not active or has expired",
			SubscriptionStatus: string(subscription.Status),
		}, nil
	}

	usage, err := s.getCurrentUsage(ctx, subscription.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage data: %w", err)
	}

	canAdd, reason := usage.CanAddResource(resourceType, subscription.Plan.Features, limitKey)
	var currentCount int64
	var maxAllowed int64

	if limitKey != "" {
		maxAllowed = subscription.Plan.Features.GetLimit(limitKey)
		switch resourceType {
		case ResourceTypeRestaurant:
			currentCount = int64(usage.RestaurantsCount)
		case ResourceTypeTeamMember:
			currentCount = int64(usage.TeamMembersCount)
		case ResourceTypeFeedback:
			currentCount = int64(usage.FeedbacksCount)
		case ResourceTypeLocation:
			currentCount = int64(usage.LocationsCount)
		case ResourceTypeQRCode:
			currentCount = int64(usage.QRCodesCount)
		}
	}

	return &PermissionResponse{
		CanCreate:          canAdd,
		Reason:             reason,
		CurrentCount:       int(currentCount),
		MaxAllowed:         int(maxAllowed),
		SubscriptionStatus: string(subscription.Status),
	}, nil
}

func (s *subscriptionService) getCurrentUsage(ctx context.Context, subscriptionID uuid.UUID) (*SubscriptionUsage, error) {
	subscription, err := s.subscriptionRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	usage, err := s.usageRepo.FindBySubscriptionAndPeriod(ctx, subscriptionID, subscription.CurrentPeriodStart, subscription.CurrentPeriodEnd)
	if err != nil {
		return &SubscriptionUsage{
			SubscriptionID:   subscriptionID,
			PeriodStart:      subscription.CurrentPeriodStart,
			PeriodEnd:        subscription.CurrentPeriodEnd,
			FeedbacksCount:   0,
			RestaurantsCount: 0,
			LocationsCount:   0,
			QRCodesCount:     0,
			TeamMembersCount: 0,
		}, nil
	}

	return usage, nil
}

type usageService struct {
	usageRepo        UsageRepository
	subscriptionRepo SubscriptionRepository
}

func NewUsageService(
	usageRepo UsageRepository,
	subscriptionRepo SubscriptionRepository,
) UsageService {
	return &usageService{
		usageRepo:        usageRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

func (s *usageService) TrackUsage(ctx context.Context, subscriptionID uuid.UUID, resourceType string, delta int) error {
	subscription, err := s.subscriptionRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	usage, err := s.usageRepo.FindBySubscriptionAndPeriod(ctx, subscriptionID, subscription.CurrentPeriodStart, subscription.CurrentPeriodEnd)
	if err != nil {
		usage = &SubscriptionUsage{
			SubscriptionID: subscriptionID,
			PeriodStart:    subscription.CurrentPeriodStart,
			PeriodEnd:      subscription.CurrentPeriodEnd,
		}
		if err := s.usageRepo.Create(ctx, usage); err != nil {
			return fmt.Errorf("failed to create usage record: %w", err)
		}
	}

	switch resourceType {
	case ResourceTypeFeedback:
		usage.FeedbacksCount += delta
	case ResourceTypeRestaurant:
		usage.RestaurantsCount += delta
	case ResourceTypeLocation:
		usage.LocationsCount += delta
	case ResourceTypeQRCode:
		usage.QRCodesCount += delta
	case ResourceTypeTeamMember:
		usage.TeamMembersCount += delta
	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}

	usage.LastUpdatedAt = time.Now()
	return s.usageRepo.Update(ctx, usage)
}

func (s *usageService) RecordUsageEvent(ctx context.Context, event *UsageEvent) error {
	return s.usageRepo.CreateEvent(ctx, event)
}

func (s *usageService) CanAddResource(ctx context.Context, subscriptionID uuid.UUID, resourceType string) (bool, string, error) {
	return s.CanAddResourceWithLimitKey(ctx, subscriptionID, resourceType, "")
}

func (s *usageService) CanAddResourceWithLimitKey(ctx context.Context, subscriptionID uuid.UUID, resourceType string, limitKey string) (bool, string, error) {
	subscription, err := s.subscriptionRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return false, "Subscription not found", err
	}

	if !subscription.IsActive() {
		return false, "Subscription is not active", nil
	}

	usage, err := s.GetCurrentUsage(ctx, subscriptionID)
	if err != nil {
		return false, "Failed to get usage data", err
	}

	canAdd, reason := usage.CanAddResource(resourceType, subscription.Plan.Features, limitKey)
	return canAdd, reason, nil
}

func (s *usageService) GetCurrentUsage(ctx context.Context, subscriptionID uuid.UUID) (*SubscriptionUsage, error) {
	subscription, err := s.subscriptionRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	usage, err := s.usageRepo.FindBySubscriptionAndPeriod(ctx, subscriptionID, subscription.CurrentPeriodStart, subscription.CurrentPeriodEnd)
	if err != nil {
		return &SubscriptionUsage{
			SubscriptionID:   subscriptionID,
			PeriodStart:      subscription.CurrentPeriodStart,
			PeriodEnd:        subscription.CurrentPeriodEnd,
			FeedbacksCount:   0,
			RestaurantsCount: 0,
			LocationsCount:   0,
			QRCodesCount:     0,
			TeamMembersCount: 0,
		}, nil
	}

	return usage, nil
}

func (s *usageService) GetUsageForPeriod(ctx context.Context, subscriptionID uuid.UUID, start, end time.Time) (*SubscriptionUsage, error) {
	return s.usageRepo.FindBySubscriptionAndPeriod(ctx, subscriptionID, start, end)
}

func (s *usageService) InitializeUsagePeriod(ctx context.Context, subscriptionID uuid.UUID, periodStart, periodEnd time.Time) error {
	existing, _ := s.usageRepo.FindBySubscriptionAndPeriod(ctx, subscriptionID, periodStart, periodEnd)
	if existing != nil {
		return nil
	}

	usage := &SubscriptionUsage{
		SubscriptionID: subscriptionID,
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
		LastUpdatedAt:  time.Now(),
	}

	return s.usageRepo.Create(ctx, usage)
}

func (s *usageService) ResetMonthlyUsage(ctx context.Context) error {
	return nil
}