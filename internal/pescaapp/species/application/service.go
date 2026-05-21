package application

import (
	"context"
	"fmt"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/species/domain"
	"github.com/lriverd/big-service/internal/platform/cache"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	log "github.com/sirupsen/logrus"
)

type SpeciesService struct {
	repo  domain.SpeciesRepository
	cache *cache.Cache
}

func NewSpeciesService(repo domain.SpeciesRepository, cache *cache.Cache) *SpeciesService {
	return &SpeciesService{repo: repo, cache: cache}
}

func (s *SpeciesService) List(ctx context.Context, limit, offset int, search string) ([]*domain.Species, int, error) {
	cacheKey := fmt.Sprintf("species:list:%d:%d:%s", limit, offset, search)
	if cached, found := s.cache.Get(cacheKey); found {
		result := cached.(*speciesListCache)
		return result.species, result.total, nil
	}
	species, total, err := s.repo.List(ctx, limit, offset, search)
	if err != nil {
		return nil, 0, err
	}
	s.cache.Set(cacheKey, &speciesListCache{species: species, total: total}, 5*time.Minute)
	return species, total, nil
}

type speciesListCache struct {
	species []*domain.Species
	total   int
}

func (s *SpeciesService) GetByID(ctx context.Context, id string) (*domain.Species, error) {
	cacheKey := fmt.Sprintf("species:%s", id)
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*domain.Species), nil
	}
	sp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.NotFound("Species")
	}
	s.cache.Set(cacheKey, sp, 10*time.Minute)
	return sp, nil
}

func (s *SpeciesService) Create(ctx context.Context, req domain.CreateSpeciesRequest) (*domain.Species, error) {
	sp := &domain.Species{
		CommonName:      req.CommonName,
		ScientificName:  req.ScientificName,
		Description:     req.Description,
		ImageURL:        req.ImageURL,
		Habitat:         req.Habitat,
		Diet:            req.Diet,
		AverageSizeCm:   req.AverageSizeCm,
		AverageWeightKg: req.AverageWeightKg,
		MaxSizeCm:       req.MaxSizeCm,
		MaxWeightKg:     req.MaxWeightKg,
		FishingTips:     req.FishingTips,
	}
	created, err := s.repo.Create(ctx, sp)
	if err != nil {
		return nil, err
	}
	s.cache.DeleteByPrefix("species:list")
	log.WithField("speciesId", created.ID).Info("Species created")
	return created, nil
}

func (s *SpeciesService) Update(ctx context.Context, id string, req domain.UpdateSpeciesRequest) (*domain.Species, error) {
	updated, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}
	s.cache.Delete(fmt.Sprintf("species:%s", id))
	s.cache.DeleteByPrefix("species:list")
	log.WithField("speciesId", id).Info("Species updated")
	return updated, nil
}

func (s *SpeciesService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.cache.Delete(fmt.Sprintf("species:%s", id))
	s.cache.DeleteByPrefix("species:list")
	log.WithField("speciesId", id).Info("Species deleted")
	return nil
}

func (s *SpeciesService) Search(ctx context.Context, query string, limit int) ([]*domain.Species, error) {
	return s.repo.Search(ctx, query, limit)
}

