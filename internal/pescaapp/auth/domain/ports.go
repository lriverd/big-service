package domain

import "context"

type AuthService interface {
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	Logout(ctx context.Context, userID string) error
}

type TokenValidator interface {
	ValidateGoogleToken(ctx context.Context, idToken string) (*GoogleClaims, error)
}

