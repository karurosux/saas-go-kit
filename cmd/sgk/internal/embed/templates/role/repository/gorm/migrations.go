package gorm

import (
	rolemodel "{{.Project.GoModule}}/internal/role/model"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&rolemodel.DefaultRole{},
		&rolemodel.DefaultUserRole{},
	)
}