package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                        string
	FirebaseProjectID           string
	CredentialsJSON             string
	JWTSecret                   string
	JWTExpiryMinutes            int
	AllowedOrigins              []string
	Environment                 string
	LogLevel                    string
	RateLimitPerMin             int
	GoogleClientID              string
	RegistrationEnabled         bool
	DailySpotLimitDefault       int
	DuplicateSearchRadiusMeters float64
	DuplicateSearchMaxResults   int
	ReportHideThreshold         int

	ReputationDeltaSpotVerified           int
	ReputationDeltaSpotHidden             int
	ReputationDeltaSpotDeleted            int
	ReputationDeltaGoodRating             int
	ReputationGoodRatingStarsThreshold    int
	ReputationDeltaRejectedContentPenalty int

	PenaltyRejectedContentRateThreshold float64
	PenaltyRejectedContentMinSampleSize int
	PenaltyDailyLimitReducedValue       int
	PenaltyDailyLimitDurationHours      int
}

func Load() *Config {
	return &Config{
		Port:                        getEnv("PORT", "8080"),
		FirebaseProjectID:           getEnv("FIREBASE_PROJECT_ID", "pescaap-35d41"),
		CredentialsJSON:             getEnv("FIREBASE_CREDENTIALS_JSON", ""),
		JWTSecret:                   getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiryMinutes:            getEnvInt("JWT_EXPIRY_MINUTES", 60),
		AllowedOrigins:              []string{getEnv("ALLOWED_ORIGINS", "*")},
		Environment:                 getEnv("ENVIRONMENT", "development"),
		LogLevel:                    getEnv("LOG_LEVEL", "info"),
		RateLimitPerMin:             getEnvInt("RATE_LIMIT_PER_MIN", 100),
		GoogleClientID:              getEnv("GOOGLE_CLIENT_ID", ""),
		RegistrationEnabled:         getEnvBool("REGISTRATION_ENABLED", false),
		DailySpotLimitDefault:       getEnvInt("DAILY_SPOT_LIMIT_DEFAULT", 3),
		DuplicateSearchRadiusMeters: getEnvFloat("DUPLICATE_SEARCH_RADIUS_METERS", 50),
		DuplicateSearchMaxResults:   getEnvInt("DUPLICATE_SEARCH_MAX_RESULTS", 5),
		ReportHideThreshold:         getEnvInt("REPORT_HIDE_THRESHOLD", 5),

		ReputationDeltaSpotVerified:           getEnvInt("REPUTATION_DELTA_SPOT_VERIFIED", 10),
		ReputationDeltaSpotHidden:             getEnvInt("REPUTATION_DELTA_SPOT_HIDDEN", -15),
		ReputationDeltaSpotDeleted:            getEnvInt("REPUTATION_DELTA_SPOT_DELETED", -20),
		ReputationDeltaGoodRating:             getEnvInt("REPUTATION_DELTA_GOOD_RATING", 2),
		ReputationGoodRatingStarsThreshold:    getEnvInt("REPUTATION_GOOD_RATING_STARS_THRESHOLD", 4),
		ReputationDeltaRejectedContentPenalty: getEnvInt("REPUTATION_DELTA_REJECTED_CONTENT_PENALTY", -25),

		PenaltyRejectedContentRateThreshold: getEnvFloat("PENALTY_REJECTED_CONTENT_RATE_THRESHOLD", 0.5),
		PenaltyRejectedContentMinSampleSize: getEnvInt("PENALTY_REJECTED_CONTENT_MIN_SAMPLE_SIZE", 3),
		PenaltyDailyLimitReducedValue:       getEnvInt("PENALTY_DAILY_LIMIT_REDUCED_VALUE", 1),
		PenaltyDailyLimitDurationHours:      getEnvInt("PENALTY_DAILY_LIMIT_DURATION_HOURS", 168),
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

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
