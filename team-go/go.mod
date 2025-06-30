module github.com/karurosux/saas-go-kit/team-go

go 1.21

require (
	github.com/google/uuid v1.6.0
	github.com/karurosux/saas-go-kit/core-go v0.0.0
	github.com/karurosux/saas-go-kit/errors-go v0.0.0
	github.com/karurosux/saas-go-kit/response-go v0.0.0
	github.com/karurosux/saas-go-kit/validator-go v0.0.0
	github.com/labstack/echo/v4 v4.11.4
	golang.org/x/crypto v0.17.0
	gorm.io/gorm v1.25.5
)

replace github.com/karurosux/saas-go-kit/core-go => ../core-go

replace github.com/karurosux/saas-go-kit/errors-go => ../errors-go

replace github.com/karurosux/saas-go-kit/response-go => ../response-go

replace github.com/karurosux/saas-go-kit/validator-go => ../validator-go