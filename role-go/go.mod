module github.com/karurosux/saas-go-kit/role-go

go 1.21

require (
	github.com/google/uuid v1.5.0
	github.com/labstack/echo/v4 v4.11.3
	github.com/karurosux/saas-go-kit/core-go v0.0.0
	github.com/karurosux/saas-go-kit/errors-go v0.0.0
	github.com/karurosux/saas-go-kit/response-go v0.0.0
	github.com/karurosux/saas-go-kit/validator-go v0.0.0
	gorm.io/gorm v1.25.5
)

require (
	github.com/labstack/gommon v0.4.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
)

replace (
	github.com/karurosux/saas-go-kit/core-go => ../core-go
	github.com/karurosux/saas-go-kit/errors-go => ../errors-go
	github.com/karurosux/saas-go-kit/response-go => ../response-go
	github.com/karurosux/saas-go-kit/validator-go => ../validator-go
)