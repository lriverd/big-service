package middleware_test

import (
	"testing"

	"github.com/lriverd/big-service/internal/platform/middleware"
)

const testSecret = "test-secret-key"

func TestGenerateAndParseJWT(t *testing.T) {
	token, err := middleware.GenerateJWT("user123", "test@test.com", "user", testSecret, 60)
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	claims, err := middleware.ParseJWT(token, testSecret)
	if err != nil {
		t.Fatalf("failed to parse JWT: %v", err)
	}
	if claims.UserID != "user123" {
		t.Errorf("expected userID user123, got %s", claims.UserID)
	}
	if claims.Email != "test@test.com" {
		t.Errorf("expected email test@test.com, got %s", claims.Email)
	}
	if claims.Role != "user" {
		t.Errorf("expected role user, got %s", claims.Role)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token, err := middleware.GenerateRefreshToken("user123", testSecret)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	claims, err := middleware.ParseJWT(token, testSecret)
	if err != nil {
		t.Fatalf("failed to parse refresh token: %v", err)
	}
	if claims.UserID != "user123" {
		t.Errorf("expected userID user123, got %s", claims.UserID)
	}
}

func TestParseJWT_InvalidToken(t *testing.T) {
	_, err := middleware.ParseJWT("invalid-token", testSecret)
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestParseJWT_WrongSecret(t *testing.T) {
	token, _ := middleware.GenerateJWT("user123", "test@test.com", "user", testSecret, 60)
	_, err := middleware.ParseJWT(token, "wrong-secret")
	if err == nil {
		t.Error("expected error for wrong secret")
	}
}

