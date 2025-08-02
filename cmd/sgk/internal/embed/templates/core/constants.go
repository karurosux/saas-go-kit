package core

const (
	EnvPort        = "PORT"
	EnvEnvironment = "ENV"

	EnvDatabaseURL  = "DATABASE_URL"
	EnvDatabaseHost = "DB_HOST"
	EnvDatabasePort = "DB_PORT"
	EnvDatabaseName = "DB_NAME"
	EnvDatabaseUser = "DB_USER"
	EnvDatabasePass = "DB_PASSWORD"

	EnvRedisURL      = "REDIS_URL"
	EnvRedisHost     = "REDIS_HOST"
	EnvRedisPort     = "REDIS_PORT"
	EnvRedisPassword = "REDIS_PASSWORD"
	EnvRedisDB       = "REDIS_DB"

	EnvJWTSecret          = "JWT_SECRET"
	EnvJWTExpirationHours = "JWT_EXPIRATION_HOURS"
	EnvJWTRefreshHours    = "JWT_REFRESH_HOURS"

	EnvSMTPHost     = "SMTP_HOST"
	EnvSMTPPort     = "SMTP_PORT"
	EnvSMTPUser     = "SMTP_USER"
	EnvSMTPPassword = "SMTP_PASSWORD"
	EnvSMTPFrom     = "SMTP_FROM"

	EnvMaxFileSize      = "MAX_FILE_SIZE"
	EnvUploadDir        = "UPLOAD_DIR"
	EnvAllowedFileTypes = "ALLOWED_FILE_TYPES"

	EnvRateLimitEnabled = "RATE_LIMIT_ENABLED"
	EnvRateLimitRPM     = "RATE_LIMIT_RPM"

	EnvCORSAllowedOrigins = "CORS_ALLOWED_ORIGINS"
	EnvCORSAllowedMethods = "CORS_ALLOWED_METHODS"
	EnvCORSAllowedHeaders = "CORS_ALLOWED_HEADERS"

	EnvLogLevel  = "LOG_LEVEL"
	EnvLogFormat = "LOG_FORMAT"
)

const (
	EnvDevelopment = "development"
	EnvDev         = "dev"
	EnvProduction  = "production"
	EnvProd        = "prod"
	EnvTesting     = "testing"
	EnvTest        = "test"
)

const (
	DefaultPort        = "8080"
	DefaultEnvironment = EnvDevelopment

	DefaultDatabaseHost = "localhost"
	DefaultDatabasePort = "5432"
	DefaultDatabaseName = "dbname"
	DefaultDatabaseUser = "user"
	DefaultDatabasePass = "password"

	DefaultRedisHost = "localhost"
	DefaultRedisPort = "6379"
	DefaultRedisDB   = 0

	DefaultJWTSecret          = "your-super-secret-jwt-key-change-this-in-production"
	DefaultJWTExpirationHours = 24
	DefaultJWTRefreshHours    = 168 // 7 days

	DefaultSMTPHost = "smtp.gmail.com"
	DefaultSMTPPort = 587
	DefaultSMTPFrom = "noreply@example.com"

	DefaultMaxFileSize = 10 * 1024 * 1024 // 10MB
	DefaultUploadDir   = "./uploads"

	DefaultRateLimitEnabled = true
	DefaultRateLimitRPM     = 100

	DefaultLogLevel  = "info"
	DefaultLogFormat = "json"
)

var (
	DefaultAllowedFileTypes = []string{"image/jpeg", "image/png", "application/pdf"}
	DefaultCORSOrigins      = []string{"*"}
	DefaultCORSMethods      = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	DefaultCORSHeaders      = []string{"*"}
)

const (
	ContainerKeyEcho   = "echo"
	ContainerKeyDB     = "db"
	ContainerKeyConfig = "config"
	ContainerKeyRedis  = "redis"
	ContainerKeyLogger = "logger"
)

const (
	DefaultPostgresURL = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
	DefaultMySQLURL    = "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	DefaultRedisURL    = "redis://localhost:6379"
)

const (
	HTTPGet     = "GET"
	HTTPPost    = "POST"
	HTTPPut     = "PUT"
	HTTPDelete  = "DELETE"
	HTTPPatch   = "PATCH"
	HTTPOptions = "OPTIONS"
	HTTPHead    = "HEAD"
)

const (
	ContentTypeJSON      = "application/json"
	ContentTypeXML       = "application/xml"
	ContentTypeForm      = "application/x-www-form-urlencoded"
	ContentTypeMultipart = "multipart/form-data"
	ContentTypeText      = "text/plain"
	ContentTypeHTML      = "text/html"
	ContentTypeJPEG      = "image/jpeg"
	ContentTypePNG       = "image/png"
	ContentTypePDF       = "application/pdf"
)

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

const (
	LogFormatJSON = "json"
	LogFormatText = "text"
)

const (
	ErrMsgRequiredEnvVar     = "Required environment variable %s is not set"
	ErrMsgInvalidJWTSecret   = "JWT_SECRET must be changed from default value in production"
	ErrMsgDatabaseConnection = "Failed to connect to database: %v"
	ErrMsgServerStart        = "Server failed to start: %v"
	ErrMsgModuleRegistration = "Failed to register %s module: %v"
	ErrMsgConfigValidation   = "Configuration validation failed: %v"
)
