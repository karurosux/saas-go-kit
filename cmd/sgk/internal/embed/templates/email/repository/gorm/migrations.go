package emailgorm

import (
	"gorm.io/gorm"
	emailinterface "{{.Project.GoModule}}/internal/email/interface"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&emailinterface.EmailMessage{},
		&emailinterface.EmailTemplate{},
	)
}