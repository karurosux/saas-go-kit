package gorm

import (
	"fmt"

	"github.com/karurosux/saas-go-kit/auth-go"
	"gorm.io/gorm"
)

// AutoMigrate runs all necessary migrations for auth module
func AutoMigrate(db *gorm.DB) error {
	// Create schema if it doesn't exist
	if err := db.Exec("CREATE SCHEMA IF NOT EXISTS auth").Error; err != nil {
		return fmt.Errorf("failed to create auth schema: %w", err)
	}

	err := db.AutoMigrate(
		&auth.DefaultAccount{},
		&auth.DefaultVerificationToken{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate auth tables: %w", err)
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
		// User indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_users_email ON auth.users(email)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_users_active ON auth.users(active)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_users_email_verified ON auth.users(email_verified)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_users_created_at ON auth.users(created_at)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_users_deactivated_at ON auth.users(deactivated_at)",

		// Verification token indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_token ON auth.verification_tokens(token)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_account_id ON auth.verification_tokens(account_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_type ON auth.verification_tokens(type)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_expires_at ON auth.verification_tokens(expires_at)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_used_at ON auth.verification_tokens(used_at)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_account_type ON auth.verification_tokens(account_id, type)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			// Log warning but don't fail - index might already exist
			fmt.Printf("Warning: Could not create index: %v\n", err)
		}
	}

	return nil
}

// DropTables drops all auth-related tables (useful for testing)
func DropTables(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&auth.DefaultVerificationToken{},
		&auth.DefaultAccount{},
	)
}