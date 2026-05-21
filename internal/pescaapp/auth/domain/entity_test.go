package domain_test

import (
	"testing"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/auth/domain"
)

func TestLoginRequest(t *testing.T) {
	req := domain.LoginRequest{IDToken: "token", Email: "test@test.com"}
	if req.IDToken != "token" {
		t.Error("unexpected IDToken")
	}
	if req.Email != "test@test.com" {
		t.Error("unexpected email")
	}
}

func TestAuthResponse(t *testing.T) {
	now := time.Now()
	resp := domain.AuthResponse{
		User:         domain.AuthUser{ID: "u1", Name: "Test", Email: "t@t.com", CreatedAt: now},
		AccessToken:  "at",
		RefreshToken: "rt",
	}
	if resp.User.ID != "u1" {
		t.Error("unexpected user ID")
	}
	if resp.AccessToken != "at" {
		t.Error("unexpected access token")
	}
}

func TestGoogleClaims(t *testing.T) {
	claims := domain.GoogleClaims{Email: "e@e.com", Name: "N", Picture: "p"}
	if claims.Email != "e@e.com" {
		t.Error("unexpected email")
	}
}

