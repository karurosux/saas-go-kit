package gorm

import (
	subscriptionmodel "{{.Project.GoModule}}/internal/subscription/model"
	"gorm.io/gorm"
)

// AutoMigrate runs database migrations for subscription models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&subscriptionmodel.SubscriptionPlan{},
		&subscriptionmodel.Subscription{},
		&subscriptionmodel.Usage{},
		&subscriptionmodel.Invoice{},
	)
}