package application

import (
	"context"
	"sync"

	"github.com/lriverd/big-service/internal/pescaapp/search/domain"
	speciesDomain "github.com/lriverd/big-service/internal/pescaapp/species/domain"
	spotsDomain "github.com/lriverd/big-service/internal/pescaapp/spots/domain"
	usersDomain "github.com/lriverd/big-service/internal/pescaapp/users/domain"
)

type SearchService struct {
	spotRepo    spotsDomain.SpotRepository
	speciesRepo speciesDomain.SpeciesRepository
	userRepo    usersDomain.UserRepository
}

func NewSearchService(
	spotRepo spotsDomain.SpotRepository,
	speciesRepo speciesDomain.SpeciesRepository,
	userRepo usersDomain.UserRepository,
) *SearchService {
	return &SearchService{
		spotRepo:    spotRepo,
		speciesRepo: speciesRepo,
		userRepo:    userRepo,
	}
}

func (s *SearchService) Search(ctx context.Context, query, searchType string, limit int) (*domain.SearchResult, error) {
	result := &domain.SearchResult{
		Spots:   []domain.SpotResult{},
		Species: []domain.SpeciesResult{},
		Users:   []domain.UserResult{},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	if searchType == "all" || searchType == "spot" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			spots, err := s.spotRepo.Search(ctx, query, limit)
			if err != nil {
				return
			}
			mu.Lock()
			for _, sp := range spots {
				result.Spots = append(result.Spots, domain.SpotResult{
					ID: sp.ID, Name: sp.Name, Region: sp.Region, Rating: sp.AverageRating,
				})
			}
			mu.Unlock()
		}()
	}

	if searchType == "all" || searchType == "species" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			species, err := s.speciesRepo.Search(ctx, query, limit)
			if err != nil {
				return
			}
			mu.Lock()
			for _, sp := range species {
				result.Species = append(result.Species, domain.SpeciesResult{
					ID: sp.ID, CommonName: sp.CommonName, ScientificName: sp.ScientificName,
				})
			}
			mu.Unlock()
		}()
	}

	if searchType == "all" || searchType == "user" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			users, _, err := s.userRepo.List(ctx, limit, 0, query)
			if err != nil {
				return
			}
			mu.Lock()
			for _, u := range users {
				result.Users = append(result.Users, domain.UserResult{
					ID: u.ID, Name: u.Name,
				})
			}
			mu.Unlock()
		}()
	}

	wg.Wait()
	return result, nil
}

