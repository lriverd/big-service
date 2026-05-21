package application_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/users/application"
	"github.com/lriverd/big-service/internal/pescaapp/users/domain"
)

type mockFavRepo struct {
	listFn   func(ctx context.Context, userID string, limit, offset int) ([]*domain.Favorite, int, error)
	addFn    func(ctx context.Context, fav *domain.Favorite) error
	removeFn func(ctx context.Context, userID, spotID string) error
	existsFn func(ctx context.Context, userID, spotID string) (bool, error)
}

func (m *mockFavRepo) List(ctx context.Context, userID string, limit, offset int) ([]*domain.Favorite, int, error) {
	return m.listFn(ctx, userID, limit, offset)
}
func (m *mockFavRepo) Add(ctx context.Context, fav *domain.Favorite) error {
	return m.addFn(ctx, fav)
}
func (m *mockFavRepo) Remove(ctx context.Context, userID, spotID string) error {
	return m.removeFn(ctx, userID, spotID)
}
func (m *mockFavRepo) Exists(ctx context.Context, userID, spotID string) (bool, error) {
	return m.existsFn(ctx, userID, spotID)
}

type mockSpotInfoProvider struct{}

func (m *mockSpotInfoProvider) GetSpotBasicInfo(ctx context.Context, spotID string) (*domain.FavoriteSpot, error) {
	return &domain.FavoriteSpot{ID: spotID, Name: "Test Spot", Region: "RM", Rating: 4.0}, nil
}

func TestFavoriteService_AddFavorite(t *testing.T) {
	favRepo := &mockFavRepo{
		existsFn: func(ctx context.Context, userID, spotID string) (bool, error) { return false, nil },
		addFn:    func(ctx context.Context, fav *domain.Favorite) error { return nil },
	}
	svc := application.NewFavoriteService(favRepo, &mockSpotInfoProvider{})

	err := svc.AddFavorite(context.Background(), "u1", "s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFavoriteService_AddFavorite_AlreadyExists(t *testing.T) {
	favRepo := &mockFavRepo{
		existsFn: func(ctx context.Context, userID, spotID string) (bool, error) { return true, nil },
	}
	svc := application.NewFavoriteService(favRepo, &mockSpotInfoProvider{})

	err := svc.AddFavorite(context.Background(), "u1", "s1")
	if err == nil {
		t.Error("expected conflict error")
	}
}

func TestFavoriteService_RemoveFavorite(t *testing.T) {
	favRepo := &mockFavRepo{
		removeFn: func(ctx context.Context, userID, spotID string) error { return nil },
	}
	svc := application.NewFavoriteService(favRepo, &mockSpotInfoProvider{})

	err := svc.RemoveFavorite(context.Background(), "u1", "s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFavoriteService_ListFavorites(t *testing.T) {
	favRepo := &mockFavRepo{
		listFn: func(ctx context.Context, userID string, limit, offset int) ([]*domain.Favorite, int, error) {
			return []*domain.Favorite{
				{ID: "f1", UserID: "u1", SpotID: "s1", CreatedAt: time.Now()},
			}, 1, nil
		},
	}
	svc := application.NewFavoriteService(favRepo, &mockSpotInfoProvider{})

	spots, total, err := svc.ListFavorites(context.Background(), "u1", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(spots) != 1 {
		t.Errorf("expected 1 spot, got %d", len(spots))
	}
	if spots[0].Name != "Test Spot" {
		t.Errorf("expected 'Test Spot', got %s", spots[0].Name)
	}
}

func TestFavoriteService_ListFavorites_Error(t *testing.T) {
	favRepo := &mockFavRepo{
		listFn: func(ctx context.Context, userID string, limit, offset int) ([]*domain.Favorite, int, error) {
			return nil, 0, fmt.Errorf("db error")
		},
	}
	svc := application.NewFavoriteService(favRepo, &mockSpotInfoProvider{})

	_, _, err := svc.ListFavorites(context.Background(), "u1", 20, 0)
	if err == nil {
		t.Error("expected error")
	}
}

func TestFavoriteService_AddFavorite_ExistsError(t *testing.T) {
	favRepo := &mockFavRepo{
		existsFn: func(ctx context.Context, userID, spotID string) (bool, error) {
			return false, fmt.Errorf("db error")
		},
	}
	svc := application.NewFavoriteService(favRepo, &mockSpotInfoProvider{})

	err := svc.AddFavorite(context.Background(), "u1", "s1")
	if err == nil {
		t.Error("expected error")
	}
}

