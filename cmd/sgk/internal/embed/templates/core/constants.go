package core

// Environment variable names
const (
	// Server Configuration
	EnvPort        = "PORT"
	EnvEnvironment = "ENV"
	
	// Database Configuration
	EnvDatabaseURL  = "DATABASE_URL"
	EnvDatabaseHost = "DB_HOST"
	EnvDatabasePort = "DB_PORT"
	EnvDatabaseName = "DB_NAME"
	EnvDatabaseUser = "DB_USER"
	EnvDatabasePass = "DB_PASSWORD"
	
	// Redis Configuration
	EnvRedisURL      = "REDIS_URL"
	EnvRedisHost     = "REDIS_HOST"
	EnvRedisPort     = "REDIS_PORT"
	EnvRedisPassword = "REDIS_PASSWORD"
	EnvRedisDB       = "REDIS_DB"
	
	// JWT Configuration
	EnvJWTSecret          = "JWT_SECRET"
	EnvJWTExpirationHours = "JWT_EXPIRATION_HOURS"
	EnvJWTRefreshHours    = "JWT_REFRESH_HOURS"
	
	// Email Configuration
	EnvSMTPHost     = "SMTP_HOST"
	EnvSMTPPort     = "SMTP_PORT"
	EnvSMTPUser     = "SMTP_USER"
	EnvSMTPPassword = "SMTP_PASSWORD"
	EnvSMTPFrom     = "SMTP_FROM"
	
	// File Upload Configuration
	EnvMaxFileSize     = "MAX_FILE_SIZE"
	EnvUploadDir       = "UPLOAD_DIR"
	EnvAllowedFileTypes = "ALLOWED_FILE_TYPES"
	
	// Rate Limiting Configuration
	EnvRateLimitEnabled = "RATE_LIMIT_ENABLED"
	EnvRateLimitRPM     = "RATE_LIMIT_RPM"
	
	// CORS Configuration
	EnvCORSAllowedOrigins = "CORS_ALLOWED_ORIGINS"
	EnvCORSAllowedMethods = "CORS_ALLOWED_METHODS"
	EnvCORSAllowedHeaders = "CORS_ALLOWED_HEADERS"
	
	// Logging Configuration
	EnvLogLevel  = "LOG_LEVEL"
	EnvLogFormat = "LOG_FORMAT"
)

// Environment values
const (
	EnvDevelopment = "development"
	EnvDev         = "dev"
	EnvProduction  = "production"
	EnvProd        = "prod"
	EnvTesting     = "testing"
	EnvTest        = "test"
)

// Default values
const (
	// Server defaults
	DefaultPort        = "8080"
	DefaultEnvironment = EnvDevelopment
	
	// Database defaults
	DefaultDatabaseHost = "localhost"
	DefaultDatabasePort = "5432"
	DefaultDatabaseName = "dbname"
	DefaultDatabaseUser = "user"
	DefaultDatabasePass = "password"
	
	// Redis defaults
	DefaultRedisHost = "localhost"
	DefaultRedisPort = "6379"
	DefaultRedisDB   = 0
	
	// JWT defaults
	DefaultJWTSecret          = "your-super-secret-jwt-key-change-this-in-production"
	DefaultJWTExpirationHours = 24
	DefaultJWTRefreshHours    = 168 // 7 days
	
	// Email defaults
	DefaultSMTPHost = "smtp.gmail.com"
	DefaultSMTPPort = 587
	DefaultSMTPFrom = "noreply@example.com"
	
	// File upload defaults
	DefaultMaxFileSize = 10 * 1024 * 1024 // 10MB
	DefaultUploadDir   = "./uploads"
	
	// Rate limiting defaults
	DefaultRateLimitEnabled = true
	DefaultRateLimitRPM     = 100
	
	// Logging defaults
	DefaultLogLevel  = "info"
	DefaultLogFormat = "json"
)

// Default slices
var (
	DefaultAllowedFileTypes = []string{"image/jpeg", "image/png", "application/pdf"}
	DefaultCORSOrigins     = []string{"*"}
	DefaultCORSMethods     = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	DefaultCORSHeaders     = []string{"*"}
)

// Container keys
const (
	ContainerKeyEcho   = "echo"
	ContainerKeyDB     = "db"
	ContainerKeyConfig = "config"
	ContainerKeyRedis  = "redis"
	ContainerKeyLogger = "logger"
)

// Database URLs
const (
	DefaultPostgresURL = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
	DefaultMySQLURL    = "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	DefaultRedisURL    = "redis://localhost:6379"
)

// HTTP methods
const (
	HTTPGet     = "GET"
	HTTPPost    = "POST"
	HTTPPut     = "PUT"
	HTTPDelete  = "DELETE"
	HTTPPatch   = "PATCH"
	HTTPOptions = "OPTIONS"
	HTTPHead    = "HEAD"
)

// Content types
const (
	ContentTypeJSON        = "application/json"
	ContentTypeXML         = "application/xml"
	ContentTypeForm        = "application/x-www-form-urlencoded"
	ContentTypeMultipart   = "multipart/form-data"
	ContentTypeText        = "text/plain"
	ContentTypeHTML        = "text/html"
	ContentTypeJPEG        = "image/jpeg"
	ContentTypePNG         = "image/png"
	ContentTypePDF         = "application/pdf"
)

// Log levels
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

// Log formats
const (
	LogFormatJSON = "json"
	LogFormatText = "text"
)

// Common error messages
const (
	ErrMsgRequiredEnvVar        = "Required environment variable %s is not set"
	ErrMsgInvalidJWTSecret      = "JWT_SECRET must be changed from default value in production"
	ErrMsgDatabaseConnection    = "Failed to connect to database: %v"
	ErrMsgServerStart           = "Server failed to start: %v"
	ErrMsgModuleRegistration    = "Failed to register %s module: %v"
	ErrMsgConfigValidation      = "Configuration validation failed: %v"
)