package domain_test

import (
	"testing"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/users/domain"
)

func TestUser(t *testing.T) {
	photo := "http://photo.url"
	u := domain.User{
		ID: "u1", Email: "e@e.com", Name: "Test", PhotoURL: &photo,
		Role: "user", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if u.ID != "u1" || u.Email != "e@e.com" || *u.PhotoURL != photo {
		t.Error("unexpected user fields")
	}
}

func TestUserPublic(t *testing.T) {
	up := domain.UserPublic{ID: "u1", Name: "Test", Stats: &domain.UserStats{SpotsCreated: 5}}
	if up.Stats.SpotsCreated != 5 {
		t.Errorf("expected 5, got %d", up.Stats.SpotsCreated)
	}
}

func TestUpdateUserRequest(t *testing.T) {
	name := "New Name"
	req := domain.UpdateUserRequest{Name: &name}
	if *req.Name != "New Name" {
		t.Error("unexpected name")
	}
}

func TestFavorite(t *testing.T) {
	f := domain.Favorite{UserID: "u1", SpotID: "s1", CreatedAt: time.Now()}
	if f.UserID != "u1" || f.SpotID != "s1" {
		t.Error("unexpected favorite fields")
	}
}

func TestFavoriteSpot(t *testing.T) {
	fs := domain.FavoriteSpot{ID: "s1", Name: "Spot", Region: "R1", Rating: 4.5, AddedAt: time.Now()}
	if fs.Rating != 4.5 {
		t.Errorf("expected 4.5, got %f", fs.Rating)
	}
}

func TestUserWithStats(t *testing.T) {
	uws := domain.UserWithStats{
		User:  domain.User{ID: "u1", Name: "Test"},
		Stats: domain.UserStats{SpotsCreated: 3, CommentsCount: 10, RatingsCount: 5, AverageRating: 4.2},
	}
	if uws.Stats.CommentsCount != 10 {
		t.Errorf("expected 10, got %d", uws.Stats.CommentsCount)
	}
}

