package infrastructure

import (
	"context"

	authDomain "github.com/lriverd/big-service/internal/pescaapp/auth/domain"
	"firebase.google.com/go/v4/auth"
)

type GoogleTokenValidator struct {
	authClient *auth.Client
}

func NewGoogleTokenValidator(authClient *auth.Client) *GoogleTokenValidator {
	return &GoogleTokenValidator{authClient: authClient}
}

func (v *GoogleTokenValidator) ValidateGoogleToken(ctx context.Context, idToken string) (*authDomain.GoogleClaims, error) {
	token, err := v.authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}

	claims := &authDomain.GoogleClaims{
		Email: token.Claims["email"].(string),
	}
	if name, ok := token.Claims["name"].(string); ok {
		claims.Name = name
	}
	if picture, ok := token.Claims["picture"].(string); ok {
		claims.Picture = picture
	}

	return claims, nil
}

