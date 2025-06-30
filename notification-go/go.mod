module github.com/karurosux/saas-go-kit/notification-go

go 1.21

require (
	github.com/karurosux/saas-go-kit/core-go v0.0.0
	github.com/karurosux/saas-go-kit/errors-go v0.0.0
	github.com/labstack/echo/v4 v4.11.4
)

replace github.com/karurosux/saas-go-kit/core-go => ../core-go

replace github.com/karurosux/saas-go-kit/errors-go => ../errors-go