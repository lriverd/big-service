package application

import (
	"context"
	"fmt"
	"time"

	fbAuth "firebase.google.com/go/v4/auth"
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
	firebaseAuth        *fbAuth.Client
	jwtSecret           string
	jwtExpiry           int
	registrationEnabled bool
}

func NewAuthService(
	tv authDomain.TokenValidator,
	userRepo userDomain.UserRepository,
	firebaseAuth *fbAuth.Client,
	jwtSecret string,
	jwtExpiry int,
	registrationEnabled bool,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		tokenValidator:      tv,
		userRepo:            userRepo,
		firebaseAuth:        firebaseAuth,
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
		s.syncFirebaseAuthUser(ctx, user.Email, user.Name, user.PhotoURL, "")
	}

	s.recordLastLogin(ctx, user.ID)
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

	s.recordLastLogin(ctx, user.ID)
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

	s.syncFirebaseAuthUser(ctx, user.Email, user.Name, nil, req.Password)
	s.recordLastLogin(ctx, user.ID)
	log.WithField("userId", user.ID).Info("New user registered with email/password")
	return s.buildAuthResponse(user)
}

func (s *AuthServiceImpl) buildAuthResponse(user *userDomain.User) (*authDomain.AuthResponse, error) {
	accessToken, err := middleware.GenerateJWT(user.ID, user.Email, user.Role, user.Name, s.jwtSecret, s.jwtExpiry)
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

// recordLastLogin updates lastLoginAt in Firestore (non-blocking, best-effort).
func (s *AuthServiceImpl) recordLastLogin(ctx context.Context, userID string) {
	now := time.Now().UTC()
	if err := s.userRepo.UpdateLastLoginAt(ctx, userID, now); err != nil {
		log.WithError(err).WithField("userId", userID).Warn("Failed to update lastLoginAt")
	}
}

// syncFirebaseAuthUser creates the user in Firebase Authentication if they don't exist yet,
// so they appear in the Firebase console. Failures are non-blocking (best-effort).
func (s *AuthServiceImpl) syncFirebaseAuthUser(ctx context.Context, email, name string, photoURL *string, password string) {
	if s.firebaseAuth == nil {
		return
	}
	// Check if already exists
	if _, err := s.firebaseAuth.GetUserByEmail(ctx, email); err == nil {
		return // already exists
	}

	params := (&fbAuth.UserToCreate{}).
		Email(email).
		DisplayName(name).
		EmailVerified(false)

	if photoURL != nil && *photoURL != "" {
		params = params.PhotoURL(*photoURL)
	}
	if password != "" {
		params = params.Password(password)
	}

	u, err := s.firebaseAuth.CreateUser(ctx, params)
	if err != nil {
		log.WithError(err).WithField("email", email).Warn("Failed to sync user to Firebase Auth")
		return
	}
	log.WithField("firebaseUID", u.UID).WithField("email", email).Info("User synced to Firebase Auth")
}
