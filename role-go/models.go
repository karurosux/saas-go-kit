package role

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DefaultRole represents a role in the system with tag-based permissions
type DefaultRole struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	Description string         `json:"description"`
	Permissions PermissionList `json:"permissions" gorm:"type:jsonb"`
	IsSystem    bool           `json:"is_system" gorm:"default:false"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// DefaultUserRole represents the assignment of a role to a user
type DefaultUserRole struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	RoleID     uuid.UUID      `json:"role_id" gorm:"type:uuid;not null;index"`
	Role       DefaultRole    `json:"role" gorm:"foreignKey:RoleID"`
	AssignedBy uuid.UUID      `json:"assigned_by" gorm:"type:uuid;not null"`
	AssignedAt time.Time      `json:"assigned_at" gorm:"default:now()"`
	ExpiresAt  *time.Time     `json:"expires_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// PermissionList is a custom type for handling JSON array of permissions
type PermissionList []string

// Value implements the driver.Valuer interface for database storage
func (p PermissionList) Value() (driver.Value, error) {
	if len(p) == 0 {
		return "[]", nil
	}
	return json.Marshal(p)
}

// Scan implements the sql.Scanner interface for database reading
func (p *PermissionList) Scan(value interface{}) error {
	if value == nil {
		*p = []string{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into PermissionList", value)
	}

	return json.Unmarshal(bytes, p)
}

// Table names with schema prefix
func (DefaultRole) TableName() string {
	return "role.roles"
}

func (DefaultUserRole) TableName() string {
	return "role.user_roles"
}

// Role interface implementations
func (r *DefaultRole) GetID() uuid.UUID {
	return r.ID
}

func (r *DefaultRole) GetName() string {
	return r.Name
}

func (r *DefaultRole) GetDescription() string {
	return r.Description
}

func (r *DefaultRole) GetPermissions() []string {
	return []string(r.Permissions)
}

func (r *DefaultRole) IsSystemRole() bool {
	return r.IsSystem
}

func (r *DefaultRole) GetCreatedAt() time.Time {
	return r.CreatedAt
}

func (r *DefaultRole) GetUpdatedAt() time.Time {
	return r.UpdatedAt
}

// Permission checking methods with wildcard support
func (r *DefaultRole) HasPermission(permission string) bool {
	for _, p := range r.Permissions {
		if matchesPermission(permission, p) {
			return true
		}
	}
	return false
}

func (r *DefaultRole) HasAnyPermission(permissions []string) bool {
	for _, permission := range permissions {
		if r.HasPermission(permission) {
			return true
		}
	}
	return false
}

func (r *DefaultRole) HasAllPermissions(permissions []string) bool {
	for _, permission := range permissions {
		if !r.HasPermission(permission) {
			return false
		}
	}
	return true
}

// UserRole interface implementations
func (ur *DefaultUserRole) GetID() uuid.UUID {
	return ur.ID
}

func (ur *DefaultUserRole) GetUserID() uuid.UUID {
	return ur.UserID
}

func (ur *DefaultUserRole) GetRoleID() uuid.UUID {
	return ur.RoleID
}

func (ur *DefaultUserRole) GetRole() Role {
	return &ur.Role
}

func (ur *DefaultUserRole) GetAssignedBy() uuid.UUID {
	return ur.AssignedBy
}

func (ur *DefaultUserRole) GetAssignedAt() time.Time {
	return ur.AssignedAt
}

func (ur *DefaultUserRole) GetExpiresAt() *time.Time {
	return ur.ExpiresAt
}

func (ur *DefaultUserRole) IsActive() bool {
	now := time.Now()
	return ur.DeletedAt.Time.IsZero() && (ur.ExpiresAt == nil || ur.ExpiresAt.After(now))
}

// Permission matching logic with wildcard support
func matchesPermission(requested, granted string) bool {
	// Exact match
	if requested == granted {
		return true
	}

	// Global wildcard
	if granted == "*" {
		return true
	}

	// Resource wildcard (e.g., "users:*" matches "users:read", "users:write", etc.)
	if strings.HasSuffix(granted, ":*") {
		resource := strings.TrimSuffix(granted, ":*")
		if strings.HasPrefix(requested, resource+":") {
			return true
		}
	}

	// Action wildcard (e.g., "*:read" matches "users:read", "posts:read", etc.)
	if strings.HasPrefix(granted, "*:") {
		action := strings.TrimPrefix(granted, "*:")
		if strings.HasSuffix(requested, ":"+action) {
			return true
		}
	}

	return false
}