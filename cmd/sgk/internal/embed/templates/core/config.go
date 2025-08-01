package core

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server Configuration
	Port        string
	Environment string
	
	// Database Configuration
	DatabaseURL  string
	DatabaseHost string
	DatabasePort string
	DatabaseName string
	DatabaseUser string
	DatabasePass string
	
	// Redis Configuration
	RedisURL      string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	
	// JWT Configuration
	JWTSecret           string
	JWTExpirationHours  int
	JWTRefreshHours     int
	
	// Email Configuration
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	
	// File Upload Configuration
	MaxFileSize   int64
	UploadDir     string
	AllowedTypes  []string
	
	// Rate Limiting
	RateLimitEnabled bool
	RateLimitRPM     int
	
	// CORS Configuration
	CORSAllowedOrigins []string
	CORSAllowedMethods []string
	CORSAllowedHeaders []string
	
	// Logging Configuration
	LogLevel  string
	LogFormat string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		// Server Configuration
		Port:        getEnv(EnvPort, DefaultPort),
		Environment: getEnv(EnvEnvironment, DefaultEnvironment),
		
		// Database Configuration
		DatabaseURL:  getEnv(EnvDatabaseURL, DefaultPostgresURL),
		DatabaseHost: getEnv(EnvDatabaseHost, DefaultDatabaseHost),
		DatabasePort: getEnv(EnvDatabasePort, DefaultDatabasePort),
		DatabaseName: getEnv(EnvDatabaseName, DefaultDatabaseName),
		DatabaseUser: getEnv(EnvDatabaseUser, DefaultDatabaseUser),
		DatabasePass: getEnv(EnvDatabasePass, DefaultDatabasePass),
		
		// Redis Configuration
		RedisURL:      getEnv(EnvRedisURL, DefaultRedisURL),
		RedisHost:     getEnv(EnvRedisHost, DefaultRedisHost),
		RedisPort:     getEnv(EnvRedisPort, DefaultRedisPort),
		RedisPassword: getEnv(EnvRedisPassword, ""),
		RedisDB:       getEnvInt(EnvRedisDB, DefaultRedisDB),
		
		// JWT Configuration
		JWTSecret:           getEnv(EnvJWTSecret, DefaultJWTSecret),
		JWTExpirationHours:  getEnvInt(EnvJWTExpirationHours, DefaultJWTExpirationHours),
		JWTRefreshHours:     getEnvInt(EnvJWTRefreshHours, DefaultJWTRefreshHours),
		
		// Email Configuration  
		SMTPHost:     getEnv(EnvSMTPHost, DefaultSMTPHost),
		SMTPPort:     getEnvInt(EnvSMTPPort, DefaultSMTPPort),
		SMTPUser:     getEnv(EnvSMTPUser, ""),
		SMTPPassword: getEnv(EnvSMTPPassword, ""),
		SMTPFrom:     getEnv(EnvSMTPFrom, DefaultSMTPFrom),
		
		// File Upload Configuration
		MaxFileSize:  getEnvInt64(EnvMaxFileSize, DefaultMaxFileSize),
		UploadDir:    getEnv(EnvUploadDir, DefaultUploadDir),
		AllowedTypes: getEnvSlice(EnvAllowedFileTypes, DefaultAllowedFileTypes),
		
		// Rate Limiting
		RateLimitEnabled: getEnvBool(EnvRateLimitEnabled, DefaultRateLimitEnabled),
		RateLimitRPM:     getEnvInt(EnvRateLimitRPM, DefaultRateLimitRPM),
		
		// CORS Configuration
		CORSAllowedOrigins: getEnvSlice(EnvCORSAllowedOrigins, DefaultCORSOrigins),
		CORSAllowedMethods: getEnvSlice(EnvCORSAllowedMethods, DefaultCORSMethods),
		CORSAllowedHeaders: getEnvSlice(EnvCORSAllowedHeaders, DefaultCORSHeaders),
		
		// Logging Configuration
		LogLevel:  getEnv(EnvLogLevel, DefaultLogLevel),
		LogFormat: getEnv(EnvLogFormat, DefaultLogFormat),
	}
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == EnvDevelopment || c.Environment == EnvDev
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == EnvProduction || c.Environment == EnvProd
}

// ValidateRequired validates that required configuration is present
func (c *Config) ValidateRequired() error {
	required := map[string]string{
		EnvDatabaseURL: c.DatabaseURL,
		EnvJWTSecret:   c.JWTSecret,
	}
	
	for key, value := range required {
		if value == "" {
			log.Fatalf(ErrMsgRequiredEnvVar, key)
		}
	}
	
	// Validate JWT secret in production
	if c.IsProduction() && c.JWTSecret == DefaultJWTSecret {
		log.Fatal(ErrMsgInvalidJWTSecret)
	}
	
	return nil
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		// For more complex parsing, consider using a proper CSV parser
		var result []string
		for _, item := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}