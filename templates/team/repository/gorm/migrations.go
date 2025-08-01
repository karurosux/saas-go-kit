package gorm

import (
	"{{.Project.GoModule}}/internal/team/model"
	"gorm.io/gorm"
)

// AutoMigrate runs database migrations for team models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&teammodel.User{},
		&teammodel.TeamMember{},
		&teammodel.InvitationToken{},
	)
}