package gorm

import (
	"{{.Project.GoModule}}/internal/auth/model"
	"gorm.io/gorm"
)

// AutoMigrate runs database migrations for auth models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&authmodel.Account{},
		&authmodel.Token{},
	)
}