package gorm

import (
	"fmt"

	"github.com/karurosux/saas-go-kit/team-go"
	"gorm.io/gorm"
)

// AutoMigrate runs all necessary migrations for team module
func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&team.User{},
		&team.TeamMember{},
		&team.InvitationToken{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate team tables: %w", err)
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
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_is_active ON users(is_active)",

		// Team member indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_account_id ON team_members(account_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_user_id ON team_members(user_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_user_account ON team_members(user_id, account_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_role ON team_members(role)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_invited_by ON team_members(invited_by)",

		// Invitation token indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_token ON invitation_tokens(token)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_member_id ON invitation_tokens(member_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_email ON invitation_tokens(email)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_expires_at ON invitation_tokens(expires_at)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_used_at ON invitation_tokens(used_at)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			// Log warning but don't fail - index might already exist
			fmt.Printf("Warning: Could not create index: %v\n", err)
		}
	}

	return nil
}

// DropTables drops all team-related tables (useful for testing)
func DropTables(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&team.InvitationToken{},
		&team.TeamMember{},
		&team.User{},
	)
}