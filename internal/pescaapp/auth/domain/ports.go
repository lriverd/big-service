package domain

import "context"

type AuthService interface {
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	LoginWithPassword(ctx context.Context, req PasswordLoginRequest) (*AuthResponse, error)
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Logout(ctx context.Context, userID string) error
}

type TokenValidator interface {
	ValidateGoogleToken(ctx context.Context, idToken string) (*GoogleClaims, error)
}

