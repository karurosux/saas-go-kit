package gorm

import (
	rolemodel "{{.Project.GoModule}}/internal/role/model"
	"gorm.io/gorm"
)

// AutoMigrate runs database migrations for role module
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&rolemodel.DefaultRole{},
		&rolemodel.DefaultUserRole{},
	)
}