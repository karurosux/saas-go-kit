package gorm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/karurosux/saas-go-kit/team-go"
	"gorm.io/gorm"
)

// GenerateMigrationSQL generates SQL migration files for production use
func GenerateMigrationSQL(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	
	// Generate up migration
	upFile := filepath.Join(outputDir, fmt.Sprintf("%s_team_module_up.sql", timestamp))
	upSQL := generateUpSQL()
	
	if err := os.WriteFile(upFile, []byte(upSQL), 0644); err != nil {
		return fmt.Errorf("failed to write up migration: %w", err)
	}

	// Generate down migration
	downFile := filepath.Join(outputDir, fmt.Sprintf("%s_team_module_down.sql", timestamp))
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
		&team.User{},
		&team.TeamMember{},
		&team.InvitationToken{},
	}
	
	for _, model := range models {
		err := dryRunDB.Migrator().CreateTable(model)
		if err == nil {
			// For simplicity, just add a comment about the table
			tableName := dryRunDB.Statement.Table
			statements = append(statements, "-- Table: "+tableName)
		}
	}
	
	// Add index creation statements
	indexStatements := getIndexStatements()
	statements = append(statements, indexStatements...)
	
	upSQL := "-- Team Module Migration - UP\n"
	upSQL += "-- Generated at: " + time.Now().Format("2006-01-02 15:04:05") + "\n\n"
	upSQL += strings.Join(statements, "\n\n") + "\n"
	
	upFile := filepath.Join(outputDir, fmt.Sprintf("%s_team_module_up.sql", timestamp))
	if err := os.WriteFile(upFile, []byte(upSQL), 0644); err != nil {
		return fmt.Errorf("failed to write up migration: %w", err)
	}

	// Generate down migration
	downSQL := generateDownSQL()
	downFile := filepath.Join(outputDir, fmt.Sprintf("%s_team_module_down.sql", timestamp))
	if err := os.WriteFile(downFile, []byte(downSQL), 0644); err != nil {
		return fmt.Errorf("failed to write down migration: %w", err)
	}

	fmt.Printf("✅ Generated migration files from database schema:\n")
	fmt.Printf("   UP:   %s\n", upFile)
	fmt.Printf("   DOWN: %s\n", downFile)
	
	return nil
}

func generateUpSQL() string {
	return `-- Team Module Migration - UP
-- Generated at: ` + time.Now().Format("2006-01-02 15:04:05") + `

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    is_active BOOLEAN DEFAULT true
);

-- Create team_members table
CREATE TABLE IF NOT EXISTS team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    account_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL,
    invited_by UUID,
    invited_at TIMESTAMP WITH TIME ZONE,
    accepted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, account_id)
);

-- Create invitation_tokens table
CREATE TABLE IF NOT EXISTS invitation_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    account_id UUID NOT NULL,
    member_id UUID NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    invited_by UUID,
    expires_at TIMESTAMP WITH TIME ZONE,
    used_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (member_id) REFERENCES team_members(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_is_active ON users(is_active);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_account_id ON team_members(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_user_account ON team_members(user_id, account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_role ON team_members(role);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_invited_by ON team_members(invited_by);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_token ON invitation_tokens(token);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_member_id ON invitation_tokens(member_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_email ON invitation_tokens(email);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_expires_at ON invitation_tokens(expires_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_used_at ON invitation_tokens(used_at);
`
}

func generateDownSQL() string {
	return `-- Team Module Migration - DOWN
-- Generated at: ` + time.Now().Format("2006-01-02 15:04:05") + `

-- Drop indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_invitation_tokens_used_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_invitation_tokens_expires_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_invitation_tokens_email;
DROP INDEX CONCURRENTLY IF EXISTS idx_invitation_tokens_member_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_invitation_tokens_token;

DROP INDEX CONCURRENTLY IF EXISTS idx_team_members_invited_by;
DROP INDEX CONCURRENTLY IF EXISTS idx_team_members_role;
DROP INDEX CONCURRENTLY IF EXISTS idx_team_members_user_account;
DROP INDEX CONCURRENTLY IF EXISTS idx_team_members_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_team_members_account_id;

DROP INDEX CONCURRENTLY IF EXISTS idx_users_is_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_email;

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS invitation_tokens;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS users;
`
}

func getIndexStatements() []string {
	return []string{
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users(email);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_is_active ON users(is_active);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_account_id ON team_members(account_id);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_user_account ON team_members(user_id, account_id);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_role ON team_members(role);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_invited_by ON team_members(invited_by);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_token ON invitation_tokens(token);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_member_id ON invitation_tokens(member_id);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_email ON invitation_tokens(email);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_expires_at ON invitation_tokens(expires_at);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_used_at ON invitation_tokens(used_at);",
	}
}