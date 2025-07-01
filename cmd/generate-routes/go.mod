module github.com/karurosux/saas-go-kit/cmd/generate-routes

go 1.21

require (
	github.com/karurosux/saas-go-kit/core-go v0.0.0
	gorm.io/driver/sqlite v1.5.4
	gorm.io/gorm v1.25.5
)

replace github.com/karurosux/saas-go-kit/core-go => ../../core-go