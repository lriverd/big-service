package application

import (
	"context"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/ratings/domain"
	spotDomain "github.com/lriverd/big-service/internal/pescaapp/spots/domain"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	log "github.com/sirupsen/logrus"
)

type RatingService struct {
	ratingRepo domain.RatingRepository
	spotRepo   spotDomain.SpotRepository
}

func NewRatingService(ratingRepo domain.RatingRepository, spotRepo spotDomain.SpotRepository) *RatingService {
	return &RatingService{ratingRepo: ratingRepo, spotRepo: spotRepo}
}

func (s *RatingService) ListBySpot(ctx context.Context, spotID string, limit, offset int) ([]*domain.Rating, int, *domain.RatingStats, error) {
	ratings, total, err := s.ratingRepo.ListBySpot(ctx, spotID, limit, offset)
	if err != nil {
		return nil, 0, nil, err
	}
	stats, err := s.ratingRepo.GetStats(ctx, spotID)
	if err != nil {
		return nil, 0, nil, err
	}
	return ratings, total, stats, nil
}

func (s *RatingService) CreateOrUpdate(ctx context.Context, spotID, userID string, stars int) (*domain.Rating, bool, error) {
	existing, err := s.ratingRepo.FindBySpotAndUser(ctx, spotID, userID)
	if err == nil && existing != nil {
		updated, err := s.ratingRepo.Update(ctx, existing.ID, stars)
		if err != nil {
			return nil, false, err
		}
		s.updateSpotRating(ctx, spotID)
		return updated, false, nil // false = updated, not created
	}

	rating := &domain.Rating{
		SpotID:    spotID,
		UserID:    userID,
		Stars:     stars,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	created, err := s.ratingRepo.Create(ctx, rating)
	if err != nil {
		return nil, false, err
	}
	s.updateSpotRating(ctx, spotID)
	log.WithFields(log.Fields{"spotId": spotID, "userId": userID, "stars": stars}).Info("Rating created")
	return created, true, nil // true = created
}

func (s *RatingService) Delete(ctx context.Context, spotID, userID string) error {
	_, err := s.ratingRepo.FindBySpotAndUser(ctx, spotID, userID)
	if err != nil {
		return apperrors.NotFound("Rating")
	}
	if err := s.ratingRepo.Delete(ctx, spotID, userID); err != nil {
		return err
	}
	s.updateSpotRating(ctx, spotID)
	return nil
}

func (s *RatingService) updateSpotRating(ctx context.Context, spotID string) {
	stats, err := s.ratingRepo.GetStats(ctx, spotID)
	if err != nil {
		return
	}
	_ = s.spotRepo.UpdateRatingStats(ctx, spotID, stats.AverageRating, stats.TotalRatings)
}

