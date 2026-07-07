package application

import (
	"context"
	"fmt"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/ratings/domain"
	spotDomain "github.com/lriverd/big-service/internal/pescaapp/spots/domain"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	log "github.com/sirupsen/logrus"
)

type RatingService struct {
	ratingRepo                         domain.RatingRepository
	spotRepo                           spotDomain.SpotRepository
	reputationRecorder                 domain.ReputationRecorder
	reputationDeltaGoodRating          int
	reputationGoodRatingStarsThreshold int
}

func NewRatingService(
	ratingRepo domain.RatingRepository,
	spotRepo spotDomain.SpotRepository,
	reputationRecorder domain.ReputationRecorder,
	reputationDeltaGoodRating int,
	reputationGoodRatingStarsThreshold int,
) *RatingService {
	return &RatingService{
		ratingRepo:                         ratingRepo,
		spotRepo:                           spotRepo,
		reputationRecorder:                 reputationRecorder,
		reputationDeltaGoodRating:          reputationDeltaGoodRating,
		reputationGoodRatingStarsThreshold: reputationGoodRatingStarsThreshold,
	}
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
	s.rewardGoodRating(ctx, spotID, userID, stars)
	log.WithFields(log.Fields{"spotId": spotID, "userId": userID, "stars": stars}).Info("Rating created")
	return created, true, nil // true = created
}

// rewardGoodRating logs a reputation event for the spot's owner when a new
// rating meets the configured "good rating" threshold. Ratings a spot's
// owner leaves on their own spot are excluded, since otherwise self-rating
// would be a free way to farm reputation.
func (s *RatingService) rewardGoodRating(ctx context.Context, spotID, raterUserID string, stars int) {
	if stars < s.reputationGoodRatingStarsThreshold {
		return
	}
	spot, err := s.spotRepo.FindByID(ctx, spotID)
	if err != nil {
		log.WithError(err).WithField("spotId", spotID).Warn("Failed to look up spot owner for good-rating reputation event")
		return
	}
	if spot.CreatedByUserID == raterUserID {
		return
	}
	reason := fmt.Sprintf("Recibió una calificación de %d estrellas", stars)
	if err := s.reputationRecorder.RecordReputationEvent(ctx, spot.CreatedByUserID, "GOOD_RATING_RECEIVED", s.reputationDeltaGoodRating, spotID, reason); err != nil {
		log.WithError(err).WithField("spotId", spotID).Warn("Failed to record good-rating reputation event")
	}
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
