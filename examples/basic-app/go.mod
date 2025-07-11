module github.com/saas-go-kit/examples/basic-app

go 1.21

require (
	github.com/google/uuid v1.5.0
	github.com/joho/godotenv v1.5.1
	github.com/labstack/echo/v4 v4.11.3
	github.com/saas-go-kit/auth-go v0.0.0
	github.com/saas-go-kit/core-go v0.0.0
	github.com/saas-go-kit/errors-go v0.0.0
	github.com/saas-go-kit/ratelimit-go v0.0.0
	github.com/saas-go-kit/response-go v0.0.0
	gorm.io/driver/sqlite v1.5.4
	gorm.io/gorm v1.25.5
)

require (
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.16.0 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang-jwt/jwt/v5 v5.2.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/labstack/gommon v0.4.0 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-sqlite3 v1.14.17 // indirect
	github.com/saas-go-kit/validator-go v0.0.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/time v0.3.0 // indirect
)

replace (
	github.com/saas-go-kit/auth-go => ../../auth-go
	github.com/saas-go-kit/core-go => ../../core-go
	github.com/saas-go-kit/errors-go => ../../errors-go
	github.com/saas-go-kit/ratelimit-go => ../../ratelimit-go
	github.com/saas-go-kit/response-go => ../../response-go
	github.com/saas-go-kit/validator-go => ../../validator-go
)
