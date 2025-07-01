package core

import "os"

// GetEnvOrDefault returns the value of an environment variable or a default value if not set
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnv returns the value of an environment variable or empty string if not set
func GetEnv(key string) string {
	return os.Getenv(key)
}

// RequireEnv returns the value of an environment variable or panics if not set
func RequireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("required environment variable not set: " + key)
	}
	return value
}