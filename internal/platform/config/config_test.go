package config_test

import (
	"os"
	"testing"

	"github.com/lriverd/big-service/internal/platform/config"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := config.Load()

	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Port)
	}
	if cfg.Environment != "development" {
		t.Errorf("expected development, got %s", cfg.Environment)
	}
	if cfg.JWTExpiryMinutes != 60 {
		t.Errorf("expected 60, got %d", cfg.JWTExpiryMinutes)
	}
	if cfg.RateLimitPerMin != 100 {
		t.Errorf("expected 100, got %d", cfg.RateLimitPerMin)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected info, got %s", cfg.LogLevel)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("JWT_EXPIRY_MINUTES", "120")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("JWT_EXPIRY_MINUTES")
	}()

	cfg := config.Load()

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.Environment != "production" {
		t.Errorf("expected production, got %s", cfg.Environment)
	}
	if cfg.JWTExpiryMinutes != 120 {
		t.Errorf("expected 120, got %d", cfg.JWTExpiryMinutes)
	}
}

func TestLoad_InvalidInt(t *testing.T) {
	os.Setenv("JWT_EXPIRY_MINUTES", "abc")
	defer os.Unsetenv("JWT_EXPIRY_MINUTES")

	cfg := config.Load()
	if cfg.JWTExpiryMinutes != 60 {
		t.Errorf("expected default 60, got %d", cfg.JWTExpiryMinutes)
	}
}

