package application

import (
	"context"
	"fmt"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/spots/domain"
	userDomain "github.com/lriverd/big-service/internal/pescaapp/users/domain"
	"github.com/lriverd/big-service/internal/platform/cache"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	log "github.com/sirupsen/logrus"
)

type SpotService struct {
	spotRepo    domain.SpotRepository
	speciesRepo domain.SpotSpeciesRepository
	cache       *cache.Cache
}

func NewSpotService(spotRepo domain.SpotRepository, speciesRepo domain.SpotSpeciesRepository, c *cache.Cache) *SpotService {
	return &SpotService{spotRepo: spotRepo, speciesRepo: speciesRepo, cache: c}
}

func (s *SpotService) List(ctx context.Context, limit, offset int, filter domain.SpotFilter) ([]*domain.Spot, int, error) {
	spots, total, err := s.spotRepo.List(ctx, limit, offset, filter)
	if err != nil {
		return nil, 0, err
	}
	for _, spot := range spots {
		species, _ := s.speciesRepo.ListBySpot(ctx, spot.ID)
		spot.Species = species
	}
	return spots, total, nil
}

func (s *SpotService) GetByID(ctx context.Context, id string) (*domain.Spot, error) {
	cacheKey := fmt.Sprintf("spot:%s", id)
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*domain.Spot), nil
	}

	spot, err := s.spotRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.NotFound("Spot")
	}

	species, _ := s.speciesRepo.ListBySpot(ctx, id)
	spot.Species = species

	// Increment views asynchronously
	go func() {
		_ = s.spotRepo.IncrementViews(context.Background(), id)
	}()

	s.cache.Set(cacheKey, spot, 5*time.Minute)
	return spot, nil
}

func (s *SpotService) Create(ctx context.Context, req domain.CreateSpotRequest, userID string) (*domain.Spot, error) {
	spot := &domain.Spot{
		Name:            req.Name,
		Description:     req.Description,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		Region:          req.Region,
		WaterType:       req.WaterType,
		BoatAllowed:     req.BoatAllowed,
		BoatRequired:    req.BoatRequired,
		Access:          req.Access,
		IsFree:          req.IsFree,
		EntryFee:        req.EntryFee,
		CreatedByUserID: userID,
	}

	created, err := s.spotRepo.Create(ctx, spot)
	if err != nil {
		return nil, err
	}

	if len(req.Species) > 0 {
		if err := s.speciesRepo.SetForSpot(ctx, created.ID, req.Species); err != nil {
			log.WithError(err).Warn("Failed to set species for spot")
		}
		created.Species = req.Species
	}

	s.cache.DeleteByPrefix("spot:list")
	log.WithField("spotId", created.ID).Info("Spot created")
	return created, nil
}

func (s *SpotService) Update(ctx context.Context, id string, req domain.UpdateSpotRequest, userID, role string) (*domain.Spot, error) {
	spot, err := s.spotRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.NotFound("Spot")
	}

	if spot.CreatedByUserID != userID && role != "admin" {
		return nil, apperrors.Forbidden("Only the creator or admin can update this spot")
	}

	updated, err := s.spotRepo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	if req.Species != nil {
		if err := s.speciesRepo.SetForSpot(ctx, id, req.Species); err != nil {
			log.WithError(err).Warn("Failed to update species for spot")
		}
	}

	s.cache.Delete(fmt.Sprintf("spot:%s", id))
	s.cache.DeleteByPrefix("spot:list")
	return updated, nil
}

func (s *SpotService) Delete(ctx context.Context, id, userID, role string) error {
	spot, err := s.spotRepo.FindByID(ctx, id)
	if err != nil {
		return apperrors.NotFound("Spot")
	}
	if spot.CreatedByUserID != userID && role != "admin" {
		return apperrors.Forbidden("Only the creator or admin can delete this spot")
	}

	if err := s.speciesRepo.DeleteBySpot(ctx, id); err != nil {
		log.WithError(err).Warn("Failed to delete spot species")
	}
	if err := s.spotRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.cache.Delete(fmt.Sprintf("spot:%s", id))
	s.cache.DeleteByPrefix("spot:list")
	log.WithField("spotId", id).Info("Spot deleted")
	return nil
}

func (s *SpotService) FindNearby(ctx context.Context, spotID string, radiusKm float64, limit int) ([]*domain.Spot, error) {
	spot, err := s.spotRepo.FindByID(ctx, spotID)
	if err != nil {
		return nil, apperrors.NotFound("Spot")
	}
	return s.spotRepo.FindNearby(ctx, spot.Latitude, spot.Longitude, radiusKm, limit)
}

// GetSpotBasicInfo implements the SpotInfoProvider interface for favorites
func (s *SpotService) GetSpotBasicInfo(ctx context.Context, spotID string) (*userDomain.FavoriteSpot, error) {
	spot, err := s.spotRepo.FindByID(ctx, spotID)
	if err != nil {
		return nil, err
	}
	return &userDomain.FavoriteSpot{
		ID:     spot.ID,
		Name:   spot.Name,
		Region: spot.Region,
		Rating: spot.AverageRating,
	}, nil
}

func (s *SpotService) Search(ctx context.Context, query string, limit int) ([]*domain.Spot, error) {
	return s.spotRepo.Search(ctx, query, limit)
}

