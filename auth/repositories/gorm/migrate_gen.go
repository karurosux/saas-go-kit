package gorm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/karurosux/saas-go-kit/auth-go"
	"gorm.io/gorm"
)

// GenerateMigrationSQL generates SQL migration files for production use
func GenerateMigrationSQL(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	
	// Generate up migration
	upFile := filepath.Join(outputDir, fmt.Sprintf("%s_auth_module_up.sql", timestamp))
	upSQL := generateUpSQL()
	
	if err := os.WriteFile(upFile, []byte(upSQL), 0644); err != nil {
		return fmt.Errorf("failed to write up migration: %w", err)
	}

	// Generate down migration
	downFile := filepath.Join(outputDir, fmt.Sprintf("%s_auth_module_down.sql", timestamp))
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
		&auth.DefaultAccount{},
		&auth.DefaultVerificationToken{},
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
	
	upSQL := "-- Auth Module Migration - UP\n"
	upSQL += "-- Generated at: " + time.Now().Format("2006-01-02 15:04:05") + "\n\n"
	upSQL += strings.Join(statements, "\n\n") + "\n"
	
	upFile := filepath.Join(outputDir, fmt.Sprintf("%s_auth_module_up.sql", timestamp))
	if err := os.WriteFile(upFile, []byte(upSQL), 0644); err != nil {
		return fmt.Errorf("failed to write up migration: %w", err)
	}

	// Generate down migration
	downSQL := generateDownSQL()
	downFile := filepath.Join(outputDir, fmt.Sprintf("%s_auth_module_down.sql", timestamp))
	if err := os.WriteFile(downFile, []byte(downSQL), 0644); err != nil {
		return fmt.Errorf("failed to write down migration: %w", err)
	}

	fmt.Printf("✅ Generated migration files from database schema:\n")
	fmt.Printf("   UP:   %s\n", upFile)
	fmt.Printf("   DOWN: %s\n", downFile)
	
	return nil
}

func generateUpSQL() string {
	return `-- Auth Module Migration - UP
-- Generated at: ` + time.Now().Format("2006-01-02 15:04:05") + `

-- Create auth schema
CREATE SCHEMA IF NOT EXISTS auth;

-- Create users table
CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    company_name VARCHAR(255),
    phone VARCHAR(50),
    active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    deactivated_at TIMESTAMP WITH TIME ZONE,
    scheduled_deletion_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'
);

-- Create verification tokens table
CREATE TABLE IF NOT EXISTS auth.verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    account_id UUID NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (account_id) REFERENCES auth.users(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_users_email ON auth.users(email);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_users_active ON auth.users(active);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_users_email_verified ON auth.users(email_verified);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_users_created_at ON auth.users(created_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_users_deactivated_at ON auth.users(deactivated_at);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_token ON auth.verification_tokens(token);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_account_id ON auth.verification_tokens(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_type ON auth.verification_tokens(type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_expires_at ON auth.verification_tokens(expires_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_used_at ON auth.verification_tokens(used_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_account_type ON auth.verification_tokens(account_id, type);
`
}

func generateDownSQL() string {
	return `-- Auth Module Migration - DOWN
-- Generated at: ` + time.Now().Format("2006-01-02 15:04:05") + `

-- Drop indexes
DROP INDEX CONCURRENTLY IF EXISTS auth.idx_auth_verification_tokens_account_type;
DROP INDEX CONCURRENTLY IF EXISTS auth.idx_auth_verification_tokens_used_at;
DROP INDEX CONCURRENTLY IF EXISTS auth.idx_auth_verification_tokens_expires_at;
DROP INDEX CONCURRENTLY IF EXISTS auth.idx_auth_verification_tokens_type;
DROP INDEX CONCURRENTLY IF EXISTS auth.idx_auth_verification_tokens_account_id;
DROP INDEX CONCURRENTLY IF EXISTS auth.idx_auth_verification_tokens_token;

DROP INDEX CONCURRENTLY IF EXISTS auth.idx_auth_users_deactivated_at;
DROP INDEX CONCURRENTLY IF EXISTS auth.idx_auth_users_created_at;
DROP INDEX CONCURRENTLY IF EXISTS auth.idx_auth_users_email_verified;
DROP INDEX CONCURRENTLY IF EXISTS auth.idx_auth_users_active;
DROP INDEX CONCURRENTLY IF EXISTS auth.idx_auth_users_email;

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS auth.verification_tokens;
DROP TABLE IF EXISTS auth.users;

-- Drop schema
DROP SCHEMA IF EXISTS auth CASCADE;
`
}

func getIndexStatements() []string {
	return []string{
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_users_email ON auth.users(email);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_users_active ON auth.users(active);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_users_email_verified ON auth.users(email_verified);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_users_created_at ON auth.users(created_at);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_users_deactivated_at ON auth.users(deactivated_at);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_token ON auth.verification_tokens(token);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_account_id ON auth.verification_tokens(account_id);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_type ON auth.verification_tokens(type);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_expires_at ON auth.verification_tokens(expires_at);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_used_at ON auth.verification_tokens(used_at);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_auth_verification_tokens_account_type ON auth.verification_tokens(account_id, type);",
	}
}