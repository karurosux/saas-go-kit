package gorm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/karurosux/saas-go-kit/subscription-go"
	"gorm.io/gorm"
)

// GenerateMigrationSQL generates SQL migration files for production use
func GenerateMigrationSQL(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	
	// Generate up migration
	upFile := filepath.Join(outputDir, fmt.Sprintf("%s_subscription_module_up.sql", timestamp))
	upSQL := generateUpSQL()
	
	if err := os.WriteFile(upFile, []byte(upSQL), 0644); err != nil {
		return fmt.Errorf("failed to write up migration: %w", err)
	}

	// Generate down migration
	downFile := filepath.Join(outputDir, fmt.Sprintf("%s_subscription_module_down.sql", timestamp))
	downSQL := generateDownSQL()
	
	if err := os.WriteFile(downFile, []byte(downSQL), 0644); err != nil {
		return fmt.Errorf("failed to write down migration: %w", err)
	}

	fmt.Printf("✅ Generated migration files:\n")
	fmt.Printf("   UP:   %s\n", upFile)
	fmt.Printf("   DOWN: %s\n", downFile)
	
	return nil
}

// GenerateMigrationSQLFromDB generates migration by comparing with existing database
func GenerateMigrationSQLFromDB(db *gorm.DB, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	
	// Use GORM's dry run to capture SQL
	dryRunDB := db.Session(&gorm.Session{DryRun: true})
	
	var statements []string
	
	// Capture create table statements
	models := []interface{}{
		&subscription.SubscriptionPlan{},
		&subscription.Subscription{},
		&subscription.SubscriptionUsage{},
		&subscription.UsageEvent{},
	}
	
	for _, model := range models {
		stmt := dryRunDB.Migrator().CreateTable(model)
		if stmt.Error == nil && stmt.Statement != nil && stmt.Statement.SQL.String() != "" {
			statements = append(statements, stmt.Statement.SQL.String()+";")
		}
	}
	
	// Add index creation statements
	indexStatements := getIndexStatements()
	statements = append(statements, indexStatements...)
	
	upSQL := "-- Subscription Module Migration - UP\n"
	upSQL += "-- Generated at: " + time.Now().Format("2006-01-02 15:04:05") + "\n\n"
	upSQL += strings.Join(statements, "\n\n") + "\n"
	
	upFile := filepath.Join(outputDir, fmt.Sprintf("%s_subscription_module_up.sql", timestamp))
	if err := os.WriteFile(upFile, []byte(upSQL), 0644); err != nil {
		return fmt.Errorf("failed to write up migration: %w", err)
	}

	// Generate down migration
	downSQL := generateDownSQL()
	downFile := filepath.Join(outputDir, fmt.Sprintf("%s_subscription_module_down.sql", timestamp))
	if err := os.WriteFile(downFile, []byte(downSQL), 0644); err != nil {
		return fmt.Errorf("failed to write down migration: %w", err)
	}

	fmt.Printf("✅ Generated migration files from database schema:\n")
	fmt.Printf("   UP:   %s\n", upFile)
	fmt.Printf("   DOWN: %s\n", downFile)
	
	return nil
}

func generateUpSQL() string {
	return `-- Subscription Module Migration - UP
-- Generated at: ` + time.Now().Format("2006-01-02 15:04:05") + `

-- Create subscription_plans table
CREATE TABLE IF NOT EXISTS subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    price DECIMAL(10,2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    interval VARCHAR(20) DEFAULT 'month',
    features JSONB,
    is_active BOOLEAN DEFAULT true,
    is_visible BOOLEAN DEFAULT true,
    trial_days INTEGER DEFAULT 0,
    stripe_price_id VARCHAR(255)
);

-- Create subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    account_id UUID NOT NULL,
    plan_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    cancel_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),
    FOREIGN KEY (plan_id) REFERENCES subscription_plans(id)
);

-- Create subscription_usages table
CREATE TABLE IF NOT EXISTS subscription_usages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    subscription_id UUID NOT NULL,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    feedbacks_count INTEGER DEFAULT 0,
    restaurants_count INTEGER DEFAULT 0,
    locations_count INTEGER DEFAULT 0,
    qr_codes_count INTEGER DEFAULT 0,
    team_members_count INTEGER DEFAULT 0,
    last_updated_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE
);

-- Create usage_events table
CREATE TABLE IF NOT EXISTS usage_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    subscription_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    metadata JSONB,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_plans_code ON subscription_plans(code);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_plans_active_visible ON subscription_plans(is_active, is_visible);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_plans_stripe_price_id ON subscription_plans(stripe_price_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_account_id ON subscriptions(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_stripe_subscription_id ON subscriptions(stripe_subscription_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_current_period ON subscriptions(current_period_start, current_period_end);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_usages_subscription_id ON subscription_usages(subscription_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_usages_period ON subscription_usages(subscription_id, period_start, period_end);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_events_subscription_id ON usage_events(subscription_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_events_resource_type ON usage_events(resource_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_events_created_at ON usage_events(created_at);
`
}

func generateDownSQL() string {
	return `-- Subscription Module Migration - DOWN
-- Generated at: ` + time.Now().Format("2006-01-02 15:04:05") + `

-- Drop indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_usage_events_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_usage_events_resource_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_usage_events_subscription_id;

DROP INDEX CONCURRENTLY IF EXISTS idx_subscription_usages_period;
DROP INDEX CONCURRENTLY IF EXISTS idx_subscription_usages_subscription_id;

DROP INDEX CONCURRENTLY IF EXISTS idx_subscriptions_current_period;
DROP INDEX CONCURRENTLY IF EXISTS idx_subscriptions_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_subscriptions_stripe_subscription_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_subscriptions_account_id;

DROP INDEX CONCURRENTLY IF EXISTS idx_subscription_plans_stripe_price_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_subscription_plans_active_visible;
DROP INDEX CONCURRENTLY IF EXISTS idx_subscription_plans_code;

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS usage_events;
DROP TABLE IF EXISTS subscription_usages;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS subscription_plans;
`
}

func getIndexStatements() []string {
	return []string{
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_plans_code ON subscription_plans(code);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_plans_active_visible ON subscription_plans(is_active, is_visible);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_plans_stripe_price_id ON subscription_plans(stripe_price_id);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_account_id ON subscriptions(account_id);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_stripe_subscription_id ON subscriptions(stripe_subscription_id);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_current_period ON subscriptions(current_period_start, current_period_end);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_usages_subscription_id ON subscription_usages(subscription_id);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_usages_period ON subscription_usages(subscription_id, period_start, period_end);",
		"CREATE INDEX CONCURRENTLY IF EXISTS idx_usage_events_subscription_id ON usage_events(subscription_id);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_events_resource_type ON usage_events(resource_type);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_events_created_at ON usage_events(created_at);",
	}
}