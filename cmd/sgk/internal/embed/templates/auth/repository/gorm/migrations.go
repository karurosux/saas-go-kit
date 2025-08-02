package gorm

import (
	authmodel "{{.Project.GoModule}}/internal/auth/model"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&authmodel.Account{},
		&authmodel.Token{},
	)
}