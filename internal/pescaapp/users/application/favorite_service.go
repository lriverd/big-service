package application

import (
	"context"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/users/domain"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	log "github.com/sirupsen/logrus"
)

type FavoriteService struct {
	favRepo  domain.FavoriteRepository
	spotRepo SpotInfoProvider
}

// SpotInfoProvider is a port to get spot info without coupling to spots domain directly
type SpotInfoProvider interface {
	GetSpotBasicInfo(ctx context.Context, spotID string) (*domain.FavoriteSpot, error)
}

func NewFavoriteService(favRepo domain.FavoriteRepository, spotRepo SpotInfoProvider) *FavoriteService {
	return &FavoriteService{favRepo: favRepo, spotRepo: spotRepo}
}

func (s *FavoriteService) ListFavorites(ctx context.Context, userID string, limit, offset int) ([]*domain.FavoriteSpot, int, error) {
	favs, total, err := s.favRepo.List(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	spots := make([]*domain.FavoriteSpot, 0, len(favs))
	for _, f := range favs {
		spot, err := s.spotRepo.GetSpotBasicInfo(ctx, f.SpotID)
		if err != nil {
			continue
		}
		spot.AddedAt = f.CreatedAt
		spots = append(spots, spot)
	}
	return spots, total, nil
}

func (s *FavoriteService) AddFavorite(ctx context.Context, userID, spotID string) error {
	exists, err := s.favRepo.Exists(ctx, userID, spotID)
	if err != nil {
		return err
	}
	if exists {
		return apperrors.New(409, "CONFLICT", "Spot already in favorites")
	}

	fav := &domain.Favorite{
		UserID:    userID,
		SpotID:    spotID,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.favRepo.Add(ctx, fav); err != nil {
		return err
	}
	log.WithFields(log.Fields{"userId": userID, "spotId": spotID}).Info("Favorite added")
	return nil
}

func (s *FavoriteService) RemoveFavorite(ctx context.Context, userID, spotID string) error {
	return s.favRepo.Remove(ctx, userID, spotID)
}

