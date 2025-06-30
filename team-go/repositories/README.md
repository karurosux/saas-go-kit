# Repository Implementations

This directory contains ready-to-use repository implementations for team management functionality, maintaining clean architecture interfaces while providing production-ready database access patterns.

## 🚀 Quick Start

### Using GORM Implementation

```go
package main

import (
    "context"
    "log"
    
    "github.com/karurosux/saas-go-kit/team-go"
    gormrepo "github.com/karurosux/saas-go-kit/team-go/repositories/gorm"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    // Database connection
    db, err := gorm.Open(postgres.Open("your-dsn"), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }
    
    // Auto-migrate models with indexes
    if err := gormrepo.AutoMigrate(db); err != nil {
        log.Fatal("Migration failed:", err)
    }
    
    // Create repository instances
    userRepo := gormrepo.NewUserRepository(db)
    memberRepo := gormrepo.NewTeamMemberRepository(db)
    tokenRepo := gormrepo.NewInvitationTokenRepository(db)
    
    // Create notification service (implement based on your needs)
    notificationService := &MyNotificationService{}
    
    // Create team service
    teamService := team.NewTeamService(
        userRepo,
        memberRepo,
        tokenRepo,
        notificationService,
    )
    
    // Use the service
    ctx := context.Background()
    accountID := uuid.New()
    
    members, err := teamService.ListMembers(ctx, accountID)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d team members\n", len(members))
}
```

## 📋 Available Implementations

### GORM Implementation (`./gorm/`)

**Supports:**
- PostgreSQL, MySQL, SQLite, SQL Server
- User management with password hashing
- Team member relationships and roles
- Invitation token management with expiration
- Team statistics and role breakdowns
- Soft deletes and audit trails

**Files:**
- `user_repository.go` - User CRUD operations with email lookup
- `team_member_repository.go` - Team membership management with statistics
- `invitation_token_repository.go` - Invitation lifecycle management

**Features:**
- ✅ All interface methods implemented
- ✅ Proper error handling with domain-specific errors
- ✅ Context support for cancellation and timeouts
- ✅ Optimized queries with preloading
- ✅ Automatic cleanup of expired tokens
- ✅ Team statistics calculation

## 🛠️ Creating Custom Implementations

### Example: Redis-Backed User Session Repository

```go
package redis

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/go-redis/redis/v8"
    "github.com/google/uuid"
    "github.com/karurosux/saas-go-kit/team-go"
)

type RedisUserRepository struct {
    primary team.UserRepository
    redis   *redis.Client
    ttl     time.Duration
}

func NewRedisUserRepository(
    primary team.UserRepository,
    redis *redis.Client,
    ttl time.Duration,
) team.UserRepository {
    return &RedisUserRepository{
        primary: primary,
        redis:   redis,
        ttl:     ttl,
    }
}

func (r *RedisUserRepository) FindByEmail(ctx context.Context, email string) (*team.User, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("user:email:%s", email)
    cached, err := r.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var user team.User
        if json.Unmarshal([]byte(cached), &user) == nil {
            return &user, nil
        }
    }
    
    // Fall back to primary repository
    user, err := r.primary.FindByEmail(ctx, email)
    if err != nil {
        return nil, err
    }
    
    // Cache the result (exclude password hash for security)
    userCache := *user
    userCache.PasswordHash = "" // Never cache password hash
    if data, err := json.Marshal(userCache); err == nil {
        r.redis.Set(ctx, cacheKey, data, r.ttl)
    }
    
    return user, nil
}

func (r *RedisUserRepository) Update(ctx context.Context, user *team.User) error {
    err := r.primary.Update(ctx, user)
    if err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("user:email:%s", user.Email)
    r.redis.Del(ctx, cacheKey)
    
    return nil
}

// Implement other methods...
```

### Example: Custom Team Analytics Repository

```go
package analytics

import (
    "context"
    "time"
    
    "github.com/google/uuid"
    "github.com/karurosux/saas-go-kit/team-go"
)

type AnalyticsTeamMemberRepository struct {
    primary team.TeamMemberRepository
    analytics AnalyticsService
}

func NewAnalyticsTeamMemberRepository(
    primary team.TeamMemberRepository,
    analytics AnalyticsService,
) team.TeamMemberRepository {
    return &AnalyticsTeamMemberRepository{
        primary: primary,
        analytics: analytics,
    }
}

func (r *AnalyticsTeamMemberRepository) Create(ctx context.Context, member *team.TeamMember) error {
    err := r.primary.Create(ctx, member)
    if err != nil {
        return err
    }
    
    // Track team member addition
    r.analytics.Track("team_member_added", map[string]interface{}{
        "account_id": member.AccountID,
        "role":       string(member.Role),
        "invited_by": member.InvitedBy,
    })
    
    return nil
}

func (r *AnalyticsTeamMemberRepository) Delete(ctx context.Context, id uuid.UUID) error {
    // Get member details before deletion for analytics
    member, err := r.primary.FindByID(ctx, id)
    if err != nil {
        return err
    }
    
    err = r.primary.Delete(ctx, id)
    if err != nil {
        return err
    }
    
    // Track team member removal
    r.analytics.Track("team_member_removed", map[string]interface{}{
        "account_id": member.AccountID,
        "role":       string(member.Role),
        "was_active": member.IsActive(),
    })
    
    return nil
}

// Implement other methods...
```

### Example: Multi-Tenant Team Repository

```go
package multitenant

import (
    "context"
    "fmt"
    
    "github.com/google/uuid"
    "github.com/karurosux/saas-go-kit/team-go"
    "gorm.io/gorm"
)

type MultiTenantTeamMemberRepository struct {
    db *gorm.DB
}

func NewMultiTenantTeamMemberRepository(db *gorm.DB) team.TeamMemberRepository {
    return &MultiTenantTeamMemberRepository{db: db}
}

func (r *MultiTenantTeamMemberRepository) FindByAccountID(ctx context.Context, accountID uuid.UUID) ([]team.TeamMember, error) {
    // Add tenant isolation
    tenantID := getTenantIDFromContext(ctx)
    if tenantID == "" {
        return nil, fmt.Errorf("tenant ID required")
    }
    
    var members []team.TeamMember
    err := r.db.WithContext(ctx).
        Preload("User").
        Where("account_id = ? AND tenant_id = ?", accountID, tenantID).
        Order("created_at ASC").
        Find(&members).Error
    
    return members, err
}

func (r *MultiTenantTeamMemberRepository) Create(ctx context.Context, member *team.TeamMember) error {
    tenantID := getTenantIDFromContext(ctx)
    if tenantID == "" {
        return fmt.Errorf("tenant ID required")
    }
    
    // Add tenant ID to the model
    member.TenantID = tenantID
    return r.db.WithContext(ctx).Create(member).Error
}

func getTenantIDFromContext(ctx context.Context) string {
    if tenantID, ok := ctx.Value("tenant_id").(string); ok {
        return tenantID
    }
    return ""
}

// Implement other methods...
```

## 🏗️ Repository Interface Design

All repository implementations must satisfy these interfaces:

```go
// UserRepository defines the interface for user data access
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id uuid.UUID, preloads ...string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id uuid.UUID) error
}

// TeamMemberRepository defines the interface for team member data access
type TeamMemberRepository interface {
    Create(ctx context.Context, member *TeamMember) error
    FindByID(ctx context.Context, id uuid.UUID, preloads ...string) (*TeamMember, error)
    FindByAccountID(ctx context.Context, accountID uuid.UUID) ([]TeamMember, error)
    FindByUserAndAccount(ctx context.Context, userID, accountID uuid.UUID) (*TeamMember, error)
    Update(ctx context.Context, member *TeamMember) error
    Delete(ctx context.Context, id uuid.UUID) error
    CountByAccountID(ctx context.Context, accountID uuid.UUID) (int64, error)
    GetTeamStats(ctx context.Context, accountID uuid.UUID) (*TeamStats, error)
}

// InvitationTokenRepository defines the interface for invitation token data access
type InvitationTokenRepository interface {
    Create(ctx context.Context, token *InvitationToken) error
    FindByToken(ctx context.Context, token string) (*InvitationToken, error)
    FindByMemberID(ctx context.Context, memberID uuid.UUID) (*InvitationToken, error)
    Update(ctx context.Context, token *InvitationToken) error
    Delete(ctx context.Context, id uuid.UUID) error
    DeleteExpired(ctx context.Context) error
}
```

## ⚡ Performance Considerations

### Optimized Queries

```go
// Good: Load user data with team memberships
members, err := repo.FindByAccountID(ctx, accountID)
// This automatically preloads User data

// Good: Get stats efficiently
stats, err := repo.GetTeamStats(ctx, accountID)
// Single query with aggregation

// Good: Check membership efficiently
member, err := repo.FindByUserAndAccount(ctx, userID, accountID)
// Indexed lookup
```

### Caching Strategies

```go
// Cache team statistics (changes infrequently)
type CachedTeamStatsRepository struct {
    primary team.TeamMemberRepository
    cache   map[uuid.UUID]*team.TeamStats
    mu      sync.RWMutex
}

func (r *CachedTeamStatsRepository) GetTeamStats(ctx context.Context, accountID uuid.UUID) (*team.TeamStats, error) {
    r.mu.RLock()
    if stats, exists := r.cache[accountID]; exists {
        r.mu.RUnlock()
        return stats, nil
    }
    r.mu.RUnlock()
    
    stats, err := r.primary.GetTeamStats(ctx, accountID)
    if err != nil {
        return nil, err
    }
    
    r.mu.Lock()
    r.cache[accountID] = stats
    r.mu.Unlock()
    
    return stats, nil
}
```

## 🔐 Security Considerations

### Password Handling

```go
// Always hash passwords before storing
func (r *UserRepository) Create(ctx context.Context, user *team.User) error {
    if user.PasswordHash == "" {
        return errors.New("password hash required")
    }
    
    // Ensure password is hashed (bcrypt handles multiple hashing gracefully)
    if err := user.SetPassword(user.PasswordHash); err != nil {
        return err
    }
    
    return r.db.WithContext(ctx).Create(user).Error
}
```

### Token Security

```go
// Generate secure invitation tokens
func (s *TeamService) InviteMember(ctx context.Context, req *InviteMemberRequest) error {
    // Generate cryptographically secure token
    token, err := team.GenerateInvitationToken()
    if err != nil {
        return err
    }
    
    invitationToken := &team.InvitationToken{
        Token:     token,
        Email:     req.Email,
        Role:      req.Role,
        ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
    }
    
    return s.tokenRepo.Create(ctx, invitationToken)
}
```

### Input Validation

```go
// Validate email format and role
func (r *TeamMemberRepository) Create(ctx context.Context, member *team.TeamMember) error {
    if !member.Role.IsValid() {
        return team.ErrInvalidRole
    }
    
    return r.db.WithContext(ctx).Create(member).Error
}
```

## 🧪 Testing Strategies

### Unit Testing with Mocks

```go
func TestTeamService_InviteMember(t *testing.T) {
    // Create mock repositories
    userRepo := &MockUserRepository{}
    memberRepo := &MockTeamMemberRepository{}
    tokenRepo := &MockInvitationTokenRepository{}
    notificationSvc := &MockNotificationService{}
    
    service := team.NewTeamService(userRepo, memberRepo, tokenRepo, notificationSvc)
    
    // Set up expectations
    userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, team.ErrUserNotFound)
    memberRepo.On("Create", mock.Anything, mock.AnythingOfType("*team.TeamMember")).Return(nil)
    tokenRepo.On("Create", mock.Anything, mock.AnythingOfType("*team.InvitationToken")).Return(nil)
    notificationSvc.On("SendTeamInvitation", mock.Anything, mock.Anything).Return(nil)
    
    // Test invitation
    req := &team.InviteMemberRequest{
        AccountID: uuid.New(),
        InviterID: uuid.New(),
        Email:     "test@example.com",
        Role:      team.RoleViewer,
    }
    
    member, err := service.InviteMember(context.Background(), req)
    assert.NoError(t, err)
    assert.NotNil(t, member)
    
    // Verify all expectations were met
    userRepo.AssertExpectations(t)
    memberRepo.AssertExpectations(t)
    tokenRepo.AssertExpectations(t)
    notificationSvc.AssertExpectations(t)
}
```

### Integration Testing

```go
func TestTeamRepositories_Integration(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    userRepo := gormrepo.NewUserRepository(db)
    memberRepo := gormrepo.NewTeamMemberRepository(db)
    tokenRepo := gormrepo.NewInvitationTokenRepository(db)
    
    ctx := context.Background()
    
    // Create test user
    user := &team.User{
        Email:     "test@example.com",
        FirstName: "Test",
        LastName:  "User",
    }
    user.SetPassword("password123")
    
    err := userRepo.Create(ctx, user)
    assert.NoError(t, err)
    
    // Create team member
    accountID := uuid.New()
    member := &team.TeamMember{
        AccountID: accountID,
        UserID:    user.ID,
        Role:      team.RoleViewer,
        InvitedBy: uuid.New(),
        InvitedAt: time.Now(),
    }
    
    err = memberRepo.Create(ctx, member)
    assert.NoError(t, err)
    
    // Test finding members
    members, err := memberRepo.FindByAccountID(ctx, accountID)
    assert.NoError(t, err)
    assert.Len(t, members, 1)
    assert.Equal(t, user.Email, members[0].User.Email)
    
    // Test team stats
    stats, err := memberRepo.GetTeamStats(ctx, accountID)
    assert.NoError(t, err)
    assert.Equal(t, 1, stats.TotalMembers)
    assert.Equal(t, 0, stats.ActiveMembers) // No accepted_at date
    assert.Equal(t, 1, stats.PendingMembers)
}
```

## 📊 Team Statistics and Reporting

### Advanced Team Analytics

```go
type AdvancedTeamMemberRepository struct {
    *gormrepo.TeamMemberRepository
}

func (r *AdvancedTeamMemberRepository) GetDetailedTeamStats(ctx context.Context, accountID uuid.UUID) (*DetailedTeamStats, error) {
    var stats DetailedTeamStats
    
    // Get basic stats
    basicStats, err := r.TeamMemberRepository.GetTeamStats(ctx, accountID)
    if err != nil {
        return nil, err
    }
    stats.TeamStats = *basicStats
    
    // Get invitation conversion rate
    var totalInvitations, acceptedInvitations int64
    r.db.Model(&team.TeamMember{}).
        Where("account_id = ?", accountID).
        Count(&totalInvitations)
    
    r.db.Model(&team.TeamMember{}).
        Where("account_id = ? AND accepted_at IS NOT NULL", accountID).
        Count(&acceptedInvitations)
    
    if totalInvitations > 0 {
        stats.ConversionRate = float64(acceptedInvitations) / float64(totalInvitations) * 100
    }
    
    // Get average time to accept invitation
    var avgAcceptTime sql.NullFloat64
    r.db.Model(&team.TeamMember{}).
        Select("AVG(EXTRACT(EPOCH FROM (accepted_at - invited_at)))").
        Where("account_id = ? AND accepted_at IS NOT NULL", accountID).
        Scan(&avgAcceptTime)
    
    if avgAcceptTime.Valid {
        stats.AvgAcceptanceTimeHours = avgAcceptTime.Float64 / 3600
    }
    
    return &stats, nil
}

type DetailedTeamStats struct {
    team.TeamStats
    ConversionRate          float64 `json:"conversion_rate"`
    AvgAcceptanceTimeHours  float64 `json:"avg_acceptance_time_hours"`
}
```

## 🔧 Configuration Examples

### Database Optimization

```go
// Configure GORM for team management
func setupTeamDatabase() *gorm.DB {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Create indexes for performance
    db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_account_id ON team_members(account_id)")
    db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_members_user_account ON team_members(user_id, account_id)")
    db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_token ON invitation_tokens(token)")
    db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invitation_tokens_expires ON invitation_tokens(expires_at)")
    
    // Set up connection pool
    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    return db
}
```

This repository pattern enables rapid team management development while maintaining the flexibility to customize for specific business requirements.