package application

import (
	"context"
	"fmt"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/users/domain"
	"github.com/lriverd/big-service/internal/platform/cache"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	log "github.com/sirupsen/logrus"
)

type UserService struct {
	repo  domain.UserRepository
	cache *cache.Cache
}

func NewUserService(repo domain.UserRepository, cache *cache.Cache) *UserService {
	return &UserService{repo: repo, cache: cache}
}

func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	cacheKey := fmt.Sprintf("user:%s", id)
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*domain.User), nil
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.NotFound("User")
	}

	s.cache.Set(cacheKey, user, 5*time.Minute)
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error) {
	user, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}
	s.cache.Delete(fmt.Sprintf("user:%s", id))
	log.WithField("userId", id).Info("User updated")
	return user, nil
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int, search string) ([]*domain.UserPublic, int, error) {
	return s.repo.List(ctx, limit, offset, search)
}

// IncrementReputationScore implements the reputation module's UserScoreWriter
// port. It also drops the cached copy of the user so a subsequent GetUser
// (or reputation summary read) reflects the new score immediately.
func (s *UserService) IncrementReputationScore(ctx context.Context, userID string, delta int) error {
	if err := s.repo.IncrementReputationScore(ctx, userID, delta); err != nil {
		return err
	}
	s.cache.Delete(fmt.Sprintf("user:%s", userID))
	return nil
}

// GetReputationScore implements the reputation module's UserScoreReader port.
func (s *UserService) GetReputationScore(ctx context.Context, userID string) (int, error) {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	return user.ReputationScore, nil
}

// SetDailySpotLimitOverride implements the reputation module's
// UserPenaltyApplier port, letting a penalty temporarily reduce a user's
// daily spot-creation limit.
func (s *UserService) SetDailySpotLimitOverride(ctx context.Context, userID string, limit int, expiresAt time.Time) error {
	if err := s.repo.SetDailySpotLimitOverride(ctx, userID, limit, expiresAt); err != nil {
		return err
	}
	s.cache.Delete(fmt.Sprintf("user:%s", userID))
	return nil
}

// GetDailySpotLimitOverride implements the spots module's DailyLimitProvider
// port: it returns the user's active temporary daily spot-limit override
// (e.g. from a penalty), or nil if none is set or it has expired.
func (s *UserService) GetDailySpotLimitOverride(ctx context.Context, userID string) (*int, error) {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user.ActiveDailySpotLimitOverride(time.Now().UTC()), nil
}
