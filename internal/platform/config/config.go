package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                string
	FirebaseProjectID   string
	CredentialsJSON     string
	JWTSecret           string
	JWTExpiryMinutes    int
	AllowedOrigins      []string
	Environment         string
	LogLevel            string
	RateLimitPerMin     int
	GoogleClientID      string
	RegistrationEnabled bool
}

func Load() *Config {
	return &Config{
		Port:                getEnv("PORT", "8080"),
		FirebaseProjectID:   getEnv("FIREBASE_PROJECT_ID", "pescaap-35d41"),
		CredentialsJSON:     getEnv("FIREBASE_CREDENTIALS_JSON", ""),
		JWTSecret:           getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiryMinutes:    getEnvInt("JWT_EXPIRY_MINUTES", 60),
		AllowedOrigins:      []string{getEnv("ALLOWED_ORIGINS", "*")},
		Environment:         getEnv("ENVIRONMENT", "development"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		RateLimitPerMin:     getEnvInt("RATE_LIMIT_PER_MIN", 100),
		GoogleClientID:      getEnv("GOOGLE_CLIENT_ID", ""),
		RegistrationEnabled: getEnvBool("REGISTRATION_ENABLED", false),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "true" || v == "1"
	}
	return fallback
}
