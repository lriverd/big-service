package application

import (
	"context"

	authDomain "github.com/lriverd/big-service/internal/pescaapp/auth/domain"
	userDomain "github.com/lriverd/big-service/internal/pescaapp/users/domain"
	"github.com/lriverd/big-service/internal/platform/middleware"
	log "github.com/sirupsen/logrus"
)

type AuthServiceImpl struct {
	tokenValidator authDomain.TokenValidator
	userRepo       userDomain.UserRepository
	jwtSecret      string
	jwtExpiry      int
}

func NewAuthService(
	tv authDomain.TokenValidator,
	userRepo userDomain.UserRepository,
	jwtSecret string,
	jwtExpiry int,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		tokenValidator: tv,
		userRepo:       userRepo,
		jwtSecret:      jwtSecret,
		jwtExpiry:      jwtExpiry,
	}
}

func (s *AuthServiceImpl) Login(ctx context.Context, req authDomain.LoginRequest) (*authDomain.AuthResponse, error) {
	claims, err := s.tokenValidator.ValidateGoogleToken(ctx, req.IDToken)
	if err != nil {
		log.WithError(err).Warn("Google token validation failed")
		return nil, err
	}

	user, err := s.userRepo.FindByEmail(ctx, claims.Email)
	if err != nil {
		photoURL := &claims.Picture
		if claims.Picture == "" {
			photoURL = nil
		}
		newUser := &userDomain.User{
			Email:    claims.Email,
			Name:     claims.Name,
			PhotoURL: photoURL,
			Role:     "user",
		}
		user, err = s.userRepo.Create(ctx, newUser)
		if err != nil {
			return nil, err
		}
		log.WithField("userId", user.ID).Info("New user created via Google login")
	}

	accessToken, err := middleware.GenerateJWT(user.ID, user.Email, user.Role, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, err := middleware.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &authDomain.AuthResponse{
		User: authDomain.AuthUser{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			PhotoURL:  user.PhotoURL,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServiceImpl) Logout(ctx context.Context, userID string) error {
	log.WithField("userId", userID).Info("User logged out")
	return nil
}

