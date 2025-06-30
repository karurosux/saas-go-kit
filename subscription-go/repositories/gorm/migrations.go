package gorm

import (
	"fmt"

	"github.com/karurosux/saas-go-kit/subscription-go"
	"gorm.io/gorm"
)

// AutoMigrate runs all necessary migrations for subscription module
func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&subscription.Subscription{},
		&subscription.SubscriptionPlan{},
		&subscription.SubscriptionUsage{},
		&subscription.UsageEvent{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate subscription tables: %w", err)
	}

	// Create indexes for better performance
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// createIndexes creates performance-critical indexes
func createIndexes(db *gorm.DB) error {
	indexes := []string{
		// Subscription indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_account_id ON subscriptions(account_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_stripe_subscription_id ON subscriptions(stripe_subscription_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_status ON subscriptions(status)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_current_period ON subscriptions(current_period_start, current_period_end)",

		// Subscription plan indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_plans_code ON subscription_plans(code)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_plans_active_visible ON subscription_plans(is_active, is_visible)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_plans_stripe_price_id ON subscription_plans(stripe_price_id)",

		// Usage indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_usages_subscription_id ON subscription_usages(subscription_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_usages_period ON subscription_usages(subscription_id, period_start, period_end)",

		// Usage event indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_events_subscription_id ON usage_events(subscription_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_events_resource_type ON usage_events(resource_type)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_events_created_at ON usage_events(created_at)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			// Log warning but don't fail - index might already exist
			fmt.Printf("Warning: Could not create index: %v\n", err)
		}
	}

	return nil
}

// DropTables drops all subscription-related tables (useful for testing)
func DropTables(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&subscription.UsageEvent{},
		&subscription.SubscriptionUsage{},
		&subscription.Subscription{},
		&subscription.SubscriptionPlan{},
	)
}