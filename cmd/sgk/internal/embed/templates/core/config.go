package core

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port        string
	Environment string

	DatabaseURL  string
	DatabaseHost string
	DatabasePort string
	DatabaseName string
	DatabaseUser string
	DatabasePass string

	RedisURL      string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	JWTSecret          string
	JWTExpirationHours int
	JWTRefreshHours    int

	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	MaxFileSize  int64
	UploadDir    string
	AllowedTypes []string

	RateLimitEnabled bool
	RateLimitRPM     int

	CORSAllowedOrigins []string
	CORSAllowedMethods []string
	CORSAllowedHeaders []string

	LogLevel  string
	LogFormat string
}

func LoadConfig() *Config {
	return &Config{
		Port:        getEnv(EnvPort, DefaultPort),
		Environment: getEnv(EnvEnvironment, DefaultEnvironment),

		DatabaseURL:  getEnv(EnvDatabaseURL, DefaultPostgresURL),
		DatabaseHost: getEnv(EnvDatabaseHost, DefaultDatabaseHost),
		DatabasePort: getEnv(EnvDatabasePort, DefaultDatabasePort),
		DatabaseName: getEnv(EnvDatabaseName, DefaultDatabaseName),
		DatabaseUser: getEnv(EnvDatabaseUser, DefaultDatabaseUser),
		DatabasePass: getEnv(EnvDatabasePass, DefaultDatabasePass),

		RedisURL:      getEnv(EnvRedisURL, DefaultRedisURL),
		RedisHost:     getEnv(EnvRedisHost, DefaultRedisHost),
		RedisPort:     getEnv(EnvRedisPort, DefaultRedisPort),
		RedisPassword: getEnv(EnvRedisPassword, ""),
		RedisDB:       getEnvInt(EnvRedisDB, DefaultRedisDB),

		JWTSecret:          getEnv(EnvJWTSecret, DefaultJWTSecret),
		JWTExpirationHours: getEnvInt(EnvJWTExpirationHours, DefaultJWTExpirationHours),
		JWTRefreshHours:    getEnvInt(EnvJWTRefreshHours, DefaultJWTRefreshHours),

		SMTPHost:     getEnv(EnvSMTPHost, DefaultSMTPHost),
		SMTPPort:     getEnvInt(EnvSMTPPort, DefaultSMTPPort),
		SMTPUser:     getEnv(EnvSMTPUser, ""),
		SMTPPassword: getEnv(EnvSMTPPassword, ""),
		SMTPFrom:     getEnv(EnvSMTPFrom, DefaultSMTPFrom),

		MaxFileSize:  getEnvInt64(EnvMaxFileSize, DefaultMaxFileSize),
		UploadDir:    getEnv(EnvUploadDir, DefaultUploadDir),
		AllowedTypes: getEnvSlice(EnvAllowedFileTypes, DefaultAllowedFileTypes),

		RateLimitEnabled: getEnvBool(EnvRateLimitEnabled, DefaultRateLimitEnabled),
		RateLimitRPM:     getEnvInt(EnvRateLimitRPM, DefaultRateLimitRPM),

		CORSAllowedOrigins: getEnvSlice(EnvCORSAllowedOrigins, DefaultCORSOrigins),
		CORSAllowedMethods: getEnvSlice(EnvCORSAllowedMethods, DefaultCORSMethods),
		CORSAllowedHeaders: getEnvSlice(EnvCORSAllowedHeaders, DefaultCORSHeaders),

		LogLevel:  getEnv(EnvLogLevel, DefaultLogLevel),
		LogFormat: getEnv(EnvLogFormat, DefaultLogFormat),
	}
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == EnvDevelopment || c.Environment == EnvDev
}

func (c *Config) IsProduction() bool {
	return c.Environment == EnvProduction || c.Environment == EnvProd
}

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

	if c.IsProduction() && c.JWTSecret == DefaultJWTSecret {
		log.Fatal(ErrMsgInvalidJWTSecret)
	}

	return nil
}

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
