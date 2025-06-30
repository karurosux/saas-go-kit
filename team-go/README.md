# Team Management Module

A comprehensive team management module for SaaS applications, featuring role-based access control, invitation system, and permission management.

## Features

- **Role-Based Access Control**: Four predefined roles with customizable permissions
- **Team Invitations**: Secure token-based invitation system with expiration
- **User Management**: Complete user lifecycle management
- **Permission System**: Granular permission checking for different actions
- **Team Statistics**: Insights into team composition and activity
- **Notification Integration**: Pluggable notification system for team events

## Installation

```bash
go get github.com/karurosux/saas-go-kit/team-go
```

## Quick Start

### 1. Basic Setup

```go
package main

import (
    "github.com/karurosux/saas-go-kit/core-go"
    "github.com/karurosux/saas-go-kit/team-go"
    "gorm.io/gorm"
)

func main() {
    // Setup your database and repositories
    var db *gorm.DB // your database connection
    
    // Initialize repositories
    userRepo := NewUserRepository(db)
    teamMemberRepo := NewTeamMemberRepository(db)
    tokenRepo := NewInvitationTokenRepository(db)
    
    // Initialize optional services
    var notificationSvc team.NotificationService // your notification service
    var usageSvc team.UsageService               // your usage tracking service
    
    // Initialize team service
    teamService := team.NewTeamService(
        teamMemberRepo,
        userRepo,
        tokenRepo,
        notificationSvc,
        usageSvc,
    )
    
    // Create and mount module
    module := team.NewModule(team.ModuleConfig{
        TeamService:   teamService,
        RoutePrefix:   "/api/team",
        RequireAuth:   true,
    })
    
    app := core.NewKit(nil, core.KitConfig{})
    app.Register(module)
    app.Mount()
}
```

### 2. Define Your Repositories

```go
// Implement the required interfaces
type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) team.UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *team.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*team.User, error) {
    var user team.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    return &user, err
}

// ... implement other required methods
```

### 3. Invite Team Members

```go
// Invite a new team member
req := &team.InviteMemberRequest{
    AccountID: accountID,
    InviterID: currentUserID,
    Email:     "new.member@company.com",
    Role:      team.RoleManager,
}

member, err := teamService.InviteMember(ctx, req)
if err != nil {
    log.Printf("Failed to invite member: %v", err)
    return
}

fmt.Printf("Invited %s as %s\n", member.User.Email, member.Role)
```

### 4. Check Permissions

```go
// Check if user can perform an action
canInvite, err := teamService.CheckPermission(ctx, accountID, userID, "invite_members")
if err != nil {
    log.Printf("Error checking permission: %v", err)
    return
}

if !canInvite {
    return errors.Forbidden("invite team members")
}

// Get user's role
role, err := teamService.GetMemberRole(ctx, accountID, userID)
if err != nil {
    log.Printf("User is not a team member: %v", err)
    return
}

fmt.Printf("User role: %s\n", role)
```

## Roles and Permissions

### Predefined Roles

| Role | Description | Level |
|------|-------------|-------|
| **OWNER** | Full account control | 100 |
| **ADMIN** | Administrative privileges | 80 |
| **MANAGER** | Management capabilities | 60 |
| **VIEWER** | Read-only access | 40 |

### Permission System

Each role has specific permissions:

#### Owner Permissions
- `manage_account` - Manage account settings
- `invite_members` - Invite team members
- `remove_members` - Remove team members
- `change_roles` - Change member roles
- `view_billing` - View billing information
- `manage_billing` - Manage billing and subscriptions
- `view_analytics` - View analytics and reports
- `manage_integrations` - Manage integrations

#### Admin Permissions
- `invite_members` - Invite team members
- `remove_members` - Remove team members (except owners)
- `change_roles` - Change member roles (except owners)
- `view_billing` - View billing information
- `view_analytics` - View analytics and reports
- `manage_integrations` - Manage integrations

#### Manager Permissions
- `view_team` - View team members
- `view_analytics` - View analytics and reports

#### Viewer Permissions
- `view_team` - View team members
- `view_basic_analytics` - View basic analytics

### Custom Permission Checking

```go
// Check specific permissions
if team.RoleAdmin.HasPermission("invite_members") {
    // Admin can invite members
}

// Get all permissions for a role
permissions := team.RoleManager.GetPermissions()
for _, perm := range permissions {
    fmt.Printf("Permission: %s - %s\n", perm.Name, perm.Description)
}
```

## API Endpoints

### Team Management

- `GET /team/members` - List all team members
- `GET /team/members/:id` - Get specific team member
- `POST /team/members/invite` - Invite new team member
- `PUT /team/members/:id/role` - Update member role
- `DELETE /team/members/:id` - Remove team member

### Invitation Management

- `POST /team/accept-invite` - Accept team invitation (public)
- `POST /team/members/:id/resend-invite` - Resend invitation
- `DELETE /team/members/:id/cancel-invite` - Cancel invitation

### Team Information

- `GET /team/stats` - Get team statistics
- `GET /team/permissions/:permission` - Check user permission
- `GET /team/roles/:role/permissions` - Get role permissions

## Team Statistics

Get insights into your team:

```go
stats, err := teamService.GetTeamStats(ctx, accountID)
if err != nil {
    log.Printf("Failed to get stats: %v", err)
    return
}

fmt.Printf("Total members: %d\n", stats.TotalMembers)
fmt.Printf("Active members: %d\n", stats.ActiveMembers)
fmt.Printf("Pending invitations: %d\n", stats.PendingMembers)

// Role breakdown
for role, count := range stats.RoleBreakdown {
    fmt.Printf("%s: %d\n", role, count)
}
```

## Invitation System

### Token Generation

Invitations use secure, time-limited tokens:

```go
// Tokens are automatically generated and expire in 7 days
token, err := team.GenerateInvitationToken()
if err != nil {
    log.Printf("Failed to generate token: %v", err)
    return
}

// Check token validity
if invitationToken.IsValid() {
    // Token is valid and unused
}

if invitationToken.IsExpired() {
    // Token has expired
}

if invitationToken.IsUsed() {
    // Token has been used
}
```

### Accepting Invitations

```go
// Accept invitation with token
err := teamService.AcceptInvitation(ctx, token)
if err != nil {
    log.Printf("Failed to accept invitation: %v", err)
    return
}
```

## Notification Integration

Implement the `NotificationService` interface to send team-related notifications:

```go
type notificationService struct {
    emailSender EmailSender
}

func (s *notificationService) SendTeamInvitation(ctx context.Context, req *team.TeamInvitationNotification) error {
    return s.emailSender.SendInvitationEmail(
        req.Email,
        req.InviterName,
        req.TeamName,
        req.Role,
        req.Token,
    )
}

func (s *notificationService) SendRoleChanged(ctx context.Context, req *team.RoleChangedNotification) error {
    return s.emailSender.SendRoleChangeEmail(
        req.Email,
        req.UserName,
        req.OldRole,
        req.NewRole,
    )
}

func (s *notificationService) SendMemberRemoved(ctx context.Context, req *team.MemberRemovedNotification) error {
    return s.emailSender.SendRemovalEmail(
        req.Email,
        req.UserName,
        req.TeamName,
    )
}
```

## Usage Tracking Integration

Integrate with subscription limits:

```go
type usageService struct {
    subscriptionService SubscriptionService
}

func (s *usageService) CanAddMember(ctx context.Context, accountID uuid.UUID) (bool, string, error) {
    subscription, err := s.subscriptionService.GetSubscription(ctx, accountID)
    if err != nil {
        return false, "No subscription found", err
    }
    
    currentCount, err := s.teamMemberRepo.CountByAccountID(ctx, accountID)
    if err != nil {
        return false, "Failed to count members", err
    }
    
    limit := subscription.Plan.Features.GetLimit("max_team_members")
    if limit == -1 {
        return true, "", nil // Unlimited
    }
    
    if currentCount >= limit {
        return false, "Team member limit reached", nil
    }
    
    return true, "", nil
}

func (s *usageService) TrackMemberAdded(ctx context.Context, accountID uuid.UUID) error {
    return s.subscriptionService.TrackUsage(ctx, accountID, "team_member", 1)
}

func (s *usageService) TrackMemberRemoved(ctx context.Context, accountID uuid.UUID) error {
    return s.subscriptionService.TrackUsage(ctx, accountID, "team_member", -1)
}
```

## Database Schema

```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR UNIQUE NOT NULL,
    password_hash VARCHAR NOT NULL,
    first_name VARCHAR,
    last_name VARCHAR,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Team members table
CREATE TABLE team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL,
    user_id UUID REFERENCES users(id),
    role VARCHAR NOT NULL,
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMP NOT NULL,
    accepted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    UNIQUE(account_id, user_id)
);

-- Invitation tokens table
CREATE TABLE invitation_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL,
    member_id UUID REFERENCES team_members(id),
    token VARCHAR UNIQUE NOT NULL,
    email VARCHAR NOT NULL,
    role VARCHAR NOT NULL,
    invited_by UUID NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_team_members_account_id ON team_members(account_id);
CREATE INDEX idx_team_members_user_id ON team_members(user_id);
CREATE INDEX idx_invitation_tokens_token ON invitation_tokens(token);
CREATE INDEX idx_invitation_tokens_member_id ON invitation_tokens(member_id);
```

## Configuration

### Module Configuration

```go
type ModuleConfig struct {
    TeamService         TeamService
    RoutePrefix         string                                    // Default: "/team"
    RequireAuth         bool                                      // Whether to require authentication
    PermissionChecker   func(c echo.Context, permission string) bool // Custom permission checker
}
```

## Testing

```bash
# Run tests for the team module
cd team-go
go test ./...

# Run with coverage
go test -cover ./...
```

## Dependencies

- `github.com/google/uuid` - UUID support
- `github.com/labstack/echo/v4` - HTTP framework
- `golang.org/x/crypto` - Password hashing
- `gorm.io/gorm` - ORM for database operations
- `github.com/karurosux/saas-go-kit/core-go` - Core module system
- `github.com/karurosux/saas-go-kit/errors-go` - Error handling
- `github.com/karurosux/saas-go-kit/response-go` - Response formatting
- `github.com/karurosux/saas-go-kit/validator-go` - Request validation

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.