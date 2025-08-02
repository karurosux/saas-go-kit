package emailgorm

import (
	"gorm.io/gorm"
	emailinterface "{{.Project.GoModule}}/internal/email/interface"
)

// AutoMigrate runs auto migrations for email models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&emailinterface.EmailMessage{},
		&emailinterface.EmailTemplate{},
	)
}