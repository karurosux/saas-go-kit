package {{.ModuleName}}model

import (
	"time"
	"gorm.io/gorm"
)

type {{.ModuleNameCap}} struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	Name        string         `json:"name" gorm:"not null"`
	Description *string        `json:"description"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func ({{.ModuleNameCap}}) TableName() string {
	return "{{.ModuleName}}s"
}