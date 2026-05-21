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

