module github.com/karurosux/saas-go-kit/subscription-go

go 1.21

require (
	github.com/google/uuid v1.6.0
	github.com/karurosux/saas-go-kit/core-go v0.0.0
	github.com/karurosux/saas-go-kit/errors-go v0.0.0
	github.com/karurosux/saas-go-kit/response-go v0.0.0
	github.com/labstack/echo/v4 v4.11.4
	github.com/stripe/stripe-go/v76 v76.25.0
	gorm.io/gorm v1.25.5
)

replace github.com/karurosux/saas-go-kit/core-go => ../core-go

replace github.com/karurosux/saas-go-kit/errors-go => ../errors-go

replace github.com/karurosux/saas-go-kit/response-go => ../response-go