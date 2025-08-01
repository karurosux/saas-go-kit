package gorm

import (
	"github.com/karurosux/saas-go-kit/job-go"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&job.JobModel{},
		&job.JobResultModel{},
	)
}