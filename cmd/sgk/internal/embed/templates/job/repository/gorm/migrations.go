package gorm

import (
	"{{.Project.GoModule}}/internal/job/model"
	"gorm.io/gorm"
)

// AutoMigrate runs database migrations for job models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&jobmodel.Job{},
		&jobmodel.JobResult{},
	)
}