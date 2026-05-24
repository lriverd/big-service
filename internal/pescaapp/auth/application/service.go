package application

import (
	"context"
	"fmt"

	authDomain "github.com/lriverd/big-service/internal/pescaapp/auth/domain"
	userDomain "github.com/lriverd/big-service/internal/pescaapp/users/domain"
	"github.com/lriverd/big-service/internal/platform/middleware"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceImpl struct {
	tokenValidator      authDomain.TokenValidator
	userRepo            userDomain.UserRepository
	jwtSecret           string
	jwtExpiry           int
	registrationEnabled bool
}

func NewAuthService(
	tv authDomain.TokenValidator,
	userRepo userDomain.UserRepository,
	jwtSecret string,
	jwtExpiry int,
	registrationEnabled bool,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		tokenValidator:      tv,
		userRepo:            userRepo,
		jwtSecret:           jwtSecret,
		jwtExpiry:           jwtExpiry,
		registrationEnabled: registrationEnabled,
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

	return s.buildAuthResponse(user)
}

func (s *AuthServiceImpl) Logout(ctx context.Context, userID string) error {
	log.WithField("userId", userID).Info("User logged out")
	return nil
}

func (s *AuthServiceImpl) LoginWithPassword(ctx context.Context, req authDomain.PasswordLoginRequest) (*authDomain.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil || user.PasswordHash == "" {
		return nil, apperrors.Unauthorized("Invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.Unauthorized("Invalid email or password")
	}

	return s.buildAuthResponse(user)
}

func (s *AuthServiceImpl) Register(ctx context.Context, req authDomain.RegisterRequest) (*authDomain.AuthResponse, error) {
	if !s.registrationEnabled {
		return nil, apperrors.Forbidden("User registration is currently disabled")
	}

	if _, err := s.userRepo.FindByEmail(ctx, req.Email); err == nil {
		return nil, apperrors.Conflict("Email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &userDomain.User{
		Email:        req.Email,
		Name:         req.Name,
		Role:         "user",
		PasswordHash: string(hash),
	}
	user, err := s.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, err
	}

	log.WithField("userId", user.ID).Info("New user registered with email/password")
	return s.buildAuthResponse(user)
}

func (s *AuthServiceImpl) buildAuthResponse(user *userDomain.User) (*authDomain.AuthResponse, error) {
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

