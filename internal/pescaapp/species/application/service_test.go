package application_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/species/application"
	"github.com/lriverd/big-service/internal/pescaapp/species/domain"
	"github.com/lriverd/big-service/internal/platform/cache"
)

type mockSpeciesRepo struct {
	findByIDFn func(ctx context.Context, id string) (*domain.Species, error)
	listFn     func(ctx context.Context, limit, offset int, search string) ([]*domain.Species, int, error)
	createFn   func(ctx context.Context, sp *domain.Species) (*domain.Species, error)
	updateFn   func(ctx context.Context, id string, req domain.UpdateSpeciesRequest) (*domain.Species, error)
	deleteFn   func(ctx context.Context, id string) error
	searchFn   func(ctx context.Context, query string, limit int) ([]*domain.Species, error)
}

func (m *mockSpeciesRepo) FindByID(ctx context.Context, id string) (*domain.Species, error) { return m.findByIDFn(ctx, id) }
func (m *mockSpeciesRepo) List(ctx context.Context, limit, offset int, search string) ([]*domain.Species, int, error) { return m.listFn(ctx, limit, offset, search) }
func (m *mockSpeciesRepo) Create(ctx context.Context, sp *domain.Species) (*domain.Species, error) { return m.createFn(ctx, sp) }
func (m *mockSpeciesRepo) Update(ctx context.Context, id string, req domain.UpdateSpeciesRequest) (*domain.Species, error) { return m.updateFn(ctx, id, req) }
func (m *mockSpeciesRepo) Delete(ctx context.Context, id string) error { return m.deleteFn(ctx, id) }
func (m *mockSpeciesRepo) Search(ctx context.Context, query string, limit int) ([]*domain.Species, error) { return m.searchFn(ctx, query, limit) }

func newTestSpecies() *domain.Species {
	return &domain.Species{ID: "sp1", CommonName: "Trucha", ScientificName: "O. mykiss", CreatedAt: time.Now()}
}

func TestSpeciesService_GetByID(t *testing.T) {
	repo := &mockSpeciesRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.Species, error) {
			return newTestSpecies(), nil
		},
	}
	c := cache.New(time.Minute, time.Minute)
	svc := application.NewSpeciesService(repo, c)

	sp, err := svc.GetByID(context.Background(), "sp1")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if sp.CommonName != "Trucha" { t.Errorf("expected Trucha, got %s", sp.CommonName) }

	// Second call hits cache
	sp2, _ := svc.GetByID(context.Background(), "sp1")
	if sp2.CommonName != "Trucha" { t.Error("cache miss") }
}

func TestSpeciesService_GetByID_NotFound(t *testing.T) {
	repo := &mockSpeciesRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.Species, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	svc := application.NewSpeciesService(repo, cache.New(time.Minute, time.Minute))
	_, err := svc.GetByID(context.Background(), "nope")
	if err == nil { t.Error("expected error") }
}

func TestSpeciesService_List(t *testing.T) {
	repo := &mockSpeciesRepo{
		listFn: func(ctx context.Context, limit, offset int, search string) ([]*domain.Species, int, error) {
			return []*domain.Species{newTestSpecies()}, 1, nil
		},
	}
	svc := application.NewSpeciesService(repo, cache.New(time.Minute, time.Minute))

	species, total, err := svc.List(context.Background(), 20, 0, "")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if total != 1 { t.Errorf("expected 1, got %d", total) }
	if len(species) != 1 { t.Errorf("expected 1 species") }

	// Second call hits cache
	species2, _, _ := svc.List(context.Background(), 20, 0, "")
	if len(species2) != 1 { t.Error("cache miss") }
}

func TestSpeciesService_Create(t *testing.T) {
	repo := &mockSpeciesRepo{
		createFn: func(ctx context.Context, sp *domain.Species) (*domain.Species, error) {
			sp.ID = "sp_new"
			return sp, nil
		},
	}
	svc := application.NewSpeciesService(repo, cache.New(time.Minute, time.Minute))

	sp, err := svc.Create(context.Background(), domain.CreateSpeciesRequest{CommonName: "New", ScientificName: "N. new", Description: "d"})
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if sp.ID != "sp_new" { t.Errorf("expected sp_new, got %s", sp.ID) }
}

func TestSpeciesService_Update(t *testing.T) {
	name := "Updated"
	repo := &mockSpeciesRepo{
		updateFn: func(ctx context.Context, id string, req domain.UpdateSpeciesRequest) (*domain.Species, error) {
			return &domain.Species{ID: id, CommonName: *req.CommonName}, nil
		},
	}
	svc := application.NewSpeciesService(repo, cache.New(time.Minute, time.Minute))

	sp, err := svc.Update(context.Background(), "sp1", domain.UpdateSpeciesRequest{CommonName: &name})
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if sp.CommonName != "Updated" { t.Errorf("expected Updated") }
}

func TestSpeciesService_Delete(t *testing.T) {
	repo := &mockSpeciesRepo{
		deleteFn: func(ctx context.Context, id string) error { return nil },
	}
	svc := application.NewSpeciesService(repo, cache.New(time.Minute, time.Minute))

	err := svc.Delete(context.Background(), "sp1")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestSpeciesService_Search(t *testing.T) {
	repo := &mockSpeciesRepo{
		searchFn: func(ctx context.Context, query string, limit int) ([]*domain.Species, error) {
			return []*domain.Species{newTestSpecies()}, nil
		},
	}
	svc := application.NewSpeciesService(repo, cache.New(time.Minute, time.Minute))

	results, err := svc.Search(context.Background(), "trucha", 10)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(results) != 1 { t.Errorf("expected 1 result") }
}

