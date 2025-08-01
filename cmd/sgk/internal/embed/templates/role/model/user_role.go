package rolemodel

import (
	"time"

	roleinterface "{{.Project.GoModule}}/internal/role/interface"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DefaultUserRole struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	RoleID     uuid.UUID      `json:"role_id" gorm:"type:uuid;not null;index"`
	Role       *DefaultRole   `json:"role" gorm:"foreignKey:RoleID"`
	AssignedBy uuid.UUID      `json:"assigned_by" gorm:"type:uuid;not null"`
	AssignedAt time.Time      `json:"assigned_at"`
	ExpiresAt  *time.Time     `json:"expires_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (DefaultUserRole) TableName() string {
	return "user_roles"
}

func (ur *DefaultUserRole) GetID() uuid.UUID { return ur.ID }
func (ur *DefaultUserRole) GetUserID() uuid.UUID { return ur.UserID }
func (ur *DefaultUserRole) GetRoleID() uuid.UUID { return ur.RoleID }
func (ur *DefaultUserRole) GetRole() roleinterface.Role { 
	if ur.Role == nil {
		return nil
	}
	return ur.Role 
}
func (ur *DefaultUserRole) GetAssignedBy() uuid.UUID { return ur.AssignedBy }
func (ur *DefaultUserRole) GetAssignedAt() time.Time { return ur.AssignedAt }
func (ur *DefaultUserRole) GetExpiresAt() *time.Time { return ur.ExpiresAt }
func (ur *DefaultUserRole) IsActive() bool {
	return ur.ExpiresAt == nil || ur.ExpiresAt.After(time.Now())
}

func (ur *DefaultUserRole) BeforeCreate(tx *gorm.DB) error {
	if ur.ID == uuid.Nil {
		ur.ID = uuid.New()
	}
	if ur.AssignedAt.IsZero() {
		ur.AssignedAt = time.Now()
	}
	return nil
}