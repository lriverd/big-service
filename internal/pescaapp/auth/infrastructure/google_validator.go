package infrastructure

import (
	"context"
	"fmt"

	authDomain "github.com/lriverd/big-service/internal/pescaapp/auth/domain"
	"firebase.google.com/go/v4/auth"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/idtoken"
)

type GoogleTokenValidator struct {
	authClient     *auth.Client
	googleClientID string
}

func NewGoogleTokenValidator(authClient *auth.Client, googleClientID string) *GoogleTokenValidator {
	return &GoogleTokenValidator{authClient: authClient, googleClientID: googleClientID}
}

func (v *GoogleTokenValidator) ValidateGoogleToken(ctx context.Context, token string) (*authDomain.GoogleClaims, error) {
	// Try Firebase ID token first
	fbToken, err := v.authClient.VerifyIDToken(ctx, token)
	if err == nil {
		claims := &authDomain.GoogleClaims{
			Email: fbToken.Claims["email"].(string),
		}
		if name, ok := fbToken.Claims["name"].(string); ok {
			claims.Name = name
		}
		if picture, ok := fbToken.Claims["picture"].(string); ok {
			claims.Picture = picture
		}
		return claims, nil
	}

	log.WithError(err).Debug("Firebase VerifyIDToken failed, trying Google idtoken fallback")

	// Fallback: validate as Google OAuth ID token (sent by google_sign_in with serverClientId)
	if v.googleClientID == "" {
		return nil, fmt.Errorf("firebase token invalid and GOOGLE_CLIENT_ID is not configured: %w", err)
	}

	payload, gErr := idtoken.Validate(ctx, token, v.googleClientID)
	if gErr != nil {
		return nil, fmt.Errorf("google idtoken validation failed (check GOOGLE_CLIENT_ID matches serverClientId in app): %w", gErr)
	}

	claims := &authDomain.GoogleClaims{}
	if email, ok := payload.Claims["email"].(string); ok {
		claims.Email = email
	}
	if name, ok := payload.Claims["name"].(string); ok {
		claims.Name = name
	}
	if picture, ok := payload.Claims["picture"].(string); ok {
		claims.Picture = picture
	}
	return claims, nil
}
