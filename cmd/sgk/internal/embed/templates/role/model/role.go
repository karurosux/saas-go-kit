package rolemodel

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionList represents a list of permission tags that can be stored in JSONB
type PermissionList []string

// Value implements driver.Valuer interface for database storage  
func (p PermissionList) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements sql.Scanner interface for database retrieval
func (p *PermissionList) Scan(value interface{}) error {
	if value == nil {
		*p = PermissionList{}
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	}
	
	return fmt.Errorf("cannot scan %T into PermissionList", value)
}

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

// TableName specifies the table name for GORM
func (DefaultRole) TableName() string {
	return "roles"
}

// Implement Role interface
func (r *DefaultRole) GetID() uuid.UUID { return r.ID }
func (r *DefaultRole) GetName() string { return r.Name }
func (r *DefaultRole) GetDescription() string { return r.Description }
func (r *DefaultRole) GetPermissions() []string { return []string(r.Permissions) }
func (r *DefaultRole) IsSystemRole() bool { return r.IsSystem }
func (r *DefaultRole) GetCreatedAt() time.Time { return r.CreatedAt }
func (r *DefaultRole) GetUpdatedAt() time.Time { return r.UpdatedAt }

func (r *DefaultRole) HasPermission(permission string) bool {
	for _, p := range r.Permissions {
		if p == permission || p == "*" {
			return true
		}
		// Check wildcard permissions (e.g., "users:*" matches "users:read")
		if strings.HasSuffix(p, ":*") && strings.HasPrefix(permission, strings.TrimSuffix(p, "*")) {
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

// BeforeCreate hook for GORM
func (r *DefaultRole) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}