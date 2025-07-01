# Auth Module Repositories

This directory contains repository implementations for the auth module.

## GORM Repository

The GORM repository provides PostgreSQL-compatible implementations for:

- `AccountStore` - Manages account persistence
- `TokenStore` - Manages verification token persistence

### Usage

```go
import (
    "github.com/karurosux/saas-go-kit/auth-go"
    authgorm "github.com/karurosux/saas-go-kit/auth-go/repositories/gorm"
    "gorm.io/gorm"
)

// Setup database connection
db, err := gorm.Open(/* your database config */)
if err != nil {
    log.Fatal(err)
}

// Run migrations
if err := authgorm.AutoMigrate(db); err != nil {
    log.Fatal(err)
}

// Create repositories
accountStore := authgorm.NewAccountRepository(db)
tokenStore := authgorm.NewTokenRepository(db)

// Create auth service
authService := auth.NewService(
    accountStore,
    tokenStore,
    emailProvider,
    configProvider,
)
```

### Migrations

The GORM repository includes auto-migration support and SQL migration generation:

```go
// Auto-migrate (development)
err := authgorm.AutoMigrate(db)

// Generate SQL migrations (production)
err := authgorm.GenerateMigrationSQL("./migrations")
```

### Database Schema

#### Accounts Table (`default_accounts`)

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| email | VARCHAR(255) | Unique email address |
| password_hash | VARCHAR(255) | Bcrypt password hash |
| company_name | VARCHAR(255) | Optional company name |
| phone | VARCHAR(50) | Optional phone number |
| active | BOOLEAN | Account active status |
| email_verified | BOOLEAN | Email verification status |
| email_verified_at | TIMESTAMP | Email verification timestamp |
| deactivated_at | TIMESTAMP | Deactivation timestamp |
| scheduled_deletion_at | TIMESTAMP | Scheduled deletion timestamp |
| metadata | JSONB | Additional account metadata |
| created_at | TIMESTAMP | Creation timestamp |
| updated_at | TIMESTAMP | Last update timestamp |

#### Verification Tokens Table (`default_verification_tokens`)

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| account_id | UUID | Foreign key to accounts |
| token | VARCHAR(255) | Unique verification token |
| type | VARCHAR(50) | Token type (EMAIL_VERIFICATION, PASSWORD_RESET, etc.) |
| expires_at | TIMESTAMP | Token expiration |
| used_at | TIMESTAMP | Token usage timestamp |
| created_at | TIMESTAMP | Creation timestamp |